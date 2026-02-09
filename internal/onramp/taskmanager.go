package onramp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/google/uuid"
)

// TaskManager orchestrates async playbook execution and tracks task lifecycle.
type TaskManager struct {
	store   state.Store
	runner  *Runner
	manager *Manager
	mu      sync.Mutex
	active  map[string]context.CancelFunc // taskID → cancel
}

// NewTaskManager creates a TaskManager with the given dependencies.
func NewTaskManager(store state.Store, runner *Runner, mgr *Manager) *TaskManager {
	return &TaskManager{
		store:   store,
		runner:  runner,
		manager: mgr,
		active:  make(map[string]context.CancelFunc),
	}
}

// StartSequence creates a task, sets deployment state, and launches the playbook
// sequence in a background goroutine. Returns the task ID.
func (tm *TaskManager) StartSequence(ctx context.Context, seq PlaybookSequence, component string) (string, error) {
	taskID := uuid.New().String()

	task := &state.DeploymentTask{
		ID:        taskID,
		Operation: seq.Name,
		Status:    state.TaskStatusPending,
	}
	if err := tm.store.CreateTask(ctx, task); err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Determine the in-progress deployment state based on operation name
	deployStatus := state.DeployStateDeploying
	for _, step := range seq.Steps {
		for _, tag := range step.Tags {
			if tag == "uninstall" || tag == "stop" {
				deployStatus = state.DeployStateUndeploying
				break
			}
		}
	}

	if component != "" {
		if err := tm.store.SetDeploymentState(ctx, component, deployStatus, taskID); err != nil {
			slog.Error("failed to set deployment state", "component", component, "error", err)
		}
	}

	// Regenerate inventory and ensure vars
	if err := tm.manager.GenerateHostsINI(ctx); err != nil {
		slog.Error("failed to regenerate hosts.ini", "error", err)
	}
	if err := tm.manager.GenerateVarsFile(ctx, ""); err != nil {
		slog.Error("failed to ensure vars file", "error", err)
	}

	// Launch background execution
	execCtx, cancel := context.WithCancel(context.Background())
	tm.mu.Lock()
	tm.active[taskID] = cancel
	tm.mu.Unlock()

	go tm.executeSequence(execCtx, taskID, seq, component)

	return taskID, nil
}

// executeSequence runs the playbook steps and updates task/deployment state.
func (tm *TaskManager) executeSequence(ctx context.Context, taskID string, seq PlaybookSequence, component string) {
	defer func() {
		tm.mu.Lock()
		delete(tm.active, taskID)
		tm.mu.Unlock()
	}()

	// Mark task as running
	if err := tm.store.UpdateTaskStatus(ctx, taskID, state.TaskStatusRunning); err != nil {
		slog.Error("failed to update task status to running", "task_id", taskID, "error", err)
		return
	}

	inventory := tm.manager.InventoryPath()
	writer := &taskOutputWriter{store: tm.store, taskID: taskID}

	for i, step := range seq.Steps {
		slog.Info("executing playbook step",
			"task_id", taskID,
			"step", i+1,
			"total", len(seq.Steps),
			"name", step.Name,
			"playbook", step.Playbook,
		)

		fmt.Fprintf(writer, "\n=== Step %d/%d: %s ===\n", i+1, len(seq.Steps), step.Name)

		if err := tm.runner.RunPlaybook(ctx, step, inventory, writer); err != nil {
			slog.Error("playbook step failed",
				"task_id", taskID,
				"step", step.Name,
				"error", err,
			)

			errMsg := fmt.Sprintf("step %q failed: %v", step.Name, err)
			if completeErr := tm.store.CompleteTask(ctx, taskID, state.TaskStatusFailed, errMsg); completeErr != nil {
				slog.Error("failed to mark task as failed", "task_id", taskID, "error", completeErr)
			}
			if component != "" {
				if stateErr := tm.store.SetDeploymentState(ctx, component, state.DeployStateFailed, taskID); stateErr != nil {
					slog.Error("failed to set deployment state to failed", "component", component, "error", stateErr)
				}
			}
			tm.logOperation(ctx, seq.Name, state.OpStatusFailure, errMsg)
			return
		}
	}

	// All steps succeeded
	if err := tm.store.CompleteTask(ctx, taskID, state.TaskStatusCompleted, ""); err != nil {
		slog.Error("failed to mark task as completed", "task_id", taskID, "error", err)
	}

	// Determine final deployment state
	if component != "" {
		finalStatus := state.DeployStateDeployed
		for _, step := range seq.Steps {
			for _, tag := range step.Tags {
				if tag == "uninstall" || tag == "stop" {
					finalStatus = state.DeployStateNotDeployed
					break
				}
			}
		}
		if err := tm.store.SetDeploymentState(ctx, component, finalStatus, taskID); err != nil {
			slog.Error("failed to set final deployment state", "component", component, "error", err)
		}
	}

	tm.logOperation(ctx, seq.Name, state.OpStatusSuccess, "")
}

// logOperation records an operations log entry for the deployment action.
func (tm *TaskManager) logOperation(ctx context.Context, operation, opStatus, errMsg string) {
	entry := &state.OperationLog{
		Operation: operation,
		Status:    opStatus,
		Error:     errMsg,
	}
	if err := tm.store.LogOperation(ctx, entry); err != nil {
		slog.Error("failed to log operation", "error", err)
	}
}

// CancelTask cancels a running task.
func (tm *TaskManager) CancelTask(taskID string) error {
	tm.mu.Lock()
	cancel, ok := tm.active[taskID]
	tm.mu.Unlock()

	if !ok {
		return fmt.Errorf("task %s is not active", taskID)
	}
	cancel()

	ctx := context.Background()
	return tm.store.CompleteTask(ctx, taskID, state.TaskStatusCancelled, "cancelled by user")
}

// GetTask retrieves a task from the store.
func (tm *TaskManager) GetTask(ctx context.Context, taskID string) (*state.DeploymentTask, error) {
	return tm.store.GetTask(ctx, taskID)
}

// taskOutputWriter implements io.Writer and appends output to the task in the store.
type taskOutputWriter struct {
	store  state.Store
	taskID string
}

func (w *taskOutputWriter) Write(p []byte) (int, error) {
	ctx := context.Background()
	if err := w.store.AppendTaskOutput(ctx, w.taskID, string(p)); err != nil {
		slog.Error("failed to append task output", "task_id", w.taskID, "error", err)
	}
	return len(p), nil
}

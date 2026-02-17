package onramp

import (
	"context"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/state"
)

func newTestStore(t *testing.T) *state.SQLiteStore {
	t.Helper()
	store, err := state.NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestTaskManagerStartSequenceCreatesTask(t *testing.T) {
	store := newTestStore(t)
	runner := NewRunner(t.TempDir())
	mgr := NewManager(Config{WorkDir: t.TempDir()}, store)
	tm := NewTaskManager(store, runner, mgr)

	// Use a simple sequence that will fail (no ansible installed) but should still create the task
	seq := PlaybookSequence{
		Name:  "test-seq",
		Steps: []PlaybookStep{{Name: "test", Playbook: "test.yml", Tags: []string{"install"}}},
	}

	taskID, err := tm.StartSequence(context.Background(), seq, "test-component")
	if err != nil {
		t.Fatalf("StartSequence() error = %v", err)
	}
	if taskID == "" {
		t.Fatal("StartSequence() returned empty task ID")
	}

	// Give the goroutine a moment to start
	time.Sleep(100 * time.Millisecond)

	// Task should exist in the store
	task, err := store.GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task.Operation != "test-seq" {
		t.Errorf("Operation = %q, want %q", task.Operation, "test-seq")
	}
}

func TestTaskManagerStartSequenceSetsDeploymentState(t *testing.T) {
	store := newTestStore(t)
	runner := NewRunner(t.TempDir())
	mgr := NewManager(Config{WorkDir: t.TempDir()}, store)
	tm := NewTaskManager(store, runner, mgr)

	seq := PlaybookSequence{
		Name:  "install-test",
		Steps: []PlaybookStep{{Name: "install", Playbook: "test.yml", Tags: []string{"install"}}},
	}

	_, err := tm.StartSequence(context.Background(), seq, "my-component")
	if err != nil {
		t.Fatalf("StartSequence() error = %v", err)
	}

	// Deployment state should be set immediately (before goroutine runs)
	ds, err := store.GetDeploymentState(context.Background(), "my-component")
	if err != nil {
		t.Fatalf("GetDeploymentState() error = %v", err)
	}
	if ds.Status != state.DeployStateDeploying {
		t.Errorf("deployment status = %q, want %q", ds.Status, state.DeployStateDeploying)
	}
}

func TestTaskManagerStartSequenceUninstallState(t *testing.T) {
	store := newTestStore(t)
	runner := NewRunner(t.TempDir())
	mgr := NewManager(Config{WorkDir: t.TempDir()}, store)
	tm := NewTaskManager(store, runner, mgr)

	seq := PlaybookSequence{
		Name:  "uninstall-test",
		Steps: []PlaybookStep{{Name: "uninstall", Playbook: "test.yml", Tags: []string{"uninstall"}}},
	}

	_, err := tm.StartSequence(context.Background(), seq, "my-component")
	if err != nil {
		t.Fatalf("StartSequence() error = %v", err)
	}

	ds, err := store.GetDeploymentState(context.Background(), "my-component")
	if err != nil {
		t.Fatalf("GetDeploymentState() error = %v", err)
	}
	if ds.Status != state.DeployStateUndeploying {
		t.Errorf("deployment status = %q, want %q", ds.Status, state.DeployStateUndeploying)
	}
}

func TestTaskManagerCancelNonexistent(t *testing.T) {
	store := newTestStore(t)
	runner := NewRunner(t.TempDir())
	mgr := NewManager(Config{WorkDir: t.TempDir()}, store)
	tm := NewTaskManager(store, runner, mgr)

	err := tm.CancelTask("nonexistent-task")
	if err == nil {
		t.Error("CancelTask() should return error for nonexistent task")
	}
}

func TestTaskManagerGetTask(t *testing.T) {
	store := newTestStore(t)
	runner := NewRunner(t.TempDir())
	mgr := NewManager(Config{WorkDir: t.TempDir()}, store)
	tm := NewTaskManager(store, runner, mgr)

	// Create a task directly in the store
	task := &state.DeploymentTask{
		ID:        "test-task-1",
		Operation: "test-op",
		Status:    state.TaskStatusPending,
	}
	if err := store.CreateTask(context.Background(), task); err != nil {
		t.Fatal(err)
	}

	got, err := tm.GetTask(context.Background(), "test-task-1")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if got.Operation != "test-op" {
		t.Errorf("Operation = %q, want %q", got.Operation, "test-op")
	}
}

func TestTaskManagerNoComponent(t *testing.T) {
	store := newTestStore(t)
	runner := NewRunner(t.TempDir())
	mgr := NewManager(Config{WorkDir: t.TempDir()}, store)
	tm := NewTaskManager(store, runner, mgr)

	seq := PlaybookSequence{
		Name:  "ping",
		Steps: []PlaybookStep{{Name: "ping", Playbook: "pingall.yml"}},
	}

	// Empty component should not set deployment state
	taskID, err := tm.StartSequence(context.Background(), seq, "")
	if err != nil {
		t.Fatalf("StartSequence() error = %v", err)
	}
	if taskID == "" {
		t.Fatal("empty task ID")
	}
}

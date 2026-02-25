package taskrunner

import (
	"errors"
	"time"
)

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusSucceeded TaskStatus = "succeeded"
	StatusFailed    TaskStatus = "failed"
	StatusCanceled  TaskStatus = "canceled"
)

// Sentinel errors returned by Runner methods.
var (
	ErrNotFound         = errors.New("task not found")
	ErrNotRunning       = errors.New("task is not running")
	ErrConcurrencyLimit = errors.New("concurrency limit reached")
)

// TaskSpec describes a command to execute. Callers provide this to Runner.Submit.
type TaskSpec struct {
	Command     string            // binary name (validated via exec.LookPath)
	Args        []string          // command arguments
	Dir         string            // working directory
	Env         []string          // extra KEY=VALUE env vars (appended to inherited env)
	Labels      map[string]string // arbitrary provider-specific metadata
	Description string            // human-readable summary
}

// task is the internal mutable state for a running or completed command.
type task struct {
	id          string
	spec        TaskSpec
	status      TaskStatus
	createdAt   time.Time
	startedAt   time.Time
	finishedAt  time.Time
	exitCode    int
	errMsg      string
	output      *OutputBuffer
	cancelFunc  func()
}

// view returns an immutable snapshot of the task's current state.
func (t *task) view() TaskView {
	return TaskView{
		ID:          t.id,
		Status:      t.status,
		Description: t.spec.Description,
		Labels:      t.spec.Labels,
		Command:     t.spec.Command,
		Args:        t.spec.Args,
		CreatedAt:   t.createdAt,
		StartedAt:   t.startedAt,
		FinishedAt:  t.finishedAt,
		ExitCode:    t.exitCode,
		Error:       t.errMsg,
	}
}

// TaskView is an immutable snapshot of a task's state, safe for returning to
// callers. Output is intentionally excluded — fetch it separately via
// Runner.Output to avoid pulling large blobs on list/get.
type TaskView struct {
	ID          string            `json:"id"`
	Status      TaskStatus        `json:"status"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels,omitempty"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	CreatedAt   time.Time         `json:"created_at"`
	StartedAt   time.Time         `json:"started_at,omitzero"`
	FinishedAt  time.Time         `json:"finished_at,omitzero"`
	ExitCode    int               `json:"exit_code"`
	Error       string            `json:"error,omitempty"`
}

// ListFilter controls which tasks Runner.List returns.
type ListFilter struct {
	Status *TaskStatus
	Label  map[string]string
}

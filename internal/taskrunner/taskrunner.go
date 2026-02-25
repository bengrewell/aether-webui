package taskrunner

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RunnerConfig controls Runner behavior.
type RunnerConfig struct {
	// MaxConcurrent limits the number of simultaneously running tasks.
	// Zero means unlimited.
	MaxConcurrent int
	Logger        *slog.Logger
}

// Runner manages the lifecycle of asynchronous command executions.
type Runner struct {
	cfg RunnerConfig
	log *slog.Logger

	mu    sync.RWMutex
	tasks map[string]*task
}

// New creates a Runner with the given configuration.
func New(cfg RunnerConfig) *Runner {
	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}
	return &Runner{
		cfg:   cfg,
		log:   log,
		tasks: make(map[string]*task),
	}
}

// Submit validates the command, creates a new task, and starts it in a
// background goroutine. The task's context is independent of the caller's
// context so it outlives the HTTP request that created it.
func (r *Runner) Submit(spec TaskSpec) (TaskView, error) {
	// Validate the command binary exists on PATH.
	if _, err := exec.LookPath(spec.Command); err != nil {
		return TaskView{}, fmt.Errorf("command %q not found: %w", spec.Command, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Enforce concurrency limit.
	if r.cfg.MaxConcurrent > 0 {
		running := 0
		for _, t := range r.tasks {
			if t.status == StatusRunning {
				running++
			}
		}
		if running >= r.cfg.MaxConcurrent {
			return TaskView{}, ErrConcurrencyLimit
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	t := &task{
		id:         uuid.NewString(),
		spec:       spec,
		status:     StatusRunning,
		createdAt:  time.Now().UTC(),
		startedAt:  time.Now().UTC(),
		output:     &OutputBuffer{},
		cancelFunc: cancel,
	}
	r.tasks[t.id] = t

	go r.run(ctx, t)

	return t.view(), nil
}

// Get returns an immutable snapshot of the task with the given ID.
func (r *Runner) Get(id string) (TaskView, error) {
	r.mu.RLock()
	t, ok := r.tasks[id]
	if !ok {
		r.mu.RUnlock()
		return TaskView{}, ErrNotFound
	}
	v := t.view()
	r.mu.RUnlock()
	return v, nil
}

// List returns snapshots of all tasks, most-recent first. An optional filter
// can narrow results by status or labels.
func (r *Runner) List(filter *ListFilter) []TaskView {
	r.mu.RLock()
	views := make([]TaskView, 0, len(r.tasks))
	for _, t := range r.tasks {
		if filter != nil && !matchFilter(t, filter) {
			continue
		}
		views = append(views, t.view())
	}
	r.mu.RUnlock()

	sort.Slice(views, func(i, j int) bool {
		return views[i].CreatedAt.After(views[j].CreatedAt)
	})
	return views
}

// Output returns an incremental chunk of the task's output starting at offset.
func (r *Runner) Output(id string, offset int) (OutputChunk, error) {
	r.mu.RLock()
	t, ok := r.tasks[id]
	r.mu.RUnlock()
	if !ok {
		return OutputChunk{}, ErrNotFound
	}

	data, newOffset := t.output.ReadFrom(offset)
	return OutputChunk{
		Data:      string(data),
		Offset:    offset,
		NewOffset: newOffset,
	}, nil
}

// Cancel sends a cancellation signal to a running task. Returns ErrNotFound if
// the task ID is unknown, or ErrNotRunning if the task has already finished.
func (r *Runner) Cancel(id string) error {
	r.mu.RLock()
	t, ok := r.tasks[id]
	r.mu.RUnlock()
	if !ok {
		return ErrNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if t.status != StatusRunning {
		return ErrNotRunning
	}
	t.cancelFunc()
	return nil
}

// run executes the command described by the task and updates its state on
// completion.
func (r *Runner) run(ctx context.Context, t *task) {
	cmd := exec.CommandContext(ctx, t.spec.Command, t.spec.Args...)
	cmd.Dir = t.spec.Dir
	if len(t.spec.Env) > 0 {
		cmd.Env = append(cmd.Environ(), t.spec.Env...)
	}
	cmd.Stdout = t.output
	cmd.Stderr = t.output

	r.log.Info("task started", "id", t.id, "command", t.spec.Command, "args", t.spec.Args)

	err := cmd.Run()

	r.mu.Lock()
	t.finishedAt = time.Now().UTC()
	if ctx.Err() != nil {
		t.status = StatusCanceled
		t.exitCode = -1
		t.errMsg = "canceled"
		r.log.Info("task canceled", "id", t.id)
	} else if err != nil {
		t.status = StatusFailed
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.exitCode = exitErr.ExitCode()
		} else {
			t.exitCode = -1
		}
		t.errMsg = err.Error()
		r.log.Error("task failed", "id", t.id, "error", err)
	} else {
		t.status = StatusSucceeded
		t.exitCode = 0
		r.log.Info("task succeeded", "id", t.id)
	}
	v := t.view()
	cb := t.spec.OnComplete
	r.mu.Unlock()

	if cb != nil {
		r.safeCallback(v, cb)
	}
}

// safeCallback invokes the OnComplete callback, recovering from panics.
func (r *Runner) safeCallback(v TaskView, cb func(TaskView)) {
	defer func() {
		if p := recover(); p != nil {
			r.log.Error("OnComplete callback panicked", "task_id", v.ID, "panic", p)
		}
	}()
	cb(v)
}

// matchFilter returns true if the task matches all non-nil filter criteria.
func matchFilter(t *task, f *ListFilter) bool {
	if f.Status != nil && t.status != *f.Status {
		return false
	}
	for k, v := range f.Label {
		if t.spec.Labels[k] != v {
			return false
		}
	}
	return true
}

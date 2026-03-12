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
// When MaxConcurrent is set, tasks beyond the limit are queued in
// submission order and started automatically as running tasks complete.
type Runner struct {
	cfg RunnerConfig
	log *slog.Logger

	mu    sync.RWMutex
	tasks map[string]*task
	queue []*task // pending tasks in FIFO order
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

// Submit validates the command, creates a new task, and either starts it
// immediately or queues it if the concurrency limit is reached. The returned
// TaskView reflects the initial state (StatusRunning or StatusPending).
func (r *Runner) Submit(spec TaskSpec) (TaskView, error) {
	// Validate the command binary exists on PATH.
	if _, err := exec.LookPath(spec.Command); err != nil {
		return TaskView{}, fmt.Errorf("command %q not found: %w", spec.Command, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := spec.ID
	if id == "" {
		id = uuid.NewString()
	}

	now := time.Now().UTC()
	t := &task{
		id:        id,
		spec:      spec,
		status:    StatusPending,
		createdAt: now,
		output:    &OutputBuffer{},
	}
	r.tasks[t.id] = t

	if r.canStartLocked() {
		r.startLocked(t)
	} else {
		r.queue = append(r.queue, t)
		r.log.Info("task queued", "id", t.id, "queue_depth", len(r.queue))
	}

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

// Cancel sends a cancellation signal to a running task. Pending tasks are
// removed from the queue and marked as canceled immediately. Returns
// ErrNotFound if the task ID is unknown, or ErrNotRunning if the task has
// already finished.
func (r *Runner) Cancel(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	if !ok {
		return ErrNotFound
	}

	switch t.status {
	case StatusPending:
		// Remove from queue.
		for i, qt := range r.queue {
			if qt.id == id {
				r.queue = append(r.queue[:i], r.queue[i+1:]...)
				break
			}
		}
		t.status = StatusCanceled
		t.finishedAt = time.Now().UTC()
		t.errMsg = "canceled"
		t.exitCode = -1
		return nil
	case StatusRunning:
		t.cancelFunc()
		return nil
	default:
		return ErrNotRunning
	}
}

// canStartLocked returns true if there is capacity to start another task.
// Caller must hold r.mu.
func (r *Runner) canStartLocked() bool {
	if r.cfg.MaxConcurrent <= 0 {
		return true
	}
	running := 0
	for _, t := range r.tasks {
		if t.status == StatusRunning {
			running++
		}
	}
	return running < r.cfg.MaxConcurrent
}

// startLocked transitions a pending task to running and spawns its goroutine.
// Caller must hold r.mu.
func (r *Runner) startLocked(t *task) {
	ctx, cancel := context.WithCancel(context.Background())
	t.status = StatusRunning
	t.startedAt = time.Now().UTC()
	t.cancelFunc = cancel

	cb := t.spec.OnStart
	v := t.view()

	go func() {
		if cb != nil {
			r.safeCallback(v, cb)
		}
		r.run(ctx, t)
	}()
}

// drainQueue starts the next pending task from the queue if capacity allows.
// Caller must hold r.mu.
func (r *Runner) drainQueue() {
	for len(r.queue) > 0 && r.canStartLocked() {
		next := r.queue[0]
		r.queue = r.queue[1:]
		// Skip tasks that were canceled while pending.
		if next.status != StatusPending {
			continue
		}
		r.startLocked(next)
		r.log.Info("dequeued task", "id", next.id, "remaining", len(r.queue))
	}
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

	// Start the next queued task before releasing the lock.
	r.drainQueue()
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

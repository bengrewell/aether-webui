package taskrunner

import (
	"strings"
	"testing"
	"time"
)

func TestSubmitAndGet(t *testing.T) {
	r := New(RunnerConfig{})
	view, err := r.Submit(TaskSpec{
		Command:     "echo",
		Args:        []string{"hello"},
		Description: "test echo",
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if view.ID == "" {
		t.Fatal("Submit returned empty ID")
	}
	if view.Status != StatusRunning {
		t.Fatalf("status = %q, want %q", view.Status, StatusRunning)
	}

	// Wait for completion.
	waitForTask(t, r, view.ID, 5*time.Second)

	got, err := r.Get(view.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != StatusSucceeded {
		t.Fatalf("status = %q, want %q", got.Status, StatusSucceeded)
	}
	if got.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", got.ExitCode)
	}
}

func TestSubmitCommandNotFound(t *testing.T) {
	r := New(RunnerConfig{})
	_, err := r.Submit(TaskSpec{
		Command: "definitely-not-a-real-command-xyz",
	})
	if err == nil {
		t.Fatal("Submit should fail for missing command")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %q, should mention 'not found'", err)
	}
}

func TestSubmitConcurrencyLimit(t *testing.T) {
	r := New(RunnerConfig{MaxConcurrent: 1})

	// Submit a long-running task.
	view, err := r.Submit(TaskSpec{
		Command: "sleep",
		Args:    []string{"10"},
	})
	if err != nil {
		t.Fatalf("first Submit: %v", err)
	}

	// Second submit should be rejected.
	_, err = r.Submit(TaskSpec{
		Command: "echo",
		Args:    []string{"second"},
	})
	if err != ErrConcurrencyLimit {
		t.Fatalf("second Submit error = %v, want ErrConcurrencyLimit", err)
	}

	// Cancel the first so the test doesn't hang.
	r.Cancel(view.ID)
}

func TestOutputStreaming(t *testing.T) {
	r := New(RunnerConfig{})

	view, err := r.Submit(TaskSpec{
		Command: "echo",
		Args:    []string{"-n", "hello world"},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	waitForTask(t, r, view.ID, 5*time.Second)

	chunk, err := r.Output(view.ID, 0)
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if chunk.Data != "hello world" {
		t.Fatalf("output = %q, want %q", chunk.Data, "hello world")
	}
	if chunk.Offset != 0 {
		t.Fatalf("offset = %d, want 0", chunk.Offset)
	}
	if chunk.NewOffset != 11 {
		t.Fatalf("new_offset = %d, want 11", chunk.NewOffset)
	}

	// Reading from the new offset should return empty.
	chunk2, err := r.Output(view.ID, chunk.NewOffset)
	if err != nil {
		t.Fatalf("Output (second): %v", err)
	}
	if chunk2.Data != "" {
		t.Fatalf("second read data = %q, want empty", chunk2.Data)
	}
}

func TestCancel(t *testing.T) {
	r := New(RunnerConfig{})
	view, err := r.Submit(TaskSpec{
		Command: "sleep",
		Args:    []string{"60"},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	if err := r.Cancel(view.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	waitForTask(t, r, view.ID, 5*time.Second)

	got, _ := r.Get(view.ID)
	if got.Status != StatusCanceled {
		t.Fatalf("status = %q, want %q", got.Status, StatusCanceled)
	}
}

func TestCancelNotRunning(t *testing.T) {
	r := New(RunnerConfig{})
	view, err := r.Submit(TaskSpec{
		Command: "echo",
		Args:    []string{"fast"},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	waitForTask(t, r, view.ID, 5*time.Second)

	if err := r.Cancel(view.ID); err != ErrNotRunning {
		t.Fatalf("Cancel after completion = %v, want ErrNotRunning", err)
	}
}

func TestCancelNotFound(t *testing.T) {
	r := New(RunnerConfig{})
	if err := r.Cancel("nonexistent"); err != ErrNotFound {
		t.Fatalf("Cancel unknown = %v, want ErrNotFound", err)
	}
}

func TestGetNotFound(t *testing.T) {
	r := New(RunnerConfig{})
	_, err := r.Get("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("Get unknown = %v, want ErrNotFound", err)
	}
}

func TestOutputNotFound(t *testing.T) {
	r := New(RunnerConfig{})
	_, err := r.Output("nonexistent", 0)
	if err != ErrNotFound {
		t.Fatalf("Output unknown = %v, want ErrNotFound", err)
	}
}

func TestListEmpty(t *testing.T) {
	r := New(RunnerConfig{})
	views := r.List(nil)
	if len(views) != 0 {
		t.Fatalf("List on empty runner = %d items, want 0", len(views))
	}
}

func TestListMostRecentFirst(t *testing.T) {
	r := New(RunnerConfig{})

	v1, _ := r.Submit(TaskSpec{Command: "echo", Args: []string{"1"}})
	waitForTask(t, r, v1.ID, 5*time.Second)

	// Small delay so CreatedAt differs.
	time.Sleep(10 * time.Millisecond)

	v2, _ := r.Submit(TaskSpec{Command: "echo", Args: []string{"2"}})
	waitForTask(t, r, v2.ID, 5*time.Second)

	views := r.List(nil)
	if len(views) != 2 {
		t.Fatalf("List = %d items, want 2", len(views))
	}
	if views[0].ID != v2.ID {
		t.Fatalf("first item = %s, want %s (most recent)", views[0].ID, v2.ID)
	}
}

func TestListFilterByStatus(t *testing.T) {
	r := New(RunnerConfig{})

	v1, _ := r.Submit(TaskSpec{Command: "echo", Args: []string{"done"}})
	waitForTask(t, r, v1.ID, 5*time.Second)

	status := StatusSucceeded
	views := r.List(&ListFilter{Status: &status})
	if len(views) != 1 {
		t.Fatalf("filtered list = %d items, want 1", len(views))
	}

	status = StatusRunning
	views = r.List(&ListFilter{Status: &status})
	if len(views) != 0 {
		t.Fatalf("filtered list (running) = %d items, want 0", len(views))
	}
}

func TestListFilterByLabel(t *testing.T) {
	r := New(RunnerConfig{})

	r.Submit(TaskSpec{
		Command: "echo", Args: []string{"a"},
		Labels: map[string]string{"component": "k8s"},
	})
	r.Submit(TaskSpec{
		Command: "echo", Args: []string{"b"},
		Labels: map[string]string{"component": "5gc"},
	})

	// Wait for both.
	for _, v := range r.List(nil) {
		waitForTask(t, r, v.ID, 5*time.Second)
	}

	views := r.List(&ListFilter{Label: map[string]string{"component": "k8s"}})
	if len(views) != 1 {
		t.Fatalf("label-filtered list = %d items, want 1", len(views))
	}
}

func TestFailedCommand(t *testing.T) {
	r := New(RunnerConfig{})
	view, err := r.Submit(TaskSpec{
		Command: "sh",
		Args:    []string{"-c", "exit 42"},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	waitForTask(t, r, view.ID, 5*time.Second)

	got, _ := r.Get(view.ID)
	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want %q", got.Status, StatusFailed)
	}
	if got.ExitCode != 42 {
		t.Fatalf("exit code = %d, want 42", got.ExitCode)
	}
}

func TestStderrCaptured(t *testing.T) {
	r := New(RunnerConfig{})
	view, err := r.Submit(TaskSpec{
		Command: "sh",
		Args:    []string{"-c", "echo -n errout >&2"},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	waitForTask(t, r, view.ID, 5*time.Second)

	chunk, _ := r.Output(view.ID, 0)
	if chunk.Data != "errout" {
		t.Fatalf("stderr output = %q, want %q", chunk.Data, "errout")
	}
}

// waitForTask polls until the task reaches a terminal state or the timeout expires.
func waitForTask(t *testing.T, r *Runner, id string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		v, err := r.Get(id)
		if err != nil {
			t.Fatalf("Get(%s): %v", id, err)
		}
		switch v.Status {
		case StatusSucceeded, StatusFailed, StatusCanceled:
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("task %s did not complete within %v", id, timeout)
}

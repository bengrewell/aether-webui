package wizard

import (
	"strings"
	"testing"

	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/store"
)

func newTestProvider(t *testing.T) *Wizard {
	t.Helper()
	ctx := t.Context()
	dbPath := t.TempDir() + "/test.db"
	st, err := store.New(ctx, dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return NewProvider(provider.WithStore(st))
}

// ---------------------------------------------------------------------------
// Constructor / registration tests
// ---------------------------------------------------------------------------

func TestNewProvider_ImplementsInterface(t *testing.T) {
	var _ provider.Provider = newTestProvider(t)
}

func TestNewProvider_EndpointCount(t *testing.T) {
	p := newTestProvider(t)
	descs := p.Base.Descriptors()
	if len(descs) != 3 {
		t.Errorf("registered %d endpoints, want 3", len(descs))
	}
}

func TestNewProvider_EndpointPaths(t *testing.T) {
	p := newTestProvider(t)

	wantOps := map[string]string{
		"wizard-get":           "/api/v1/wizard",
		"wizard-complete-step": "/api/v1/wizard/steps/{step}",
		"wizard-reset":         "/api/v1/wizard",
	}

	descs := p.Base.Descriptors()
	for _, d := range descs {
		want, ok := wantOps[d.OperationID]
		if !ok {
			t.Errorf("unexpected operation %q", d.OperationID)
			continue
		}
		if d.HTTP.Path != want {
			t.Errorf("operation %q path = %q, want %q", d.OperationID, d.HTTP.Path, want)
		}
		delete(wantOps, d.OperationID)
	}
	for op := range wantOps {
		t.Errorf("missing operation %q", op)
	}
}

// ---------------------------------------------------------------------------
// Get handler — empty state
// ---------------------------------------------------------------------------

func TestHandleGet_Empty(t *testing.T) {
	p := newTestProvider(t)
	out, err := p.handleGet(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGet: %v", err)
	}
	if out.Body.Completed {
		t.Error("expected completed=false for empty state")
	}
	if len(out.Body.Steps) != len(validSteps) {
		t.Errorf("expected %d steps, got %d", len(validSteps), len(out.Body.Steps))
	}
	for name, status := range out.Body.Steps {
		if status.Completed {
			t.Errorf("step %q should not be completed", name)
		}
		if status.CompletedAt != nil {
			t.Errorf("step %q completed_at should be nil", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Complete step handler
// ---------------------------------------------------------------------------

func TestHandleCompleteStep(t *testing.T) {
	p := newTestProvider(t)
	out, err := p.handleCompleteStep(t.Context(), &StepCompleteInput{Step: "nodes"})
	if err != nil {
		t.Fatalf("handleCompleteStep: %v", err)
	}
	if !out.Body.Completed {
		t.Error("expected completed=true")
	}
	if out.Body.CompletedAt == nil {
		t.Error("expected non-nil completed_at")
	}
}

func TestHandleCompleteStep_InvalidStep(t *testing.T) {
	p := newTestProvider(t)
	_, err := p.handleCompleteStep(t.Context(), &StepCompleteInput{Step: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid step")
	}
	if !strings.Contains(err.Error(), "invalid step") {
		t.Errorf("error = %q, should mention 'invalid step'", err)
	}
}

func TestHandleCompleteStep_Idempotent(t *testing.T) {
	p := newTestProvider(t)
	ctx := t.Context()

	_, err := p.handleCompleteStep(ctx, &StepCompleteInput{Step: "preflight"})
	if err != nil {
		t.Fatalf("first complete: %v", err)
	}
	_, err = p.handleCompleteStep(ctx, &StepCompleteInput{Step: "preflight"})
	if err != nil {
		t.Fatalf("second complete: %v", err)
	}

	out, err := p.handleGet(ctx, nil)
	if err != nil {
		t.Fatalf("handleGet: %v", err)
	}
	if !out.Body.Steps["preflight"].Completed {
		t.Error("preflight should still be completed")
	}
}

// ---------------------------------------------------------------------------
// Get handler — partial and full completion
// ---------------------------------------------------------------------------

func TestHandleGet_PartialCompletion(t *testing.T) {
	p := newTestProvider(t)
	ctx := t.Context()

	if _, err := p.handleCompleteStep(ctx, &StepCompleteInput{Step: "nodes"}); err != nil {
		t.Fatalf("complete nodes: %v", err)
	}
	if _, err := p.handleCompleteStep(ctx, &StepCompleteInput{Step: "preflight"}); err != nil {
		t.Fatalf("complete preflight: %v", err)
	}

	out, err := p.handleGet(ctx, nil)
	if err != nil {
		t.Fatalf("handleGet: %v", err)
	}
	if out.Body.Completed {
		t.Error("expected completed=false for partial completion")
	}
	if !out.Body.Steps["nodes"].Completed {
		t.Error("nodes should be completed")
	}
	if !out.Body.Steps["preflight"].Completed {
		t.Error("preflight should be completed")
	}
	if out.Body.Steps["roles"].Completed {
		t.Error("roles should not be completed")
	}
	if out.Body.Steps["config"].Completed {
		t.Error("config should not be completed")
	}
	if out.Body.Steps["deployment"].Completed {
		t.Error("deployment should not be completed")
	}
}

func TestHandleGet_AllCompleted(t *testing.T) {
	p := newTestProvider(t)
	ctx := t.Context()

	for step := range validSteps {
		if _, err := p.handleCompleteStep(ctx, &StepCompleteInput{Step: step}); err != nil {
			t.Fatalf("complete %s: %v", step, err)
		}
	}

	out, err := p.handleGet(ctx, nil)
	if err != nil {
		t.Fatalf("handleGet: %v", err)
	}
	if !out.Body.Completed {
		t.Error("expected completed=true when all steps are done")
	}
	for name, status := range out.Body.Steps {
		if !status.Completed {
			t.Errorf("step %q should be completed", name)
		}
		if status.CompletedAt == nil {
			t.Errorf("step %q should have completed_at", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Reset handler
// ---------------------------------------------------------------------------

func TestHandleReset(t *testing.T) {
	p := newTestProvider(t)
	ctx := t.Context()

	// Complete all steps first.
	for step := range validSteps {
		if _, err := p.handleCompleteStep(ctx, &StepCompleteInput{Step: step}); err != nil {
			t.Fatalf("complete %s: %v", step, err)
		}
	}

	out, err := p.handleReset(ctx, nil)
	if err != nil {
		t.Fatalf("handleReset: %v", err)
	}
	if !strings.Contains(out.Body.Message, "reset") {
		t.Errorf("message = %q, should mention 'reset'", out.Body.Message)
	}

	// Verify all steps are now incomplete.
	state, err := p.handleGet(ctx, nil)
	if err != nil {
		t.Fatalf("handleGet after reset: %v", err)
	}
	if state.Body.Completed {
		t.Error("expected completed=false after reset")
	}
	for name, status := range state.Body.Steps {
		if status.Completed {
			t.Errorf("step %q should not be completed after reset", name)
		}
	}
}

func TestHandleReset_Empty(t *testing.T) {
	p := newTestProvider(t)
	_, err := p.handleReset(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleReset on empty state: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Config step
// ---------------------------------------------------------------------------

func TestHandleCompleteStep_Config(t *testing.T) {
	p := newTestProvider(t)
	out, err := p.handleCompleteStep(t.Context(), &StepCompleteInput{Step: "config"})
	if err != nil {
		t.Fatalf("handleCompleteStep config: %v", err)
	}
	if !out.Body.Completed {
		t.Error("expected completed=true")
	}
}

func TestHandleGet_ConfigNotComplete_MeansNotAllDone(t *testing.T) {
	p := newTestProvider(t)
	ctx := t.Context()

	// Complete everything except config.
	for step := range validSteps {
		if step == "config" {
			continue
		}
		if _, err := p.handleCompleteStep(ctx, &StepCompleteInput{Step: step}); err != nil {
			t.Fatalf("complete %s: %v", step, err)
		}
	}

	out, err := p.handleGet(ctx, nil)
	if err != nil {
		t.Fatalf("handleGet: %v", err)
	}
	if out.Body.Completed {
		t.Error("expected completed=false when config is not done")
	}
}

// ---------------------------------------------------------------------------
// Active task
// ---------------------------------------------------------------------------

func TestHandleGet_ActiveTask_None(t *testing.T) {
	p := newTestProvider(t)
	out, err := p.handleGet(t.Context(), nil)
	if err != nil {
		t.Fatalf("handleGet: %v", err)
	}
	if out.Body.ActiveTask != nil {
		t.Error("expected nil active_task with no running actions")
	}
}

func TestHandleGet_ActiveTask_Running(t *testing.T) {
	p := newTestProvider(t)
	ctx := t.Context()

	// Insert a running action into the store.
	rec := store.ActionRecord{
		ID:        "task-123",
		Component: "k8s",
		Action:    "install",
		Target:    "aether-k8s-install",
		Status:    "running",
		ExitCode:  -1,
	}
	if err := p.Store().InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	out, err := p.handleGet(ctx, nil)
	if err != nil {
		t.Fatalf("handleGet: %v", err)
	}
	if out.Body.ActiveTask == nil {
		t.Fatal("expected non-nil active_task")
	}
	if out.Body.ActiveTask.ID != "task-123" {
		t.Errorf("active_task.id = %q, want %q", out.Body.ActiveTask.ID, "task-123")
	}
	if out.Body.ActiveTask.Component != "k8s" {
		t.Errorf("active_task.component = %q, want %q", out.Body.ActiveTask.Component, "k8s")
	}
}

package onramp

import (
	"fmt"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/store"
	"github.com/bengrewell/aether-webui/internal/taskrunner"
)

// ---------------------------------------------------------------------------
// Ordering tests
// ---------------------------------------------------------------------------

func TestOrderActions_Install(t *testing.T) {
	input := []ComponentActionPair{
		{Component: "srsran", Action: "gnb-install"},
		{Component: "5gc", Action: "install"},
		{Component: "k8s", Action: "install"},
	}
	got := orderActions(input)

	want := []string{"k8s", "5gc", "srsran"}
	for i, w := range want {
		if got[i].Component != w {
			t.Errorf("position %d: got %s, want %s", i, got[i].Component, w)
		}
	}
}

func TestOrderActions_Uninstall(t *testing.T) {
	input := []ComponentActionPair{
		{Component: "k8s", Action: "uninstall"},
		{Component: "5gc", Action: "uninstall"},
		{Component: "srsran", Action: "gnb-uninstall"},
	}
	got := orderActions(input)

	// Reverse order: srsran (tier 3) > 5gc (tier 1) > k8s (tier 0)
	want := []string{"srsran", "5gc", "k8s"}
	for i, w := range want {
		if got[i].Component != w {
			t.Errorf("position %d: got %s, want %s", i, got[i].Component, w)
		}
	}
}

func TestOrderActions_Mixed(t *testing.T) {
	input := []ComponentActionPair{
		{Component: "srsran", Action: "gnb-install"},
		{Component: "k8s", Action: "uninstall"},
		{Component: "5gc", Action: "install"},
	}
	got := orderActions(input)

	// Mixed uses install order: k8s (0), 5gc (1), srsran (3)
	want := []string{"k8s", "5gc", "srsran"}
	for i, w := range want {
		if got[i].Component != w {
			t.Errorf("position %d: got %s, want %s", i, got[i].Component, w)
		}
	}
}

func TestOrderActions_StableSort(t *testing.T) {
	input := []ComponentActionPair{
		{Component: "amp", Action: "install"},
		{Component: "sdran", Action: "install"},
		{Component: "oscric", Action: "ric-install"},
	}
	got := orderActions(input)

	// All tier 2 — stable sort preserves original order.
	want := []string{"amp", "sdran", "oscric"}
	for i, w := range want {
		if got[i].Component != w {
			t.Errorf("position %d: got %s, want %s", i, got[i].Component, w)
		}
	}
}

// ---------------------------------------------------------------------------
// Handler tests
// ---------------------------------------------------------------------------

func TestHandleDeploy_EmptyActions(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()

	_, err := o.HandleDeploy(ctx, &DeployInput{
		Body: DeployBody{Actions: nil},
	})
	if err == nil {
		t.Fatal("expected error for empty actions")
	}
}

func TestHandleDeploy_InvalidComponent(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()

	_, err := o.HandleDeploy(ctx, &DeployInput{
		Body: DeployBody{Actions: []ComponentActionPair{
			{Component: "nonexistent", Action: "install"},
		}},
	})
	if err == nil {
		t.Fatal("expected error for invalid component")
	}
}

func TestHandleDeploy_InvalidAction(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()

	_, err := o.HandleDeploy(ctx, &DeployInput{
		Body: DeployBody{Actions: []ComponentActionPair{
			{Component: "k8s", Action: "nonexistent"},
		}},
	})
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

// TestSubmitDeploymentAction_Chain exercises the full chaining mechanism by
// submitting actions that use "echo" (always succeeds) directly via
// submitDeploymentAction, bypassing the make command requirement.
func TestSubmitDeploymentAction_Chain(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	dep := store.Deployment{
		ID:        "chain-test",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		StartedAt: time.Now().UTC(),
		Actions: []store.DeploymentAction{
			{DeploymentID: "chain-test", Seq: 0, ActionID: "chain-a0", Component: "k8s", Action: "install"},
			{DeploymentID: "chain-test", Seq: 1, ActionID: "chain-a1", Component: "5gc", Action: "install"},
		},
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	// Pre-insert action records for both actions.
	for _, a := range dep.Actions {
		rec := store.ActionRecord{
			ID:        a.ActionID,
			Component: a.Component,
			Action:    a.Action,
			Target:    resolveTarget(a.Component, a.Action),
			Status:    "pending",
			ExitCode:  -1,
			StartedAt: time.Now().UTC(),
		}
		if err := st.InsertAction(ctx, rec); err != nil {
			t.Fatalf("InsertAction(%s): %v", a.ActionID, err)
		}
	}

	// Submit the first action using "echo" instead of "make".
	first := dep.Actions[0]
	_, err := o.runner.Submit(taskrunner.TaskSpec{
		ID:          first.ActionID,
		Command:     "echo",
		Args:        []string{"ok"},
		Description: "test chain",
		Labels: map[string]string{
			"component":     first.Component,
			"action":        first.Action,
			"target":        resolveTarget(first.Component, first.Action),
			"deployment_id": dep.ID,
		},
		OnStart: buildOnStart(st, o.Log(), first.ActionID, first.Component, first.Action),
		OnComplete: func(v taskrunner.TaskView) {
			buildOnComplete(st, o.Log(), first.ActionID, first.Component, first.Action)(v)
			// Simulate the chaining: on success, submit next action.
			if v.Status == taskrunner.StatusSucceeded && len(dep.Actions) > 1 {
				next := dep.Actions[1]
				o.runner.Submit(taskrunner.TaskSpec{
					ID:          next.ActionID,
					Command:     "echo",
					Args:        []string{"ok"},
					Description: "test chain 2",
					OnComplete: func(v2 taskrunner.TaskView) {
						buildOnComplete(st, o.Log(), next.ActionID, next.Component, next.Action)(v2)
						if v2.Status == taskrunner.StatusSucceeded {
							st.UpdateDeploymentStatus(ctx, dep.ID, "succeeded", "", time.Now().UTC())
						}
					},
				})
			}
		},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	// Poll until deployment reaches terminal state.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		got, _, err := st.GetDeployment(ctx, "chain-test")
		if err != nil {
			t.Fatalf("GetDeployment: %v", err)
		}
		if got.Status == "succeeded" {
			// Verify both actions completed.
			a0, _, _ := st.GetAction(ctx, "chain-a0")
			a1, _, _ := st.GetAction(ctx, "chain-a1")
			if a0.Status != "succeeded" {
				t.Errorf("action 0 status = %q, want succeeded", a0.Status)
			}
			if a1.Status != "succeeded" {
				t.Errorf("action 1 status = %q, want succeeded", a1.Status)
			}
			return
		}
		if got.Status == "failed" || got.Status == "canceled" {
			t.Fatalf("deployment ended with status %q, error: %s", got.Status, got.Error)
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("deployment did not reach terminal state within timeout")
}

// TestSubmitDeploymentAction_FailFast verifies that when an action fails,
// subsequent actions are canceled and the deployment is marked failed.
func TestSubmitDeploymentAction_FailFast(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	dep := store.Deployment{
		ID:        "failfast-test",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		StartedAt: time.Now().UTC(),
		Actions: []store.DeploymentAction{
			{DeploymentID: "failfast-test", Seq: 0, ActionID: "ff-a0", Component: "k8s", Action: "install"},
			{DeploymentID: "failfast-test", Seq: 1, ActionID: "ff-a1", Component: "5gc", Action: "install"},
		},
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	for _, a := range dep.Actions {
		rec := store.ActionRecord{
			ID:        a.ActionID,
			Component: a.Component,
			Action:    a.Action,
			Target:    resolveTarget(a.Component, a.Action),
			Status:    "pending",
			ExitCode:  -1,
			StartedAt: time.Now().UTC(),
		}
		if err := st.InsertAction(ctx, rec); err != nil {
			t.Fatalf("InsertAction(%s): %v", a.ActionID, err)
		}
	}

	// Submit using "false" which always exits with code 1.
	first := dep.Actions[0]
	baseOnComplete := buildOnComplete(st, o.Log(), first.ActionID, first.Component, first.Action)
	_, err := o.runner.Submit(taskrunner.TaskSpec{
		ID:      first.ActionID,
		Command: "false",
		OnComplete: func(v taskrunner.TaskView) {
			baseOnComplete(v)
			if v.Status == taskrunner.StatusFailed {
				st.UpdateDeploymentStatus(ctx, dep.ID, "failed", v.Error, time.Now().UTC())
				o.cancelRemainingActions(dep, 1)
			}
		},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		got, _, _ := st.GetDeployment(ctx, "failfast-test")
		if got.Status == "failed" {
			// Second action should be canceled.
			a1, _, _ := st.GetAction(ctx, "ff-a1")
			if a1.Status != "canceled" {
				t.Errorf("action 1 status = %q, want canceled", a1.Status)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("deployment did not reach failed state within timeout")
}

func TestHandleListDeployments(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	for i := range 3 {
		dep := store.Deployment{
			ID:        fmt.Sprintf("list-dep-%d", i),
			Status:    "succeeded",
			CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
		}
		if err := st.InsertDeployment(ctx, dep); err != nil {
			t.Fatalf("InsertDeployment: %v", err)
		}
	}

	out, err := o.HandleListDeployments(ctx, &DeploymentListInput{Limit: 10})
	if err != nil {
		t.Fatalf("HandleListDeployments: %v", err)
	}
	if len(out.Body) != 3 {
		t.Errorf("len = %d, want 3", len(out.Body))
	}
}

func TestHandleGetDeployment_Found(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	now := time.Now().UTC()
	dep := store.Deployment{
		ID:        "get-found",
		Status:    "succeeded",
		CreatedAt: now,
		StartedAt: now,
		Actions: []store.DeploymentAction{
			{DeploymentID: "get-found", Seq: 0, ActionID: "gf-a0", Component: "k8s", Action: "install"},
		},
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}
	rec := store.ActionRecord{
		ID:        "gf-a0",
		Component: "k8s",
		Action:    "install",
		Target:    "aether-k8s-install",
		Status:    "succeeded",
		ExitCode:  0,
		StartedAt: now,
	}
	if err := st.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	out, err := o.HandleGetDeployment(ctx, &DeploymentGetInput{ID: "get-found"})
	if err != nil {
		t.Fatalf("HandleGetDeployment: %v", err)
	}
	if out.Body.Status != "succeeded" {
		t.Errorf("status = %q, want succeeded", out.Body.Status)
	}
	if len(out.Body.Actions) != 1 {
		t.Fatalf("actions = %d, want 1", len(out.Body.Actions))
	}
	if out.Body.Actions[0].Status != "succeeded" {
		t.Errorf("action status = %q, want succeeded", out.Body.Actions[0].Status)
	}
	if out.Body.StartedAt == 0 {
		t.Error("started_at should be set")
	}
}

func TestHandleGetDeployment_NotFound(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()

	_, err := o.HandleGetDeployment(ctx, &DeploymentGetInput{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing deployment")
	}
}

func TestHandleCancelDeployment_AlreadyTerminal(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	dep := store.Deployment{
		ID:        "cancel-term",
		Status:    "succeeded",
		CreatedAt: time.Now().UTC(),
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	_, err := o.HandleCancelDeployment(ctx, &DeploymentCancelInput{ID: "cancel-term"})
	if err == nil {
		t.Fatal("expected error for terminal deployment")
	}
}

func TestHandleCancelDeployment_NotFound(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()

	_, err := o.HandleCancelDeployment(ctx, &DeploymentCancelInput{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing deployment")
	}
}

func TestHandleCancelDeployment_Pending(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	rec := store.ActionRecord{
		ID:        "cancel-act-1",
		Component: "k8s",
		Action:    "install",
		Target:    "aether-k8s-install",
		Status:    "pending",
		ExitCode:  -1,
		StartedAt: time.Now().UTC(),
	}
	if err := st.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	dep := store.Deployment{
		ID:        "cancel-pend",
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
		Actions: []store.DeploymentAction{
			{DeploymentID: "cancel-pend", Seq: 0, ActionID: "cancel-act-1", Component: "k8s", Action: "install"},
		},
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	out, err := o.HandleCancelDeployment(ctx, &DeploymentCancelInput{ID: "cancel-pend"})
	if err != nil {
		t.Fatalf("HandleCancelDeployment: %v", err)
	}
	if out.Body.Message == "" {
		t.Error("expected non-empty message")
	}

	got, _, _ := st.GetDeployment(ctx, "cancel-pend")
	if got.Status != "canceled" {
		t.Errorf("status = %q, want canceled", got.Status)
	}

	action, _, _ := st.GetAction(ctx, "cancel-act-1")
	if action.Status != "canceled" {
		t.Errorf("action status = %q, want canceled", action.Status)
	}
}

// TestHandleCancelDeployment_RunningAction cancels a deployment that has
// a running action in the task runner.
func TestHandleCancelDeployment_RunningAction(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	rec := store.ActionRecord{
		ID:        "cancel-run-act",
		Component: "k8s",
		Action:    "install",
		Target:    "aether-k8s-install",
		Status:    "running",
		ExitCode:  -1,
		StartedAt: time.Now().UTC(),
	}
	if err := st.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	// Submit a long-running task to the runner.
	_, err := o.runner.Submit(taskrunner.TaskSpec{
		ID:      "cancel-run-act",
		Command: "sleep",
		Args:    []string{"60"},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	dep := store.Deployment{
		ID:        "cancel-running",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		Actions: []store.DeploymentAction{
			{DeploymentID: "cancel-running", Seq: 0, ActionID: "cancel-run-act", Component: "k8s", Action: "install"},
		},
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	out, err := o.HandleCancelDeployment(ctx, &DeploymentCancelInput{ID: "cancel-running"})
	if err != nil {
		t.Fatalf("HandleCancelDeployment: %v", err)
	}
	if out.Body.Message == "" {
		t.Error("expected non-empty message")
	}

	got, _, _ := st.GetDeployment(ctx, "cancel-running")
	if got.Status != "canceled" {
		t.Errorf("deployment status = %q, want canceled", got.Status)
	}
}

func TestResolveTarget(t *testing.T) {
	tests := []struct {
		comp, action, want string
	}{
		{"k8s", "install", "aether-k8s-install"},
		{"5gc", "uninstall", "aether-5gc-uninstall"},
		{"srsran", "gnb-install", "aether-srsran-gnb-install"},
		{"nonexistent", "install", ""},
		{"k8s", "nonexistent", ""},
	}
	for _, tt := range tests {
		got := resolveTarget(tt.comp, tt.action)
		if got != tt.want {
			t.Errorf("resolveTarget(%q, %q) = %q, want %q", tt.comp, tt.action, got, tt.want)
		}
	}
}

func TestBuildDeploymentItem_EnrichesStatus(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	now := time.Now().UTC()
	dep := store.Deployment{
		ID:        "enrich-test",
		Status:    "running",
		CreatedAt: now,
		StartedAt: now,
		Actions: []store.DeploymentAction{
			{DeploymentID: "enrich-test", Seq: 0, ActionID: "en-a0", Component: "k8s", Action: "install"},
			{DeploymentID: "enrich-test", Seq: 1, ActionID: "en-a1", Component: "5gc", Action: "install"},
		},
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	// First action succeeded, second still pending.
	for _, rec := range []store.ActionRecord{
		{ID: "en-a0", Component: "k8s", Action: "install", Target: "t", Status: "succeeded", StartedAt: now},
		{ID: "en-a1", Component: "5gc", Action: "install", Target: "t", Status: "pending", StartedAt: now},
	} {
		if err := st.InsertAction(ctx, rec); err != nil {
			t.Fatalf("InsertAction: %v", err)
		}
	}

	item := o.buildDeploymentItem(ctx, dep)
	if item.Actions[0].Status != "succeeded" {
		t.Errorf("action 0 status = %q, want succeeded", item.Actions[0].Status)
	}
	if item.Actions[1].Status != "pending" {
		t.Errorf("action 1 status = %q, want pending", item.Actions[1].Status)
	}
	if item.CreatedAt == 0 {
		t.Error("created_at should be set")
	}
	if item.StartedAt == 0 {
		t.Error("started_at should be set")
	}
}

func TestCancelRemainingActions(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	now := time.Now().UTC()
	for _, id := range []string{"cr-a0", "cr-a1", "cr-a2"} {
		rec := store.ActionRecord{
			ID:        id,
			Component: "k8s",
			Action:    "install",
			Target:    "t",
			Status:    "pending",
			ExitCode:  -1,
			StartedAt: now,
		}
		if err := st.InsertAction(ctx, rec); err != nil {
			t.Fatalf("InsertAction: %v", err)
		}
	}

	dep := store.Deployment{
		ID:        "cr-test",
		Status:    "running",
		CreatedAt: now,
		Actions: []store.DeploymentAction{
			{DeploymentID: "cr-test", Seq: 0, ActionID: "cr-a0", Component: "k8s", Action: "install"},
			{DeploymentID: "cr-test", Seq: 1, ActionID: "cr-a1", Component: "k8s", Action: "install"},
			{DeploymentID: "cr-test", Seq: 2, ActionID: "cr-a2", Component: "k8s", Action: "install"},
		},
	}

	// Cancel from seq 1 onward (a1, a2 should be canceled, a0 left alone).
	o.cancelRemainingActions(dep, 1)

	a0, _, _ := st.GetAction(ctx, "cr-a0")
	if a0.Status != "pending" {
		t.Errorf("a0 status = %q, want pending (unchanged)", a0.Status)
	}
	a1, _, _ := st.GetAction(ctx, "cr-a1")
	if a1.Status != "canceled" {
		t.Errorf("a1 status = %q, want canceled", a1.Status)
	}
	a2, _, _ := st.GetAction(ctx, "cr-a2")
	if a2.Status != "canceled" {
		t.Errorf("a2 status = %q, want canceled", a2.Status)
	}
}

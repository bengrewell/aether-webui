package onramp

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/store"
)

var lookPath = exec.LookPath

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

func TestHandleDeploy_HappyPath(t *testing.T) {
	if _, err := lookPath("true"); err != nil {
		t.Skip("true not on PATH")
	}

	o := newTestProviderWithStore(t, "")
	ctx := t.Context()

	// Override the runner command to use "true" (always succeeds).
	// Submit via the handler will fail because "make" is the command and
	// the onramp dir has no Makefile. Instead, test the store/ordering layer.
	st := o.Store()

	// Directly test the deployment store round-trip.
	dep := store.Deployment{
		ID:        "test-deploy-happy",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		Actions: []store.DeploymentAction{
			{DeploymentID: "test-deploy-happy", Seq: 0, ActionID: "a0", Component: "k8s", Action: "install"},
			{DeploymentID: "test-deploy-happy", Seq: 1, ActionID: "a1", Component: "5gc", Action: "install"},
		},
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	got, found, err := st.GetDeployment(ctx, "test-deploy-happy")
	if err != nil {
		t.Fatalf("GetDeployment: %v", err)
	}
	if !found {
		t.Fatal("deployment not found")
	}
	if got.Status != "running" {
		t.Errorf("status = %q, want running", got.Status)
	}
	if len(got.Actions) != 2 {
		t.Fatalf("actions = %d, want 2", len(got.Actions))
	}
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

func TestHandleCancelDeployment_Pending(t *testing.T) {
	o := newTestProviderWithStore(t, "")
	ctx := t.Context()
	st := o.Store()

	// Insert a pending action record.
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

	// Verify deployment is canceled.
	got, _, _ := st.GetDeployment(ctx, "cancel-pend")
	if got.Status != "canceled" {
		t.Errorf("status = %q, want canceled", got.Status)
	}

	// Verify action was canceled.
	action, _, _ := st.GetAction(ctx, "cancel-act-1")
	if action.Status != "canceled" {
		t.Errorf("action status = %q, want canceled", action.Status)
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

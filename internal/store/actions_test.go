package store

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Action history
// ---------------------------------------------------------------------------

func TestInsertAction_RoundTrip(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	now := time.Now().UTC().Truncate(time.Second)
	rec := ActionRecord{
		ID:        "act-1",
		Component: "k8s",
		Action:    "install",
		Target:    "aether-k8s-install",
		Status:    "running",
		ExitCode:  -1,
		Labels:    map[string]string{"env": "staging"},
		Tags:      []string{"nightly", "canary"},
		StartedAt: now,
	}

	if err := st.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	got, ok, err := st.GetAction(ctx, "act-1")
	if err != nil {
		t.Fatalf("GetAction: %v", err)
	}
	if !ok {
		t.Fatal("expected action to exist")
	}

	if got.Component != "k8s" {
		t.Errorf("Component = %q, want %q", got.Component, "k8s")
	}
	if got.Action != "install" {
		t.Errorf("Action = %q, want %q", got.Action, "install")
	}
	if got.Target != "aether-k8s-install" {
		t.Errorf("Target = %q, want %q", got.Target, "aether-k8s-install")
	}
	if got.Status != "running" {
		t.Errorf("Status = %q, want %q", got.Status, "running")
	}
	if got.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1", got.ExitCode)
	}
	if got.Labels["env"] != "staging" {
		t.Errorf("Labels[env] = %q, want %q", got.Labels["env"], "staging")
	}
	if len(got.Tags) != 2 || got.Tags[0] != "nightly" {
		t.Errorf("Tags = %v, want [nightly canary]", got.Tags)
	}
	if !got.StartedAt.Equal(now) {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, now)
	}
	if !got.FinishedAt.IsZero() {
		t.Errorf("FinishedAt = %v, want zero", got.FinishedAt)
	}
}

func TestInsertAction_Validation(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	tests := []struct {
		name string
		rec  ActionRecord
	}{
		{"empty ID", ActionRecord{Component: "k8s", Action: "install", Target: "t"}},
		{"empty Component", ActionRecord{ID: "id", Action: "install", Target: "t"}},
		{"empty Action", ActionRecord{ID: "id", Component: "k8s", Target: "t"}},
		{"empty Target", ActionRecord{ID: "id", Component: "k8s", Action: "install"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.rec.StartedAt = time.Now()
			if err := st.InsertAction(ctx, tt.rec); err != ErrInvalidArgument {
				t.Errorf("expected ErrInvalidArgument, got %v", err)
			}
		})
	}
}

func TestInsertAction_NilLabelsAndTags(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	rec := ActionRecord{
		ID:        "act-nil",
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

	got, ok, _ := st.GetAction(ctx, "act-nil")
	if !ok {
		t.Fatal("expected action to exist")
	}
	if got.Labels != nil {
		t.Errorf("Labels = %v, want nil", got.Labels)
	}
	if got.Tags != nil {
		t.Errorf("Tags = %v, want nil", got.Tags)
	}
}

func TestUpdateActionResult(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	now := time.Now().UTC().Truncate(time.Second)
	rec := ActionRecord{
		ID:        "act-update",
		Component: "5gc",
		Action:    "install",
		Target:    "aether-5gc-install",
		Status:    "running",
		ExitCode:  -1,
		StartedAt: now,
	}
	if err := st.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	finished := now.Add(30 * time.Second)
	result := ActionResult{
		Status:     "succeeded",
		ExitCode:   0,
		FinishedAt: finished,
	}
	if err := st.UpdateActionResult(ctx, "act-update", result); err != nil {
		t.Fatalf("UpdateActionResult: %v", err)
	}

	got, _, _ := st.GetAction(ctx, "act-update")
	if got.Status != "succeeded" {
		t.Errorf("Status = %q, want %q", got.Status, "succeeded")
	}
	if got.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", got.ExitCode)
	}
	if !got.FinishedAt.Equal(finished) {
		t.Errorf("FinishedAt = %v, want %v", got.FinishedAt, finished)
	}
}

func TestUpdateActionResult_WithError(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	rec := ActionRecord{
		ID:        "act-fail",
		Component: "5gc",
		Action:    "install",
		Target:    "aether-5gc-install",
		Status:    "running",
		ExitCode:  -1,
		StartedAt: time.Now().UTC(),
	}
	if err := st.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	result := ActionResult{
		Status:     "failed",
		ExitCode:   1,
		Error:      "make target failed",
		FinishedAt: time.Now().UTC(),
	}
	if err := st.UpdateActionResult(ctx, "act-fail", result); err != nil {
		t.Fatalf("UpdateActionResult: %v", err)
	}

	got, _, _ := st.GetAction(ctx, "act-fail")
	if got.Error != "make target failed" {
		t.Errorf("Error = %q, want %q", got.Error, "make target failed")
	}
}

func TestUpdateActionResult_NotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	err := st.UpdateActionResult(ctx, "nonexistent", ActionResult{Status: "failed"})
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateActionResult_EmptyID(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	err := st.UpdateActionResult(ctx, "", ActionResult{Status: "failed"})
	if err != ErrInvalidArgument {
		t.Errorf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGetAction_NotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	_, ok, err := st.GetAction(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetAction: %v", err)
	}
	if ok {
		t.Error("expected ok=false for missing action")
	}
}

func TestGetAction_EmptyID(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	_, _, err := st.GetAction(ctx, "")
	if err != ErrInvalidArgument {
		t.Errorf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestListActions_Empty(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	list, err := st.ListActions(ctx, ActionFilter{})
	if err != nil {
		t.Fatalf("ListActions: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestListActions_OrderAndFilter(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	base := time.Now().UTC().Truncate(time.Second)
	for i, r := range []ActionRecord{
		{ID: "a1", Component: "k8s", Action: "install", Target: "t1", Status: "succeeded", StartedAt: base},
		{ID: "a2", Component: "5gc", Action: "install", Target: "t2", Status: "running", StartedAt: base.Add(time.Second)},
		{ID: "a3", Component: "k8s", Action: "uninstall", Target: "t3", Status: "failed", StartedAt: base.Add(2 * time.Second)},
	} {
		if err := st.InsertAction(ctx, r); err != nil {
			t.Fatalf("InsertAction[%d]: %v", i, err)
		}
	}

	// No filter: most recent first.
	all, err := st.ListActions(ctx, ActionFilter{})
	if err != nil {
		t.Fatalf("ListActions (all): %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(all))
	}
	if all[0].ID != "a3" {
		t.Errorf("first action = %q, want a3 (most recent)", all[0].ID)
	}

	// Filter by component.
	k8s, _ := st.ListActions(ctx, ActionFilter{Component: "k8s"})
	if len(k8s) != 2 {
		t.Fatalf("expected 2 k8s actions, got %d", len(k8s))
	}

	// Filter by status.
	running, _ := st.ListActions(ctx, ActionFilter{Status: "running"})
	if len(running) != 1 || running[0].ID != "a2" {
		t.Errorf("running filter: got %v, want [a2]", running)
	}

	// Filter by component + action.
	k8sInstall, _ := st.ListActions(ctx, ActionFilter{Component: "k8s", Action: "install"})
	if len(k8sInstall) != 1 || k8sInstall[0].ID != "a1" {
		t.Errorf("k8s+install filter: got %v, want [a1]", k8sInstall)
	}
}

func TestListActions_Pagination(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	base := time.Now().UTC().Truncate(time.Second)
	for i := range 5 {
		r := ActionRecord{
			ID:        "p" + string(rune('0'+i)),
			Component: "k8s",
			Action:    "install",
			Target:    "t",
			Status:    "succeeded",
			StartedAt: base.Add(time.Duration(i) * time.Second),
		}
		if err := st.InsertAction(ctx, r); err != nil {
			t.Fatalf("InsertAction[%d]: %v", i, err)
		}
	}

	// Page 1: limit 2.
	page1, _ := st.ListActions(ctx, ActionFilter{Limit: 2})
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}

	// Page 2: offset 2, limit 2.
	page2, _ := st.ListActions(ctx, ActionFilter{Limit: 2, Offset: 2})
	if len(page2) != 2 {
		t.Fatalf("page2 len = %d, want 2", len(page2))
	}

	// Pages should not overlap.
	if page1[0].ID == page2[0].ID {
		t.Error("pages should not overlap")
	}

	// Page 3: offset 4, limit 2.
	page3, _ := st.ListActions(ctx, ActionFilter{Limit: 2, Offset: 4})
	if len(page3) != 1 {
		t.Fatalf("page3 len = %d, want 1", len(page3))
	}
}

func TestListActions_DefaultLimit(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	// Insert 60 records.
	base := time.Now().UTC()
	for i := range 60 {
		r := ActionRecord{
			ID:        "dl" + string(rune('A'+i/26)) + string(rune('a'+i%26)),
			Component: "k8s",
			Action:    "install",
			Target:    "t",
			Status:    "succeeded",
			StartedAt: base.Add(time.Duration(i) * time.Second),
		}
		if err := st.InsertAction(ctx, r); err != nil {
			t.Fatalf("InsertAction[%d]: %v", i, err)
		}
	}

	// Default limit should be 50.
	list, _ := st.ListActions(ctx, ActionFilter{})
	if len(list) != 50 {
		t.Errorf("default limit: got %d, want 50", len(list))
	}
}

// ---------------------------------------------------------------------------
// Component state
// ---------------------------------------------------------------------------

func TestUpsertComponentState_RoundTrip(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	now := time.Now().UTC().Truncate(time.Second)
	cs := ComponentState{
		Component:  "k8s",
		Status:     "installed",
		LastAction: "install",
		ActionID:   "act-1",
		UpdatedAt:  now,
	}
	if err := st.UpsertComponentState(ctx, cs); err != nil {
		t.Fatalf("UpsertComponentState: %v", err)
	}

	got, ok, err := st.GetComponentState(ctx, "k8s")
	if err != nil {
		t.Fatalf("GetComponentState: %v", err)
	}
	if !ok {
		t.Fatal("expected component state to exist")
	}
	if got.Status != "installed" {
		t.Errorf("Status = %q, want %q", got.Status, "installed")
	}
	if got.LastAction != "install" {
		t.Errorf("LastAction = %q, want %q", got.LastAction, "install")
	}
	if got.ActionID != "act-1" {
		t.Errorf("ActionID = %q, want %q", got.ActionID, "act-1")
	}
	if !got.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, now)
	}
}

func TestUpsertComponentState_Update(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	now := time.Now().UTC().Truncate(time.Second)
	cs := ComponentState{
		Component: "k8s",
		Status:    "installing",
		ActionID:  "act-1",
		UpdatedAt: now,
	}
	if err := st.UpsertComponentState(ctx, cs); err != nil {
		t.Fatalf("UpsertComponentState: %v", err)
	}

	// Update to installed.
	cs.Status = "installed"
	cs.LastAction = "install"
	cs.UpdatedAt = now.Add(30 * time.Second)
	if err := st.UpsertComponentState(ctx, cs); err != nil {
		t.Fatalf("UpsertComponentState (update): %v", err)
	}

	got, _, _ := st.GetComponentState(ctx, "k8s")
	if got.Status != "installed" {
		t.Errorf("Status = %q, want %q", got.Status, "installed")
	}
	if got.LastAction != "install" {
		t.Errorf("LastAction = %q, want %q", got.LastAction, "install")
	}
}

func TestUpsertComponentState_Validation(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	tests := []struct {
		name string
		cs   ComponentState
	}{
		{"empty Component", ComponentState{Status: "installed"}},
		{"empty Status", ComponentState{Component: "k8s"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := st.UpsertComponentState(ctx, tt.cs); err != ErrInvalidArgument {
				t.Errorf("expected ErrInvalidArgument, got %v", err)
			}
		})
	}
}

func TestGetComponentState_NotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	_, ok, err := st.GetComponentState(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetComponentState: %v", err)
	}
	if ok {
		t.Error("expected ok=false for missing component state")
	}
}

func TestGetComponentState_EmptyName(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	_, _, err := st.GetComponentState(ctx, "")
	if err != ErrInvalidArgument {
		t.Errorf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestListComponentStates(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	// Start empty.
	list, err := st.ListComponentStates(ctx)
	if err != nil {
		t.Fatalf("ListComponentStates: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}

	// Add two components.
	now := time.Now().UTC().Truncate(time.Second)
	for _, cs := range []ComponentState{
		{Component: "k8s", Status: "installed", UpdatedAt: now},
		{Component: "5gc", Status: "not_installed", UpdatedAt: now},
	} {
		if err := st.UpsertComponentState(ctx, cs); err != nil {
			t.Fatalf("UpsertComponentState(%s): %v", cs.Component, err)
		}
	}

	list, err = st.ListComponentStates(ctx)
	if err != nil {
		t.Fatalf("ListComponentStates: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 states, got %d", len(list))
	}
	// Ordered by component name.
	if list[0].Component != "5gc" {
		t.Errorf("first = %q, want %q", list[0].Component, "5gc")
	}
	if list[1].Component != "k8s" {
		t.Errorf("second = %q, want %q", list[1].Component, "k8s")
	}
}

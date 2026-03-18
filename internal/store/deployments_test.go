package store

import (
	"fmt"
	"testing"
	"time"
)

func TestInsertDeployment_RoundTrip(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	now := time.Now().UTC().Truncate(time.Second)
	dep := Deployment{
		ID:        "dep-1",
		Status:    "pending",
		CreatedAt: now,
		Actions: []DeploymentAction{
			{DeploymentID: "dep-1", Seq: 0, ActionID: "act-0", Component: "k8s", Action: "install"},
			{DeploymentID: "dep-1", Seq: 1, ActionID: "act-1", Component: "5gc", Action: "install"},
		},
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	got, found, err := st.GetDeployment(ctx, "dep-1")
	if err != nil {
		t.Fatalf("GetDeployment: %v", err)
	}
	if !found {
		t.Fatal("deployment not found")
	}
	if got.ID != "dep-1" || got.Status != "pending" {
		t.Errorf("ID=%q Status=%q, want dep-1/pending", got.ID, got.Status)
	}
	if got.CreatedAt.Unix() != now.Unix() {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, now)
	}
	if len(got.Actions) != 2 {
		t.Fatalf("len(Actions) = %d, want 2", len(got.Actions))
	}
	if got.Actions[0].Component != "k8s" || got.Actions[1].Component != "5gc" {
		t.Errorf("actions = %+v", got.Actions)
	}
}

func TestUpdateDeploymentStatus(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	now := time.Now().UTC().Truncate(time.Second)
	dep := Deployment{
		ID:        "dep-2",
		Status:    "pending",
		CreatedAt: now,
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	fin := now.Add(30 * time.Second)
	if err := st.UpdateDeploymentStatus(ctx, "dep-2", "succeeded", "", fin); err != nil {
		t.Fatalf("UpdateDeploymentStatus: %v", err)
	}

	got, _, err := st.GetDeployment(ctx, "dep-2")
	if err != nil {
		t.Fatalf("GetDeployment: %v", err)
	}
	if got.Status != "succeeded" {
		t.Errorf("Status = %q, want succeeded", got.Status)
	}
	if got.FinishedAt.Unix() != fin.Unix() {
		t.Errorf("FinishedAt = %v, want %v", got.FinishedAt, fin)
	}
}

func TestUpdateDeploymentStatus_WithError(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	dep := Deployment{
		ID:        "dep-err",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	fin := time.Now().UTC()
	if err := st.UpdateDeploymentStatus(ctx, "dep-err", "failed", "action k8s/install failed", fin); err != nil {
		t.Fatalf("UpdateDeploymentStatus: %v", err)
	}

	got, _, _ := st.GetDeployment(ctx, "dep-err")
	if got.Error != "action k8s/install failed" {
		t.Errorf("Error = %q, want 'action k8s/install failed'", got.Error)
	}
}

func TestListDeployments_Pagination(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	for i := range 5 {
		dep := Deployment{
			ID:        fmt.Sprintf("dep-%d", i),
			Status:    "succeeded",
			CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
		}
		if err := st.InsertDeployment(ctx, dep); err != nil {
			t.Fatalf("InsertDeployment(%d): %v", i, err)
		}
	}

	// Page 1: limit 2
	page1, err := st.ListDeployments(ctx, DeploymentFilter{Limit: 2})
	if err != nil {
		t.Fatalf("ListDeployments page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}

	// Page 2: limit 2, offset 2
	page2, err := st.ListDeployments(ctx, DeploymentFilter{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("ListDeployments page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page2 len = %d, want 2", len(page2))
	}

	// IDs should not overlap
	if page1[0].ID == page2[0].ID {
		t.Error("page1 and page2 overlap")
	}
}

func TestListDeployments_StatusFilter(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	for i, status := range []string{"pending", "running", "succeeded", "failed"} {
		dep := Deployment{
			ID:        fmt.Sprintf("dep-f-%d", i),
			Status:    status,
			CreatedAt: time.Now().UTC().Add(time.Duration(i) * time.Second),
		}
		if err := st.InsertDeployment(ctx, dep); err != nil {
			t.Fatalf("InsertDeployment: %v", err)
		}
	}

	got, err := st.ListDeployments(ctx, DeploymentFilter{Status: "running"})
	if err != nil {
		t.Fatalf("ListDeployments: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Status != "running" {
		t.Errorf("Status = %q, want running", got[0].Status)
	}
}

func TestGetDeployment_NotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	_, found, err := st.GetDeployment(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetDeployment: %v", err)
	}
	if found {
		t.Error("expected not found")
	}
}

func TestUpdateDeploymentStatus_SetsStartedAt(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	dep := Deployment{
		ID:        "dep-start",
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	// First update should set started_at via COALESCE.
	if err := st.UpdateDeploymentStatus(ctx, "dep-start", "running", "", time.Time{}); err != nil {
		t.Fatalf("UpdateDeploymentStatus (running): %v", err)
	}

	got, _, _ := st.GetDeployment(ctx, "dep-start")
	if got.StartedAt.IsZero() {
		t.Error("started_at should be set after transitioning to running")
	}
	startedAt := got.StartedAt

	// Second update should not overwrite started_at.
	fin := time.Now().UTC()
	if err := st.UpdateDeploymentStatus(ctx, "dep-start", "succeeded", "", fin); err != nil {
		t.Fatalf("UpdateDeploymentStatus (succeeded): %v", err)
	}

	got2, _, _ := st.GetDeployment(ctx, "dep-start")
	if got2.StartedAt.Unix() != startedAt.Unix() {
		t.Errorf("started_at changed from %v to %v", startedAt, got2.StartedAt)
	}
}

func TestInsertDeployment_InvalidID(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	err := st.InsertDeployment(ctx, Deployment{})
	if err != ErrInvalidArgument {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestUpdateDeploymentStatus_InvalidArgs(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	if err := st.UpdateDeploymentStatus(ctx, "", "running", "", time.Time{}); err != ErrInvalidArgument {
		t.Errorf("empty id: err = %v, want ErrInvalidArgument", err)
	}
	if err := st.UpdateDeploymentStatus(ctx, "x", "", "", time.Time{}); err != ErrInvalidArgument {
		t.Errorf("empty status: err = %v, want ErrInvalidArgument", err)
	}
}

func TestGetDeployment_InvalidID(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	_, _, err := st.GetDeployment(ctx, "")
	if err != ErrInvalidArgument {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestListDeployments_DefaultLimit(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	// Insert 1 deployment and verify default limit works.
	dep := Deployment{
		ID:        "default-limit",
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	got, err := st.ListDeployments(ctx, DeploymentFilter{})
	if err != nil {
		t.Fatalf("ListDeployments: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestInsertDeployment_WithStartedAt(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	now := time.Now().UTC().Truncate(time.Second)
	dep := Deployment{
		ID:        "dep-started",
		Status:    "running",
		CreatedAt: now,
		StartedAt: now,
	}
	if err := st.InsertDeployment(ctx, dep); err != nil {
		t.Fatalf("InsertDeployment: %v", err)
	}

	got, _, _ := st.GetDeployment(ctx, "dep-started")
	if got.StartedAt.Unix() != now.Unix() {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, now)
	}
}

func TestUpdateDeploymentStatus_NotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()

	err := st.UpdateDeploymentStatus(ctx, "nonexistent", "failed", "oops", time.Now())
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

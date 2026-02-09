package aether

import (
	"context"
	"errors"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator"
	"github.com/bengrewell/aether-webui/internal/state"
)

// newTestStore creates a SQLiteStore backed by a temp directory for testing.
func newTestStore(t *testing.T) *state.SQLiteStore {
	t.Helper()
	store, err := state.NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestOperatorImplementsAetherOperator(t *testing.T) {
	var _ AetherOperator = (*Operator)(nil)
}

func TestNew(t *testing.T) {
	op := New(nil, newTestStore(t))
	if op == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDomain(t *testing.T) {
	op := New(nil, newTestStore(t))
	if got := op.Domain(); got != operator.DomainAether {
		t.Errorf("Domain() = %q, want %q", got, operator.DomainAether)
	}
}

func TestHealthWithNilTaskMgr(t *testing.T) {
	op := New(nil, newTestStore(t))
	health, err := op.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "unavailable" {
		t.Errorf("Status = %q, want %q", health.Status, "unavailable")
	}
}

func TestUpdateCoreReturnsNotImplemented(t *testing.T) {
	op := New(nil, newTestStore(t))
	err := op.UpdateCore(context.Background(), "5gc", &CoreConfig{})
	if !errors.Is(err, operator.ErrNotImplemented) {
		t.Errorf("UpdateCore() error = %v, want ErrNotImplemented", err)
	}
}

func TestUpdateGNBReturnsNotImplemented(t *testing.T) {
	op := New(nil, newTestStore(t))
	err := op.UpdateGNB(context.Background(), "srsran-gnb", &GNBConfig{})
	if !errors.Is(err, operator.ErrNotImplemented) {
		t.Errorf("UpdateGNB() error = %v, want ErrNotImplemented", err)
	}
}

func TestGetCoreStatusNotDeployed(t *testing.T) {
	store := newTestStore(t)
	op := New(nil, store)
	status, err := op.GetCoreStatus(context.Background(), "5gc")
	if err != nil {
		t.Fatalf("GetCoreStatus() error = %v", err)
	}
	if status.State != StateNotDeployed {
		t.Errorf("State = %q, want %q", status.State, StateNotDeployed)
	}
}

func TestGetCoreStatusWithState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	if err := store.SetDeploymentState(ctx, "5gc", state.DeployStateDeployed, "task-1"); err != nil {
		t.Fatal(err)
	}
	op := New(nil, store)
	status, err := op.GetCoreStatus(ctx, "5gc")
	if err != nil {
		t.Fatalf("GetCoreStatus() error = %v", err)
	}
	if status.State != StateDeployed {
		t.Errorf("State = %q, want %q", status.State, StateDeployed)
	}
}

func TestGetGNBStatusNotDeployed(t *testing.T) {
	store := newTestStore(t)
	op := New(nil, store)
	status, err := op.GetGNBStatus(context.Background(), "srsran-gnb")
	if err != nil {
		t.Fatalf("GetGNBStatus() error = %v", err)
	}
	if status.State != StateNotDeployed {
		t.Errorf("State = %q, want %q", status.State, StateNotDeployed)
	}
}

func TestListCoresEmpty(t *testing.T) {
	store := newTestStore(t)
	op := New(nil, store)
	list, err := op.ListCores(context.Background())
	if err != nil {
		t.Fatalf("ListCores() error = %v", err)
	}
	if len(list.Cores) != 0 {
		t.Errorf("expected 0 cores, got %d", len(list.Cores))
	}
}

func TestListCoresWithDeployed(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	if err := store.SetDeploymentState(ctx, "5gc", state.DeployStateDeployed, "task-1"); err != nil {
		t.Fatal(err)
	}
	op := New(nil, store)
	list, err := op.ListCores(ctx)
	if err != nil {
		t.Fatalf("ListCores() error = %v", err)
	}
	if len(list.Cores) != 1 {
		t.Errorf("expected 1 core, got %d", len(list.Cores))
	}
}

func TestListGNBsEmpty(t *testing.T) {
	store := newTestStore(t)
	op := New(nil, store)
	list, err := op.ListGNBs(context.Background())
	if err != nil {
		t.Fatalf("ListGNBs() error = %v", err)
	}
	if len(list.GNBs) != 0 {
		t.Errorf("expected 0 gnbs, got %d", len(list.GNBs))
	}
}

func TestGetCoreKnownID(t *testing.T) {
	op := New(nil, newTestStore(t))
	cfg, err := op.GetCore(context.Background(), "5gc")
	if err != nil {
		t.Fatalf("GetCore() error = %v", err)
	}
	if cfg.ID != "5gc" {
		t.Errorf("ID = %q, want %q", cfg.ID, "5gc")
	}
}

func TestGetCoreUnknownID(t *testing.T) {
	op := New(nil, newTestStore(t))
	_, err := op.GetCore(context.Background(), "unknown")
	if !errors.Is(err, operator.ErrNotImplemented) {
		t.Errorf("GetCore(unknown) error = %v, want ErrNotImplemented", err)
	}
}

func TestGetGNBKnownID(t *testing.T) {
	op := New(nil, newTestStore(t))
	cfg, err := op.GetGNB(context.Background(), "srsran-gnb")
	if err != nil {
		t.Fatalf("GetGNB() error = %v", err)
	}
	if cfg.ID != "srsran-gnb" {
		t.Errorf("ID = %q, want %q", cfg.ID, "srsran-gnb")
	}
}

func TestGetGNBUnknownID(t *testing.T) {
	op := New(nil, newTestStore(t))
	_, err := op.GetGNB(context.Background(), "unknown")
	if !errors.Is(err, operator.ErrNotImplemented) {
		t.Errorf("GetGNB(unknown) error = %v, want ErrNotImplemented", err)
	}
}

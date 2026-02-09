package aether

import (
	"context"
	"errors"
	"testing"

	"github.com/bengrewell/aether-webui/internal/executor"
	"github.com/bengrewell/aether-webui/internal/operator"
)

func TestOperatorImplementsAetherOperator(t *testing.T) {
	var _ AetherOperator = (*Operator)(nil)
}

func TestNew(t *testing.T) {
	op := New(executor.NewMockExecutor())
	if op == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDomain(t *testing.T) {
	op := New(executor.NewMockExecutor())
	if got := op.Domain(); got != operator.DomainAether {
		t.Errorf("Domain() = %q, want %q", got, operator.DomainAether)
	}
}

func TestHealth(t *testing.T) {
	op := New(executor.NewMockExecutor())
	health, err := op.Health(context.Background())

	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health == nil {
		t.Fatal("Health() returned nil")
	}
	if health.Status != "unavailable" {
		t.Errorf("Status = %q, want %q", health.Status, "unavailable")
	}
	if health.Message != "not implemented" {
		t.Errorf("Message = %q, want %q", health.Message, "not implemented")
	}
}

func TestCoreMethodsReturnErrNotImplemented(t *testing.T) {
	op := New(executor.NewMockExecutor())
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"ListCores", func() error { _, err := op.ListCores(ctx); return err }},
		{"GetCore", func() error { _, err := op.GetCore(ctx, "core-1"); return err }},
		{"DeployCore", func() error { _, err := op.DeployCore(ctx, &CoreConfig{}); return err }},
		{"UpdateCore", func() error { return op.UpdateCore(ctx, "core-1", &CoreConfig{}) }},
		{"UndeployCore", func() error { _, err := op.UndeployCore(ctx, "core-1"); return err }},
		{"GetCoreStatus", func() error { _, err := op.GetCoreStatus(ctx, "core-1"); return err }},
		{"ListCoreStatuses", func() error { _, err := op.ListCoreStatuses(ctx); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("%s() error = %v, want ErrNotImplemented", tc.name, err)
			}
		})
	}
}

func TestGNBMethodsReturnErrNotImplemented(t *testing.T) {
	op := New(executor.NewMockExecutor())
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"ListGNBs", func() error { _, err := op.ListGNBs(ctx); return err }},
		{"GetGNB", func() error { _, err := op.GetGNB(ctx, "gnb-1"); return err }},
		{"DeployGNB", func() error { _, err := op.DeployGNB(ctx, &GNBConfig{}); return err }},
		{"UpdateGNB", func() error { return op.UpdateGNB(ctx, "gnb-1", &GNBConfig{}) }},
		{"UndeployGNB", func() error { _, err := op.UndeployGNB(ctx, "gnb-1"); return err }},
		{"GetGNBStatus", func() error { _, err := op.GetGNBStatus(ctx, "gnb-1"); return err }},
		{"ListGNBStatuses", func() error { _, err := op.ListGNBStatuses(ctx); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("%s() error = %v, want ErrNotImplemented", tc.name, err)
			}
		})
	}
}

func TestCoreMethodsReturnNilData(t *testing.T) {
	op := New(executor.NewMockExecutor())
	ctx := context.Background()

	t.Run("ListCores", func(t *testing.T) {
		data, _ := op.ListCores(ctx)
		if data != nil {
			t.Error("ListCores() returned non-nil data")
		}
	})

	t.Run("GetCore", func(t *testing.T) {
		data, _ := op.GetCore(ctx, "core-1")
		if data != nil {
			t.Error("GetCore() returned non-nil data")
		}
	})

	t.Run("DeployCore", func(t *testing.T) {
		data, _ := op.DeployCore(ctx, nil)
		if data != nil {
			t.Error("DeployCore() returned non-nil data")
		}
	})

	t.Run("UndeployCore", func(t *testing.T) {
		data, _ := op.UndeployCore(ctx, "core-1")
		if data != nil {
			t.Error("UndeployCore() returned non-nil data")
		}
	})

	t.Run("GetCoreStatus", func(t *testing.T) {
		data, _ := op.GetCoreStatus(ctx, "core-1")
		if data != nil {
			t.Error("GetCoreStatus() returned non-nil data")
		}
	})

	t.Run("ListCoreStatuses", func(t *testing.T) {
		data, _ := op.ListCoreStatuses(ctx)
		if data != nil {
			t.Error("ListCoreStatuses() returned non-nil data")
		}
	})
}

func TestGNBMethodsReturnNilData(t *testing.T) {
	op := New(executor.NewMockExecutor())
	ctx := context.Background()

	t.Run("ListGNBs", func(t *testing.T) {
		data, _ := op.ListGNBs(ctx)
		if data != nil {
			t.Error("ListGNBs() returned non-nil data")
		}
	})

	t.Run("GetGNB", func(t *testing.T) {
		data, _ := op.GetGNB(ctx, "gnb-1")
		if data != nil {
			t.Error("GetGNB() returned non-nil data")
		}
	})

	t.Run("DeployGNB", func(t *testing.T) {
		data, _ := op.DeployGNB(ctx, nil)
		if data != nil {
			t.Error("DeployGNB() returned non-nil data")
		}
	})

	t.Run("UndeployGNB", func(t *testing.T) {
		data, _ := op.UndeployGNB(ctx, "gnb-1")
		if data != nil {
			t.Error("UndeployGNB() returned non-nil data")
		}
	})

	t.Run("GetGNBStatus", func(t *testing.T) {
		data, _ := op.GetGNBStatus(ctx, "gnb-1")
		if data != nil {
			t.Error("GetGNBStatus() returned non-nil data")
		}
	})

	t.Run("ListGNBStatuses", func(t *testing.T) {
		data, _ := op.ListGNBStatuses(ctx)
		if data != nil {
			t.Error("ListGNBStatuses() returned non-nil data")
		}
	})
}

func TestMethodsWithDifferentIDs(t *testing.T) {
	op := New(executor.NewMockExecutor())
	ctx := context.Background()

	ids := []string{"", "core-1", "gnb-1", "test-id-12345", "id-with-dashes"}

	for _, id := range ids {
		t.Run("GetCore_"+id, func(t *testing.T) {
			_, err := op.GetCore(ctx, id)
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("GetCore(%q) error = %v, want ErrNotImplemented", id, err)
			}
		})

		t.Run("GetGNB_"+id, func(t *testing.T) {
			_, err := op.GetGNB(ctx, id)
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("GetGNB(%q) error = %v, want ErrNotImplemented", id, err)
			}
		})
	}
}

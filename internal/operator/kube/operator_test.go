package kube

import (
	"context"
	"errors"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator"
)

func TestOperatorImplementsKubeOperator(t *testing.T) {
	var _ KubeOperator = (*Operator)(nil)
}

func TestNew(t *testing.T) {
	op := New()
	if op == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDomain(t *testing.T) {
	op := New()
	if got := op.Domain(); got != operator.DomainKube {
		t.Errorf("Domain() = %q, want %q", got, operator.DomainKube)
	}
}

func TestHealth(t *testing.T) {
	op := New()
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

func TestMethodsReturnErrNotImplemented(t *testing.T) {
	op := New()
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"GetClusterHealth", func() error { _, err := op.GetClusterHealth(ctx); return err }},
		{"GetNodes", func() error { _, err := op.GetNodes(ctx); return err }},
		{"GetNamespaces", func() error { _, err := op.GetNamespaces(ctx); return err }},
		{"GetEvents", func() error { _, err := op.GetEvents(ctx, "default", 50); return err }},
		{"GetPods", func() error { _, err := op.GetPods(ctx, "default"); return err }},
		{"GetDeployments", func() error { _, err := op.GetDeployments(ctx, "default"); return err }},
		{"GetServices", func() error { _, err := op.GetServices(ctx, "default"); return err }},
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

func TestMethodsReturnNilData(t *testing.T) {
	op := New()
	ctx := context.Background()

	t.Run("GetClusterHealth", func(t *testing.T) {
		data, _ := op.GetClusterHealth(ctx)
		if data != nil {
			t.Error("GetClusterHealth() returned non-nil data")
		}
	})

	t.Run("GetNodes", func(t *testing.T) {
		data, _ := op.GetNodes(ctx)
		if data != nil {
			t.Error("GetNodes() returned non-nil data")
		}
	})

	t.Run("GetNamespaces", func(t *testing.T) {
		data, _ := op.GetNamespaces(ctx)
		if data != nil {
			t.Error("GetNamespaces() returned non-nil data")
		}
	})

	t.Run("GetEvents", func(t *testing.T) {
		data, _ := op.GetEvents(ctx, "", 100)
		if data != nil {
			t.Error("GetEvents() returned non-nil data")
		}
	})

	t.Run("GetPods", func(t *testing.T) {
		data, _ := op.GetPods(ctx, "")
		if data != nil {
			t.Error("GetPods() returned non-nil data")
		}
	})

	t.Run("GetDeployments", func(t *testing.T) {
		data, _ := op.GetDeployments(ctx, "")
		if data != nil {
			t.Error("GetDeployments() returned non-nil data")
		}
	})

	t.Run("GetServices", func(t *testing.T) {
		data, _ := op.GetServices(ctx, "")
		if data != nil {
			t.Error("GetServices() returned non-nil data")
		}
	})
}

func TestMethodsWithDifferentParameters(t *testing.T) {
	op := New()
	ctx := context.Background()

	// Test GetEvents with different limits
	t.Run("GetEventsVaryingLimits", func(t *testing.T) {
		limits := []int{0, 1, 50, 100, 500}
		for _, limit := range limits {
			_, err := op.GetEvents(ctx, "default", limit)
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("GetEvents(limit=%d) error = %v, want ErrNotImplemented", limit, err)
			}
		}
	})

	// Test namespace-scoped methods with different namespaces
	t.Run("NamespaceScopedMethods", func(t *testing.T) {
		namespaces := []string{"", "default", "kube-system", "custom-namespace"}
		for _, ns := range namespaces {
			_, err := op.GetPods(ctx, ns)
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("GetPods(namespace=%q) error = %v, want ErrNotImplemented", ns, err)
			}
			_, err = op.GetDeployments(ctx, ns)
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("GetDeployments(namespace=%q) error = %v, want ErrNotImplemented", ns, err)
			}
			_, err = op.GetServices(ctx, ns)
			if !errors.Is(err, operator.ErrNotImplemented) {
				t.Errorf("GetServices(namespace=%q) error = %v, want ErrNotImplemented", ns, err)
			}
		}
	})
}

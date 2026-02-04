package host

import (
	"context"
	"errors"
	"testing"

	"github.com/bengrewell/aether-webui/internal/operator"
)

func TestOperatorImplementsHostOperator(t *testing.T) {
	var _ HostOperator = (*Operator)(nil)
}

func TestNew(t *testing.T) {
	op := New()
	if op == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDomain(t *testing.T) {
	op := New()
	if got := op.Domain(); got != operator.DomainHost {
		t.Errorf("Domain() = %q, want %q", got, operator.DomainHost)
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
		{"GetCPUInfo", func() error { _, err := op.GetCPUInfo(ctx); return err }},
		{"GetMemoryInfo", func() error { _, err := op.GetMemoryInfo(ctx); return err }},
		{"GetDiskInfo", func() error { _, err := op.GetDiskInfo(ctx); return err }},
		{"GetNICInfo", func() error { _, err := op.GetNICInfo(ctx); return err }},
		{"GetOSInfo", func() error { _, err := op.GetOSInfo(ctx); return err }},
		{"GetCPUUsage", func() error { _, err := op.GetCPUUsage(ctx); return err }},
		{"GetMemoryUsage", func() error { _, err := op.GetMemoryUsage(ctx); return err }},
		{"GetDiskUsage", func() error { _, err := op.GetDiskUsage(ctx); return err }},
		{"GetNICUsage", func() error { _, err := op.GetNICUsage(ctx); return err }},
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

	t.Run("GetCPUInfo", func(t *testing.T) {
		data, _ := op.GetCPUInfo(ctx)
		if data != nil {
			t.Error("GetCPUInfo() returned non-nil data")
		}
	})

	t.Run("GetMemoryInfo", func(t *testing.T) {
		data, _ := op.GetMemoryInfo(ctx)
		if data != nil {
			t.Error("GetMemoryInfo() returned non-nil data")
		}
	})

	t.Run("GetDiskInfo", func(t *testing.T) {
		data, _ := op.GetDiskInfo(ctx)
		if data != nil {
			t.Error("GetDiskInfo() returned non-nil data")
		}
	})

	t.Run("GetNICInfo", func(t *testing.T) {
		data, _ := op.GetNICInfo(ctx)
		if data != nil {
			t.Error("GetNICInfo() returned non-nil data")
		}
	})

	t.Run("GetOSInfo", func(t *testing.T) {
		data, _ := op.GetOSInfo(ctx)
		if data != nil {
			t.Error("GetOSInfo() returned non-nil data")
		}
	})

	t.Run("GetCPUUsage", func(t *testing.T) {
		data, _ := op.GetCPUUsage(ctx)
		if data != nil {
			t.Error("GetCPUUsage() returned non-nil data")
		}
	})

	t.Run("GetMemoryUsage", func(t *testing.T) {
		data, _ := op.GetMemoryUsage(ctx)
		if data != nil {
			t.Error("GetMemoryUsage() returned non-nil data")
		}
	})

	t.Run("GetDiskUsage", func(t *testing.T) {
		data, _ := op.GetDiskUsage(ctx)
		if data != nil {
			t.Error("GetDiskUsage() returned non-nil data")
		}
	})

	t.Run("GetNICUsage", func(t *testing.T) {
		data, _ := op.GetNICUsage(ctx)
		if data != nil {
			t.Error("GetNICUsage() returned non-nil data")
		}
	})
}

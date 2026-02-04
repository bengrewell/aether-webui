package metrics

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/operator"
	"github.com/bengrewell/aether-webui/internal/operator/host"
	"github.com/bengrewell/aether-webui/internal/state"
)

// mockHostOperator implements host.HostOperator for testing.
type mockHostOperator struct {
	cpuUsage    *host.CPUUsage
	memUsage    *host.MemoryUsage
	diskUsage   *host.DiskUsage
	nicUsage    *host.NICUsage
	callCount   int
	cpuCallErr  error
	memCallErr  error
	diskCallErr error
	nicCallErr  error
}

func (m *mockHostOperator) Domain() operator.Domain {
	return operator.DomainHost
}

func (m *mockHostOperator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{Status: "healthy"}, nil
}

func (m *mockHostOperator) GetCPUInfo(_ context.Context) (*host.CPUInfo, error) {
	return nil, nil
}

func (m *mockHostOperator) GetMemoryInfo(_ context.Context) (*host.MemoryInfo, error) {
	return nil, nil
}

func (m *mockHostOperator) GetDiskInfo(_ context.Context) (*host.DiskInfo, error) {
	return nil, nil
}

func (m *mockHostOperator) GetNICInfo(_ context.Context) (*host.NICInfo, error) {
	return nil, nil
}

func (m *mockHostOperator) GetOSInfo(_ context.Context) (*host.OSInfo, error) {
	return nil, nil
}

func (m *mockHostOperator) GetCPUUsage(_ context.Context) (*host.CPUUsage, error) {
	m.callCount++
	if m.cpuCallErr != nil {
		return nil, m.cpuCallErr
	}
	return m.cpuUsage, nil
}

func (m *mockHostOperator) GetMemoryUsage(_ context.Context) (*host.MemoryUsage, error) {
	if m.memCallErr != nil {
		return nil, m.memCallErr
	}
	return m.memUsage, nil
}

func (m *mockHostOperator) GetDiskUsage(_ context.Context) (*host.DiskUsage, error) {
	if m.diskCallErr != nil {
		return nil, m.diskCallErr
	}
	return m.diskUsage, nil
}

func (m *mockHostOperator) GetNICUsage(_ context.Context) (*host.NICUsage, error) {
	if m.nicCallErr != nil {
		return nil, m.nicCallErr
	}
	return m.nicUsage, nil
}

func newTestStore(t *testing.T) state.Store {
	t.Helper()
	store, err := state.NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Interval != 10*time.Second {
		t.Errorf("expected 10s interval, got %v", cfg.Interval)
	}
	if cfg.Retention != 24*time.Hour {
		t.Errorf("expected 24h retention, got %v", cfg.Retention)
	}
}

func TestNewCollectorDefaults(t *testing.T) {
	mockOp := &mockHostOperator{}
	store := newTestStore(t)

	// Empty config should use defaults
	collector := NewCollector(mockOp, store, Config{})
	if collector.config.Interval != 10*time.Second {
		t.Errorf("expected default interval, got %v", collector.config.Interval)
	}
	if collector.config.Retention != 24*time.Hour {
		t.Errorf("expected default retention, got %v", collector.config.Retention)
	}
}

func TestNewCollectorCustomConfig(t *testing.T) {
	mockOp := &mockHostOperator{}
	store := newTestStore(t)

	cfg := Config{
		Interval:  30 * time.Second,
		Retention: 48 * time.Hour,
	}
	collector := NewCollector(mockOp, store, cfg)
	if collector.config.Interval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", collector.config.Interval)
	}
	if collector.config.Retention != 48*time.Hour {
		t.Errorf("expected 48h retention, got %v", collector.config.Retention)
	}
}

func TestCollectorStartAndStop(t *testing.T) {
	mockOp := &mockHostOperator{
		cpuUsage:  &host.CPUUsage{UsagePercent: 50.0},
		memUsage:  &host.MemoryUsage{UsagePercent: 60.0},
		diskUsage: &host.DiskUsage{},
		nicUsage:  &host.NICUsage{},
	}
	store := newTestStore(t)

	cfg := Config{
		Interval:  50 * time.Millisecond,
		Retention: time.Hour,
	}
	collector := NewCollector(mockOp, store, cfg)

	ctx, cancel := context.WithCancel(context.Background())

	// Start collector in background
	done := make(chan error, 1)
	go func() {
		done <- collector.Start(ctx)
	}()

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Stop via context cancellation
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("collector did not stop within timeout")
	}

	// Verify some metrics were collected
	if mockOp.callCount == 0 {
		t.Error("expected at least one collection call")
	}
}

func TestCollectorStopMethod(t *testing.T) {
	mockOp := &mockHostOperator{
		cpuUsage:  &host.CPUUsage{UsagePercent: 50.0},
		memUsage:  &host.MemoryUsage{UsagePercent: 60.0},
		diskUsage: &host.DiskUsage{},
		nicUsage:  &host.NICUsage{},
	}
	store := newTestStore(t)

	cfg := Config{
		Interval:  100 * time.Millisecond,
		Retention: time.Hour,
	}
	collector := NewCollector(mockOp, store, cfg)

	ctx := context.Background()

	// Start collector in background
	done := make(chan error, 1)
	go func() {
		done <- collector.Start(ctx)
	}()

	// Let it start
	time.Sleep(50 * time.Millisecond)

	// Stop via Stop method
	collector.Stop()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Fatal("collector did not stop within timeout")
	}
}

func TestCollectorStoresMetrics(t *testing.T) {
	mockOp := &mockHostOperator{
		cpuUsage:  &host.CPUUsage{UsagePercent: 42.5},
		memUsage:  &host.MemoryUsage{UsedBytes: 1234567890, UsagePercent: 55.0},
		diskUsage: &host.DiskUsage{Disks: []host.DiskUsageEntry{{MountPoint: "/", UsagePercent: 30.0}}},
		nicUsage:  &host.NICUsage{Interfaces: []host.NICUsageEntry{{Name: "eth0", RxBytes: 1000}}},
	}
	store := newTestStore(t)

	cfg := Config{
		Interval:  50 * time.Millisecond,
		Retention: time.Hour,
	}
	collector := NewCollector(mockOp, store, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start collector
	go collector.Start(ctx)

	// Wait for at least one collection
	time.Sleep(100 * time.Millisecond)

	// Check that metrics were stored
	cpuSnapshots, err := store.GetMetricsHistory(ctx, string(MetricTypeCPU), 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory failed: %v", err)
	}
	if len(cpuSnapshots) == 0 {
		t.Error("expected at least one CPU snapshot")
	} else {
		var cpu host.CPUUsage
		if err := json.Unmarshal(cpuSnapshots[0].Data, &cpu); err != nil {
			t.Errorf("failed to unmarshal CPU data: %v", err)
		} else if cpu.UsagePercent != 42.5 {
			t.Errorf("expected UsagePercent 42.5, got %f", cpu.UsagePercent)
		}
	}

	memSnapshots, err := store.GetMetricsHistory(ctx, string(MetricTypeMemory), 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory failed: %v", err)
	}
	if len(memSnapshots) == 0 {
		t.Error("expected at least one memory snapshot")
	}

	diskSnapshots, err := store.GetMetricsHistory(ctx, string(MetricTypeDisk), 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory failed: %v", err)
	}
	if len(diskSnapshots) == 0 {
		t.Error("expected at least one disk snapshot")
	}

	nicSnapshots, err := store.GetMetricsHistory(ctx, string(MetricTypeNIC), 10)
	if err != nil {
		t.Fatalf("GetMetricsHistory failed: %v", err)
	}
	if len(nicSnapshots) == 0 {
		t.Error("expected at least one NIC snapshot")
	}

	collector.Stop()
}

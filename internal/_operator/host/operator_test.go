package host

import (
	"context"
	"testing"
	"time"

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
	if op.lastNICStats == nil {
		t.Error("expected lastNICStats map to be initialized")
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
	if health.Status != "healthy" {
		t.Errorf("Status = %q, want %q", health.Status, "healthy")
	}
}

func TestGetCPUInfo(t *testing.T) {
	op := New()
	info, err := op.GetCPUInfo(context.Background())
	if err != nil {
		t.Fatalf("GetCPUInfo failed: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil CPUInfo")
	}
	// Basic sanity checks
	if info.Threads <= 0 {
		t.Errorf("expected at least 1 thread, got %d", info.Threads)
	}
}

func TestGetCPUInfoCaching(t *testing.T) {
	op := New()
	ctx := context.Background()

	// First call
	info1, err := op.GetCPUInfo(ctx)
	if err != nil {
		t.Fatalf("GetCPUInfo (1) failed: %v", err)
	}

	// Second call should return cached value
	info2, err := op.GetCPUInfo(ctx)
	if err != nil {
		t.Fatalf("GetCPUInfo (2) failed: %v", err)
	}

	// Should be the same pointer (cached)
	if info1 != info2 {
		t.Error("expected cached value to be returned")
	}
}

func TestGetMemoryInfo(t *testing.T) {
	op := New()
	info, err := op.GetMemoryInfo(context.Background())
	if err != nil {
		t.Fatalf("GetMemoryInfo failed: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil MemoryInfo")
	}
	if info.TotalBytes == 0 {
		t.Error("expected non-zero TotalBytes")
	}
}

func TestGetMemoryInfoCaching(t *testing.T) {
	op := New()
	ctx := context.Background()

	info1, err := op.GetMemoryInfo(ctx)
	if err != nil {
		t.Fatalf("GetMemoryInfo (1) failed: %v", err)
	}

	info2, err := op.GetMemoryInfo(ctx)
	if err != nil {
		t.Fatalf("GetMemoryInfo (2) failed: %v", err)
	}

	if info1 != info2 {
		t.Error("expected cached value to be returned")
	}
}

func TestGetDiskInfo(t *testing.T) {
	op := New()
	info, err := op.GetDiskInfo(context.Background())
	if err != nil {
		t.Fatalf("GetDiskInfo failed: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil DiskInfo")
	}
	// Most systems have at least one disk
	if len(info.Disks) == 0 {
		t.Log("no disks found - this may be expected in some environments")
	}
}

func TestGetNICInfo(t *testing.T) {
	op := New()
	info, err := op.GetNICInfo(context.Background())
	if err != nil {
		t.Fatalf("GetNICInfo failed: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil NICInfo")
	}
}

func TestGetOSInfo(t *testing.T) {
	op := New()
	info, err := op.GetOSInfo(context.Background())
	if err != nil {
		t.Fatalf("GetOSInfo failed: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil OSInfo")
	}
	if info.Hostname == "" {
		t.Error("expected non-empty Hostname")
	}
	if info.Architecture == "" {
		t.Error("expected non-empty Architecture")
	}
}

func TestGetCPUUsage(t *testing.T) {
	op := New()
	usage, err := op.GetCPUUsage(context.Background())
	if err != nil {
		t.Fatalf("GetCPUUsage failed: %v", err)
	}
	if usage == nil {
		t.Fatal("expected non-nil CPUUsage")
	}
	// Usage percent should be between 0 and 100
	if usage.UsagePercent < 0 || usage.UsagePercent > 100 {
		t.Errorf("UsagePercent %f out of range [0, 100]", usage.UsagePercent)
	}
}

func TestGetMemoryUsage(t *testing.T) {
	op := New()
	usage, err := op.GetMemoryUsage(context.Background())
	if err != nil {
		t.Fatalf("GetMemoryUsage failed: %v", err)
	}
	if usage == nil {
		t.Fatal("expected non-nil MemoryUsage")
	}
	// UsedBytes should be non-zero on most systems
	if usage.UsedBytes == 0 {
		t.Error("expected non-zero UsedBytes")
	}
	// Usage percent should be between 0 and 100
	if usage.UsagePercent < 0 || usage.UsagePercent > 100 {
		t.Errorf("UsagePercent %f out of range [0, 100]", usage.UsagePercent)
	}
}

func TestGetDiskUsage(t *testing.T) {
	op := New()
	usage, err := op.GetDiskUsage(context.Background())
	if err != nil {
		t.Fatalf("GetDiskUsage failed: %v", err)
	}
	if usage == nil {
		t.Fatal("expected non-nil DiskUsage")
	}
}

func TestGetNICUsage(t *testing.T) {
	op := New()
	usage, err := op.GetNICUsage(context.Background())
	if err != nil {
		t.Fatalf("GetNICUsage failed: %v", err)
	}
	if usage == nil {
		t.Fatal("expected non-nil NICUsage")
	}
}

func TestGetNICUsageRateCalculation(t *testing.T) {
	op := New()
	ctx := context.Background()

	// First call to establish baseline
	usage1, err := op.GetNICUsage(ctx)
	if err != nil {
		t.Fatalf("GetNICUsage (1) failed: %v", err)
	}

	// Small delay
	time.Sleep(100 * time.Millisecond)

	// Second call should have rate calculations
	usage2, err := op.GetNICUsage(ctx)
	if err != nil {
		t.Fatalf("GetNICUsage (2) failed: %v", err)
	}

	// Just verify we got results; rates may be 0 if no traffic
	if usage1 == nil || usage2 == nil {
		t.Fatal("expected non-nil results")
	}
}

func TestDetectDiskType(t *testing.T) {
	tests := []struct {
		device string
		fstype string
		want   string
	}{
		{"/dev/nvme0n1p1", "ext4", "nvme"},
		{"/dev/sda1", "ext4", "ssd"},
		{"/dev/hda1", "ext4", "hdd"},
		{"/dev/vda1", "ext4", "virtual"},
		{"/dev/xvda1", "xfs", "virtual"},
		{"/dev/unknown", "ext4", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.device, func(t *testing.T) {
			got := detectDiskType(tt.device, tt.fstype)
			if got != tt.want {
				t.Errorf("detectDiskType(%q, %q) = %q, want %q", tt.device, tt.fstype, got, tt.want)
			}
		})
	}
}

func TestCacheExpiration(t *testing.T) {
	// This test verifies the cache logic works
	cache := cachedItem[CPUInfo]{
		value:     &CPUInfo{Model: "test"},
		expiresAt: time.Now().Add(-time.Second), // Already expired
	}

	if cache.isValid() {
		t.Error("expected expired cache to be invalid")
	}

	cache.expiresAt = time.Now().Add(time.Hour) // Not expired
	if !cache.isValid() {
		t.Error("expected non-expired cache to be valid")
	}

	cache.value = nil // Nil value
	if cache.isValid() {
		t.Error("expected nil value cache to be invalid")
	}
}

func TestConcurrentCacheAccess(t *testing.T) {
	op := New()
	ctx := context.Background()

	// Run concurrent reads to test cache thread safety
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := op.GetCPUInfo(ctx)
			if err != nil {
				t.Errorf("concurrent GetCPUInfo failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

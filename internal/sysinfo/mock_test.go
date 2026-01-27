package sysinfo

import (
	"context"
	"testing"
)

func TestNewMockProvider(t *testing.T) {
	provider := NewMockProvider()
	if provider == nil {
		t.Fatal("NewMockProvider returned nil")
	}
}

func TestMockProvider_GetCPUInfo(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	info, err := provider.GetCPUInfo(ctx)
	if err != nil {
		t.Fatalf("GetCPUInfo returned error: %v", err)
	}
	if info == nil {
		t.Fatal("GetCPUInfo returned nil")
	}

	// Verify required fields are populated
	if info.Model == "" {
		t.Error("CPUInfo.Model is empty")
	}
	if info.Vendor == "" {
		t.Error("CPUInfo.Vendor is empty")
	}
	if info.Cores <= 0 {
		t.Errorf("CPUInfo.Cores should be positive, got %d", info.Cores)
	}
	if info.Threads <= 0 {
		t.Errorf("CPUInfo.Threads should be positive, got %d", info.Threads)
	}
	if info.FrequencyMHz <= 0 {
		t.Errorf("CPUInfo.FrequencyMHz should be positive, got %f", info.FrequencyMHz)
	}
}

func TestMockProvider_GetMemoryInfo(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	info, err := provider.GetMemoryInfo(ctx)
	if err != nil {
		t.Fatalf("GetMemoryInfo returned error: %v", err)
	}
	if info == nil {
		t.Fatal("GetMemoryInfo returned nil")
	}

	if info.TotalBytes == 0 {
		t.Error("MemoryInfo.TotalBytes is zero")
	}
	if info.Type == "" {
		t.Error("MemoryInfo.Type is empty")
	}
	if info.SpeedMHz <= 0 {
		t.Errorf("MemoryInfo.SpeedMHz should be positive, got %d", info.SpeedMHz)
	}
}

func TestMockProvider_GetDiskInfo(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	info, err := provider.GetDiskInfo(ctx)
	if err != nil {
		t.Fatalf("GetDiskInfo returned error: %v", err)
	}
	if info == nil {
		t.Fatal("GetDiskInfo returned nil")
	}

	if len(info.Disks) == 0 {
		t.Error("DiskInfo.Disks is empty")
	}

	for i, disk := range info.Disks {
		if disk.Device == "" {
			t.Errorf("Disk[%d].Device is empty", i)
		}
		if disk.SizeBytes == 0 {
			t.Errorf("Disk[%d].SizeBytes is zero", i)
		}
		if disk.Type == "" {
			t.Errorf("Disk[%d].Type is empty", i)
		}
	}
}

func TestMockProvider_GetNICInfo(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	info, err := provider.GetNICInfo(ctx)
	if err != nil {
		t.Fatalf("GetNICInfo returned error: %v", err)
	}
	if info == nil {
		t.Fatal("GetNICInfo returned nil")
	}

	if len(info.Interfaces) == 0 {
		t.Error("NICInfo.Interfaces is empty")
	}

	for i, nic := range info.Interfaces {
		if nic.Name == "" {
			t.Errorf("Interface[%d].Name is empty", i)
		}
		if nic.MACAddress == "" {
			t.Errorf("Interface[%d].MACAddress is empty", i)
		}
		if nic.MTU <= 0 {
			t.Errorf("Interface[%d].MTU should be positive, got %d", i, nic.MTU)
		}
	}
}

func TestMockProvider_GetOSInfo(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	info, err := provider.GetOSInfo(ctx)
	if err != nil {
		t.Fatalf("GetOSInfo returned error: %v", err)
	}
	if info == nil {
		t.Fatal("GetOSInfo returned nil")
	}

	if info.Name == "" {
		t.Error("OSInfo.Name is empty")
	}
	if info.Version == "" {
		t.Error("OSInfo.Version is empty")
	}
	if info.Kernel == "" {
		t.Error("OSInfo.Kernel is empty")
	}
	if info.Architecture == "" {
		t.Error("OSInfo.Architecture is empty")
	}
	if info.Hostname == "" {
		t.Error("OSInfo.Hostname is empty")
	}
}

func TestMockProvider_GetCPUUsage(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	usage, err := provider.GetCPUUsage(ctx)
	if err != nil {
		t.Fatalf("GetCPUUsage returned error: %v", err)
	}
	if usage == nil {
		t.Fatal("GetCPUUsage returned nil")
	}

	if usage.UsagePercent < 0 || usage.UsagePercent > 100 {
		t.Errorf("CPUUsage.UsagePercent out of range: %f", usage.UsagePercent)
	}
	if usage.IdlePercent < 0 || usage.IdlePercent > 100 {
		t.Errorf("CPUUsage.IdlePercent out of range: %f", usage.IdlePercent)
	}
	if len(usage.PerCoreUsage) == 0 {
		t.Error("CPUUsage.PerCoreUsage is empty")
	}
	if usage.LoadAverage1 < 0 {
		t.Errorf("CPUUsage.LoadAverage1 should be non-negative, got %f", usage.LoadAverage1)
	}
}

func TestMockProvider_GetMemoryUsage(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	usage, err := provider.GetMemoryUsage(ctx)
	if err != nil {
		t.Fatalf("GetMemoryUsage returned error: %v", err)
	}
	if usage == nil {
		t.Fatal("GetMemoryUsage returned nil")
	}

	if usage.UsagePercent < 0 || usage.UsagePercent > 100 {
		t.Errorf("MemoryUsage.UsagePercent out of range: %f", usage.UsagePercent)
	}
	// UsedBytes + FreeBytes should roughly equal total (allowing for cached/buffers)
	if usage.UsedBytes == 0 {
		t.Error("MemoryUsage.UsedBytes is zero")
	}
}

func TestMockProvider_GetDiskUsage(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	usage, err := provider.GetDiskUsage(ctx)
	if err != nil {
		t.Fatalf("GetDiskUsage returned error: %v", err)
	}
	if usage == nil {
		t.Fatal("GetDiskUsage returned nil")
	}

	if len(usage.Disks) == 0 {
		t.Error("DiskUsage.Disks is empty")
	}

	for i, disk := range usage.Disks {
		if disk.Device == "" {
			t.Errorf("DiskUsage[%d].Device is empty", i)
		}
		if disk.MountPoint == "" {
			t.Errorf("DiskUsage[%d].MountPoint is empty", i)
		}
		if disk.UsagePercent < 0 || disk.UsagePercent > 100 {
			t.Errorf("DiskUsage[%d].UsagePercent out of range: %f", i, disk.UsagePercent)
		}
		if disk.UsedBytes+disk.FreeBytes != disk.TotalBytes {
			t.Errorf("DiskUsage[%d]: UsedBytes + FreeBytes != TotalBytes", i)
		}
	}
}

func TestMockProvider_GetNICUsage(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	usage, err := provider.GetNICUsage(ctx)
	if err != nil {
		t.Fatalf("GetNICUsage returned error: %v", err)
	}
	if usage == nil {
		t.Fatal("GetNICUsage returned nil")
	}

	if len(usage.Interfaces) == 0 {
		t.Error("NICUsage.Interfaces is empty")
	}

	for i, nic := range usage.Interfaces {
		if nic.Name == "" {
			t.Errorf("NICUsage[%d].Name is empty", i)
		}
		// Bytes and packets should be non-negative
		if nic.RxBytesPerSec < 0 {
			t.Errorf("NICUsage[%d].RxBytesPerSec is negative", i)
		}
		if nic.TxBytesPerSec < 0 {
			t.Errorf("NICUsage[%d].TxBytesPerSec is negative", i)
		}
	}
}

// TestMockProviderImplementsInterface verifies the interface is properly implemented
func TestMockProviderImplementsInterface(t *testing.T) {
	var _ Provider = (*MockProvider)(nil)
}

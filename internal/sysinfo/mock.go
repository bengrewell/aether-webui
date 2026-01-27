package sysinfo

import "context"

// MockProvider returns static mock data for all system information.
// This is used during development before real implementations are added.
type MockProvider struct{}

// NewMockProvider creates a new MockProvider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) GetCPUInfo(ctx context.Context) (*CPUInfo, error) {
	return &CPUInfo{
		Model:        "Intel(R) Core(TM) i9-13900K",
		Vendor:       "GenuineIntel",
		Cores:        24,
		Threads:      32,
		FrequencyMHz: 3000.0,
		CacheKB:      36864,
	}, nil
}

func (m *MockProvider) GetMemoryInfo(ctx context.Context) (*MemoryInfo, error) {
	return &MemoryInfo{
		TotalBytes: 68719476736, // 64 GB
		Type:       "DDR5",
		SpeedMHz:   5600,
		Slots:      4,
		SlotsUsed:  2,
	}, nil
}

func (m *MockProvider) GetDiskInfo(ctx context.Context) (*DiskInfo, error) {
	return &DiskInfo{
		Disks: []Disk{
			{
				Device:     "/dev/nvme0n1",
				Model:      "Samsung 990 PRO 2TB",
				SizeBytes:  2000398934016,
				Type:       "nvme",
				MountPoint: "/",
			},
			{
				Device:     "/dev/sda",
				Model:      "WD Red Plus 4TB",
				SizeBytes:  4000787030016,
				Type:       "hdd",
				MountPoint: "/data",
			},
		},
	}, nil
}

func (m *MockProvider) GetNICInfo(ctx context.Context) (*NICInfo, error) {
	return &NICInfo{
		Interfaces: []NetworkInterface{
			{
				Name:       "eth0",
				MACAddress: "00:1a:2b:3c:4d:5e",
				Driver:     "e1000e",
				SpeedMbps:  1000,
				MTU:        1500,
				IPAddresses: []string{"192.168.1.100/24", "fe80::1a2b:3cff:fe4d:5e00/64"},
			},
			{
				Name:       "eth1",
				MACAddress: "00:1a:2b:3c:4d:5f",
				Driver:     "igb",
				SpeedMbps:  10000,
				MTU:        9000,
				IPAddresses: []string{"10.0.0.1/24"},
			},
		},
	}, nil
}

func (m *MockProvider) GetOSInfo(ctx context.Context) (*OSInfo, error) {
	return &OSInfo{
		Name:         "Ubuntu",
		Version:      "24.04 LTS",
		Kernel:       "6.8.0-90-generic",
		Architecture: "x86_64",
		Hostname:     "aether-node-01",
		Uptime:       864000, // 10 days
	}, nil
}

func (m *MockProvider) GetCPUUsage(ctx context.Context) (*CPUUsage, error) {
	return &CPUUsage{
		UsagePercent:   23.5,
		UserPercent:    15.2,
		SystemPercent:  8.3,
		IdlePercent:    76.5,
		IOWaitPercent:  0.8,
		PerCoreUsage:   []float64{25.1, 22.3, 28.7, 19.2, 24.0, 21.5, 26.3, 23.1},
		LoadAverage1:   1.25,
		LoadAverage5:   1.10,
		LoadAverage15:  0.95,
	}, nil
}

func (m *MockProvider) GetMemoryUsage(ctx context.Context) (*MemoryUsage, error) {
	return &MemoryUsage{
		UsedBytes:      32212254720,  // ~30 GB
		FreeBytes:      17179869184,  // ~16 GB
		AvailableBytes: 34359738368,  // ~32 GB
		CachedBytes:    15032385536,  // ~14 GB
		BuffersBytes:   2147483648,   // ~2 GB
		SwapTotalBytes: 8589934592,   // 8 GB
		SwapUsedBytes:  104857600,    // 100 MB
		UsagePercent:   46.9,
	}, nil
}

func (m *MockProvider) GetDiskUsage(ctx context.Context) (*DiskUsage, error) {
	return &DiskUsage{
		Disks: []DiskUsageEntry{
			{
				Device:       "/dev/nvme0n1p1",
				MountPoint:   "/",
				TotalBytes:   2000398934016,
				UsedBytes:    500000000000,
				FreeBytes:    1500398934016,
				UsagePercent: 25.0,
				InodesTotal:  122068992,
				InodesUsed:   1250000,
			},
			{
				Device:       "/dev/sda1",
				MountPoint:   "/data",
				TotalBytes:   4000787030016,
				UsedBytes:    2400472218009,
				FreeBytes:    1600314812007,
				UsagePercent: 60.0,
				InodesTotal:  244137984,
				InodesUsed:   5000000,
			},
		},
	}, nil
}

func (m *MockProvider) GetNICUsage(ctx context.Context) (*NICUsage, error) {
	return &NICUsage{
		Interfaces: []NICUsageEntry{
			{
				Name:          "eth0",
				RxBytes:       1099511627776, // ~1 TB
				TxBytes:       549755813888,  // ~512 GB
				RxPackets:     750000000,
				TxPackets:     500000000,
				RxErrors:      12,
				TxErrors:      3,
				RxDropped:     45,
				TxDropped:     8,
				RxBytesPerSec: 52428800,  // ~50 MB/s
				TxBytesPerSec: 26214400,  // ~25 MB/s
			},
			{
				Name:          "eth1",
				RxBytes:       5497558138880,  // ~5 TB
				TxBytes:       2748779069440,  // ~2.5 TB
				RxPackets:     3750000000,
				TxPackets:     2500000000,
				RxErrors:      0,
				TxErrors:      0,
				RxDropped:     100,
				TxDropped:     25,
				RxBytesPerSec: 524288000,  // ~500 MB/s
				TxBytesPerSec: 262144000,  // ~250 MB/s
			},
		},
	}, nil
}

// Ensure MockProvider implements Provider
var _ Provider = (*MockProvider)(nil)

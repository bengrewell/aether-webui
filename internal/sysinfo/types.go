package sysinfo

// CPUInfo contains static CPU information.
type CPUInfo struct {
	Model       string  `json:"model"`
	Vendor      string  `json:"vendor"`
	Cores       int     `json:"cores"`
	Threads     int     `json:"threads"`
	FrequencyMHz float64 `json:"frequency_mhz"`
	CacheKB     int     `json:"cache_kb"`
}

// MemoryInfo contains static memory information.
type MemoryInfo struct {
	TotalBytes uint64 `json:"total_bytes"`
	Type       string `json:"type"`
	SpeedMHz   int    `json:"speed_mhz"`
	Slots      int    `json:"slots"`
	SlotsUsed  int    `json:"slots_used"`
}

// DiskInfo contains static disk information.
type DiskInfo struct {
	Disks []Disk `json:"disks"`
}

// Disk represents a single disk device.
type Disk struct {
	Device     string `json:"device"`
	Model      string `json:"model"`
	SizeBytes  uint64 `json:"size_bytes"`
	Type       string `json:"type"` // ssd, hdd, nvme
	MountPoint string `json:"mount_point,omitempty"`
}

// NICInfo contains static network interface information.
type NICInfo struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

// NetworkInterface represents a single network interface.
type NetworkInterface struct {
	Name       string   `json:"name"`
	MACAddress string   `json:"mac_address"`
	Driver     string   `json:"driver"`
	SpeedMbps  int      `json:"speed_mbps"`
	MTU        int      `json:"mtu"`
	IPAddresses []string `json:"ip_addresses,omitempty"`
}

// OSInfo contains static operating system information.
type OSInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	Hostname     string `json:"hostname"`
	Uptime       int64  `json:"uptime_seconds"`
}

// CPUUsage contains dynamic CPU usage metrics.
type CPUUsage struct {
	UsagePercent    float64    `json:"usage_percent"`
	UserPercent     float64    `json:"user_percent"`
	SystemPercent   float64    `json:"system_percent"`
	IdlePercent     float64    `json:"idle_percent"`
	IOWaitPercent   float64    `json:"iowait_percent"`
	PerCoreUsage    []float64  `json:"per_core_usage"`
	LoadAverage1    float64    `json:"load_average_1"`
	LoadAverage5    float64    `json:"load_average_5"`
	LoadAverage15   float64    `json:"load_average_15"`
}

// MemoryUsage contains dynamic memory usage metrics.
type MemoryUsage struct {
	UsedBytes      uint64  `json:"used_bytes"`
	FreeBytes      uint64  `json:"free_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	CachedBytes    uint64  `json:"cached_bytes"`
	BuffersBytes   uint64  `json:"buffers_bytes"`
	SwapTotalBytes uint64  `json:"swap_total_bytes"`
	SwapUsedBytes  uint64  `json:"swap_used_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
}

// DiskUsage contains dynamic disk usage metrics.
type DiskUsage struct {
	Disks []DiskUsageEntry `json:"disks"`
}

// DiskUsageEntry represents usage for a single disk/mount.
type DiskUsageEntry struct {
	Device       string  `json:"device"`
	MountPoint   string  `json:"mount_point"`
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	FreeBytes    uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	InodesTotal  uint64  `json:"inodes_total"`
	InodesUsed   uint64  `json:"inodes_used"`
}

// NICUsage contains dynamic network interface usage metrics.
type NICUsage struct {
	Interfaces []NICUsageEntry `json:"interfaces"`
}

// NICUsageEntry represents usage for a single network interface.
type NICUsageEntry struct {
	Name          string  `json:"name"`
	RxBytes       uint64  `json:"rx_bytes"`
	TxBytes       uint64  `json:"tx_bytes"`
	RxPackets     uint64  `json:"rx_packets"`
	TxPackets     uint64  `json:"tx_packets"`
	RxErrors      uint64  `json:"rx_errors"`
	TxErrors      uint64  `json:"tx_errors"`
	RxDropped     uint64  `json:"rx_dropped"`
	TxDropped     uint64  `json:"tx_dropped"`
	RxBytesPerSec float64 `json:"rx_bytes_per_sec"`
	TxBytesPerSec float64 `json:"tx_bytes_per_sec"`
}

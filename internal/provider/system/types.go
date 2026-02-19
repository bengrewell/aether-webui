package system

// Static info types

type CPUInfo struct {
	Model         string   `json:"model" example:"Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz" doc:"CPU model name"`
	PhysicalCores int      `json:"physical_cores" example:"8" doc:"Number of physical CPU cores"`
	LogicalCores  int      `json:"logical_cores" example:"16" doc:"Number of logical CPU cores (includes hyperthreading)"`
	FrequencyMHz  float64  `json:"frequency_mhz" example:"3800" doc:"Base clock frequency in MHz"`
	CacheSizeKB   int32    `json:"cache_size_kb" example:"16384" doc:"CPU cache size in kilobytes"`
	Flags         []string `json:"flags" doc:"CPU feature flags"`
}

type CPUInfoOutput struct {
	Body CPUInfo
}

type MemoryInfo struct {
	TotalBytes     uint64  `json:"total_bytes" example:"34359738368" doc:"Total physical memory in bytes"`
	AvailableBytes uint64  `json:"available_bytes" example:"17179869184" doc:"Available physical memory in bytes"`
	UsedBytes      uint64  `json:"used_bytes" example:"17179869184" doc:"Used physical memory in bytes"`
	UsagePercent   float64 `json:"usage_percent" example:"50.0" doc:"Physical memory usage as a percentage"`
	SwapTotalBytes uint64  `json:"swap_total_bytes" example:"8589934592" doc:"Total swap space in bytes"`
	SwapUsedBytes  uint64  `json:"swap_used_bytes" example:"0" doc:"Used swap space in bytes"`
	SwapPercent    float64 `json:"swap_percent" example:"0.0" doc:"Swap usage as a percentage"`
}

type MemoryInfoOutput struct {
	Body MemoryInfo
}

type Partition struct {
	Device       string  `json:"device" example:"/dev/sda1" doc:"Device path"`
	Mountpoint   string  `json:"mountpoint" example:"/" doc:"Filesystem mount point"`
	FSType       string  `json:"fs_type" example:"ext4" doc:"Filesystem type"`
	TotalBytes   uint64  `json:"total_bytes" example:"512110190592" doc:"Total partition size in bytes"`
	UsedBytes    uint64  `json:"used_bytes" example:"128027547648" doc:"Used space in bytes"`
	FreeBytes    uint64  `json:"free_bytes" example:"384082642944" doc:"Free space in bytes"`
	UsagePercent float64 `json:"usage_percent" example:"25.0" doc:"Disk usage as a percentage"`
}

type DiskInfo struct {
	Partitions []Partition `json:"partitions"`
}

type DiskInfoOutput struct {
	Body DiskInfo
}

type OSInfo struct {
	Hostname        string `json:"hostname" example:"aether-node-01" doc:"System hostname"`
	OS              string `json:"os" example:"linux" doc:"Operating system name"`
	Platform        string `json:"platform" example:"ubuntu" doc:"OS distribution or platform"`
	PlatformVersion string `json:"platform_version" example:"22.04" doc:"Platform version"`
	KernelVersion   string `json:"kernel_version" example:"6.8.0-100-generic" doc:"Kernel version string"`
	KernelArch      string `json:"kernel_arch" example:"x86_64" doc:"Kernel architecture"`
	UptimeSeconds   uint64 `json:"uptime_seconds" example:"86400" doc:"System uptime in seconds"`
}

type OSInfoOutput struct {
	Body OSInfo
}

type NetworkInterface struct {
	Name      string   `json:"name" example:"eth0" doc:"Interface name"`
	MAC       string   `json:"mac" example:"00:1a:2b:3c:4d:5e" doc:"Hardware MAC address"`
	MTU       int      `json:"mtu" example:"1500" doc:"Maximum transmission unit"`
	Flags     []string `json:"flags" doc:"Interface flags (e.g. up, broadcast, multicast)"`
	Addresses []string `json:"addresses" doc:"Assigned IP addresses with CIDR prefix"`
}

type NetworkInterfacesOutput struct {
	Body []NetworkInterface
}

type NetworkConfig struct {
	DNSServers    []string `json:"dns_servers" doc:"Configured DNS nameservers"`
	SearchDomains []string `json:"search_domains" doc:"DNS search domains"`
}

type NetworkConfigOutput struct {
	Body NetworkConfig
}

type ListeningPort struct {
	Protocol    string `json:"protocol" example:"tcp" doc:"Network protocol (tcp or udp)"`
	LocalAddr   string `json:"local_addr" example:"0.0.0.0" doc:"Local bind address"`
	LocalPort   uint32 `json:"local_port" example:"8186" doc:"Local port number"`
	PID         int32  `json:"pid" example:"1234" doc:"Process ID of the listener"`
	ProcessName string `json:"process_name" example:"aether-webd" doc:"Name of the listening process"`
	State       string `json:"state" example:"LISTEN" doc:"Connection state"`
}

type ListeningPortsOutput struct {
	Body []ListeningPort
}

// Metrics query types

type MetricsQueryInput struct {
	Metric      string `query:"metric" example:"system.cpu.usage_percent" doc:"Metric name to query" required:"true"`
	From        string `query:"from" example:"2026-02-18T21:00:00Z" doc:"Start time (RFC 3339)"`
	To          string `query:"to" example:"2026-02-18T21:30:00Z" doc:"End time (RFC 3339)"`
	Labels      string `query:"labels" example:"cpu=total" doc:"Comma-separated key=val label filters"`
	Aggregation string `query:"aggregation" example:"1m" enum:"raw,10s,1m,5m,1h" doc:"Time bucket aggregation"`
}

type PointResult struct {
	Timestamp string  `json:"timestamp" example:"2026-02-18T21:00:00Z" doc:"Sample timestamp (RFC 3339)"`
	Value     float64 `json:"value" example:"23.5" doc:"Metric value at this timestamp"`
}

type SeriesResult struct {
	Metric string            `json:"metric" example:"system.cpu.usage_percent" doc:"Metric name"`
	Labels map[string]string `json:"labels" doc:"Label key-value pairs identifying the series"`
	Points []PointResult     `json:"points" doc:"Time-ordered data points"`
}

type MetricsResult struct {
	Series []SeriesResult `json:"series"`
}

type MetricsQueryOutput struct {
	Body MetricsResult
}

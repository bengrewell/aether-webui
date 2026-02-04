package host

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bengrewell/aether-webui/internal/operator"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// HostOperator handles host/system information and metrics.
type HostOperator interface {
	operator.Operator

	// Static info
	GetCPUInfo(ctx context.Context) (*CPUInfo, error)
	GetMemoryInfo(ctx context.Context) (*MemoryInfo, error)
	GetDiskInfo(ctx context.Context) (*DiskInfo, error)
	GetNICInfo(ctx context.Context) (*NICInfo, error)
	GetOSInfo(ctx context.Context) (*OSInfo, error)

	// Dynamic metrics
	GetCPUUsage(ctx context.Context) (*CPUUsage, error)
	GetMemoryUsage(ctx context.Context) (*MemoryUsage, error)
	GetDiskUsage(ctx context.Context) (*DiskUsage, error)
	GetNICUsage(ctx context.Context) (*NICUsage, error)
}

const cacheTTL = 5 * time.Minute

type cachedItem[T any] struct {
	value     *T
	expiresAt time.Time
}

func (c *cachedItem[T]) isValid() bool {
	return c.value != nil && time.Now().Before(c.expiresAt)
}

// Operator implements HostOperator using gopsutil.
type Operator struct {
	mu sync.RWMutex

	// Static info caches (5 minute TTL)
	cpuInfoCache    cachedItem[CPUInfo]
	memInfoCache    cachedItem[MemoryInfo]
	diskInfoCache   cachedItem[DiskInfo]
	nicInfoCache    cachedItem[NICInfo]
	osInfoCache     cachedItem[OSInfo]

	// For NIC rate calculation
	lastNICStats map[string]nicSnapshot
	lastNICTime  time.Time
}

type nicSnapshot struct {
	rxBytes   uint64
	txBytes   uint64
	rxPackets uint64
	txPackets uint64
}

// New creates a new host operator.
func New() *Operator {
	return &Operator{
		lastNICStats: make(map[string]nicSnapshot),
	}
}

// Domain returns the operator's domain.
func (o *Operator) Domain() operator.Domain {
	return operator.DomainHost
}

// Health returns the operator's health status.
func (o *Operator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{
		Status:  "healthy",
		Message: "gopsutil operational",
	}, nil
}

// GetCPUInfo returns static CPU information.
func (o *Operator) GetCPUInfo(ctx context.Context) (*CPUInfo, error) {
	o.mu.RLock()
	if o.cpuInfoCache.isValid() {
		defer o.mu.RUnlock()
		return o.cpuInfoCache.value, nil
	}
	o.mu.RUnlock()

	o.mu.Lock()
	defer o.mu.Unlock()

	// Double-check after acquiring write lock
	if o.cpuInfoCache.isValid() {
		return o.cpuInfoCache.value, nil
	}

	infos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	info := &CPUInfo{}
	if len(infos) > 0 {
		first := infos[0]
		info.Model = first.ModelName
		info.Vendor = first.VendorID
		info.FrequencyMHz = first.Mhz
		info.CacheKB = int(first.CacheSize)

		// Count physical cores and threads
		coreMap := make(map[string]bool)
		for _, i := range infos {
			coreMap[i.CoreID] = true
		}
		info.Cores = len(coreMap)
		if info.Cores == 0 {
			info.Cores = len(infos)
		}
		info.Threads = len(infos)
	}

	o.cpuInfoCache = cachedItem[CPUInfo]{
		value:     info,
		expiresAt: time.Now().Add(cacheTTL),
	}
	return info, nil
}

// GetMemoryInfo returns static memory information.
func (o *Operator) GetMemoryInfo(ctx context.Context) (*MemoryInfo, error) {
	o.mu.RLock()
	if o.memInfoCache.isValid() {
		defer o.mu.RUnlock()
		return o.memInfoCache.value, nil
	}
	o.mu.RUnlock()

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.memInfoCache.isValid() {
		return o.memInfoCache.value, nil
	}

	vmem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	info := &MemoryInfo{
		TotalBytes: vmem.Total,
		// Type and speed not available via gopsutil
		Type:      "unknown",
		SpeedMHz:  0,
		Slots:     0,
		SlotsUsed: 0,
	}

	o.memInfoCache = cachedItem[MemoryInfo]{
		value:     info,
		expiresAt: time.Now().Add(cacheTTL),
	}
	return info, nil
}

// GetDiskInfo returns static disk information.
func (o *Operator) GetDiskInfo(ctx context.Context) (*DiskInfo, error) {
	o.mu.RLock()
	if o.diskInfoCache.isValid() {
		defer o.mu.RUnlock()
		return o.diskInfoCache.value, nil
	}
	o.mu.RUnlock()

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.diskInfoCache.isValid() {
		return o.diskInfoCache.value, nil
	}

	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	var disks []Disk
	seen := make(map[string]bool)

	for _, p := range partitions {
		// Skip duplicate devices
		if seen[p.Device] {
			continue
		}
		seen[p.Device] = true

		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			continue
		}

		diskType := detectDiskType(p.Device, p.Fstype)

		disks = append(disks, Disk{
			Device:     p.Device,
			Model:      "", // Not available via gopsutil partitions
			SizeBytes:  usage.Total,
			Type:       diskType,
			MountPoint: p.Mountpoint,
		})
	}

	info := &DiskInfo{Disks: disks}
	o.diskInfoCache = cachedItem[DiskInfo]{
		value:     info,
		expiresAt: time.Now().Add(cacheTTL),
	}
	return info, nil
}

func detectDiskType(device, fstype string) string {
	deviceLower := strings.ToLower(device)
	if strings.Contains(deviceLower, "nvme") {
		return "nvme"
	}
	if strings.Contains(deviceLower, "sd") {
		// Could be SSD or HDD - default to unknown without more info
		return "ssd"
	}
	if strings.Contains(deviceLower, "hd") {
		return "hdd"
	}
	if strings.Contains(deviceLower, "vd") || strings.Contains(deviceLower, "xvd") {
		return "virtual"
	}
	return "unknown"
}

// GetNICInfo returns static network interface information.
func (o *Operator) GetNICInfo(ctx context.Context) (*NICInfo, error) {
	o.mu.RLock()
	if o.nicInfoCache.isValid() {
		defer o.mu.RUnlock()
		return o.nicInfoCache.value, nil
	}
	o.mu.RUnlock()

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.nicInfoCache.isValid() {
		return o.nicInfoCache.value, nil
	}

	ifaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var interfaces []NetworkInterface
	for _, iface := range ifaces {
		// Skip loopback
		if strings.HasPrefix(iface.Name, "lo") {
			continue
		}

		var ips []string
		for _, addr := range iface.Addrs {
			ips = append(ips, addr.Addr)
		}

		interfaces = append(interfaces, NetworkInterface{
			Name:        iface.Name,
			MACAddress:  iface.HardwareAddr,
			Driver:      "", // Not available via gopsutil
			SpeedMbps:   0,  // Not available via gopsutil
			MTU:         iface.MTU,
			IPAddresses: ips,
		})
	}

	info := &NICInfo{Interfaces: interfaces}
	o.nicInfoCache = cachedItem[NICInfo]{
		value:     info,
		expiresAt: time.Now().Add(cacheTTL),
	}
	return info, nil
}

// GetOSInfo returns static operating system information.
func (o *Operator) GetOSInfo(ctx context.Context) (*OSInfo, error) {
	o.mu.RLock()
	if o.osInfoCache.isValid() {
		defer o.mu.RUnlock()
		return o.osInfoCache.value, nil
	}
	o.mu.RUnlock()

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.osInfoCache.isValid() {
		return o.osInfoCache.value, nil
	}

	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	info := &OSInfo{
		Name:         hostInfo.Platform,
		Version:      hostInfo.PlatformVersion,
		Kernel:       hostInfo.KernelVersion,
		Architecture: runtime.GOARCH,
		Hostname:     hostInfo.Hostname,
		Uptime:       int64(hostInfo.Uptime),
	}

	o.osInfoCache = cachedItem[OSInfo]{
		value:     info,
		expiresAt: time.Now().Add(cacheTTL),
	}
	return info, nil
}

// GetCPUUsage returns current CPU usage metrics.
func (o *Operator) GetCPUUsage(ctx context.Context) (*CPUUsage, error) {
	// Get per-CPU percentages (100ms sample)
	perCPU, err := cpu.PercentWithContext(ctx, 100*time.Millisecond, true)
	if err != nil {
		return nil, err
	}

	// Get overall percentage
	overall, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, err
	}

	// Get CPU times for user/system/idle breakdown
	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	// Get load averages
	loadAvg, err := load.AvgWithContext(ctx)
	if err != nil {
		// Load average may not be available on all platforms
		loadAvg = &load.AvgStat{}
	}

	usage := &CPUUsage{
		PerCoreUsage:  perCPU,
		LoadAverage1:  loadAvg.Load1,
		LoadAverage5:  loadAvg.Load5,
		LoadAverage15: loadAvg.Load15,
	}

	if len(overall) > 0 {
		usage.UsagePercent = overall[0]
	}

	if len(times) > 0 {
		t := times[0]
		total := t.User + t.System + t.Idle + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal
		if total > 0 {
			usage.UserPercent = (t.User + t.Nice) / total * 100
			usage.SystemPercent = t.System / total * 100
			usage.IdlePercent = t.Idle / total * 100
			usage.IOWaitPercent = t.Iowait / total * 100
		}
	}

	return usage, nil
}

// GetMemoryUsage returns current memory usage metrics.
func (o *Operator) GetMemoryUsage(ctx context.Context) (*MemoryUsage, error) {
	vmem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	swap, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		// Swap may not be available
		swap = &mem.SwapMemoryStat{}
	}

	return &MemoryUsage{
		UsedBytes:      vmem.Used,
		FreeBytes:      vmem.Free,
		AvailableBytes: vmem.Available,
		CachedBytes:    vmem.Cached,
		BuffersBytes:   vmem.Buffers,
		SwapTotalBytes: swap.Total,
		SwapUsedBytes:  swap.Used,
		UsagePercent:   vmem.UsedPercent,
	}, nil
}

// GetDiskUsage returns current disk usage metrics.
func (o *Operator) GetDiskUsage(ctx context.Context) (*DiskUsage, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	var entries []DiskUsageEntry
	seen := make(map[string]bool)

	for _, p := range partitions {
		// Skip duplicate mountpoints
		if seen[p.Mountpoint] {
			continue
		}
		seen[p.Mountpoint] = true

		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			continue
		}

		entries = append(entries, DiskUsageEntry{
			Device:       p.Device,
			MountPoint:   p.Mountpoint,
			TotalBytes:   usage.Total,
			UsedBytes:    usage.Used,
			FreeBytes:    usage.Free,
			UsagePercent: usage.UsedPercent,
			InodesTotal:  usage.InodesTotal,
			InodesUsed:   usage.InodesUsed,
		})
	}

	return &DiskUsage{Disks: entries}, nil
}

// GetNICUsage returns current network interface usage metrics.
func (o *Operator) GetNICUsage(ctx context.Context) (*NICUsage, error) {
	counters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	o.mu.Lock()
	defer o.mu.Unlock()

	var entries []NICUsageEntry
	elapsed := now.Sub(o.lastNICTime).Seconds()

	for _, c := range counters {
		// Skip loopback
		if strings.HasPrefix(c.Name, "lo") {
			continue
		}

		entry := NICUsageEntry{
			Name:      c.Name,
			RxBytes:   c.BytesRecv,
			TxBytes:   c.BytesSent,
			RxPackets: c.PacketsRecv,
			TxPackets: c.PacketsSent,
			RxErrors:  c.Errin,
			TxErrors:  c.Errout,
			RxDropped: c.Dropin,
			TxDropped: c.Dropout,
		}

		// Calculate rates if we have previous data
		if last, ok := o.lastNICStats[c.Name]; ok && elapsed > 0 {
			entry.RxBytesPerSec = float64(c.BytesRecv-last.rxBytes) / elapsed
			entry.TxBytesPerSec = float64(c.BytesSent-last.txBytes) / elapsed

			// Handle counter wrap-around
			if entry.RxBytesPerSec < 0 {
				entry.RxBytesPerSec = 0
			}
			if entry.TxBytesPerSec < 0 {
				entry.TxBytesPerSec = 0
			}
		}

		// Store current values for next calculation
		o.lastNICStats[c.Name] = nicSnapshot{
			rxBytes:   c.BytesRecv,
			txBytes:   c.BytesSent,
			rxPackets: c.PacketsRecv,
			txPackets: c.PacketsSent,
		}

		entries = append(entries, entry)
	}

	o.lastNICTime = now

	return &NICUsage{Interfaces: entries}, nil
}

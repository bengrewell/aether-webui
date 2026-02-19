package system

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/bengrewell/aether-webui/internal/store"
)

func (s *System) handleCPU(ctx context.Context, _ *struct{}) (*CPUInfoOutput, error) {
	infos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("cpu info: %w", err)
	}

	physical, _ := cpu.CountsWithContext(ctx, false)
	logical, _ := cpu.CountsWithContext(ctx, true)

	out := CPUInfo{
		PhysicalCores: physical,
		LogicalCores:  logical,
	}
	if len(infos) > 0 {
		out.Model = infos[0].ModelName
		out.FrequencyMHz = infos[0].Mhz
		out.CacheSizeKB = infos[0].CacheSize
		out.Flags = infos[0].Flags
	}
	if out.Flags == nil {
		out.Flags = []string{}
	}
	return &CPUInfoOutput{Body: out}, nil
}

func (s *System) handleMemory(ctx context.Context, _ *struct{}) (*MemoryInfoOutput, error) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("virtual memory: %w", err)
	}

	out := MemoryInfo{
		TotalBytes:     vm.Total,
		AvailableBytes: vm.Available,
		UsedBytes:      vm.Used,
		UsagePercent:   vm.UsedPercent,
	}

	if sw, err := mem.SwapMemoryWithContext(ctx); err == nil {
		out.SwapTotalBytes = sw.Total
		out.SwapUsedBytes = sw.Used
		out.SwapPercent = sw.UsedPercent
	}

	return &MemoryInfoOutput{Body: out}, nil
}

func (s *System) handleDisks(ctx context.Context, _ *struct{}) (*DiskInfoOutput, error) {
	parts, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("disk partitions: %w", err)
	}

	partitions := make([]Partition, 0, len(parts))
	for _, p := range parts {
		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			continue
		}
		partitions = append(partitions, Partition{
			Device:       p.Device,
			Mountpoint:   p.Mountpoint,
			FSType:       p.Fstype,
			TotalBytes:   usage.Total,
			UsedBytes:    usage.Used,
			FreeBytes:    usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}
	return &DiskInfoOutput{Body: DiskInfo{Partitions: partitions}}, nil
}

func (s *System) handleOS(ctx context.Context, _ *struct{}) (*OSInfoOutput, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("host info: %w", err)
	}
	return &OSInfoOutput{Body: OSInfo{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		UptimeSeconds:   info.Uptime,
	}}, nil
}

func (s *System) handleNetworkInterfaces(ctx context.Context, _ *struct{}) (*NetworkInterfacesOutput, error) {
	ifaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("network interfaces: %w", err)
	}

	out := make([]NetworkInterface, len(ifaces))
	for i, iface := range ifaces {
		addrs := make([]string, len(iface.Addrs))
		for j, a := range iface.Addrs {
			addrs[j] = a.Addr
		}
		out[i] = NetworkInterface{
			Name:      iface.Name,
			MAC:       iface.HardwareAddr,
			MTU:       iface.MTU,
			Flags:     iface.Flags,
			Addresses: addrs,
		}
		if out[i].Flags == nil {
			out[i].Flags = []string{}
		}
		if out[i].Addresses == nil {
			out[i].Addresses = []string{}
		}
	}
	return &NetworkInterfacesOutput{Body: out}, nil
}

func (s *System) handleNetworkConfig(_ context.Context, _ *struct{}) (*NetworkConfigOutput, error) {
	servers, search, err := parseResolvConf("/etc/resolv.conf")
	if err != nil {
		// Non-fatal: return empty config rather than erroring.
		s.Base.Log().Warn("failed to parse resolv.conf", "error", err)
		return &NetworkConfigOutput{Body: NetworkConfig{
			DNSServers:    []string{},
			SearchDomains: []string{},
		}}, nil
	}
	if servers == nil {
		servers = []string{}
	}
	if search == nil {
		search = []string{}
	}
	return &NetworkConfigOutput{Body: NetworkConfig{
		DNSServers:    servers,
		SearchDomains: search,
	}}, nil
}

func (s *System) handleListeningPorts(ctx context.Context, _ *struct{}) (*ListeningPortsOutput, error) {
	conns, err := net.ConnectionsWithContext(ctx, "inet")
	if err != nil {
		return nil, fmt.Errorf("network connections: %w", err)
	}

	ports := make([]ListeningPort, 0, len(conns))
	for _, c := range conns {
		if c.Status != "LISTEN" {
			continue
		}
		protocol := "tcp"
		if c.Type == syscall.SOCK_DGRAM {
			protocol = "udp"
		}
		lp := ListeningPort{
			Protocol:  protocol,
			LocalAddr: c.Laddr.IP,
			LocalPort: c.Laddr.Port,
			PID:       c.Pid,
			State:     c.Status,
		}
		if c.Pid > 0 {
			if p, err := process.NewProcessWithContext(ctx, c.Pid); err == nil {
				if name, err := p.NameWithContext(ctx); err == nil {
					lp.ProcessName = name
				}
			}
		}
		ports = append(ports, lp)
	}
	return &ListeningPortsOutput{Body: ports}, nil
}

func (s *System) handleMetricsQuery(ctx context.Context, in *MetricsQueryInput) (*MetricsQueryOutput, error) {
	if in.Metric == "" {
		return &MetricsQueryOutput{Body: MetricsResult{Series: []SeriesResult{}}}, nil
	}

	st := s.Store()
	if st == (store.Client{}) {
		return nil, fmt.Errorf("metrics store not configured")
	}

	now := time.Now()
	from := now.Add(-1 * time.Hour)
	to := now

	if in.From != "" {
		if t, err := time.Parse(time.RFC3339, in.From); err == nil {
			from = t
		} else {
			s.Base.Log().Warn("invalid 'from' timestamp, using default", "from", in.From, "error", err)
		}
	}
	if in.To != "" {
		if t, err := time.Parse(time.RFC3339, in.To); err == nil {
			to = t
		} else {
			s.Base.Log().Warn("invalid 'to' timestamp, using default", "to", in.To, "error", err)
		}
	}

	labels := parseLabels(in.Labels)
	agg := parseAggregation(in.Aggregation)

	q := store.RangeQuery{
		Metric:      in.Metric,
		Range:       store.TimeRange{From: from, To: to},
		LabelsExact: labels,
		Agg:         agg,
	}

	series, err := st.QueryRange(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}

	results := make([]SeriesResult, len(series))
	for i, sr := range series {
		points := make([]PointResult, len(sr.Points))
		for j, p := range sr.Points {
			points[j] = PointResult{
				Timestamp: p.TS.Format(time.RFC3339),
				Value:     p.Value,
			}
		}
		results[i] = SeriesResult{
			Metric: sr.Metric,
			Labels: sr.Labels,
			Points: points,
		}
		if results[i].Labels == nil {
			results[i].Labels = map[string]string{}
		}
	}

	return &MetricsQueryOutput{Body: MetricsResult{Series: results}}, nil
}

// parseLabels parses a comma-separated "key=val,key2=val2" string into a map.
func parseLabels(s string) map[string]string {
	if s == "" {
		return nil
	}
	m := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if ok {
			m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// parseAggregation converts a string aggregation name to the store Agg constant.
func parseAggregation(s string) store.Agg {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "10s":
		return store.Agg10s
	case "1m":
		return store.Agg1m
	case "5m":
		return store.Agg5m
	case "1h":
		return store.Agg1h
	default:
		return store.AggRaw
	}
}

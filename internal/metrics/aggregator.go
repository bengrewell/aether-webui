package metrics

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/bengrewell/aether-webui/internal/operator/host"
	"github.com/bengrewell/aether-webui/internal/state"
)

// Aggregate bins snapshots into time buckets and aggregates values.
// If granularity is 0, returns raw points without aggregation.
func Aggregate(snapshots []state.MetricsSnapshot, metricType MetricType, granularity time.Duration) []AggregatedPoint {
	if len(snapshots) == 0 {
		return nil
	}

	// Sort by time ascending
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].RecordedAt.Before(snapshots[j].RecordedAt)
	})

	// No aggregation - return raw points
	if granularity == 0 {
		return rawPoints(snapshots, metricType)
	}

	// Group snapshots into time buckets
	buckets := groupIntoBuckets(snapshots, granularity)

	var points []AggregatedPoint
	for _, bucket := range buckets {
		if len(bucket.snapshots) == 0 {
			continue
		}

		point := AggregatedPoint{
			Timestamp: bucket.timestamp,
		}

		switch metricType {
		case MetricTypeCPU:
			point.Data = aggregateCPU(bucket.snapshots)
		case MetricTypeMemory:
			point.Data = aggregateMemory(bucket.snapshots)
		case MetricTypeDisk:
			point.Data = aggregateDisk(bucket.snapshots)
		case MetricTypeNIC:
			point.Data = aggregateNIC(bucket.snapshots)
		default:
			// Unknown type, skip
			continue
		}

		points = append(points, point)
	}

	return points
}

type bucket struct {
	timestamp time.Time
	snapshots []state.MetricsSnapshot
}

func groupIntoBuckets(snapshots []state.MetricsSnapshot, granularity time.Duration) []bucket {
	if len(snapshots) == 0 {
		return nil
	}

	// Determine time range
	start := snapshots[0].RecordedAt.Truncate(granularity)
	end := snapshots[len(snapshots)-1].RecordedAt

	var buckets []bucket
	for t := start; !t.After(end); t = t.Add(granularity) {
		buckets = append(buckets, bucket{timestamp: t})
	}

	// Assign snapshots to buckets
	for _, snap := range snapshots {
		bucketIdx := int(snap.RecordedAt.Sub(start) / granularity)
		if bucketIdx >= 0 && bucketIdx < len(buckets) {
			buckets[bucketIdx].snapshots = append(buckets[bucketIdx].snapshots, snap)
		}
	}

	return buckets
}

func rawPoints(snapshots []state.MetricsSnapshot, metricType MetricType) []AggregatedPoint {
	var points []AggregatedPoint
	for _, snap := range snapshots {
		point := AggregatedPoint{
			Timestamp: snap.RecordedAt,
		}

		switch metricType {
		case MetricTypeCPU:
			var data host.CPUUsage
			if err := json.Unmarshal(snap.Data, &data); err == nil {
				point.Data = data
			}
		case MetricTypeMemory:
			var data host.MemoryUsage
			if err := json.Unmarshal(snap.Data, &data); err == nil {
				point.Data = data
			}
		case MetricTypeDisk:
			var data host.DiskUsage
			if err := json.Unmarshal(snap.Data, &data); err == nil {
				point.Data = data
			}
		case MetricTypeNIC:
			var data host.NICUsage
			if err := json.Unmarshal(snap.Data, &data); err == nil {
				point.Data = data
			}
		}

		if point.Data != nil {
			points = append(points, point)
		}
	}
	return points
}

func aggregateCPU(snapshots []state.MetricsSnapshot) *AggregatedCPU {
	var usages []host.CPUUsage
	for _, snap := range snapshots {
		var u host.CPUUsage
		if err := json.Unmarshal(snap.Data, &u); err == nil {
			usages = append(usages, u)
		}
	}

	if len(usages) == 0 {
		return nil
	}

	agg := &AggregatedCPU{SampleCount: len(usages)}

	for _, u := range usages {
		agg.UsagePercent += u.UsagePercent
		agg.UserPercent += u.UserPercent
		agg.SystemPercent += u.SystemPercent
		agg.IdlePercent += u.IdlePercent
		agg.IOWaitPercent += u.IOWaitPercent
		agg.LoadAverage1 += u.LoadAverage1
		agg.LoadAverage5 += u.LoadAverage5
		agg.LoadAverage15 += u.LoadAverage15
	}

	n := float64(len(usages))
	agg.UsagePercent /= n
	agg.UserPercent /= n
	agg.SystemPercent /= n
	agg.IdlePercent /= n
	agg.IOWaitPercent /= n
	agg.LoadAverage1 /= n
	agg.LoadAverage5 /= n
	agg.LoadAverage15 /= n

	// Average per-core usage if present
	if len(usages[0].PerCoreUsage) > 0 {
		agg.PerCoreUsage = make([]float64, len(usages[0].PerCoreUsage))
		for _, u := range usages {
			for i, v := range u.PerCoreUsage {
				if i < len(agg.PerCoreUsage) {
					agg.PerCoreUsage[i] += v
				}
			}
		}
		for i := range agg.PerCoreUsage {
			agg.PerCoreUsage[i] /= n
		}
	}

	return agg
}

func aggregateMemory(snapshots []state.MetricsSnapshot) *AggregatedMemory {
	var usages []host.MemoryUsage
	for _, snap := range snapshots {
		var u host.MemoryUsage
		if err := json.Unmarshal(snap.Data, &u); err == nil {
			usages = append(usages, u)
		}
	}

	if len(usages) == 0 {
		return nil
	}

	agg := &AggregatedMemory{SampleCount: len(usages)}

	for _, u := range usages {
		agg.UsedBytesAvg += u.UsedBytes
		agg.FreeBytesAvg += u.FreeBytes
		agg.AvailableBytesAvg += u.AvailableBytes
		agg.CachedBytesAvg += u.CachedBytes
		agg.UsagePercentAvg += u.UsagePercent
		agg.SwapUsedBytesAvg += u.SwapUsedBytes

		if u.UsedBytes > agg.UsedBytesMax {
			agg.UsedBytesMax = u.UsedBytes
		}
		if u.UsagePercent > agg.UsagePercentMax {
			agg.UsagePercentMax = u.UsagePercent
		}
	}

	n := uint64(len(usages))
	agg.UsedBytesAvg /= n
	agg.FreeBytesAvg /= n
	agg.AvailableBytesAvg /= n
	agg.CachedBytesAvg /= n
	agg.UsagePercentAvg /= float64(n)
	agg.SwapUsedBytesAvg /= n

	return agg
}

func aggregateDisk(snapshots []state.MetricsSnapshot) *AggregatedDisk {
	// For disk, we take the latest snapshot
	if len(snapshots) == 0 {
		return nil
	}

	var latest host.DiskUsage
	if err := json.Unmarshal(snapshots[len(snapshots)-1].Data, &latest); err != nil {
		return nil
	}

	return &AggregatedDisk{
		Disks:       latest.Disks,
		SampleCount: len(snapshots),
	}
}

func aggregateNIC(snapshots []state.MetricsSnapshot) *AggregatedNIC {
	var usages []host.NICUsage
	for _, snap := range snapshots {
		var u host.NICUsage
		if err := json.Unmarshal(snap.Data, &u); err == nil {
			usages = append(usages, u)
		}
	}

	if len(usages) == 0 {
		return nil
	}

	// Aggregate per interface
	ifaceStats := make(map[string]*AggregatedNICEntry)
	ifaceCounts := make(map[string]int)

	for _, u := range usages {
		for _, iface := range u.Interfaces {
			if _, ok := ifaceStats[iface.Name]; !ok {
				ifaceStats[iface.Name] = &AggregatedNICEntry{Name: iface.Name}
			}
			stats := ifaceStats[iface.Name]
			stats.RxBytesPerSecAvg += iface.RxBytesPerSec
			stats.TxBytesPerSecAvg += iface.TxBytesPerSec
			ifaceCounts[iface.Name]++
		}
	}

	// Calculate averages and get totals from first and last
	if len(usages) >= 2 {
		first := usages[0]
		last := usages[len(usages)-1]

		firstMap := make(map[string]host.NICUsageEntry)
		for _, iface := range first.Interfaces {
			firstMap[iface.Name] = iface
		}

		for _, iface := range last.Interfaces {
			if stats, ok := ifaceStats[iface.Name]; ok {
				if firstEntry, found := firstMap[iface.Name]; found {
					// Calculate total bytes transferred in this window
					stats.RxBytesTotal = iface.RxBytes - firstEntry.RxBytes
					stats.TxBytesTotal = iface.TxBytes - firstEntry.TxBytes
					stats.RxPacketsTotal = iface.RxPackets - firstEntry.RxPackets
					stats.TxPacketsTotal = iface.TxPackets - firstEntry.TxPackets
				}
			}
		}
	}

	// Finalize averages
	var interfaces []AggregatedNICEntry
	for name, stats := range ifaceStats {
		if count := ifaceCounts[name]; count > 0 {
			stats.RxBytesPerSecAvg /= float64(count)
			stats.TxBytesPerSecAvg /= float64(count)
		}
		interfaces = append(interfaces, *stats)
	}

	// Sort by name for consistent output
	sort.Slice(interfaces, func(i, j int) bool {
		return interfaces[i].Name < interfaces[j].Name
	})

	return &AggregatedNIC{
		Interfaces:  interfaces,
		SampleCount: len(usages),
	}
}

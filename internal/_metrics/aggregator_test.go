package metrics

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/operator/host"
	"github.com/bengrewell/aether-webui/internal/state"
)

func makeSnapshot(metricType string, data interface{}, recordedAt time.Time) state.MetricsSnapshot {
	b, _ := json.Marshal(data)
	return state.MetricsSnapshot{
		MetricType: metricType,
		Data:       b,
		RecordedAt: recordedAt,
	}
}

func TestAggregateEmpty(t *testing.T) {
	points := Aggregate(nil, MetricTypeCPU, time.Minute)
	if len(points) != 0 {
		t.Errorf("expected empty result, got %d points", len(points))
	}
}

func TestAggregateRawPointsCPU(t *testing.T) {
	now := time.Now()
	snapshots := []state.MetricsSnapshot{
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 40.0}, now.Add(-2*time.Minute)),
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 60.0}, now.Add(-time.Minute)),
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 50.0}, now),
	}

	// No aggregation (granularity = 0)
	points := Aggregate(snapshots, MetricTypeCPU, 0)
	if len(points) != 3 {
		t.Fatalf("expected 3 raw points, got %d", len(points))
	}

	// Verify data is preserved
	cpu, ok := points[0].Data.(host.CPUUsage)
	if !ok {
		t.Fatalf("expected host.CPUUsage, got %T", points[0].Data)
	}
	if cpu.UsagePercent != 40.0 {
		t.Errorf("expected 40.0, got %f", cpu.UsagePercent)
	}
}

func TestAggregateCPUWithGranularity(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	snapshots := []state.MetricsSnapshot{
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 40.0, LoadAverage1: 1.0}, base),
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 60.0, LoadAverage1: 2.0}, base.Add(10*time.Second)),
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 50.0, LoadAverage1: 1.5}, base.Add(20*time.Second)),
		// Next bucket
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 80.0, LoadAverage1: 3.0}, base.Add(time.Minute)),
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 70.0, LoadAverage1: 2.5}, base.Add(70*time.Second)),
	}

	// 1 minute granularity
	points := Aggregate(snapshots, MetricTypeCPU, time.Minute)
	if len(points) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(points))
	}

	// First bucket: average of 40, 60, 50 = 50
	agg1, ok := points[0].Data.(*AggregatedCPU)
	if !ok {
		t.Fatalf("expected *AggregatedCPU, got %T", points[0].Data)
	}
	if agg1.UsagePercent != 50.0 {
		t.Errorf("first bucket avg: expected 50.0, got %f", agg1.UsagePercent)
	}
	if agg1.SampleCount != 3 {
		t.Errorf("first bucket samples: expected 3, got %d", agg1.SampleCount)
	}

	// Second bucket: average of 80, 70 = 75
	agg2, ok := points[1].Data.(*AggregatedCPU)
	if !ok {
		t.Fatalf("expected *AggregatedCPU, got %T", points[1].Data)
	}
	if agg2.UsagePercent != 75.0 {
		t.Errorf("second bucket avg: expected 75.0, got %f", agg2.UsagePercent)
	}
	if agg2.SampleCount != 2 {
		t.Errorf("second bucket samples: expected 2, got %d", agg2.SampleCount)
	}
}

func TestAggregateMemory(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	snapshots := []state.MetricsSnapshot{
		makeSnapshot("memory", host.MemoryUsage{UsedBytes: 1000, UsagePercent: 50.0}, base),
		makeSnapshot("memory", host.MemoryUsage{UsedBytes: 2000, UsagePercent: 60.0}, base.Add(10*time.Second)),
		makeSnapshot("memory", host.MemoryUsage{UsedBytes: 1500, UsagePercent: 55.0}, base.Add(20*time.Second)),
	}

	points := Aggregate(snapshots, MetricTypeMemory, time.Minute)
	if len(points) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(points))
	}

	agg, ok := points[0].Data.(*AggregatedMemory)
	if !ok {
		t.Fatalf("expected *AggregatedMemory, got %T", points[0].Data)
	}

	// Average: (1000+2000+1500)/3 = 1500
	if agg.UsedBytesAvg != 1500 {
		t.Errorf("UsedBytesAvg: expected 1500, got %d", agg.UsedBytesAvg)
	}
	// Max: 2000
	if agg.UsedBytesMax != 2000 {
		t.Errorf("UsedBytesMax: expected 2000, got %d", agg.UsedBytesMax)
	}
	// Usage percent avg: (50+60+55)/3 = 55
	if agg.UsagePercentAvg != 55.0 {
		t.Errorf("UsagePercentAvg: expected 55.0, got %f", agg.UsagePercentAvg)
	}
	// Usage percent max: 60
	if agg.UsagePercentMax != 60.0 {
		t.Errorf("UsagePercentMax: expected 60.0, got %f", agg.UsagePercentMax)
	}
}

func TestAggregateDisk(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	snapshots := []state.MetricsSnapshot{
		makeSnapshot("disk", host.DiskUsage{Disks: []host.DiskUsageEntry{{MountPoint: "/", UsagePercent: 30.0}}}, base),
		makeSnapshot("disk", host.DiskUsage{Disks: []host.DiskUsageEntry{{MountPoint: "/", UsagePercent: 40.0}}}, base.Add(10*time.Second)),
		makeSnapshot("disk", host.DiskUsage{Disks: []host.DiskUsageEntry{{MountPoint: "/", UsagePercent: 50.0}}}, base.Add(20*time.Second)),
	}

	points := Aggregate(snapshots, MetricTypeDisk, time.Minute)
	if len(points) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(points))
	}

	agg, ok := points[0].Data.(*AggregatedDisk)
	if !ok {
		t.Fatalf("expected *AggregatedDisk, got %T", points[0].Data)
	}

	// Disk should take latest value (50%)
	if len(agg.Disks) != 1 {
		t.Fatalf("expected 1 disk, got %d", len(agg.Disks))
	}
	if agg.Disks[0].UsagePercent != 50.0 {
		t.Errorf("expected latest value 50.0, got %f", agg.Disks[0].UsagePercent)
	}
	if agg.SampleCount != 3 {
		t.Errorf("expected 3 samples, got %d", agg.SampleCount)
	}
}

func TestAggregateNIC(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	snapshots := []state.MetricsSnapshot{
		makeSnapshot("nic", host.NICUsage{Interfaces: []host.NICUsageEntry{
			{Name: "eth0", RxBytes: 1000, TxBytes: 500, RxBytesPerSec: 100, TxBytesPerSec: 50},
		}}, base),
		makeSnapshot("nic", host.NICUsage{Interfaces: []host.NICUsageEntry{
			{Name: "eth0", RxBytes: 2000, TxBytes: 1000, RxBytesPerSec: 200, TxBytesPerSec: 100},
		}}, base.Add(10*time.Second)),
		makeSnapshot("nic", host.NICUsage{Interfaces: []host.NICUsageEntry{
			{Name: "eth0", RxBytes: 3000, TxBytes: 1500, RxBytesPerSec: 150, TxBytesPerSec: 75},
		}}, base.Add(20*time.Second)),
	}

	points := Aggregate(snapshots, MetricTypeNIC, time.Minute)
	if len(points) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(points))
	}

	agg, ok := points[0].Data.(*AggregatedNIC)
	if !ok {
		t.Fatalf("expected *AggregatedNIC, got %T", points[0].Data)
	}

	if len(agg.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(agg.Interfaces))
	}

	iface := agg.Interfaces[0]
	if iface.Name != "eth0" {
		t.Errorf("expected eth0, got %s", iface.Name)
	}

	// Total bytes transferred: 3000-1000=2000 rx, 1500-500=1000 tx
	if iface.RxBytesTotal != 2000 {
		t.Errorf("RxBytesTotal: expected 2000, got %d", iface.RxBytesTotal)
	}
	if iface.TxBytesTotal != 1000 {
		t.Errorf("TxBytesTotal: expected 1000, got %d", iface.TxBytesTotal)
	}

	// Average rate: (100+200+150)/3 = 150 rx, (50+100+75)/3 = 75 tx
	if iface.RxBytesPerSecAvg != 150.0 {
		t.Errorf("RxBytesPerSecAvg: expected 150.0, got %f", iface.RxBytesPerSecAvg)
	}
	if iface.TxBytesPerSecAvg != 75.0 {
		t.Errorf("TxBytesPerSecAvg: expected 75.0, got %f", iface.TxBytesPerSecAvg)
	}
}

func TestAggregatePreservesOrder(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create snapshots out of order
	snapshots := []state.MetricsSnapshot{
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 50.0}, base.Add(2*time.Minute)),
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 30.0}, base),
		makeSnapshot("cpu", host.CPUUsage{UsagePercent: 40.0}, base.Add(time.Minute)),
	}

	points := Aggregate(snapshots, MetricTypeCPU, 0) // Raw points

	// Should be sorted ascending by time
	if len(points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(points))
	}

	expected := []float64{30.0, 40.0, 50.0}
	for i, p := range points {
		cpu := p.Data.(host.CPUUsage)
		if cpu.UsagePercent != expected[i] {
			t.Errorf("point %d: expected %f, got %f", i, expected[i], cpu.UsagePercent)
		}
	}
}

func TestAggregateCPUPerCore(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	snapshots := []state.MetricsSnapshot{
		makeSnapshot("cpu", host.CPUUsage{
			UsagePercent: 50.0,
			PerCoreUsage: []float64{40.0, 60.0},
		}, base),
		makeSnapshot("cpu", host.CPUUsage{
			UsagePercent: 70.0,
			PerCoreUsage: []float64{60.0, 80.0},
		}, base.Add(10*time.Second)),
	}

	points := Aggregate(snapshots, MetricTypeCPU, time.Minute)
	if len(points) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(points))
	}

	agg := points[0].Data.(*AggregatedCPU)

	// Per-core averages: [(40+60)/2, (60+80)/2] = [50, 70]
	if len(agg.PerCoreUsage) != 2 {
		t.Fatalf("expected 2 per-core values, got %d", len(agg.PerCoreUsage))
	}
	if agg.PerCoreUsage[0] != 50.0 {
		t.Errorf("core 0: expected 50.0, got %f", agg.PerCoreUsage[0])
	}
	if agg.PerCoreUsage[1] != 70.0 {
		t.Errorf("core 1: expected 70.0, got %f", agg.PerCoreUsage[1])
	}
}

func TestAggregateMultipleNICs(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	snapshots := []state.MetricsSnapshot{
		makeSnapshot("nic", host.NICUsage{Interfaces: []host.NICUsageEntry{
			{Name: "eth0", RxBytesPerSec: 100},
			{Name: "eth1", RxBytesPerSec: 200},
		}}, base),
		makeSnapshot("nic", host.NICUsage{Interfaces: []host.NICUsageEntry{
			{Name: "eth0", RxBytesPerSec: 150},
			{Name: "eth1", RxBytesPerSec: 250},
		}}, base.Add(10*time.Second)),
	}

	points := Aggregate(snapshots, MetricTypeNIC, time.Minute)
	if len(points) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(points))
	}

	agg := points[0].Data.(*AggregatedNIC)
	if len(agg.Interfaces) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(agg.Interfaces))
	}

	// Interfaces should be sorted by name
	if agg.Interfaces[0].Name != "eth0" {
		t.Errorf("expected first interface to be eth0, got %s", agg.Interfaces[0].Name)
	}
	if agg.Interfaces[1].Name != "eth1" {
		t.Errorf("expected second interface to be eth1, got %s", agg.Interfaces[1].Name)
	}

	// eth0: (100+150)/2 = 125
	if agg.Interfaces[0].RxBytesPerSecAvg != 125.0 {
		t.Errorf("eth0 RxBytesPerSecAvg: expected 125.0, got %f", agg.Interfaces[0].RxBytesPerSecAvg)
	}
	// eth1: (200+250)/2 = 225
	if agg.Interfaces[1].RxBytesPerSecAvg != 225.0 {
		t.Errorf("eth1 RxBytesPerSecAvg: expected 225.0, got %f", agg.Interfaces[1].RxBytesPerSecAvg)
	}
}

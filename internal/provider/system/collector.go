package system

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"github.com/bengrewell/aether-webui/internal/store"
)

// Start overrides Base.Start to spawn the background metrics collector.
func (s *System) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.done = make(chan struct{})

	go s.run(ctx)
	s.Base.SetRunning(true)
	return nil
}

// Stop overrides Base.Stop to cancel the collector and wait for it to exit.
// Safe to call multiple times; only the first call performs cleanup.
func (s *System) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
			<-s.done
		}
		s.Base.SetRunning(false)
	})
	return nil
}

func (s *System) run(ctx context.Context) {
	defer close(s.done)

	ticker := time.NewTicker(s.config.CollectInterval)
	defer ticker.Stop()

	// Collect immediately on start. Note: the first cpu.Percent(0, ...) call
	// may return inaccurate or zero values because there is no prior measurement
	// to compare against. Subsequent ticks will produce accurate readings.
	s.collect(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collect(ctx)
		}
	}
}

func (s *System) collect(ctx context.Context) {
	st := s.Store()
	if st == (store.Client{}) {
		s.Base.Log().Warn("skipping metrics collection: store not configured")
		return
	}

	now := time.Now()
	var samples []store.Sample

	// CPU usage (total + per-core)
	if totals, err := cpu.PercentWithContext(ctx, 0, false); err == nil && len(totals) > 0 {
		samples = append(samples, store.Sample{
			Metric: "system.cpu.usage_percent",
			TS:     now,
			Value:  totals[0],
			Labels: map[string]string{"cpu": "total"},
			Unit:   "percent",
		})
	}
	if perCPU, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
		for i, v := range perCPU {
			samples = append(samples, store.Sample{
				Metric: "system.cpu.usage_percent",
				TS:     now,
				Value:  v,
				Labels: map[string]string{"cpu": itoa(i)},
				Unit:   "percent",
			})
		}
	}

	// Memory
	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		samples = append(samples,
			store.Sample{Metric: "system.memory.used_bytes", TS: now, Value: float64(vm.Used), Unit: "bytes"},
			store.Sample{Metric: "system.memory.available_bytes", TS: now, Value: float64(vm.Available), Unit: "bytes"},
			store.Sample{Metric: "system.memory.usage_percent", TS: now, Value: vm.UsedPercent, Unit: "percent"},
		)
	}

	// Swap
	if sw, err := mem.SwapMemoryWithContext(ctx); err == nil {
		samples = append(samples,
			store.Sample{Metric: "system.swap.used_bytes", TS: now, Value: float64(sw.Used), Unit: "bytes"},
		)
	}

	// Disk usage per partition
	if parts, err := disk.PartitionsWithContext(ctx, false); err == nil {
		for _, p := range parts {
			usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
			if err != nil {
				continue
			}
			labels := map[string]string{"device": p.Device, "mount": p.Mountpoint}
			samples = append(samples,
				store.Sample{Metric: "system.disk.used_bytes", TS: now, Value: float64(usage.Used), Labels: labels, Unit: "bytes"},
				store.Sample{Metric: "system.disk.usage_percent", TS: now, Value: usage.UsedPercent, Labels: labels, Unit: "percent"},
			)
		}
	}

	// Disk I/O
	if counters, err := disk.IOCountersWithContext(ctx); err == nil {
		for name, c := range counters {
			labels := map[string]string{"device": name}
			samples = append(samples,
				store.Sample{Metric: "system.disk.read_bytes", TS: now, Value: float64(c.ReadBytes), Labels: labels, Unit: "bytes"},
				store.Sample{Metric: "system.disk.write_bytes", TS: now, Value: float64(c.WriteBytes), Labels: labels, Unit: "bytes"},
			)
		}
	}

	// Network I/O per interface
	if counters, err := net.IOCountersWithContext(ctx, true); err == nil {
		for _, c := range counters {
			labels := map[string]string{"interface": c.Name}
			samples = append(samples,
				store.Sample{Metric: "system.net.bytes_sent", TS: now, Value: float64(c.BytesSent), Labels: labels, Unit: "bytes"},
				store.Sample{Metric: "system.net.bytes_recv", TS: now, Value: float64(c.BytesRecv), Labels: labels, Unit: "bytes"},
			)
		}
	}

	// Load averages
	if avg, err := load.AvgWithContext(ctx); err == nil {
		samples = append(samples,
			store.Sample{Metric: "system.load.1m", TS: now, Value: avg.Load1},
			store.Sample{Metric: "system.load.5m", TS: now, Value: avg.Load5},
			store.Sample{Metric: "system.load.15m", TS: now, Value: avg.Load15},
		)
	}

	if len(samples) > 0 {
		if err := st.AppendSamples(ctx, samples); err != nil {
			s.Base.Log().Error("failed to store metrics samples", "error", err, "count", len(samples))
		}
	}
}

func itoa(i int) string {
	// Avoid importing strconv for a simple int-to-string.
	if i < 10 {
		return string(rune('0' + i))
	}
	digits := make([]byte, 0, 4)
	for i > 0 {
		digits = append(digits, byte('0'+i%10))
		i /= 10
	}
	for l, r := 0, len(digits)-1; l < r; l, r = l+1, r-1 {
		digits[l], digits[r] = digits[r], digits[l]
	}
	return string(digits)
}

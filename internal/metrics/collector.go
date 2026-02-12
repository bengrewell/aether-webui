package metrics

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/bengrewell/aether-webui/internal/operator/host"
	"github.com/bengrewell/aether-webui/internal/state"
)

// Collector periodically polls metrics from a host operator and stores them.
type Collector struct {
	hostOp host.HostOperator
	store  state.Store
	config Config

	mu      sync.Mutex
	stopped bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewCollector creates a new metrics collector.
func NewCollector(hostOp host.HostOperator, store state.Store, cfg Config) *Collector {
	if cfg.Interval == 0 {
		cfg.Interval = DefaultConfig().Interval
	}
	if cfg.Retention == 0 {
		cfg.Retention = DefaultConfig().Retention
	}

	return &Collector{
		hostOp: hostOp,
		store:  store,
		config: cfg,
		stopCh: make(chan struct{}),
	}
}

// Start begins the metrics collection loop. It blocks until Stop is called
// or the context is cancelled.
func (c *Collector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return nil
	}
	c.wg.Add(1)
	c.mu.Unlock()

	slog.Info("metrics collector starting",
		"interval", c.config.Interval,
		"retention", c.config.Retention,
	)

	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	// Start pruning goroutine
	go c.pruneLoop(ctx)

	// Collect immediately on start
	c.collectAll(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("metrics collector stopping (context cancelled)")
			return ctx.Err()
		case <-c.stopCh:
			slog.Info("metrics collector stopping (stop called)")
			return nil
		case <-ticker.C:
			c.collectAll(ctx)
		}
	}
}

// Stop signals the collector to stop.
func (c *Collector) Stop() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.stopped = true
	close(c.stopCh)
	c.mu.Unlock()

	c.wg.Wait()
}

func (c *Collector) collectAll(ctx context.Context) {
	c.collectCPU(ctx)
	c.collectMemory(ctx)
	c.collectDisk(ctx)
	c.collectNIC(ctx)
}

func (c *Collector) collectCPU(ctx context.Context) {
	usage, err := c.hostOp.GetCPUUsage(ctx)
	if err != nil {
		slog.Warn("failed to collect CPU metrics", "error", err)
		return
	}

	data, err := json.Marshal(usage)
	if err != nil {
		slog.Warn("failed to marshal CPU metrics", "error", err)
		return
	}

	if err := c.store.RecordMetrics(ctx, string(MetricTypeCPU), data); err != nil {
		slog.Warn("failed to store CPU metrics", "error", err)
	}
}

func (c *Collector) collectMemory(ctx context.Context) {
	usage, err := c.hostOp.GetMemoryUsage(ctx)
	if err != nil {
		slog.Warn("failed to collect memory metrics", "error", err)
		return
	}

	data, err := json.Marshal(usage)
	if err != nil {
		slog.Warn("failed to marshal memory metrics", "error", err)
		return
	}

	if err := c.store.RecordMetrics(ctx, string(MetricTypeMemory), data); err != nil {
		slog.Warn("failed to store memory metrics", "error", err)
	}
}

func (c *Collector) collectDisk(ctx context.Context) {
	usage, err := c.hostOp.GetDiskUsage(ctx)
	if err != nil {
		slog.Warn("failed to collect disk metrics", "error", err)
		return
	}

	data, err := json.Marshal(usage)
	if err != nil {
		slog.Warn("failed to marshal disk metrics", "error", err)
		return
	}

	if err := c.store.RecordMetrics(ctx, string(MetricTypeDisk), data); err != nil {
		slog.Warn("failed to store disk metrics", "error", err)
	}
}

func (c *Collector) collectNIC(ctx context.Context) {
	usage, err := c.hostOp.GetNICUsage(ctx)
	if err != nil {
		slog.Warn("failed to collect NIC metrics", "error", err)
		return
	}

	data, err := json.Marshal(usage)
	if err != nil {
		slog.Warn("failed to marshal NIC metrics", "error", err)
		return
	}

	if err := c.store.RecordMetrics(ctx, string(MetricTypeNIC), data); err != nil {
		slog.Warn("failed to store NIC metrics", "error", err)
	}
}

func (c *Collector) pruneLoop(ctx context.Context) {
	defer c.wg.Done()

	// Prune every hour
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	// Prune immediately on start
	c.prune(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.prune(ctx)
		}
	}
}

func (c *Collector) prune(ctx context.Context) {
	if err := c.store.PruneOldMetrics(ctx, c.config.Retention); err != nil {
		slog.Warn("failed to prune old metrics", "error", err)
	} else {
		slog.Debug("pruned old metrics", "retention", c.config.Retention)
	}
}

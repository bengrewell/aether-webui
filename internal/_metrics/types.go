package metrics

import (
	"time"

	"github.com/bengrewell/aether-webui/internal/operator/host"
)

// Config holds configuration for the metrics collector.
type Config struct {
	Interval  time.Duration // How often to collect metrics (default: 10s)
	Retention time.Duration // How long to keep historical data (default: 24h)
}

// DefaultConfig returns the default collector configuration.
func DefaultConfig() Config {
	return Config{
		Interval:  10 * time.Second,
		Retention: 24 * time.Hour,
	}
}

// MetricType identifies different metric categories.
type MetricType string

const (
	MetricTypeCPU    MetricType = "cpu"
	MetricTypeMemory MetricType = "memory"
	MetricTypeDisk   MetricType = "disk"
	MetricTypeNIC    MetricType = "nic"
)

// AggregatedPoint represents a single point in an aggregated time series.
type AggregatedPoint struct {
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// AggregatedCPU represents aggregated CPU metrics.
type AggregatedCPU struct {
	UsagePercent  float64   `json:"usage_percent"`
	UserPercent   float64   `json:"user_percent"`
	SystemPercent float64   `json:"system_percent"`
	IdlePercent   float64   `json:"idle_percent"`
	IOWaitPercent float64   `json:"iowait_percent"`
	PerCoreUsage  []float64 `json:"per_core_usage,omitempty"`
	LoadAverage1  float64   `json:"load_average_1"`
	LoadAverage5  float64   `json:"load_average_5"`
	LoadAverage15 float64   `json:"load_average_15"`
	SampleCount   int       `json:"sample_count"`
}

// AggregatedMemory represents aggregated memory metrics.
type AggregatedMemory struct {
	UsedBytesAvg      uint64  `json:"used_bytes_avg"`
	UsedBytesMax      uint64  `json:"used_bytes_max"`
	FreeBytesAvg      uint64  `json:"free_bytes_avg"`
	AvailableBytesAvg uint64  `json:"available_bytes_avg"`
	CachedBytesAvg    uint64  `json:"cached_bytes_avg"`
	UsagePercentAvg   float64 `json:"usage_percent_avg"`
	UsagePercentMax   float64 `json:"usage_percent_max"`
	SwapUsedBytesAvg  uint64  `json:"swap_used_bytes_avg"`
	SampleCount       int     `json:"sample_count"`
}

// AggregatedDisk represents aggregated disk metrics (latest per mount).
type AggregatedDisk struct {
	Disks       []host.DiskUsageEntry `json:"disks"`
	SampleCount int                   `json:"sample_count"`
}

// AggregatedNIC represents aggregated NIC metrics.
type AggregatedNIC struct {
	Interfaces  []AggregatedNICEntry `json:"interfaces"`
	SampleCount int                  `json:"sample_count"`
}

// AggregatedNICEntry represents aggregated metrics for a single NIC.
type AggregatedNICEntry struct {
	Name             string  `json:"name"`
	RxBytesTotal     uint64  `json:"rx_bytes_total"`     // Sum of deltas
	TxBytesTotal     uint64  `json:"tx_bytes_total"`     // Sum of deltas
	RxPacketsTotal   uint64  `json:"rx_packets_total"`   // Sum of deltas
	TxPacketsTotal   uint64  `json:"tx_packets_total"`   // Sum of deltas
	RxBytesPerSecAvg float64 `json:"rx_bytes_per_sec_avg"`
	TxBytesPerSecAvg float64 `json:"tx_bytes_per_sec_avg"`
}

// HistoricalResponse wraps historical metrics data for API responses.
type HistoricalResponse struct {
	Points      []AggregatedPoint `json:"points"`
	Window      string            `json:"window"`
	Granularity string            `json:"granularity,omitempty"`
	SampleCount int               `json:"sample_count"`
}

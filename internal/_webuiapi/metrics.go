package webuiapi

import (
	"context"
	"time"

	"github.com/bengrewell/aether-webui/internal/metrics"
	"github.com/bengrewell/aether-webui/internal/operator/host"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/danielgtaylor/huma/v2"
)

// MetricsInput extends NodeInput with optional window and granularity parameters.
type MetricsInput struct {
	Node        string `query:"node" default:"local" doc:"Target node identifier. Use 'local' or empty for the local node."`
	Window      string `query:"window" doc:"Time window for historical data, e.g. '60m', '24h'. If empty, returns current value."`
	Granularity string `query:"granularity" doc:"Aggregation granularity for historical data, e.g. '30s', '5m'. If empty with window set, returns raw points."`
}

// CPUUsageOutput is the response for GET /api/v1/metrics/cpu (current)
type CPUUsageOutput struct {
	Body host.CPUUsage
}

// CPUHistoricalOutput is the response for GET /api/v1/metrics/cpu with window
type CPUHistoricalOutput struct {
	Body metrics.HistoricalResponse
}

// MemoryUsageOutput is the response for GET /api/v1/metrics/memory (current)
type MemoryUsageOutput struct {
	Body host.MemoryUsage
}

// MemoryHistoricalOutput is the response for GET /api/v1/metrics/memory with window
type MemoryHistoricalOutput struct {
	Body metrics.HistoricalResponse
}

// DiskUsageOutput is the response for GET /api/v1/metrics/disk (current)
type DiskUsageOutput struct {
	Body host.DiskUsage
}

// DiskHistoricalOutput is the response for GET /api/v1/metrics/disk with window
type DiskHistoricalOutput struct {
	Body metrics.HistoricalResponse
}

// NICUsageOutput is the response for GET /api/v1/metrics/nic (current)
type NICUsageOutput struct {
	Body host.NICUsage
}

// NICHistoricalOutput is the response for GET /api/v1/metrics/nic with window
type NICHistoricalOutput struct {
	Body metrics.HistoricalResponse
}

// MetricsRoutesDeps contains dependencies for metrics routes.
type MetricsRoutesDeps struct {
	Resolver provider.ProviderResolver
	Store    state.Store
}

// RegisterMetricsRoutes registers dynamic metrics routes.
func RegisterMetricsRoutes(api huma.API, resolver provider.ProviderResolver) {
	RegisterMetricsRoutesWithStore(api, MetricsRoutesDeps{Resolver: resolver, Store: nil})
}

// RegisterMetricsRoutesWithStore registers metrics routes with optional store for historical queries.
func RegisterMetricsRoutesWithStore(api huma.API, deps MetricsRoutesDeps) {
	huma.Register(api, huma.Operation{
		OperationID: "get-cpu-usage",
		Method:      "GET",
		Path:        "/api/v1/metrics/cpu",
		Summary:     "Get CPU usage",
		Description: "Returns current CPU usage metrics for the specified node. With window parameter, returns historical data.",
		Tags:        []string{"Metrics"},
	}, func(ctx context.Context, input *MetricsInput) (*CPUUsageOutput, error) {
		// Check for historical query
		if input.Window != "" && deps.Store != nil {
			return nil, nil // Will be handled by historical handler
		}

		hostOp, err := getHostOperator(deps.Resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		usage, err := hostOp.GetCPUUsage(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get CPU usage", err)
		}
		return &CPUUsageOutput{Body: *usage}, nil
	})

	// Register historical CPU endpoint if store is available
	if deps.Store != nil {
		huma.Register(api, huma.Operation{
			OperationID:   "get-cpu-usage-historical",
			Method:        "GET",
			Path:          "/api/v1/metrics/cpu/history",
			Summary:       "Get CPU usage history",
			Description:   "Returns historical CPU usage metrics for the specified time window.",
			Tags:          []string{"Metrics"},
			DefaultStatus: 200,
		}, func(ctx context.Context, input *MetricsInput) (*CPUHistoricalOutput, error) {
			return getHistoricalMetrics[CPUHistoricalOutput](ctx, deps.Store, input, metrics.MetricTypeCPU)
		})
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-memory-usage",
		Method:      "GET",
		Path:        "/api/v1/metrics/memory",
		Summary:     "Get memory usage",
		Description: "Returns current memory usage metrics for the specified node.",
		Tags:        []string{"Metrics"},
	}, func(ctx context.Context, input *MetricsInput) (*MemoryUsageOutput, error) {
		hostOp, err := getHostOperator(deps.Resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		usage, err := hostOp.GetMemoryUsage(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get memory usage", err)
		}
		return &MemoryUsageOutput{Body: *usage}, nil
	})

	if deps.Store != nil {
		huma.Register(api, huma.Operation{
			OperationID:   "get-memory-usage-historical",
			Method:        "GET",
			Path:          "/api/v1/metrics/memory/history",
			Summary:       "Get memory usage history",
			Description:   "Returns historical memory usage metrics for the specified time window.",
			Tags:          []string{"Metrics"},
			DefaultStatus: 200,
		}, func(ctx context.Context, input *MetricsInput) (*MemoryHistoricalOutput, error) {
			return getHistoricalMetrics[MemoryHistoricalOutput](ctx, deps.Store, input, metrics.MetricTypeMemory)
		})
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-disk-usage",
		Method:      "GET",
		Path:        "/api/v1/metrics/disk",
		Summary:     "Get disk usage",
		Description: "Returns current disk usage metrics for the specified node.",
		Tags:        []string{"Metrics"},
	}, func(ctx context.Context, input *MetricsInput) (*DiskUsageOutput, error) {
		hostOp, err := getHostOperator(deps.Resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		usage, err := hostOp.GetDiskUsage(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get disk usage", err)
		}
		return &DiskUsageOutput{Body: *usage}, nil
	})

	if deps.Store != nil {
		huma.Register(api, huma.Operation{
			OperationID:   "get-disk-usage-historical",
			Method:        "GET",
			Path:          "/api/v1/metrics/disk/history",
			Summary:       "Get disk usage history",
			Description:   "Returns historical disk usage metrics for the specified time window.",
			Tags:          []string{"Metrics"},
			DefaultStatus: 200,
		}, func(ctx context.Context, input *MetricsInput) (*DiskHistoricalOutput, error) {
			return getHistoricalMetrics[DiskHistoricalOutput](ctx, deps.Store, input, metrics.MetricTypeDisk)
		})
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-nic-usage",
		Method:      "GET",
		Path:        "/api/v1/metrics/nic",
		Summary:     "Get network interface usage",
		Description: "Returns current network interface usage metrics for the specified node.",
		Tags:        []string{"Metrics"},
	}, func(ctx context.Context, input *MetricsInput) (*NICUsageOutput, error) {
		hostOp, err := getHostOperator(deps.Resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		usage, err := hostOp.GetNICUsage(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get NIC usage", err)
		}
		return &NICUsageOutput{Body: *usage}, nil
	})

	if deps.Store != nil {
		huma.Register(api, huma.Operation{
			OperationID:   "get-nic-usage-historical",
			Method:        "GET",
			Path:          "/api/v1/metrics/nic/history",
			Summary:       "Get network interface usage history",
			Description:   "Returns historical network interface usage metrics for the specified time window.",
			Tags:          []string{"Metrics"},
			DefaultStatus: 200,
		}, func(ctx context.Context, input *MetricsInput) (*NICHistoricalOutput, error) {
			return getHistoricalMetrics[NICHistoricalOutput](ctx, deps.Store, input, metrics.MetricTypeNIC)
		})
	}
}

// historicalOutput is a constraint for historical output types.
type historicalOutput interface {
	CPUHistoricalOutput | MemoryHistoricalOutput | DiskHistoricalOutput | NICHistoricalOutput
}

func getHistoricalMetrics[T historicalOutput](ctx context.Context, store state.Store, input *MetricsInput, metricType metrics.MetricType) (*T, error) {
	// Parse window duration
	window, err := time.ParseDuration(input.Window)
	if err != nil || input.Window == "" {
		// Default to 1 hour if not specified or invalid
		window = time.Hour
	}

	// Parse granularity if specified
	var granularity time.Duration
	if input.Granularity != "" {
		granularity, err = time.ParseDuration(input.Granularity)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid granularity format", err)
		}
	}

	// Calculate time range
	end := time.Now()
	start := end.Add(-window)

	// Get metrics from store
	snapshots, err := store.GetMetricsRange(ctx, string(metricType), start, end)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get metrics history", err)
	}

	// Aggregate
	points := metrics.Aggregate(snapshots, metricType, granularity)

	resp := metrics.HistoricalResponse{
		Points:      points,
		Window:      input.Window,
		SampleCount: len(snapshots),
	}
	if input.Granularity != "" {
		resp.Granularity = input.Granularity
	}

	// Create the appropriate output type
	result := new(T)
	switch any(result).(type) {
	case *CPUHistoricalOutput:
		*any(result).(*CPUHistoricalOutput) = CPUHistoricalOutput{Body: resp}
	case *MemoryHistoricalOutput:
		*any(result).(*MemoryHistoricalOutput) = MemoryHistoricalOutput{Body: resp}
	case *DiskHistoricalOutput:
		*any(result).(*DiskHistoricalOutput) = DiskHistoricalOutput{Body: resp}
	case *NICHistoricalOutput:
		*any(result).(*NICHistoricalOutput) = NICHistoricalOutput{Body: resp}
	}

	return result, nil
}

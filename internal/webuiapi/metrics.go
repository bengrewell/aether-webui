package webuiapi

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/operator/host"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/danielgtaylor/huma/v2"
)

// CPUUsageOutput is the response for GET /api/v1/metrics/cpu
type CPUUsageOutput struct {
	Body host.CPUUsage
}

// MemoryUsageOutput is the response for GET /api/v1/metrics/memory
type MemoryUsageOutput struct {
	Body host.MemoryUsage
}

// DiskUsageOutput is the response for GET /api/v1/metrics/disk
type DiskUsageOutput struct {
	Body host.DiskUsage
}

// NICUsageOutput is the response for GET /api/v1/metrics/nic
type NICUsageOutput struct {
	Body host.NICUsage
}

// RegisterMetricsRoutes registers dynamic metrics routes.
func RegisterMetricsRoutes(api huma.API, resolver provider.ProviderResolver) {
	huma.Register(api, huma.Operation{
		OperationID: "get-cpu-usage",
		Method:      "GET",
		Path:        "/api/v1/metrics/cpu",
		Summary:     "Get CPU usage",
		Description: "Returns current CPU usage metrics for the specified node",
		Tags:        []string{"Metrics"},
	}, func(ctx context.Context, input *NodeInput) (*CPUUsageOutput, error) {
		hostOp, err := getHostOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		usage, err := hostOp.GetCPUUsage(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get CPU usage", err)
		}
		return &CPUUsageOutput{Body: *usage}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-memory-usage",
		Method:      "GET",
		Path:        "/api/v1/metrics/memory",
		Summary:     "Get memory usage",
		Description: "Returns current memory usage metrics for the specified node",
		Tags:        []string{"Metrics"},
	}, func(ctx context.Context, input *NodeInput) (*MemoryUsageOutput, error) {
		hostOp, err := getHostOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		usage, err := hostOp.GetMemoryUsage(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get memory usage", err)
		}
		return &MemoryUsageOutput{Body: *usage}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-disk-usage",
		Method:      "GET",
		Path:        "/api/v1/metrics/disk",
		Summary:     "Get disk usage",
		Description: "Returns current disk usage metrics for the specified node",
		Tags:        []string{"Metrics"},
	}, func(ctx context.Context, input *NodeInput) (*DiskUsageOutput, error) {
		hostOp, err := getHostOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		usage, err := hostOp.GetDiskUsage(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get disk usage", err)
		}
		return &DiskUsageOutput{Body: *usage}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-nic-usage",
		Method:      "GET",
		Path:        "/api/v1/metrics/nic",
		Summary:     "Get network interface usage",
		Description: "Returns current network interface usage metrics for the specified node",
		Tags:        []string{"Metrics"},
	}, func(ctx context.Context, input *NodeInput) (*NICUsageOutput, error) {
		hostOp, err := getHostOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		usage, err := hostOp.GetNICUsage(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get NIC usage", err)
		}
		return &NICUsageOutput{Body: *usage}, nil
	})
}

package webuiapi

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/sysinfo"
	"github.com/danielgtaylor/huma/v2"
)

// NodeInput is the common input for endpoints that accept a node parameter.
type NodeInput struct {
	Node string `query:"node" default:"local" doc:"Target node identifier. Use 'local' or empty for the local node."`
}

// CPUInfoOutput is the response for GET /api/v1/system/cpu
type CPUInfoOutput struct {
	Body sysinfo.CPUInfo
}

// MemoryInfoOutput is the response for GET /api/v1/system/memory
type MemoryInfoOutput struct {
	Body sysinfo.MemoryInfo
}

// DiskInfoOutput is the response for GET /api/v1/system/disk
type DiskInfoOutput struct {
	Body sysinfo.DiskInfo
}

// NICInfoOutput is the response for GET /api/v1/system/nic
type NICInfoOutput struct {
	Body sysinfo.NICInfo
}

// OSInfoOutput is the response for GET /api/v1/system/os
type OSInfoOutput struct {
	Body sysinfo.OSInfo
}

// RegisterSystemRoutes registers static system information routes.
func RegisterSystemRoutes(api huma.API, resolver sysinfo.NodeResolver) {
	huma.Register(api, huma.Operation{
		OperationID: "get-cpu-info",
		Method:      "GET",
		Path:        "/api/v1/system/cpu",
		Summary:     "Get CPU information",
		Description: "Returns static CPU information for the specified node",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *NodeInput) (*CPUInfoOutput, error) {
		provider, err := resolver.Resolve(input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		info, err := provider.GetCPUInfo(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get CPU info", err)
		}
		return &CPUInfoOutput{Body: *info}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-memory-info",
		Method:      "GET",
		Path:        "/api/v1/system/memory",
		Summary:     "Get memory information",
		Description: "Returns static memory information for the specified node",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *NodeInput) (*MemoryInfoOutput, error) {
		provider, err := resolver.Resolve(input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		info, err := provider.GetMemoryInfo(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get memory info", err)
		}
		return &MemoryInfoOutput{Body: *info}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-disk-info",
		Method:      "GET",
		Path:        "/api/v1/system/disk",
		Summary:     "Get disk information",
		Description: "Returns static disk information for the specified node",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *NodeInput) (*DiskInfoOutput, error) {
		provider, err := resolver.Resolve(input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		info, err := provider.GetDiskInfo(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get disk info", err)
		}
		return &DiskInfoOutput{Body: *info}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-nic-info",
		Method:      "GET",
		Path:        "/api/v1/system/nic",
		Summary:     "Get network interface information",
		Description: "Returns static network interface information for the specified node",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *NodeInput) (*NICInfoOutput, error) {
		provider, err := resolver.Resolve(input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		info, err := provider.GetNICInfo(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get NIC info", err)
		}
		return &NICInfoOutput{Body: *info}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-os-info",
		Method:      "GET",
		Path:        "/api/v1/system/os",
		Summary:     "Get operating system information",
		Description: "Returns static operating system information for the specified node",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *NodeInput) (*OSInfoOutput, error) {
		provider, err := resolver.Resolve(input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		info, err := provider.GetOSInfo(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get OS info", err)
		}
		return &OSInfoOutput{Body: *info}, nil
	})
}

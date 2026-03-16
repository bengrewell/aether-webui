package mcp

import (
	"context"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/bengrewell/aether-webui/internal/provider/system"
)

func (s *Server) registerSystemTools() {
	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "system_overview",
		Description: "Get system overview: CPU, memory, disk, and OS information",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ SystemOverviewInput) (*gomcp.CallToolResult, any, error) {
		type overview struct {
			CPU    any `json:"cpu"`
			Memory any `json:"memory"`
			Disks  any `json:"disks"`
			OS     any `json:"os"`
		}
		var result overview

		if cpu, err := s.system.HandleCPU(ctx, nil); err == nil {
			result.CPU = cpu.Body
		}
		if mem, err := s.system.HandleMemory(ctx, nil); err == nil {
			result.Memory = mem.Body
		}
		if disks, err := s.system.HandleDisks(ctx, nil); err == nil {
			result.Disks = disks.Body
		}
		if osInfo, err := s.system.HandleOS(ctx, nil); err == nil {
			result.OS = osInfo.Body
		}

		return jsonResult(result), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "system_network",
		Description: "Get network interfaces, DNS configuration, and listening ports",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, _ SystemNetworkInput) (*gomcp.CallToolResult, any, error) {
		type netInfo struct {
			Interfaces any `json:"interfaces"`
			Config     any `json:"config"`
			Ports      any `json:"ports"`
		}
		var result netInfo

		if ifaces, err := s.system.HandleNetworkInterfaces(ctx, nil); err == nil {
			result.Interfaces = ifaces.Body
		}
		if cfg, err := s.system.HandleNetworkConfig(ctx, nil); err == nil {
			result.Config = cfg.Body
		}
		if ports, err := s.system.HandleListeningPorts(ctx, nil); err == nil {
			result.Ports = ports.Body
		}

		return jsonResult(result), nil, nil
	})

	gomcp.AddTool(s.srv, &gomcp.Tool{
		Name:        "system_metrics",
		Description: "Query system time-series metrics with optional time range, label filters, and aggregation",
	}, func(ctx context.Context, _ *gomcp.CallToolRequest, args SystemMetricsInput) (*gomcp.CallToolResult, any, error) {
		out, err := s.system.HandleMetricsQuery(ctx, &system.MetricsQueryInput{
			Metric:      args.Metric,
			From:        args.From,
			To:          args.To,
			Labels:      args.Labels,
			Aggregation: args.Aggregation,
		})
		if err != nil {
			return errorResult(err), nil, nil
		}
		return jsonResult(out.Body), nil, nil
	})
}

package sysinfo

import "context"

// Provider defines the interface for retrieving system information.
// Implementations can target local systems or remote nodes.
type Provider interface {
	// Static information
	GetCPUInfo(ctx context.Context) (*CPUInfo, error)
	GetMemoryInfo(ctx context.Context) (*MemoryInfo, error)
	GetDiskInfo(ctx context.Context) (*DiskInfo, error)
	GetNICInfo(ctx context.Context) (*NICInfo, error)
	GetOSInfo(ctx context.Context) (*OSInfo, error)

	// Dynamic usage metrics
	GetCPUUsage(ctx context.Context) (*CPUUsage, error)
	GetMemoryUsage(ctx context.Context) (*MemoryUsage, error)
	GetDiskUsage(ctx context.Context) (*DiskUsage, error)
	GetNICUsage(ctx context.Context) (*NICUsage, error)
}

// NodeResolver resolves a node identifier to a Provider.
// When node is empty or "local", it returns the local provider.
type NodeResolver interface {
	Resolve(node string) (Provider, error)
}

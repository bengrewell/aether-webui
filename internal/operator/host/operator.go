package host

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/operator"
)

// HostOperator handles host/system information and metrics.
type HostOperator interface {
	operator.Operator

	// Static info
	GetCPUInfo(ctx context.Context) (*CPUInfo, error)
	GetMemoryInfo(ctx context.Context) (*MemoryInfo, error)
	GetDiskInfo(ctx context.Context) (*DiskInfo, error)
	GetNICInfo(ctx context.Context) (*NICInfo, error)
	GetOSInfo(ctx context.Context) (*OSInfo, error)

	// Dynamic metrics
	GetCPUUsage(ctx context.Context) (*CPUUsage, error)
	GetMemoryUsage(ctx context.Context) (*MemoryUsage, error)
	GetDiskUsage(ctx context.Context) (*DiskUsage, error)
	GetNICUsage(ctx context.Context) (*NICUsage, error)
}

// Operator returns "not implemented" for all methods.
type Operator struct{}

// New creates a new host operator.
func New() *Operator {
	return &Operator{}
}

// Domain returns the operator's domain.
func (o *Operator) Domain() operator.Domain {
	return operator.DomainHost
}

// Health returns the operator's health status.
func (o *Operator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{
		Status:  "unavailable",
		Message: "not implemented",
	}, nil
}

// GetCPUInfo returns CPU information.
func (o *Operator) GetCPUInfo(_ context.Context) (*CPUInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetMemoryInfo returns memory information.
func (o *Operator) GetMemoryInfo(_ context.Context) (*MemoryInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetDiskInfo returns disk information.
func (o *Operator) GetDiskInfo(_ context.Context) (*DiskInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetNICInfo returns network interface information.
func (o *Operator) GetNICInfo(_ context.Context) (*NICInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetOSInfo returns operating system information.
func (o *Operator) GetOSInfo(_ context.Context) (*OSInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetCPUUsage returns current CPU usage metrics.
func (o *Operator) GetCPUUsage(_ context.Context) (*CPUUsage, error) {
	return nil, operator.ErrNotImplemented
}

// GetMemoryUsage returns current memory usage metrics.
func (o *Operator) GetMemoryUsage(_ context.Context) (*MemoryUsage, error) {
	return nil, operator.ErrNotImplemented
}

// GetDiskUsage returns current disk usage metrics.
func (o *Operator) GetDiskUsage(_ context.Context) (*DiskUsage, error) {
	return nil, operator.ErrNotImplemented
}

// GetNICUsage returns current network interface usage metrics.
func (o *Operator) GetNICUsage(_ context.Context) (*NICUsage, error) {
	return nil, operator.ErrNotImplemented
}

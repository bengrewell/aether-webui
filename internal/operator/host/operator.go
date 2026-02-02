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

// StubOperator returns "not implemented" for all methods.
type StubOperator struct{}

// NewStubOperator creates a new stub host operator.
func NewStubOperator() *StubOperator {
	return &StubOperator{}
}

// Domain returns the operator's domain.
func (o *StubOperator) Domain() operator.Domain {
	return operator.DomainHost
}

// Health returns the operator's health status.
func (o *StubOperator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{
		Status:  "unavailable",
		Message: "not implemented",
	}, nil
}

// GetCPUInfo returns CPU information.
func (o *StubOperator) GetCPUInfo(_ context.Context) (*CPUInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetMemoryInfo returns memory information.
func (o *StubOperator) GetMemoryInfo(_ context.Context) (*MemoryInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetDiskInfo returns disk information.
func (o *StubOperator) GetDiskInfo(_ context.Context) (*DiskInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetNICInfo returns network interface information.
func (o *StubOperator) GetNICInfo(_ context.Context) (*NICInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetOSInfo returns operating system information.
func (o *StubOperator) GetOSInfo(_ context.Context) (*OSInfo, error) {
	return nil, operator.ErrNotImplemented
}

// GetCPUUsage returns current CPU usage metrics.
func (o *StubOperator) GetCPUUsage(_ context.Context) (*CPUUsage, error) {
	return nil, operator.ErrNotImplemented
}

// GetMemoryUsage returns current memory usage metrics.
func (o *StubOperator) GetMemoryUsage(_ context.Context) (*MemoryUsage, error) {
	return nil, operator.ErrNotImplemented
}

// GetDiskUsage returns current disk usage metrics.
func (o *StubOperator) GetDiskUsage(_ context.Context) (*DiskUsage, error) {
	return nil, operator.ErrNotImplemented
}

// GetNICUsage returns current network interface usage metrics.
func (o *StubOperator) GetNICUsage(_ context.Context) (*NICUsage, error) {
	return nil, operator.ErrNotImplemented
}

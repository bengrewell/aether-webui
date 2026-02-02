package aether

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/operator"
)

// AetherOperator handles Aether 5G component management.
type AetherOperator interface {
	operator.Operator

	// Core management
	ListCores(ctx context.Context) (*CoreList, error)
	GetCore(ctx context.Context, id string) (*CoreConfig, error)
	DeployCore(ctx context.Context, config *CoreConfig) (*DeploymentResponse, error)
	UpdateCore(ctx context.Context, id string, config *CoreConfig) error
	UndeployCore(ctx context.Context, id string) (*DeploymentResponse, error)
	GetCoreStatus(ctx context.Context, id string) (*CoreStatus, error)
	ListCoreStatuses(ctx context.Context) (*CoreStatusList, error)

	// gNB management
	ListGNBs(ctx context.Context) (*GNBList, error)
	GetGNB(ctx context.Context, id string) (*GNBConfig, error)
	DeployGNB(ctx context.Context, config *GNBConfig) (*DeploymentResponse, error)
	UpdateGNB(ctx context.Context, id string, config *GNBConfig) error
	UndeployGNB(ctx context.Context, id string) (*DeploymentResponse, error)
	GetGNBStatus(ctx context.Context, id string) (*GNBStatus, error)
	ListGNBStatuses(ctx context.Context) (*GNBStatusList, error)
}

// StubOperator returns "not implemented" for all methods.
type StubOperator struct{}

// NewStubOperator creates a new stub Aether operator.
func NewStubOperator() *StubOperator {
	return &StubOperator{}
}

// Domain returns the operator's domain.
func (o *StubOperator) Domain() operator.Domain {
	return operator.DomainAether
}

// Health returns the operator's health status.
func (o *StubOperator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{
		Status:  "unavailable",
		Message: "not implemented",
	}, nil
}

// ListCores returns all SD-Core deployments.
func (o *StubOperator) ListCores(_ context.Context) (*CoreList, error) {
	return nil, operator.ErrNotImplemented
}

// GetCore returns a specific SD-Core configuration.
func (o *StubOperator) GetCore(_ context.Context, _ string) (*CoreConfig, error) {
	return nil, operator.ErrNotImplemented
}

// DeployCore deploys a new SD-Core instance.
func (o *StubOperator) DeployCore(_ context.Context, _ *CoreConfig) (*DeploymentResponse, error) {
	return nil, operator.ErrNotImplemented
}

// UpdateCore updates an SD-Core configuration.
func (o *StubOperator) UpdateCore(_ context.Context, _ string, _ *CoreConfig) error {
	return operator.ErrNotImplemented
}

// UndeployCore removes an SD-Core deployment.
func (o *StubOperator) UndeployCore(_ context.Context, _ string) (*DeploymentResponse, error) {
	return nil, operator.ErrNotImplemented
}

// GetCoreStatus returns the status of a specific SD-Core.
func (o *StubOperator) GetCoreStatus(_ context.Context, _ string) (*CoreStatus, error) {
	return nil, operator.ErrNotImplemented
}

// ListCoreStatuses returns status for all SD-Core deployments.
func (o *StubOperator) ListCoreStatuses(_ context.Context) (*CoreStatusList, error) {
	return nil, operator.ErrNotImplemented
}

// ListGNBs returns all gNB configurations.
func (o *StubOperator) ListGNBs(_ context.Context) (*GNBList, error) {
	return nil, operator.ErrNotImplemented
}

// GetGNB returns a specific gNB configuration.
func (o *StubOperator) GetGNB(_ context.Context, _ string) (*GNBConfig, error) {
	return nil, operator.ErrNotImplemented
}

// DeployGNB deploys a new gNB instance.
func (o *StubOperator) DeployGNB(_ context.Context, _ *GNBConfig) (*DeploymentResponse, error) {
	return nil, operator.ErrNotImplemented
}

// UpdateGNB updates a gNB configuration.
func (o *StubOperator) UpdateGNB(_ context.Context, _ string, _ *GNBConfig) error {
	return operator.ErrNotImplemented
}

// UndeployGNB removes a gNB deployment.
func (o *StubOperator) UndeployGNB(_ context.Context, _ string) (*DeploymentResponse, error) {
	return nil, operator.ErrNotImplemented
}

// GetGNBStatus returns the status of a specific gNB.
func (o *StubOperator) GetGNBStatus(_ context.Context, _ string) (*GNBStatus, error) {
	return nil, operator.ErrNotImplemented
}

// ListGNBStatuses returns status for all gNBs.
func (o *StubOperator) ListGNBStatuses(_ context.Context) (*GNBStatusList, error) {
	return nil, operator.ErrNotImplemented
}

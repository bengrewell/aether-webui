package aether

import (
	"context"
	"errors"

	"github.com/bengrewell/aether-webui/internal/onramp"
	"github.com/bengrewell/aether-webui/internal/operator"
	"github.com/bengrewell/aether-webui/internal/state"
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

// Operator implements AetherOperator backed by OnRamp playbook execution.
type Operator struct {
	taskMgr *onramp.TaskManager
	store   state.Store
}

// New creates a new Aether operator backed by the OnRamp task manager.
func New(taskMgr *onramp.TaskManager, store state.Store) *Operator {
	return &Operator{taskMgr: taskMgr, store: store}
}

// Domain returns the operator's domain.
func (o *Operator) Domain() operator.Domain {
	return operator.DomainAether
}

// Health reports whether the OnRamp subsystem is available.
func (o *Operator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	if o.taskMgr == nil {
		return &operator.OperatorHealth{
			Status:  "unavailable",
			Message: "task manager not configured",
		}, nil
	}
	return &operator.OperatorHealth{
		Status:  "healthy",
		Message: "OnRamp task manager available",
	}, nil
}

// DeployCore starts the 5gc-install playbook sequence.
func (o *Operator) DeployCore(ctx context.Context, _ *CoreConfig) (*DeploymentResponse, error) {
	seq, ok := onramp.Sequences["5gc-install"]
	if !ok {
		return nil, errors.New("5gc-install sequence not defined")
	}
	taskID, err := o.taskMgr.StartSequence(ctx, seq, "5gc")
	if err != nil {
		return nil, err
	}
	return &DeploymentResponse{
		Success: true,
		Message: "5G Core deployment started",
		TaskID:  taskID,
	}, nil
}

// UndeployCore starts the 5gc-uninstall playbook sequence.
func (o *Operator) UndeployCore(ctx context.Context, _ string) (*DeploymentResponse, error) {
	seq, ok := onramp.Sequences["5gc-uninstall"]
	if !ok {
		return nil, errors.New("5gc-uninstall sequence not defined")
	}
	taskID, err := o.taskMgr.StartSequence(ctx, seq, "5gc")
	if err != nil {
		return nil, err
	}
	return &DeploymentResponse{
		Success: true,
		Message: "5G Core undeployment started",
		TaskID:  taskID,
	}, nil
}

// GetCoreStatus returns the deployment state of the 5GC component.
func (o *Operator) GetCoreStatus(ctx context.Context, _ string) (*CoreStatus, error) {
	ds, err := o.store.GetDeploymentState(ctx, "5gc")
	if err != nil {
		if errors.Is(err, state.ErrNotFound) {
			return &CoreStatus{
				ID:    "5gc",
				State: StateNotDeployed,
			}, nil
		}
		return nil, err
	}
	return &CoreStatus{
		ID:         "5gc",
		Host:       "local",
		State:      DeploymentState(ds.Status),
		DeployedAt: ds.DeployedAt,
	}, nil
}

// ListCoreStatuses returns status for all 5GC-related deployment states.
func (o *Operator) ListCoreStatuses(ctx context.Context) (*CoreStatusList, error) {
	status, err := o.GetCoreStatus(ctx, "5gc")
	if err != nil {
		return nil, err
	}
	return &CoreStatusList{Cores: []CoreStatus{*status}}, nil
}

// ListCores returns the list of cores (single deployment model for now).
func (o *Operator) ListCores(ctx context.Context) (*CoreList, error) {
	ds, err := o.store.GetDeploymentState(ctx, "5gc")
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return nil, err
	}
	if ds != nil && ds.Status != state.DeployStateNotDeployed {
		return &CoreList{Cores: []CoreConfig{{ID: "5gc", Name: "SD-Core"}}}, nil
	}
	return &CoreList{Cores: []CoreConfig{}}, nil
}

// GetCore returns the core configuration.
func (o *Operator) GetCore(_ context.Context, id string) (*CoreConfig, error) {
	if id != "5gc" {
		return nil, operator.ErrNotImplemented
	}
	return &CoreConfig{ID: "5gc", Name: "SD-Core"}, nil
}

// UpdateCore is not yet supported for OnRamp-backed deployments.
func (o *Operator) UpdateCore(_ context.Context, _ string, _ *CoreConfig) error {
	return operator.ErrNotImplemented
}

// DeployGNB starts the srsran-gnb-install playbook sequence.
func (o *Operator) DeployGNB(ctx context.Context, _ *GNBConfig) (*DeploymentResponse, error) {
	seq, ok := onramp.Sequences["srsran-gnb-install"]
	if !ok {
		return nil, errors.New("srsran-gnb-install sequence not defined")
	}
	taskID, err := o.taskMgr.StartSequence(ctx, seq, "srsran-gnb")
	if err != nil {
		return nil, err
	}
	return &DeploymentResponse{
		Success: true,
		Message: "srsRAN gNB deployment started",
		TaskID:  taskID,
	}, nil
}

// UndeployGNB starts the srsran-gnb-uninstall playbook sequence.
func (o *Operator) UndeployGNB(ctx context.Context, _ string) (*DeploymentResponse, error) {
	seq, ok := onramp.Sequences["srsran-gnb-uninstall"]
	if !ok {
		return nil, errors.New("srsran-gnb-uninstall sequence not defined")
	}
	taskID, err := o.taskMgr.StartSequence(ctx, seq, "srsran-gnb")
	if err != nil {
		return nil, err
	}
	return &DeploymentResponse{
		Success: true,
		Message: "srsRAN gNB undeployment started",
		TaskID:  taskID,
	}, nil
}

// GetGNBStatus returns the deployment state of srsRAN gNB.
func (o *Operator) GetGNBStatus(ctx context.Context, _ string) (*GNBStatus, error) {
	ds, err := o.store.GetDeploymentState(ctx, "srsran-gnb")
	if err != nil {
		if errors.Is(err, state.ErrNotFound) {
			return &GNBStatus{
				ID:    "srsran-gnb",
				Type:  "srsran",
				State: StateNotDeployed,
			}, nil
		}
		return nil, err
	}
	return &GNBStatus{
		ID:         "srsran-gnb",
		Host:       "local",
		Type:       "srsran",
		State:      DeploymentState(ds.Status),
		DeployedAt: ds.DeployedAt,
	}, nil
}

// ListGNBStatuses returns status for all gNB deployments.
func (o *Operator) ListGNBStatuses(ctx context.Context) (*GNBStatusList, error) {
	status, err := o.GetGNBStatus(ctx, "srsran-gnb")
	if err != nil {
		return nil, err
	}
	return &GNBStatusList{GNBs: []GNBStatus{*status}}, nil
}

// ListGNBs returns configured gNBs.
func (o *Operator) ListGNBs(ctx context.Context) (*GNBList, error) {
	ds, err := o.store.GetDeploymentState(ctx, "srsran-gnb")
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return nil, err
	}
	if ds != nil && ds.Status != state.DeployStateNotDeployed {
		return &GNBList{GNBs: []GNBConfig{{ID: "srsran-gnb", Name: "srsRAN gNB", Type: "srsran"}}}, nil
	}
	return &GNBList{GNBs: []GNBConfig{}}, nil
}

// GetGNB returns a gNB configuration.
func (o *Operator) GetGNB(_ context.Context, id string) (*GNBConfig, error) {
	if id != "srsran-gnb" {
		return nil, operator.ErrNotImplemented
	}
	return &GNBConfig{ID: "srsran-gnb", Name: "srsRAN gNB", Type: "srsran"}, nil
}

// UpdateGNB is not yet supported for OnRamp-backed deployments.
func (o *Operator) UpdateGNB(_ context.Context, _ string, _ *GNBConfig) error {
	return operator.ErrNotImplemented
}

package kube

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/operator"
)

// KubeOperator handles Kubernetes cluster monitoring and information.
type KubeOperator interface {
	operator.Operator

	GetClusterHealth(ctx context.Context) (*ClusterHealth, error)
	GetNodes(ctx context.Context) (*NodeList, error)
	GetNamespaces(ctx context.Context) (*NamespaceList, error)
	GetEvents(ctx context.Context, namespace string, limit int) (*EventList, error)
	GetPods(ctx context.Context, namespace string) (*PodList, error)
	GetDeployments(ctx context.Context, namespace string) (*DeploymentList, error)
	GetServices(ctx context.Context, namespace string) (*ServiceList, error)
}

// StubOperator returns "not implemented" for all methods.
type StubOperator struct{}

// NewStubOperator creates a new stub Kubernetes operator.
func NewStubOperator() *StubOperator {
	return &StubOperator{}
}

// Domain returns the operator's domain.
func (o *StubOperator) Domain() operator.Domain {
	return operator.DomainKube
}

// Health returns the operator's health status.
func (o *StubOperator) Health(_ context.Context) (*operator.OperatorHealth, error) {
	return &operator.OperatorHealth{
		Status:  "unavailable",
		Message: "not implemented",
	}, nil
}

// GetClusterHealth returns cluster health information.
func (o *StubOperator) GetClusterHealth(_ context.Context) (*ClusterHealth, error) {
	return nil, operator.ErrNotImplemented
}

// GetNodes returns cluster node information.
func (o *StubOperator) GetNodes(_ context.Context) (*NodeList, error) {
	return nil, operator.ErrNotImplemented
}

// GetNamespaces returns namespace information.
func (o *StubOperator) GetNamespaces(_ context.Context) (*NamespaceList, error) {
	return nil, operator.ErrNotImplemented
}

// GetEvents returns cluster events.
func (o *StubOperator) GetEvents(_ context.Context, _ string, _ int) (*EventList, error) {
	return nil, operator.ErrNotImplemented
}

// GetPods returns pod information.
func (o *StubOperator) GetPods(_ context.Context, _ string) (*PodList, error) {
	return nil, operator.ErrNotImplemented
}

// GetDeployments returns deployment information.
func (o *StubOperator) GetDeployments(_ context.Context, _ string) (*DeploymentList, error) {
	return nil, operator.ErrNotImplemented
}

// GetServices returns service information.
func (o *StubOperator) GetServices(_ context.Context, _ string) (*ServiceList, error) {
	return nil, operator.ErrNotImplemented
}

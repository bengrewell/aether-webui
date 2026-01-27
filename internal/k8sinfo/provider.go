package k8sinfo

import "context"

// Provider defines the interface for retrieving Kubernetes cluster information.
type Provider interface {
	// Cluster-level information
	GetClusterHealth(ctx context.Context) (*ClusterHealth, error)
	GetNodes(ctx context.Context) (*NodeList, error)
	GetNamespaces(ctx context.Context) (*NamespaceList, error)
	GetEvents(ctx context.Context, namespace string, limit int) (*EventList, error)

	// Workload information
	GetPods(ctx context.Context, namespace string) (*PodList, error)
	GetDeployments(ctx context.Context, namespace string) (*DeploymentList, error)
	GetServices(ctx context.Context, namespace string) (*ServiceList, error)
}

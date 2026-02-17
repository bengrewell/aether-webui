package webuiapi

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/operator"
	"github.com/bengrewell/aether-webui/internal/operator/kube"
	"github.com/bengrewell/aether-webui/internal/provider"
	"github.com/danielgtaylor/huma/v2"
)

// NamespaceInput is the common input for endpoints that accept a namespace parameter.
type NamespaceInput struct {
	Node      string `query:"node" default:"local" doc:"Target node identifier. Use 'local' or empty for the local node."`
	Namespace string `query:"namespace" default:"" doc:"Filter by namespace. Empty or 'all' returns resources from all namespaces."`
}

// EventsInput extends NamespaceInput with event-specific parameters.
type EventsInput struct {
	Node      string `query:"node" default:"local" doc:"Target node identifier. Use 'local' or empty for the local node."`
	Namespace string `query:"namespace" default:"" doc:"Filter by namespace. Empty or 'all' returns events from all namespaces."`
	Limit     int    `query:"limit" default:"50" minimum:"1" maximum:"500" doc:"Maximum number of events to return."`
}

// ClusterHealthOutput is the response for GET /api/v1/kubernetes/health
type ClusterHealthOutput struct {
	Body kube.ClusterHealth
}

// NodeListOutput is the response for GET /api/v1/kubernetes/nodes
type NodeListOutput struct {
	Body kube.NodeList
}

// NamespaceListOutput is the response for GET /api/v1/kubernetes/namespaces
type NamespaceListOutput struct {
	Body kube.NamespaceList
}

// EventListOutput is the response for GET /api/v1/kubernetes/events
type EventListOutput struct {
	Body kube.EventList
}

// PodListOutput is the response for GET /api/v1/kubernetes/pods
type PodListOutput struct {
	Body kube.PodList
}

// DeploymentListOutput is the response for GET /api/v1/kubernetes/deployments
type DeploymentListOutput struct {
	Body kube.DeploymentList
}

// ServiceListOutput is the response for GET /api/v1/kubernetes/services
type ServiceListOutput struct {
	Body kube.ServiceList
}

// getKubeOperator resolves a node and returns its KubeOperator.
func getKubeOperator(resolver provider.ProviderResolver, node string) (kube.KubeOperator, error) {
	p, err := resolver.Resolve(provider.NodeID(node))
	if err != nil {
		return nil, err
	}
	kubeOp, ok := p.Operator(operator.DomainKube).(kube.KubeOperator)
	if !ok || kubeOp == nil {
		return nil, huma.Error503ServiceUnavailable("kubernetes operator not available")
	}
	return kubeOp, nil
}

// RegisterKubernetesRoutes registers Kubernetes monitoring routes.
func RegisterKubernetesRoutes(api huma.API, resolver provider.ProviderResolver) {
	// Cluster-level endpoints
	huma.Register(api, huma.Operation{
		OperationID: "get-cluster-health",
		Method:      "GET",
		Path:        "/api/v1/kubernetes/health",
		Summary:     "Get cluster health",
		Description: "Returns overall health status of the Kubernetes cluster",
		Tags:        []string{"Kubernetes"},
	}, func(ctx context.Context, input *NodeInput) (*ClusterHealthOutput, error) {
		kubeOp, err := getKubeOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		health, err := kubeOp.GetClusterHealth(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get cluster health", err)
		}
		return &ClusterHealthOutput{Body: *health}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-nodes",
		Method:      "GET",
		Path:        "/api/v1/kubernetes/nodes",
		Summary:     "List cluster nodes",
		Description: "Returns information about all nodes in the cluster",
		Tags:        []string{"Kubernetes"},
	}, func(ctx context.Context, input *NodeInput) (*NodeListOutput, error) {
		kubeOp, err := getKubeOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		nodes, err := kubeOp.GetNodes(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get nodes", err)
		}
		return &NodeListOutput{Body: *nodes}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-namespaces",
		Method:      "GET",
		Path:        "/api/v1/kubernetes/namespaces",
		Summary:     "List namespaces",
		Description: "Returns information about all namespaces in the cluster",
		Tags:        []string{"Kubernetes"},
	}, func(ctx context.Context, input *NodeInput) (*NamespaceListOutput, error) {
		kubeOp, err := getKubeOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		namespaces, err := kubeOp.GetNamespaces(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get namespaces", err)
		}
		return &NamespaceListOutput{Body: *namespaces}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-events",
		Method:      "GET",
		Path:        "/api/v1/kubernetes/events",
		Summary:     "List cluster events",
		Description: "Returns recent events from the cluster, optionally filtered by namespace",
		Tags:        []string{"Kubernetes"},
	}, func(ctx context.Context, input *EventsInput) (*EventListOutput, error) {
		kubeOp, err := getKubeOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		events, err := kubeOp.GetEvents(ctx, input.Namespace, input.Limit)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get events", err)
		}
		return &EventListOutput{Body: *events}, nil
	})

	// Workload endpoints
	huma.Register(api, huma.Operation{
		OperationID: "get-pods",
		Method:      "GET",
		Path:        "/api/v1/kubernetes/pods",
		Summary:     "List pods",
		Description: "Returns information about pods, optionally filtered by namespace",
		Tags:        []string{"Kubernetes"},
	}, func(ctx context.Context, input *NamespaceInput) (*PodListOutput, error) {
		kubeOp, err := getKubeOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		pods, err := kubeOp.GetPods(ctx, input.Namespace)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get pods", err)
		}
		return &PodListOutput{Body: *pods}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-deployments",
		Method:      "GET",
		Path:        "/api/v1/kubernetes/deployments",
		Summary:     "List deployments",
		Description: "Returns information about deployments, optionally filtered by namespace",
		Tags:        []string{"Kubernetes"},
	}, func(ctx context.Context, input *NamespaceInput) (*DeploymentListOutput, error) {
		kubeOp, err := getKubeOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		deployments, err := kubeOp.GetDeployments(ctx, input.Namespace)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get deployments", err)
		}
		return &DeploymentListOutput{Body: *deployments}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-services",
		Method:      "GET",
		Path:        "/api/v1/kubernetes/services",
		Summary:     "List services",
		Description: "Returns information about services, optionally filtered by namespace",
		Tags:        []string{"Kubernetes"},
	}, func(ctx context.Context, input *NamespaceInput) (*ServiceListOutput, error) {
		kubeOp, err := getKubeOperator(resolver, input.Node)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid node", err)
		}
		services, err := kubeOp.GetServices(ctx, input.Namespace)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get services", err)
		}
		return &ServiceListOutput{Body: *services}, nil
	})
}

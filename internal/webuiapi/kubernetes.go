package webuiapi

import (
	"context"

	"github.com/bengrewell/aether-webui/internal/k8sinfo"
	"github.com/danielgtaylor/huma/v2"
)

// NamespaceInput is the common input for endpoints that accept a namespace parameter.
type NamespaceInput struct {
	Namespace string `query:"namespace" default:"" doc:"Filter by namespace. Empty or 'all' returns resources from all namespaces."`
}

// EventsInput extends NamespaceInput with event-specific parameters.
type EventsInput struct {
	Namespace string `query:"namespace" default:"" doc:"Filter by namespace. Empty or 'all' returns events from all namespaces."`
	Limit     int    `query:"limit" default:"50" minimum:"1" maximum:"500" doc:"Maximum number of events to return."`
}

// ClusterHealthOutput is the response for GET /api/v1/kubernetes/health
type ClusterHealthOutput struct {
	Body k8sinfo.ClusterHealth
}

// NodeListOutput is the response for GET /api/v1/kubernetes/nodes
type NodeListOutput struct {
	Body k8sinfo.NodeList
}

// NamespaceListOutput is the response for GET /api/v1/kubernetes/namespaces
type NamespaceListOutput struct {
	Body k8sinfo.NamespaceList
}

// EventListOutput is the response for GET /api/v1/kubernetes/events
type EventListOutput struct {
	Body k8sinfo.EventList
}

// PodListOutput is the response for GET /api/v1/kubernetes/pods
type PodListOutput struct {
	Body k8sinfo.PodList
}

// DeploymentListOutput is the response for GET /api/v1/kubernetes/deployments
type DeploymentListOutput struct {
	Body k8sinfo.DeploymentList
}

// ServiceListOutput is the response for GET /api/v1/kubernetes/services
type ServiceListOutput struct {
	Body k8sinfo.ServiceList
}

// RegisterKubernetesRoutes registers Kubernetes monitoring routes.
func RegisterKubernetesRoutes(api huma.API, provider k8sinfo.Provider) {
	// Cluster-level endpoints
	huma.Register(api, huma.Operation{
		OperationID: "get-cluster-health",
		Method:      "GET",
		Path:        "/api/v1/kubernetes/health",
		Summary:     "Get cluster health",
		Description: "Returns overall health status of the Kubernetes cluster",
		Tags:        []string{"Kubernetes"},
	}, func(ctx context.Context, input *struct{}) (*ClusterHealthOutput, error) {
		health, err := provider.GetClusterHealth(ctx)
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
	}, func(ctx context.Context, input *struct{}) (*NodeListOutput, error) {
		nodes, err := provider.GetNodes(ctx)
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
	}, func(ctx context.Context, input *struct{}) (*NamespaceListOutput, error) {
		namespaces, err := provider.GetNamespaces(ctx)
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
		events, err := provider.GetEvents(ctx, input.Namespace, input.Limit)
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
		pods, err := provider.GetPods(ctx, input.Namespace)
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
		deployments, err := provider.GetDeployments(ctx, input.Namespace)
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
		services, err := provider.GetServices(ctx, input.Namespace)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get services", err)
		}
		return &ServiceListOutput{Body: *services}, nil
	})
}

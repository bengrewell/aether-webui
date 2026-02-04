package kube

import "time"

// Op represents a specific operation the kube operator can perform.
type Op string

// Query operations
const (
	// ClusterHealthOp queries overall cluster health.
	ClusterHealthOp Op = "cluster_health"
	// NodesOp queries cluster nodes.
	NodesOp Op = "nodes"
	// NamespacesOp queries namespaces.
	NamespacesOp Op = "namespaces"
	// EventsOp queries cluster events.
	EventsOp Op = "events"
	// PodsOp queries pods.
	PodsOp Op = "pods"
	// DeploymentsOp queries deployments.
	DeploymentsOp Op = "deployments"
	// ServicesOp queries services.
	ServicesOp Op = "services"
	// PodLogsOp queries pod logs.
	PodLogsOp Op = "pod_logs"
)

// Action operations
const (
	// ApplyManifest applies a Kubernetes manifest.
	ApplyManifest Op = "apply_manifest"
	// DeleteResource deletes a Kubernetes resource.
	DeleteResource Op = "delete_resource"
	// ScaleDeployment scales a deployment.
	ScaleDeployment Op = "scale_deployment"
	// RestartDeployment restarts a deployment.
	RestartDeployment Op = "restart_deployment"
)

// ClusterHealth represents the overall health status of a Kubernetes cluster.
type ClusterHealth struct {
	Status            string    `json:"status"` // healthy, degraded, unhealthy
	KubernetesVersion string    `json:"kubernetes_version"`
	NodeCount         int       `json:"node_count"`
	ReadyNodes        int       `json:"ready_nodes"`
	PodCount          int       `json:"pod_count"`
	RunningPods       int       `json:"running_pods"`
	FailedPods        int       `json:"failed_pods"`
	PendingPods       int       `json:"pending_pods"`
	LastChecked       time.Time `json:"last_checked"`
}

// NodeList contains information about cluster nodes.
type NodeList struct {
	Nodes []NodeInfo `json:"nodes"`
}

// NodeInfo represents a Kubernetes node.
type NodeInfo struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"` // Ready, NotReady, Unknown
	Roles            []string          `json:"roles"`
	KubeletVersion   string            `json:"kubelet_version"`
	ContainerRuntime string            `json:"container_runtime"`
	InternalIP       string            `json:"internal_ip"`
	ExternalIP       string            `json:"external_ip,omitempty"`
	OSImage          string            `json:"os_image"`
	Architecture     string            `json:"architecture"`
	CPUCapacity      string            `json:"cpu_capacity"`
	MemoryCapacity   string            `json:"memory_capacity"`
	PodCapacity      int               `json:"pod_capacity"`
	AllocatedPods    int               `json:"allocated_pods"`
	Conditions       []NodeCondition   `json:"conditions"`
	Labels           map[string]string `json:"labels,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
}

// NodeCondition represents a node's condition status.
type NodeCondition struct {
	Type    string `json:"type"`    // Ready, MemoryPressure, DiskPressure, PIDPressure, NetworkUnavailable
	Status  string `json:"status"`  // True, False, Unknown
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// NamespaceList contains information about namespaces.
type NamespaceList struct {
	Namespaces []NamespaceInfo `json:"namespaces"`
}

// NamespaceInfo represents a Kubernetes namespace.
type NamespaceInfo struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // Active, Terminating
	PodCount  int       `json:"pod_count"`
	CreatedAt time.Time `json:"created_at"`
}

// PodList contains information about pods.
type PodList struct {
	Pods []PodInfo `json:"pods"`
}

// PodInfo represents a Kubernetes pod.
type PodInfo struct {
	Name         string          `json:"name"`
	Namespace    string          `json:"namespace"`
	Status       string          `json:"status"` // Pending, Running, Succeeded, Failed, Unknown
	Phase        string          `json:"phase"`
	NodeName     string          `json:"node_name"`
	PodIP        string          `json:"pod_ip,omitempty"`
	HostIP       string          `json:"host_ip,omitempty"`
	RestartCount int             `json:"restart_count"`
	Containers   []ContainerInfo `json:"containers"`
	StartTime    *time.Time      `json:"start_time,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	Conditions   []PodCondition  `json:"conditions,omitempty"`
}

// ContainerInfo represents a container within a pod.
type ContainerInfo struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int    `json:"restart_count"`
	State        string `json:"state"` // running, waiting, terminated
	StateReason  string `json:"state_reason,omitempty"`
}

// PodCondition represents a pod's condition status.
type PodCondition struct {
	Type   string `json:"type"`   // Initialized, Ready, ContainersReady, PodScheduled
	Status string `json:"status"` // True, False, Unknown
}

// DeploymentList contains information about deployments.
type DeploymentList struct {
	Deployments []DeploymentInfo `json:"deployments"`
}

// DeploymentInfo represents a Kubernetes deployment.
type DeploymentInfo struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	Replicas          int       `json:"replicas"`
	ReadyReplicas     int       `json:"ready_replicas"`
	UpdatedReplicas   int       `json:"updated_replicas"`
	AvailableReplicas int       `json:"available_replicas"`
	Strategy          string    `json:"strategy"` // RollingUpdate, Recreate
	Status            string    `json:"status"`   // healthy, progressing, degraded
	CreatedAt         time.Time `json:"created_at"`
}

// ServiceList contains information about services.
type ServiceList struct {
	Services []ServiceInfo `json:"services"`
}

// ServiceInfo represents a Kubernetes service.
type ServiceInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"` // ClusterIP, NodePort, LoadBalancer, ExternalName
	ClusterIP  string            `json:"cluster_ip"`
	ExternalIP string            `json:"external_ip,omitempty"`
	Ports      []ServicePort     `json:"ports"`
	Selector   map[string]string `json:"selector,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// ServicePort represents a port exposed by a service.
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	TargetPort string `json:"target_port"`
	NodePort   int    `json:"node_port,omitempty"`
}

// EventList contains recent cluster events.
type EventList struct {
	Events []EventInfo `json:"events"`
}

// EventInfo represents a Kubernetes event.
type EventInfo struct {
	Type      string    `json:"type"` // Normal, Warning
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Object    string    `json:"object"` // e.g., "Pod/my-pod"
	Namespace string    `json:"namespace"`
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

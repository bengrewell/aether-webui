package k8sinfo

import (
	"context"
	"time"
)

// MockProvider returns static mock data for Kubernetes information.
type MockProvider struct{}

// NewMockProvider creates a new MockProvider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) GetClusterHealth(ctx context.Context) (*ClusterHealth, error) {
	return &ClusterHealth{
		Status:            "healthy",
		KubernetesVersion: "v1.31.2",
		NodeCount:         3,
		ReadyNodes:        3,
		PodCount:          42,
		RunningPods:       40,
		FailedPods:        0,
		PendingPods:       2,
		LastChecked:       time.Now(),
	}, nil
}

func (m *MockProvider) GetNodes(ctx context.Context) (*NodeList, error) {
	baseTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	return &NodeList{
		Nodes: []NodeInfo{
			{
				Name:             "control-plane-1",
				Status:           "Ready",
				Roles:            []string{"control-plane", "master"},
				KubeletVersion:   "v1.31.2",
				ContainerRuntime: "containerd://1.7.11",
				InternalIP:       "10.0.0.10",
				OSImage:          "Ubuntu 24.04 LTS",
				Architecture:     "amd64",
				CPUCapacity:      "8",
				MemoryCapacity:   "32Gi",
				PodCapacity:      110,
				AllocatedPods:    25,
				Conditions: []NodeCondition{
					{Type: "Ready", Status: "True"},
					{Type: "MemoryPressure", Status: "False"},
					{Type: "DiskPressure", Status: "False"},
					{Type: "PIDPressure", Status: "False"},
				},
				Labels: map[string]string{
					"node-role.kubernetes.io/control-plane": "",
					"topology.kubernetes.io/zone":           "us-west-2a",
				},
				CreatedAt: baseTime,
			},
			{
				Name:             "worker-1",
				Status:           "Ready",
				Roles:            []string{"worker"},
				KubeletVersion:   "v1.31.2",
				ContainerRuntime: "containerd://1.7.11",
				InternalIP:       "10.0.0.11",
				OSImage:          "Ubuntu 24.04 LTS",
				Architecture:     "amd64",
				CPUCapacity:      "16",
				MemoryCapacity:   "64Gi",
				PodCapacity:      110,
				AllocatedPods:    45,
				Conditions: []NodeCondition{
					{Type: "Ready", Status: "True"},
					{Type: "MemoryPressure", Status: "False"},
					{Type: "DiskPressure", Status: "False"},
					{Type: "PIDPressure", Status: "False"},
				},
				Labels: map[string]string{
					"node-role.kubernetes.io/worker": "",
					"topology.kubernetes.io/zone":    "us-west-2a",
				},
				CreatedAt: baseTime,
			},
			{
				Name:             "worker-2",
				Status:           "Ready",
				Roles:            []string{"worker"},
				KubeletVersion:   "v1.31.2",
				ContainerRuntime: "containerd://1.7.11",
				InternalIP:       "10.0.0.12",
				OSImage:          "Ubuntu 24.04 LTS",
				Architecture:     "amd64",
				CPUCapacity:      "16",
				MemoryCapacity:   "64Gi",
				PodCapacity:      110,
				AllocatedPods:    38,
				Conditions: []NodeCondition{
					{Type: "Ready", Status: "True"},
					{Type: "MemoryPressure", Status: "False"},
					{Type: "DiskPressure", Status: "False"},
					{Type: "PIDPressure", Status: "False"},
				},
				Labels: map[string]string{
					"node-role.kubernetes.io/worker": "",
					"topology.kubernetes.io/zone":    "us-west-2b",
				},
				CreatedAt: baseTime,
			},
		},
	}, nil
}

func (m *MockProvider) GetNamespaces(ctx context.Context) (*NamespaceList, error) {
	baseTime := time.Now().Add(-30 * 24 * time.Hour)
	return &NamespaceList{
		Namespaces: []NamespaceInfo{
			{Name: "default", Status: "Active", PodCount: 2, CreatedAt: baseTime},
			{Name: "kube-system", Status: "Active", PodCount: 12, CreatedAt: baseTime},
			{Name: "kube-public", Status: "Active", PodCount: 0, CreatedAt: baseTime},
			{Name: "kube-node-lease", Status: "Active", PodCount: 0, CreatedAt: baseTime},
			{Name: "sdcore", Status: "Active", PodCount: 18, CreatedAt: baseTime.Add(7 * 24 * time.Hour)},
			{Name: "monitoring", Status: "Active", PodCount: 6, CreatedAt: baseTime.Add(7 * 24 * time.Hour)},
			{Name: "ingress-nginx", Status: "Active", PodCount: 4, CreatedAt: baseTime.Add(7 * 24 * time.Hour)},
		},
	}, nil
}

func (m *MockProvider) GetEvents(ctx context.Context, namespace string, limit int) (*EventList, error) {
	now := time.Now()
	events := []EventInfo{
		{
			Type:      "Normal",
			Reason:    "Scheduled",
			Message:   "Successfully assigned sdcore/amf-0 to worker-1",
			Object:    "Pod/amf-0",
			Namespace: "sdcore",
			Count:     1,
			FirstSeen: now.Add(-10 * time.Minute),
			LastSeen:  now.Add(-10 * time.Minute),
		},
		{
			Type:      "Normal",
			Reason:    "Pulled",
			Message:   "Container image \"registry.opennetworking.org/sdcore/amf:v1.4.0\" already present on machine",
			Object:    "Pod/amf-0",
			Namespace: "sdcore",
			Count:     1,
			FirstSeen: now.Add(-9 * time.Minute),
			LastSeen:  now.Add(-9 * time.Minute),
		},
		{
			Type:      "Normal",
			Reason:    "Started",
			Message:   "Started container amf",
			Object:    "Pod/amf-0",
			Namespace: "sdcore",
			Count:     1,
			FirstSeen: now.Add(-9 * time.Minute),
			LastSeen:  now.Add(-9 * time.Minute),
		},
		{
			Type:      "Warning",
			Reason:    "BackOff",
			Message:   "Back-off restarting failed container",
			Object:    "Pod/test-pod-123",
			Namespace: "default",
			Count:     5,
			FirstSeen: now.Add(-1 * time.Hour),
			LastSeen:  now.Add(-5 * time.Minute),
		},
	}

	// Filter by namespace if specified
	if namespace != "" && namespace != "all" {
		filtered := make([]EventInfo, 0)
		for _, e := range events {
			if e.Namespace == namespace {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}

	// Apply limit
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return &EventList{Events: events}, nil
}

func (m *MockProvider) GetPods(ctx context.Context, namespace string) (*PodList, error) {
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)

	allPods := []PodInfo{
		// SD-Core pods
		{
			Name:         "amf-0",
			Namespace:    "sdcore",
			Status:       "Running",
			Phase:        "Running",
			NodeName:     "worker-1",
			PodIP:        "10.244.1.15",
			HostIP:       "10.0.0.11",
			RestartCount: 0,
			Containers: []ContainerInfo{
				{Name: "amf", Image: "registry.opennetworking.org/sdcore/amf:v1.4.0", Ready: true, RestartCount: 0, State: "running"},
			},
			StartTime: &startTime,
			CreatedAt: startTime,
			Conditions: []PodCondition{
				{Type: "Ready", Status: "True"},
				{Type: "ContainersReady", Status: "True"},
			},
		},
		{
			Name:         "smf-0",
			Namespace:    "sdcore",
			Status:       "Running",
			Phase:        "Running",
			NodeName:     "worker-1",
			PodIP:        "10.244.1.16",
			HostIP:       "10.0.0.11",
			RestartCount: 0,
			Containers: []ContainerInfo{
				{Name: "smf", Image: "registry.opennetworking.org/sdcore/smf:v1.4.0", Ready: true, RestartCount: 0, State: "running"},
			},
			StartTime: &startTime,
			CreatedAt: startTime,
			Conditions: []PodCondition{
				{Type: "Ready", Status: "True"},
				{Type: "ContainersReady", Status: "True"},
			},
		},
		{
			Name:         "upf-0",
			Namespace:    "sdcore",
			Status:       "Running",
			Phase:        "Running",
			NodeName:     "worker-2",
			PodIP:        "10.244.2.20",
			HostIP:       "10.0.0.12",
			RestartCount: 1,
			Containers: []ContainerInfo{
				{Name: "upf", Image: "registry.opennetworking.org/sdcore/upf:v1.4.0", Ready: true, RestartCount: 1, State: "running"},
			},
			StartTime: &startTime,
			CreatedAt: startTime,
			Conditions: []PodCondition{
				{Type: "Ready", Status: "True"},
				{Type: "ContainersReady", Status: "True"},
			},
		},
		{
			Name:         "mongodb-0",
			Namespace:    "sdcore",
			Status:       "Running",
			Phase:        "Running",
			NodeName:     "worker-1",
			PodIP:        "10.244.1.10",
			HostIP:       "10.0.0.11",
			RestartCount: 0,
			Containers: []ContainerInfo{
				{Name: "mongodb", Image: "mongo:6.0", Ready: true, RestartCount: 0, State: "running"},
			},
			StartTime: &startTime,
			CreatedAt: startTime,
			Conditions: []PodCondition{
				{Type: "Ready", Status: "True"},
				{Type: "ContainersReady", Status: "True"},
			},
		},
		// kube-system pods
		{
			Name:         "coredns-5dd5756b68-abc12",
			Namespace:    "kube-system",
			Status:       "Running",
			Phase:        "Running",
			NodeName:     "control-plane-1",
			PodIP:        "10.244.0.5",
			HostIP:       "10.0.0.10",
			RestartCount: 0,
			Containers: []ContainerInfo{
				{Name: "coredns", Image: "registry.k8s.io/coredns/coredns:v1.11.1", Ready: true, RestartCount: 0, State: "running"},
			},
			StartTime: &startTime,
			CreatedAt: startTime,
			Conditions: []PodCondition{
				{Type: "Ready", Status: "True"},
			},
		},
		// Monitoring pods
		{
			Name:         "prometheus-0",
			Namespace:    "monitoring",
			Status:       "Running",
			Phase:        "Running",
			NodeName:     "worker-2",
			PodIP:        "10.244.2.30",
			HostIP:       "10.0.0.12",
			RestartCount: 0,
			Containers: []ContainerInfo{
				{Name: "prometheus", Image: "prom/prometheus:v2.48.0", Ready: true, RestartCount: 0, State: "running"},
			},
			StartTime: &startTime,
			CreatedAt: startTime,
			Conditions: []PodCondition{
				{Type: "Ready", Status: "True"},
			},
		},
		// A pending pod example
		{
			Name:         "new-workload-xyz",
			Namespace:    "default",
			Status:       "Pending",
			Phase:        "Pending",
			NodeName:     "",
			RestartCount: 0,
			Containers: []ContainerInfo{
				{Name: "app", Image: "nginx:latest", Ready: false, RestartCount: 0, State: "waiting", StateReason: "ContainerCreating"},
			},
			CreatedAt: now.Add(-2 * time.Minute),
			Conditions: []PodCondition{
				{Type: "PodScheduled", Status: "True"},
				{Type: "Ready", Status: "False"},
			},
		},
	}

	// Filter by namespace
	if namespace != "" && namespace != "all" {
		filtered := make([]PodInfo, 0)
		for _, p := range allPods {
			if p.Namespace == namespace {
				filtered = append(filtered, p)
			}
		}
		return &PodList{Pods: filtered}, nil
	}

	return &PodList{Pods: allPods}, nil
}

func (m *MockProvider) GetDeployments(ctx context.Context, namespace string) (*DeploymentList, error) {
	baseTime := time.Now().Add(-7 * 24 * time.Hour)

	allDeployments := []DeploymentInfo{
		{
			Name:              "amf",
			Namespace:         "sdcore",
			Replicas:          1,
			ReadyReplicas:     1,
			UpdatedReplicas:   1,
			AvailableReplicas: 1,
			Strategy:          "RollingUpdate",
			Status:            "healthy",
			CreatedAt:         baseTime,
		},
		{
			Name:              "smf",
			Namespace:         "sdcore",
			Replicas:          1,
			ReadyReplicas:     1,
			UpdatedReplicas:   1,
			AvailableReplicas: 1,
			Strategy:          "RollingUpdate",
			Status:            "healthy",
			CreatedAt:         baseTime,
		},
		{
			Name:              "webui",
			Namespace:         "sdcore",
			Replicas:          2,
			ReadyReplicas:     2,
			UpdatedReplicas:   2,
			AvailableReplicas: 2,
			Strategy:          "RollingUpdate",
			Status:            "healthy",
			CreatedAt:         baseTime,
		},
		{
			Name:              "coredns",
			Namespace:         "kube-system",
			Replicas:          2,
			ReadyReplicas:     2,
			UpdatedReplicas:   2,
			AvailableReplicas: 2,
			Strategy:          "RollingUpdate",
			Status:            "healthy",
			CreatedAt:         baseTime.Add(-23 * 24 * time.Hour),
		},
		{
			Name:              "prometheus",
			Namespace:         "monitoring",
			Replicas:          1,
			ReadyReplicas:     1,
			UpdatedReplicas:   1,
			AvailableReplicas: 1,
			Strategy:          "Recreate",
			Status:            "healthy",
			CreatedAt:         baseTime,
		},
		{
			Name:              "grafana",
			Namespace:         "monitoring",
			Replicas:          1,
			ReadyReplicas:     1,
			UpdatedReplicas:   1,
			AvailableReplicas: 1,
			Strategy:          "RollingUpdate",
			Status:            "healthy",
			CreatedAt:         baseTime,
		},
	}

	if namespace != "" && namespace != "all" {
		filtered := make([]DeploymentInfo, 0)
		for _, d := range allDeployments {
			if d.Namespace == namespace {
				filtered = append(filtered, d)
			}
		}
		return &DeploymentList{Deployments: filtered}, nil
	}

	return &DeploymentList{Deployments: allDeployments}, nil
}

func (m *MockProvider) GetServices(ctx context.Context, namespace string) (*ServiceList, error) {
	baseTime := time.Now().Add(-7 * 24 * time.Hour)

	allServices := []ServiceInfo{
		{
			Name:      "amf",
			Namespace: "sdcore",
			Type:      "ClusterIP",
			ClusterIP: "10.96.100.10",
			Ports: []ServicePort{
				{Name: "sbi", Protocol: "TCP", Port: 8080, TargetPort: "8080"},
				{Name: "ngap", Protocol: "SCTP", Port: 38412, TargetPort: "38412"},
			},
			Selector:  map[string]string{"app": "amf"},
			CreatedAt: baseTime,
		},
		{
			Name:      "smf",
			Namespace: "sdcore",
			Type:      "ClusterIP",
			ClusterIP: "10.96.100.11",
			Ports: []ServicePort{
				{Name: "sbi", Protocol: "TCP", Port: 8080, TargetPort: "8080"},
				{Name: "pfcp", Protocol: "UDP", Port: 8805, TargetPort: "8805"},
			},
			Selector:  map[string]string{"app": "smf"},
			CreatedAt: baseTime,
		},
		{
			Name:       "webui",
			Namespace:  "sdcore",
			Type:       "LoadBalancer",
			ClusterIP:  "10.96.100.50",
			ExternalIP: "192.168.1.200",
			Ports: []ServicePort{
				{Name: "http", Protocol: "TCP", Port: 80, TargetPort: "5000", NodePort: 30080},
			},
			Selector:  map[string]string{"app": "webui"},
			CreatedAt: baseTime,
		},
		{
			Name:      "kube-dns",
			Namespace: "kube-system",
			Type:      "ClusterIP",
			ClusterIP: "10.96.0.10",
			Ports: []ServicePort{
				{Name: "dns", Protocol: "UDP", Port: 53, TargetPort: "53"},
				{Name: "dns-tcp", Protocol: "TCP", Port: 53, TargetPort: "53"},
			},
			Selector:  map[string]string{"k8s-app": "kube-dns"},
			CreatedAt: baseTime.Add(-23 * 24 * time.Hour),
		},
		{
			Name:      "prometheus",
			Namespace: "monitoring",
			Type:      "ClusterIP",
			ClusterIP: "10.96.200.10",
			Ports: []ServicePort{
				{Name: "http", Protocol: "TCP", Port: 9090, TargetPort: "9090"},
			},
			Selector:  map[string]string{"app": "prometheus"},
			CreatedAt: baseTime,
		},
		{
			Name:       "grafana",
			Namespace:  "monitoring",
			Type:       "NodePort",
			ClusterIP:  "10.96.200.20",
			Ports: []ServicePort{
				{Name: "http", Protocol: "TCP", Port: 3000, TargetPort: "3000", NodePort: 30300},
			},
			Selector:  map[string]string{"app": "grafana"},
			CreatedAt: baseTime,
		},
	}

	if namespace != "" && namespace != "all" {
		filtered := make([]ServiceInfo, 0)
		for _, s := range allServices {
			if s.Namespace == namespace {
				filtered = append(filtered, s)
			}
		}
		return &ServiceList{Services: filtered}, nil
	}

	return &ServiceList{Services: allServices}, nil
}

// Ensure MockProvider implements Provider
var _ Provider = (*MockProvider)(nil)

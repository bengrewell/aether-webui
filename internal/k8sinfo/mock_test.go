package k8sinfo

import (
	"context"
	"testing"
)

func TestNewMockProvider(t *testing.T) {
	provider := NewMockProvider()
	if provider == nil {
		t.Fatal("NewMockProvider returned nil")
	}
}

func TestMockProvider_GetClusterHealth(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	health, err := provider.GetClusterHealth(ctx)
	if err != nil {
		t.Fatalf("GetClusterHealth returned error: %v", err)
	}
	if health == nil {
		t.Fatal("GetClusterHealth returned nil")
	}

	// Verify required fields
	validStatuses := map[string]bool{"healthy": true, "degraded": true, "unhealthy": true}
	if !validStatuses[health.Status] {
		t.Errorf("ClusterHealth.Status has invalid value: %s", health.Status)
	}
	if health.KubernetesVersion == "" {
		t.Error("ClusterHealth.KubernetesVersion is empty")
	}
	if health.NodeCount <= 0 {
		t.Errorf("ClusterHealth.NodeCount should be positive, got %d", health.NodeCount)
	}
	if health.ReadyNodes < 0 || health.ReadyNodes > health.NodeCount {
		t.Errorf("ClusterHealth.ReadyNodes out of range: %d (NodeCount: %d)", health.ReadyNodes, health.NodeCount)
	}
	if health.LastChecked.IsZero() {
		t.Error("ClusterHealth.LastChecked is zero")
	}
}

func TestMockProvider_GetNodes(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	nodes, err := provider.GetNodes(ctx)
	if err != nil {
		t.Fatalf("GetNodes returned error: %v", err)
	}
	if nodes == nil {
		t.Fatal("GetNodes returned nil")
	}
	if len(nodes.Nodes) == 0 {
		t.Error("NodeList.Nodes is empty")
	}

	for i, node := range nodes.Nodes {
		if node.Name == "" {
			t.Errorf("Node[%d].Name is empty", i)
		}
		validStatuses := map[string]bool{"Ready": true, "NotReady": true, "Unknown": true}
		if !validStatuses[node.Status] {
			t.Errorf("Node[%d].Status has invalid value: %s", i, node.Status)
		}
		if node.KubeletVersion == "" {
			t.Errorf("Node[%d].KubeletVersion is empty", i)
		}
		if node.InternalIP == "" {
			t.Errorf("Node[%d].InternalIP is empty", i)
		}
		if node.PodCapacity <= 0 {
			t.Errorf("Node[%d].PodCapacity should be positive, got %d", i, node.PodCapacity)
		}
		if len(node.Conditions) == 0 {
			t.Errorf("Node[%d].Conditions is empty", i)
		}
	}
}

func TestMockProvider_GetNamespaces(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	namespaces, err := provider.GetNamespaces(ctx)
	if err != nil {
		t.Fatalf("GetNamespaces returned error: %v", err)
	}
	if namespaces == nil {
		t.Fatal("GetNamespaces returned nil")
	}
	if len(namespaces.Namespaces) == 0 {
		t.Error("NamespaceList.Namespaces is empty")
	}

	// Verify standard namespaces exist
	foundDefault := false
	foundKubeSystem := false
	for _, ns := range namespaces.Namespaces {
		if ns.Name == "" {
			t.Error("Namespace.Name is empty")
		}
		if ns.Name == "default" {
			foundDefault = true
		}
		if ns.Name == "kube-system" {
			foundKubeSystem = true
		}
		validStatuses := map[string]bool{"Active": true, "Terminating": true}
		if !validStatuses[ns.Status] {
			t.Errorf("Namespace %s has invalid status: %s", ns.Name, ns.Status)
		}
	}
	if !foundDefault {
		t.Error("default namespace not found")
	}
	if !foundKubeSystem {
		t.Error("kube-system namespace not found")
	}
}

func TestMockProvider_GetEvents(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	t.Run("all namespaces", func(t *testing.T) {
		events, err := provider.GetEvents(ctx, "", 50)
		if err != nil {
			t.Fatalf("GetEvents returned error: %v", err)
		}
		if events == nil {
			t.Fatal("GetEvents returned nil")
		}
		if len(events.Events) == 0 {
			t.Error("EventList.Events is empty")
		}

		for i, event := range events.Events {
			if event.Type != "Normal" && event.Type != "Warning" {
				t.Errorf("Event[%d].Type has invalid value: %s", i, event.Type)
			}
			if event.Reason == "" {
				t.Errorf("Event[%d].Reason is empty", i)
			}
			if event.Message == "" {
				t.Errorf("Event[%d].Message is empty", i)
			}
			if event.Object == "" {
				t.Errorf("Event[%d].Object is empty", i)
			}
		}
	})

	t.Run("filtered by namespace", func(t *testing.T) {
		events, err := provider.GetEvents(ctx, "sdcore", 50)
		if err != nil {
			t.Fatalf("GetEvents returned error: %v", err)
		}
		for _, event := range events.Events {
			if event.Namespace != "sdcore" {
				t.Errorf("Event should be in sdcore namespace, got %s", event.Namespace)
			}
		}
	})

	t.Run("with limit", func(t *testing.T) {
		events, err := provider.GetEvents(ctx, "", 1)
		if err != nil {
			t.Fatalf("GetEvents returned error: %v", err)
		}
		if len(events.Events) > 1 {
			t.Errorf("Expected at most 1 event, got %d", len(events.Events))
		}
	})
}

func TestMockProvider_GetPods(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	t.Run("all namespaces", func(t *testing.T) {
		pods, err := provider.GetPods(ctx, "")
		if err != nil {
			t.Fatalf("GetPods returned error: %v", err)
		}
		if pods == nil {
			t.Fatal("GetPods returned nil")
		}
		if len(pods.Pods) == 0 {
			t.Error("PodList.Pods is empty")
		}

		for i, pod := range pods.Pods {
			if pod.Name == "" {
				t.Errorf("Pod[%d].Name is empty", i)
			}
			if pod.Namespace == "" {
				t.Errorf("Pod[%d].Namespace is empty", i)
			}
			validPhases := map[string]bool{
				"Pending": true, "Running": true, "Succeeded": true,
				"Failed": true, "Unknown": true,
			}
			if !validPhases[pod.Phase] {
				t.Errorf("Pod[%d].Phase has invalid value: %s", i, pod.Phase)
			}
			if len(pod.Containers) == 0 {
				t.Errorf("Pod[%d].Containers is empty", i)
			}
		}
	})

	t.Run("filtered by namespace", func(t *testing.T) {
		pods, err := provider.GetPods(ctx, "sdcore")
		if err != nil {
			t.Fatalf("GetPods returned error: %v", err)
		}
		if len(pods.Pods) == 0 {
			t.Error("Expected pods in sdcore namespace")
		}
		for _, pod := range pods.Pods {
			if pod.Namespace != "sdcore" {
				t.Errorf("Pod should be in sdcore namespace, got %s", pod.Namespace)
			}
		}
	})

	t.Run("non-existent namespace", func(t *testing.T) {
		pods, err := provider.GetPods(ctx, "non-existent")
		if err != nil {
			t.Fatalf("GetPods returned error: %v", err)
		}
		if len(pods.Pods) != 0 {
			t.Errorf("Expected no pods in non-existent namespace, got %d", len(pods.Pods))
		}
	})
}

func TestMockProvider_GetDeployments(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	t.Run("all namespaces", func(t *testing.T) {
		deployments, err := provider.GetDeployments(ctx, "")
		if err != nil {
			t.Fatalf("GetDeployments returned error: %v", err)
		}
		if deployments == nil {
			t.Fatal("GetDeployments returned nil")
		}
		if len(deployments.Deployments) == 0 {
			t.Error("DeploymentList.Deployments is empty")
		}

		for i, dep := range deployments.Deployments {
			if dep.Name == "" {
				t.Errorf("Deployment[%d].Name is empty", i)
			}
			if dep.Namespace == "" {
				t.Errorf("Deployment[%d].Namespace is empty", i)
			}
			if dep.Replicas < 0 {
				t.Errorf("Deployment[%d].Replicas is negative", i)
			}
			validStrategies := map[string]bool{"RollingUpdate": true, "Recreate": true}
			if !validStrategies[dep.Strategy] {
				t.Errorf("Deployment[%d].Strategy has invalid value: %s", i, dep.Strategy)
			}
		}
	})

	t.Run("filtered by namespace", func(t *testing.T) {
		deployments, err := provider.GetDeployments(ctx, "monitoring")
		if err != nil {
			t.Fatalf("GetDeployments returned error: %v", err)
		}
		for _, dep := range deployments.Deployments {
			if dep.Namespace != "monitoring" {
				t.Errorf("Deployment should be in monitoring namespace, got %s", dep.Namespace)
			}
		}
	})
}

func TestMockProvider_GetServices(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	t.Run("all namespaces", func(t *testing.T) {
		services, err := provider.GetServices(ctx, "")
		if err != nil {
			t.Fatalf("GetServices returned error: %v", err)
		}
		if services == nil {
			t.Fatal("GetServices returned nil")
		}
		if len(services.Services) == 0 {
			t.Error("ServiceList.Services is empty")
		}

		for i, svc := range services.Services {
			if svc.Name == "" {
				t.Errorf("Service[%d].Name is empty", i)
			}
			if svc.Namespace == "" {
				t.Errorf("Service[%d].Namespace is empty", i)
			}
			validTypes := map[string]bool{
				"ClusterIP": true, "NodePort": true,
				"LoadBalancer": true, "ExternalName": true,
			}
			if !validTypes[svc.Type] {
				t.Errorf("Service[%d].Type has invalid value: %s", i, svc.Type)
			}
			if len(svc.Ports) == 0 {
				t.Errorf("Service[%d].Ports is empty", i)
			}
		}
	})

	t.Run("filtered by namespace", func(t *testing.T) {
		services, err := provider.GetServices(ctx, "kube-system")
		if err != nil {
			t.Fatalf("GetServices returned error: %v", err)
		}
		for _, svc := range services.Services {
			if svc.Namespace != "kube-system" {
				t.Errorf("Service should be in kube-system namespace, got %s", svc.Namespace)
			}
		}
	})
}

func TestMockProviderImplementsInterface(t *testing.T) {
	var _ Provider = (*MockProvider)(nil)
}

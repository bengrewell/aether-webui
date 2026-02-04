package webuiapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/operator/kube"
)

func TestGetClusterHealthSuccess(t *testing.T) {
	kubeOp := &mockKubeOperator{
		clusterHealth: &kube.ClusterHealth{
			Status:            "healthy",
			KubernetesVersion: "v1.28.0",
			NodeCount:         3,
			ReadyNodes:        3,
			PodCount:          50,
			RunningPods:       48,
			FailedPods:        0,
			PendingPods:       2,
			LastChecked:       time.Now(),
		},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp kube.ClusterHealth
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.Status != "healthy" {
		t.Errorf("Status = %q, want %q", resp.Status, "healthy")
	}
	if resp.NodeCount != 3 {
		t.Errorf("NodeCount = %d, want %d", resp.NodeCount, 3)
	}
}

func TestGetClusterHealthError(t *testing.T) {
	kubeOp := &mockKubeOperator{
		clusterHealthErr: errors.New("cluster connection failed"),
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetNodesSuccess(t *testing.T) {
	kubeOp := &mockKubeOperator{
		nodes: &kube.NodeList{
			Nodes: []kube.NodeInfo{
				{
					Name:           "node-1",
					Status:         "Ready",
					Roles:          []string{"control-plane", "master"},
					KubeletVersion: "v1.28.0",
					InternalIP:     "10.0.0.1",
					Architecture:   "amd64",
				},
				{
					Name:           "node-2",
					Status:         "Ready",
					Roles:          []string{"worker"},
					KubeletVersion: "v1.28.0",
					InternalIP:     "10.0.0.2",
					Architecture:   "amd64",
				},
			},
		},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/nodes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp kube.NodeList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(resp.Nodes))
	}
	if resp.Nodes[0].Name != "node-1" {
		t.Errorf("Name = %q, want %q", resp.Nodes[0].Name, "node-1")
	}
}

func TestGetNodesError(t *testing.T) {
	kubeOp := &mockKubeOperator{
		nodesErr: errors.New("failed to list nodes"),
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/nodes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetNamespacesSuccess(t *testing.T) {
	kubeOp := &mockKubeOperator{
		namespaces: &kube.NamespaceList{
			Namespaces: []kube.NamespaceInfo{
				{Name: "default", Status: "Active", PodCount: 5},
				{Name: "kube-system", Status: "Active", PodCount: 15},
				{Name: "kube-public", Status: "Active", PodCount: 0},
			},
		},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/namespaces", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp kube.NamespaceList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Namespaces) != 3 {
		t.Fatalf("expected 3 namespaces, got %d", len(resp.Namespaces))
	}
}

func TestGetNamespacesError(t *testing.T) {
	kubeOp := &mockKubeOperator{
		namespacesErr: errors.New("failed to list namespaces"),
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/namespaces", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetEventsSuccess(t *testing.T) {
	kubeOp := &mockKubeOperator{
		events: &kube.EventList{
			Events: []kube.EventInfo{
				{
					Type:      "Normal",
					Reason:    "Scheduled",
					Message:   "Successfully assigned pod",
					Object:    "Pod/test-pod",
					Namespace: "default",
					Count:     1,
				},
				{
					Type:      "Warning",
					Reason:    "BackOff",
					Message:   "Back-off restarting failed container",
					Object:    "Pod/failing-pod",
					Namespace: "default",
					Count:     5,
				},
			},
		},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp kube.EventList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(resp.Events))
	}
}

func TestGetEventsWithNamespace(t *testing.T) {
	kubeOp := &mockKubeOperator{
		events: &kube.EventList{
			Events: []kube.EventInfo{
				{Type: "Normal", Reason: "Scheduled", Namespace: "kube-system"},
			},
		},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/events?namespace=kube-system", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetEventsWithLimit(t *testing.T) {
	kubeOp := &mockKubeOperator{
		events: &kube.EventList{Events: []kube.EventInfo{}},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/events?limit=100", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetEventsError(t *testing.T) {
	kubeOp := &mockKubeOperator{
		eventsErr: errors.New("failed to list events"),
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetPodsSuccess(t *testing.T) {
	kubeOp := &mockKubeOperator{
		pods: &kube.PodList{
			Pods: []kube.PodInfo{
				{
					Name:      "nginx-abc123",
					Namespace: "default",
					Status:    "Running",
					Phase:     "Running",
					NodeName:  "node-1",
					PodIP:     "10.244.0.5",
				},
			},
		},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/pods", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp kube.PodList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Pods) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(resp.Pods))
	}
	if resp.Pods[0].Status != "Running" {
		t.Errorf("Status = %q, want %q", resp.Pods[0].Status, "Running")
	}
}

func TestGetPodsWithNamespace(t *testing.T) {
	kubeOp := &mockKubeOperator{
		pods: &kube.PodList{Pods: []kube.PodInfo{}},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/pods?namespace=kube-system", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetPodsError(t *testing.T) {
	kubeOp := &mockKubeOperator{
		podsErr: errors.New("failed to list pods"),
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/pods", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetDeploymentsSuccess(t *testing.T) {
	kubeOp := &mockKubeOperator{
		deployments: &kube.DeploymentList{
			Deployments: []kube.DeploymentInfo{
				{
					Name:              "nginx",
					Namespace:         "default",
					Replicas:          3,
					ReadyReplicas:     3,
					UpdatedReplicas:   3,
					AvailableReplicas: 3,
					Strategy:          "RollingUpdate",
					Status:            "healthy",
				},
			},
		},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/deployments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp kube.DeploymentList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Deployments) != 1 {
		t.Fatalf("expected 1 deployment, got %d", len(resp.Deployments))
	}
	if resp.Deployments[0].Name != "nginx" {
		t.Errorf("Name = %q, want %q", resp.Deployments[0].Name, "nginx")
	}
}

func TestGetDeploymentsWithNamespace(t *testing.T) {
	kubeOp := &mockKubeOperator{
		deployments: &kube.DeploymentList{Deployments: []kube.DeploymentInfo{}},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/deployments?namespace=production", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetDeploymentsError(t *testing.T) {
	kubeOp := &mockKubeOperator{
		deploymentsErr: errors.New("failed to list deployments"),
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/deployments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetServicesSuccess(t *testing.T) {
	kubeOp := &mockKubeOperator{
		services: &kube.ServiceList{
			Services: []kube.ServiceInfo{
				{
					Name:      "nginx-svc",
					Namespace: "default",
					Type:      "ClusterIP",
					ClusterIP: "10.96.0.100",
					Ports: []kube.ServicePort{
						{Name: "http", Protocol: "TCP", Port: 80, TargetPort: "8080"},
					},
				},
			},
		},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/services", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp kube.ServiceList
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(resp.Services))
	}
	if resp.Services[0].Type != "ClusterIP" {
		t.Errorf("Type = %q, want %q", resp.Services[0].Type, "ClusterIP")
	}
}

func TestGetServicesWithNamespace(t *testing.T) {
	kubeOp := &mockKubeOperator{
		services: &kube.ServiceList{Services: []kube.ServiceInfo{}},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/services?namespace=monitoring", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetServicesError(t *testing.T) {
	kubeOp := &mockKubeOperator{
		servicesErr: errors.New("failed to list services"),
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/services", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestKubernetesEndpointsInvalidNode(t *testing.T) {
	kubeOp := &mockKubeOperator{
		clusterHealth: &kube.ClusterHealth{Status: "healthy"},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/health?node=unknown-node", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with invalid node, got %d", w.Code)
	}
}

func TestKubernetesEndpointsOperatorUnavailable(t *testing.T) {
	router := newKubernetesTestRouterNoOperator(t)

	endpoints := []string{
		"/api/v1/kubernetes/health",
		"/api/v1/kubernetes/nodes",
		"/api/v1/kubernetes/namespaces",
		"/api/v1/kubernetes/events",
		"/api/v1/kubernetes/pods",
		"/api/v1/kubernetes/deployments",
		"/api/v1/kubernetes/services",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestKubernetesEndpointsContentType(t *testing.T) {
	kubeOp := &mockKubeOperator{
		clusterHealth: &kube.ClusterHealth{Status: "healthy"},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func TestKubernetesEndpointsMethodNotAllowed(t *testing.T) {
	kubeOp := &mockKubeOperator{
		clusterHealth: &kube.ClusterHealth{Status: "healthy"},
	}
	router := newKubernetesTestRouter(t, kubeOp)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	endpoints := []string{
		"/api/v1/kubernetes/health",
		"/api/v1/kubernetes/nodes",
		"/api/v1/kubernetes/namespaces",
		"/api/v1/kubernetes/events",
		"/api/v1/kubernetes/pods",
		"/api/v1/kubernetes/deployments",
		"/api/v1/kubernetes/services",
	}

	for _, endpoint := range endpoints {
		for _, method := range methods {
			t.Run(method+" "+endpoint, func(t *testing.T) {
				req := httptest.NewRequest(method, endpoint, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != http.StatusMethodNotAllowed {
					t.Fatalf("expected 405, got %d", w.Code)
				}
			})
		}
	}
}

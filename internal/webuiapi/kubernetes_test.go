package webuiapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bengrewell/aether-webui/internal/k8sinfo"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func setupKubernetesTestAPI() (http.Handler, huma.API) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	mockProvider := k8sinfo.NewMockProvider()
	RegisterKubernetesRoutes(api, mockProvider)
	return router, api
}

func TestGetClusterHealth(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response k8sinfo.ClusterHealth
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status == "" {
		t.Error("Expected non-empty Status")
	}
	if response.KubernetesVersion == "" {
		t.Error("Expected non-empty KubernetesVersion")
	}
	if response.NodeCount <= 0 {
		t.Errorf("Expected positive NodeCount, got %d", response.NodeCount)
	}
}

func TestGetNodes(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/nodes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response k8sinfo.NodeList
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Nodes) == 0 {
		t.Error("Expected non-empty Nodes array")
	}
	for i, node := range response.Nodes {
		if node.Name == "" {
			t.Errorf("Node[%d].Name is empty", i)
		}
	}
}

func TestGetNamespaces(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/namespaces", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response k8sinfo.NamespaceList
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Namespaces) == 0 {
		t.Error("Expected non-empty Namespaces array")
	}

	// Check for expected namespaces
	foundDefault := false
	for _, ns := range response.Namespaces {
		if ns.Name == "default" {
			foundDefault = true
		}
	}
	if !foundDefault {
		t.Error("Expected to find 'default' namespace")
	}
}

func TestGetEvents(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

	t.Run("all events", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/events", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.EventList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Events) == 0 {
			t.Error("Expected non-empty Events array")
		}
	})

	t.Run("filtered by namespace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/events?namespace=sdcore", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.EventList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, event := range response.Events {
			if event.Namespace != "sdcore" {
				t.Errorf("Expected namespace 'sdcore', got '%s'", event.Namespace)
			}
		}
	})

	t.Run("with limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/events?limit=1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.EventList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Events) > 1 {
			t.Errorf("Expected at most 1 event, got %d", len(response.Events))
		}
	})
}

func TestGetPods(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

	t.Run("all pods", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/pods", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.PodList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Pods) == 0 {
			t.Error("Expected non-empty Pods array")
		}
	})

	t.Run("filtered by namespace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/pods?namespace=sdcore", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.PodList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, pod := range response.Pods {
			if pod.Namespace != "sdcore" {
				t.Errorf("Expected namespace 'sdcore', got '%s'", pod.Namespace)
			}
		}
	})

	t.Run("non-existent namespace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/pods?namespace=nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.PodList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Pods) != 0 {
			t.Errorf("Expected empty Pods array for non-existent namespace, got %d", len(response.Pods))
		}
	})
}

func TestGetDeployments(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

	t.Run("all deployments", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/deployments", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.DeploymentList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Deployments) == 0 {
			t.Error("Expected non-empty Deployments array")
		}
	})

	t.Run("filtered by namespace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/deployments?namespace=monitoring", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.DeploymentList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, dep := range response.Deployments {
			if dep.Namespace != "monitoring" {
				t.Errorf("Expected namespace 'monitoring', got '%s'", dep.Namespace)
			}
		}
	})
}

func TestGetServices(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

	t.Run("all services", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/services", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.ServiceList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Services) == 0 {
			t.Error("Expected non-empty Services array")
		}
	})

	t.Run("filtered by namespace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/kubernetes/services?namespace=sdcore", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response k8sinfo.ServiceList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, svc := range response.Services {
			if svc.Namespace != "sdcore" {
				t.Errorf("Expected namespace 'sdcore', got '%s'", svc.Namespace)
			}
		}
	})
}

func TestKubernetesEndpointsMethodNotAllowed(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

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
			req := httptest.NewRequest(http.MethodPost, endpoint, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for POST to %s, got %d", http.StatusMethodNotAllowed, endpoint, w.Code)
			}
		})
	}
}

func TestKubernetesEndpointsContentType(t *testing.T) {
	router, _ := setupKubernetesTestAPI()

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

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json' for %s, got '%s'", endpoint, contentType)
			}
		})
	}
}

package webuiapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bengrewell/aether-webui/internal/aether"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func setupAetherTestAPI() (http.Handler, aether.HostResolver) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	mockProvider := aether.NewMockProvider("local")
	resolver := aether.NewDefaultHostResolver(mockProvider)
	RegisterAetherRoutes(api, resolver)
	return router, resolver
}

func TestListHosts(t *testing.T) {
	router, _ := setupAetherTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/hosts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response struct {
		Hosts []string `json:"hosts"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Hosts) == 0 {
		t.Error("Expected at least one host")
	}
}

func TestGetCoreConfig(t *testing.T) {
	router, _ := setupAetherTestAPI()

	t.Run("default host", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/config", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response aether.CoreConfig
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Helm.ChartRef == "" {
			t.Error("Expected non-empty ChartRef")
		}
	})

	t.Run("explicit local host", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/config?host=local", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("invalid host", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/config?host=unknown", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestUpdateCoreConfig(t *testing.T) {
	router, _ := setupAetherTestAPI()

	config := aether.CoreConfig{
		Standalone: true,
		DataIface:  "eth0",
		Helm: aether.HelmConfig{
			ChartRef:     "custom/chart",
			ChartVersion: "2.0.0",
		},
	}
	body, _ := json.Marshal(config)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/aether/core/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetCoreStatus(t *testing.T) {
	router, _ := setupAetherTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/core/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response aether.CoreStatus
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.State == "" {
		t.Error("Expected non-empty State")
	}
	if response.Host == "" {
		t.Error("Expected non-empty Host")
	}
}

func TestDeployCore(t *testing.T) {
	router, _ := setupAetherTestAPI()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/aether/core/deploy", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response aether.DeploymentResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected Success to be true")
	}
}

func TestUndeployCore(t *testing.T) {
	router, _ := setupAetherTestAPI()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aether/core", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response aether.DeploymentResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected Success to be true")
	}
}

func TestListGNBs(t *testing.T) {
	router, _ := setupAetherTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response aether.GNBList
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.GNBs) == 0 {
		t.Error("Expected at least one gNB")
	}
}

func TestGetGNB(t *testing.T) {
	router, _ := setupAetherTestAPI()

	t.Run("existing gNB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/gnb-0", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response aether.GNBConfig
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.ID != "gnb-0" {
			t.Errorf("Expected ID 'gnb-0', got '%s'", response.ID)
		}
	})

	t.Run("non-existent gNB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/non-existent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestCreateGNB(t *testing.T) {
	router, _ := setupAetherTestAPI()

	gnb := aether.GNBConfig{
		ID:         "gnb-new",
		Name:       "New gNB",
		Type:       "srsran",
		IP:         "10.0.0.100",
		Simulation: true,
	}
	body, _ := json.Marshal(gnb)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/aether/gnb", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response aether.DeploymentResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected Success to be true")
	}
}

func TestUpdateGNB(t *testing.T) {
	router, _ := setupAetherTestAPI()

	gnb := aether.GNBConfig{
		Name:       "Updated gNB",
		Type:       "srsran",
		IP:         "10.0.0.200",
		Simulation: false,
	}
	body, _ := json.Marshal(gnb)

	t.Run("existing gNB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/aether/gnb/gnb-0", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("non-existent gNB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/aether/gnb/non-existent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestDeleteGNB(t *testing.T) {
	router, _ := setupAetherTestAPI()

	t.Run("existing gNB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/aether/gnb/gnb-0", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("non-existent gNB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/aether/gnb/non-existent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestGetGNBStatus(t *testing.T) {
	router, _ := setupAetherTestAPI()

	t.Run("existing gNB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/gnb-0/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response aether.GNBStatus
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.ID != "gnb-0" {
			t.Errorf("Expected ID 'gnb-0', got '%s'", response.ID)
		}
	})

	t.Run("non-existent gNB", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/non-existent/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestListGNBStatuses(t *testing.T) {
	router, _ := setupAetherTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/aether/gnb/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response aether.GNBStatusList
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func TestAetherEndpointsContentType(t *testing.T) {
	router, _ := setupAetherTestAPI()

	endpoints := []string{
		"/api/v1/aether/hosts",
		"/api/v1/aether/core/config",
		"/api/v1/aether/core/status",
		"/api/v1/aether/gnb",
		"/api/v1/aether/gnb/gnb-0",
		"/api/v1/aether/gnb/status",
		"/api/v1/aether/gnb/gnb-0/status",
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

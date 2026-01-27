package webuiapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bengrewell/aether-webui/internal/sysinfo"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func setupSystemTestAPI() (http.Handler, huma.API) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	mockProvider := sysinfo.NewMockProvider()
	resolver := sysinfo.NewDefaultNodeResolver(mockProvider)
	RegisterSystemRoutes(api, resolver)
	return router, api
}

func TestGetCPUInfo(t *testing.T) {
	router, _ := setupSystemTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/cpu", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.CPUInfo
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Model == "" {
		t.Error("Expected non-empty Model")
	}
	if response.Cores <= 0 {
		t.Errorf("Expected positive Cores, got %d", response.Cores)
	}
}

func TestGetCPUInfoWithNodeParam(t *testing.T) {
	router, _ := setupSystemTestAPI()

	t.Run("local node", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/cpu?node=local", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("invalid node", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/cpu?node=remote-node", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestGetMemoryInfo(t *testing.T) {
	router, _ := setupSystemTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/memory", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.MemoryInfo
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.TotalBytes == 0 {
		t.Error("Expected non-zero TotalBytes")
	}
	if response.Type == "" {
		t.Error("Expected non-empty Type")
	}
}

func TestGetDiskInfo(t *testing.T) {
	router, _ := setupSystemTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/disk", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.DiskInfo
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Disks) == 0 {
		t.Error("Expected non-empty Disks array")
	}
}

func TestGetNICInfo(t *testing.T) {
	router, _ := setupSystemTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/nic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.NICInfo
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Interfaces) == 0 {
		t.Error("Expected non-empty Interfaces array")
	}
}

func TestGetOSInfo(t *testing.T) {
	router, _ := setupSystemTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/os", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.OSInfo
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name == "" {
		t.Error("Expected non-empty Name")
	}
	if response.Kernel == "" {
		t.Error("Expected non-empty Kernel")
	}
	if response.Hostname == "" {
		t.Error("Expected non-empty Hostname")
	}
}

func TestSystemEndpointsMethodNotAllowed(t *testing.T) {
	router, _ := setupSystemTestAPI()

	endpoints := []string{
		"/api/v1/system/cpu",
		"/api/v1/system/memory",
		"/api/v1/system/disk",
		"/api/v1/system/nic",
		"/api/v1/system/os",
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

func TestSystemEndpointsContentType(t *testing.T) {
	router, _ := setupSystemTestAPI()

	endpoints := []string{
		"/api/v1/system/cpu",
		"/api/v1/system/memory",
		"/api/v1/system/disk",
		"/api/v1/system/nic",
		"/api/v1/system/os",
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

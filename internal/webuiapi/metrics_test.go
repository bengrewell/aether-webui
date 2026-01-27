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

func setupMetricsTestAPI() (http.Handler, huma.API) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	mockProvider := sysinfo.NewMockProvider()
	resolver := sysinfo.NewDefaultNodeResolver(mockProvider)
	RegisterMetricsRoutes(api, resolver)
	return router, api
}

func TestGetCPUUsage(t *testing.T) {
	router, _ := setupMetricsTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.CPUUsage
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.UsagePercent < 0 || response.UsagePercent > 100 {
		t.Errorf("UsagePercent out of range: %f", response.UsagePercent)
	}
	if len(response.PerCoreUsage) == 0 {
		t.Error("Expected non-empty PerCoreUsage")
	}
}

func TestGetCPUUsageWithNodeParam(t *testing.T) {
	router, _ := setupMetricsTestAPI()

	t.Run("local node", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu?node=local", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("empty node", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu?node=", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("invalid node", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/cpu?node=remote-server", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestGetMemoryUsage(t *testing.T) {
	router, _ := setupMetricsTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/memory", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.MemoryUsage
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.UsagePercent < 0 || response.UsagePercent > 100 {
		t.Errorf("UsagePercent out of range: %f", response.UsagePercent)
	}
	if response.UsedBytes == 0 {
		t.Error("Expected non-zero UsedBytes")
	}
}

func TestGetDiskUsage(t *testing.T) {
	router, _ := setupMetricsTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/disk", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.DiskUsage
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Disks) == 0 {
		t.Error("Expected non-empty Disks array")
	}
	for i, disk := range response.Disks {
		if disk.UsagePercent < 0 || disk.UsagePercent > 100 {
			t.Errorf("Disk[%d].UsagePercent out of range: %f", i, disk.UsagePercent)
		}
	}
}

func TestGetNICUsage(t *testing.T) {
	router, _ := setupMetricsTestAPI()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/nic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response sysinfo.NICUsage
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Interfaces) == 0 {
		t.Error("Expected non-empty Interfaces array")
	}
}

func TestMetricsEndpointsMethodNotAllowed(t *testing.T) {
	router, _ := setupMetricsTestAPI()

	endpoints := []string{
		"/api/v1/metrics/cpu",
		"/api/v1/metrics/memory",
		"/api/v1/metrics/disk",
		"/api/v1/metrics/nic",
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

func TestMetricsEndpointsContentType(t *testing.T) {
	router, _ := setupMetricsTestAPI()

	endpoints := []string{
		"/api/v1/metrics/cpu",
		"/api/v1/metrics/memory",
		"/api/v1/metrics/disk",
		"/api/v1/metrics/nic",
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

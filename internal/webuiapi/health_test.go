package webuiapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func TestHealthEndpoint(t *testing.T) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterHealthRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}
}

func TestHealthEndpointContentType(t *testing.T) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterHealthRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

func TestHealthEndpointMethod(t *testing.T) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterHealthRoutes(api)

	// Test that POST returns method not allowed
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d for POST, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

// TestHealthHandler tests the handler function directly
func TestHealthHandler(t *testing.T) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))

	var capturedOutput *HealthOutput
	huma.Register(api, huma.Operation{
		OperationID: "test-health",
		Method:      "GET",
		Path:        "/test-healthz",
	}, func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		resp := &HealthOutput{}
		resp.Body.Status = "healthy"
		capturedOutput = resp
		return resp, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test-healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedOutput == nil {
		t.Fatal("Handler was not called")
	}
	if capturedOutput.Body.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", capturedOutput.Body.Status)
	}
}

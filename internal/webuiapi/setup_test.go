package webuiapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bengrewell/aether-webui/internal/state"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// newSetupTestRouter creates a chi router with setup routes backed by a real SQLiteStore.
func newSetupTestRouter(t *testing.T) http.Handler {
	t.Helper()
	store, err := state.NewSQLiteStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterSetupRoutes(api, store)
	return router
}

func TestGetSetupStatusDefault(t *testing.T) {
	router := newSetupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Completed   bool       `json:"completed"`
		CompletedAt *time.Time `json:"completed_at"`
		Steps       []string   `json:"steps"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.Completed {
		t.Error("expected completed=false on fresh store")
	}
}

func TestCompleteSetup(t *testing.T) {
	router := newSetupTestRouter(t)

	body := `{"steps":["network","storage"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success     bool      `json:"success"`
		CompletedAt time.Time `json:"completed_at"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.CompletedAt.IsZero() {
		t.Error("expected completed_at to be set")
	}
}

func TestCompleteSetupWithoutSteps(t *testing.T) {
	router := newSetupTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestGetSetupStatusAfterComplete(t *testing.T) {
	router := newSetupTestRouter(t)

	// Complete setup with steps.
	postBody := `{"steps":["network","dns"]}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", strings.NewReader(postBody))
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)
	if postW.Code != http.StatusOK {
		t.Fatalf("POST failed: %d", postW.Code)
	}

	// GET status.
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getW.Code)
	}

	var resp struct {
		Completed   bool       `json:"completed"`
		CompletedAt *time.Time `json:"completed_at"`
		Steps       []string   `json:"steps"`
	}
	if err := json.NewDecoder(getW.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !resp.Completed {
		t.Error("expected completed=true")
	}
	if resp.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
	if len(resp.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(resp.Steps))
	}
}

func TestResetSetupStatus(t *testing.T) {
	router := newSetupTestRouter(t)

	// Complete first.
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", strings.NewReader(`{"steps":["a"]}`))
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)
	if postW.Code != http.StatusOK {
		t.Fatalf("POST failed: %d", postW.Code)
	}

	// Reset.
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/setup/status", nil)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", delW.Code)
	}

	var resp struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(delW.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestGetSetupStatusAfterReset(t *testing.T) {
	router := newSetupTestRouter(t)

	// Complete, then reset.
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/setup/complete", strings.NewReader(`{"steps":["x"]}`))
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/setup/status", nil)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)

	// Verify status is back to default.
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getW.Code)
	}

	var resp struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(getW.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.Completed {
		t.Error("expected completed=false after reset")
	}
}

func TestSetupEndpointsContentType(t *testing.T) {
	router := newSetupTestRouter(t)

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/setup/status", ""},
		{http.MethodPost, "/api/v1/setup/complete", `{}`},
		{http.MethodDelete, "/api/v1/setup/status", ""},
	}

	for _, tc := range tests {
		var req *http.Request
		if tc.body != "" {
			req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(tc.method, tc.path, nil)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("%s %s: expected Content-Type 'application/json', got %q", tc.method, tc.path, ct)
		}
	}
}

func TestSetupEndpointsMethodNotAllowed(t *testing.T) {
	router := newSetupTestRouter(t)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPut, "/api/v1/setup/status"},
		{http.MethodPatch, "/api/v1/setup/status"},
		{http.MethodPut, "/api/v1/setup/complete"},
		{http.MethodDelete, "/api/v1/setup/complete"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: expected 405, got %d", tc.method, tc.path, w.Code)
		}
	}
}

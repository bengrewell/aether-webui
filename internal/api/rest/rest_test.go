package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityScheme_Enabled(t *testing.T) {
	tr := NewTransport(Config{
		APITitle:         "Test API",
		APIVersion:       "0.0.0",
		TokenAuthEnabled: true,
	})

	spec := tr.API().OpenAPI()

	if spec.Components == nil {
		t.Fatal("expected Components to be non-nil")
	}
	scheme, ok := spec.Components.SecuritySchemes["bearerAuth"]
	if !ok {
		t.Fatal("expected bearerAuth security scheme to be defined")
	}
	if scheme.Type != "http" {
		t.Errorf("scheme type = %q, want %q", scheme.Type, "http")
	}
	if scheme.Scheme != "bearer" {
		t.Errorf("scheme scheme = %q, want %q", scheme.Scheme, "bearer")
	}

	if len(spec.Security) == 0 {
		t.Fatal("expected global security requirement to be set")
	}
	if _, ok := spec.Security[0]["bearerAuth"]; !ok {
		t.Error("expected global security to reference bearerAuth")
	}
}

func TestHandleFunc(t *testing.T) {
	tr := NewTransport(Config{
		APITitle:   "Test API",
		APIVersion: "0.0.0",
	})
	tr.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	tr.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "pong" {
		t.Errorf("body = %q, want %q", got, "pong")
	}
}

func TestSecurityScheme_Disabled(t *testing.T) {
	tr := NewTransport(Config{
		APITitle:         "Test API",
		APIVersion:       "0.0.0",
		TokenAuthEnabled: false,
	})

	spec := tr.API().OpenAPI()

	if spec.Components != nil && len(spec.Components.SecuritySchemes) > 0 {
		t.Error("expected no security schemes when token auth is disabled")
	}
	if len(spec.Security) > 0 {
		t.Error("expected no global security requirement when token auth is disabled")
	}
}

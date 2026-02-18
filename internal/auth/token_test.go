package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func TestTokenAuthNoopWhenEmpty(t *testing.T) {
	mw := TokenAuth("", nil)
	handler := mw(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTokenAuthValidToken(t *testing.T) {
	mw := TokenAuth("secret123", DefaultSkipPaths)
	handler := mw(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTokenAuthInvalidToken(t *testing.T) {
	mw := TokenAuth("secret123", DefaultSkipPaths)
	handler := mw(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer wrongtoken")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	body, _ := io.ReadAll(rec.Body)
	if len(body) == 0 {
		t.Error("expected JSON error body")
	}
}

func TestTokenAuthMissingHeader(t *testing.T) {
	mw := TokenAuth("secret123", DefaultSkipPaths)
	handler := mw(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTokenAuthCaseInsensitiveBearer(t *testing.T) {
	mw := TokenAuth("secret123", DefaultSkipPaths)
	handler := mw(okHandler())

	for _, prefix := range []string{"Bearer", "bearer", "BEARER", "bEaReR"} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		req.Header.Set("Authorization", prefix+" secret123")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("prefix %q: status = %d, want %d", prefix, rec.Code, http.StatusOK)
		}
	}
}

func TestTokenAuthInvalidFormat(t *testing.T) {
	mw := TokenAuth("secret123", DefaultSkipPaths)
	handler := mw(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestDefaultSkipPaths(t *testing.T) {
	tests := []struct {
		path string
		skip bool
	}{
		{"/healthz", true},
		{"/openapi.json", true},
		{"/docs", true},
		{"/docs/security", true},
		{"/", true},         // not under /api/
		{"/index.html", true}, // not under /api/
		{"/api/v1/test", false},
		{"/api/v1/meta/version", false},
	}

	for _, tt := range tests {
		got := DefaultSkipPaths(tt.path)
		if got != tt.skip {
			t.Errorf("DefaultSkipPaths(%q) = %v, want %v", tt.path, got, tt.skip)
		}
	}
}

func TestTokenAuthSkipPaths(t *testing.T) {
	mw := TokenAuth("secret123", DefaultSkipPaths)
	handler := mw(okHandler())

	// /healthz should be accessible without token.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d for skipped path", rec.Code, http.StatusOK)
	}
}

func TestTokenAuthCustomSkip(t *testing.T) {
	skip := func(path string) bool { return path == "/custom" }
	mw := TokenAuth("secret123", skip)
	handler := mw(okHandler())

	// /custom should be accessible without token.
	req := httptest.NewRequest(http.MethodGet, "/custom", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d for custom skipped path", rec.Code, http.StatusOK)
	}

	// /api/v1/test should require token.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d for non-skipped path", rec.Code, http.StatusUnauthorized)
	}
}

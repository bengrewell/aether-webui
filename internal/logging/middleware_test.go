package logging

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLoggerInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, NoColor: true})

	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !containsSubstring(out, "GET") {
		t.Error("expected method in log output")
	}
	if !containsSubstring(out, "/api/v1/test") {
		t.Error("expected path in log output")
	}
	if !containsSubstring(out, "200") {
		t.Error("expected status in log output")
	}
}

func TestRequestLoggerHealthzSuppressedAtInfo(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, NoColor: true})

	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if buf.Len() > 0 {
		t.Error("expected /healthz to be suppressed at INFO level")
	}
}

func TestRequestLoggerHealthzVisibleAtDebug(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Level: slog.LevelDebug, Writer: &buf, NoColor: true})

	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !containsSubstring(buf.String(), "/healthz") {
		t.Error("expected /healthz to be logged at DEBUG level")
	}
}

func TestRequestLogger5xxError(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, NoColor: true})

	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fail", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !containsSubstring(out, "500") {
		t.Error("expected 500 status in log output")
	}
	if !containsSubstring(out, "ERR") {
		t.Error("expected ERROR level for 5xx responses")
	}
}

func TestRequestLogger4xxWarn(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, NoColor: true})

	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !containsSubstring(out, "404") {
		t.Error("expected 404 status in log output")
	}
	if !containsSubstring(out, "WRN") {
		t.Error("expected WARN level for 4xx responses")
	}
}

func TestRequestLoggerDebugExtraFields(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Level: slog.LevelDebug, Writer: &buf, NoColor: true})

	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test?foo=bar", nil)
	req.Header.Set("User-Agent", "test-agent")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !containsSubstring(out, "test-agent") {
		t.Error("expected user_agent in debug output")
	}
	if !containsSubstring(out, "foo=bar") {
		t.Error("expected query in debug output")
	}
}

func TestResponseWriterUnwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	if rw.Unwrap() != rec {
		t.Error("Unwrap should return the underlying ResponseWriter")
	}
}

func TestResponseWriterCapturesBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, status: http.StatusOK}

	rw.Write([]byte("hello"))
	rw.Write([]byte(" world"))

	if rw.bytes != 11 {
		t.Errorf("expected 11 bytes written, got %d", rw.bytes)
	}
}

func TestRequestLoggerErrorBodyCapture(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, NoColor: true})

	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"something went wrong"}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fail", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !containsSubstring(out, "something went wrong") {
		t.Error("expected error body content to appear in log output")
	}
}

func TestResponseWriterErrorBodyTruncation(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	// Trigger error body capture
	rw.WriteHeader(http.StatusInternalServerError)

	// Write more than 1KB in a single call
	largeBody := make([]byte, 2048)
	for i := range largeBody {
		largeBody[i] = 'x'
	}
	rw.Write(largeBody)

	if rw.errorBody.Len() != 1024 {
		t.Errorf("expected errorBody to be truncated to 1024 bytes, got %d", rw.errorBody.Len())
	}
}

func TestResponseWriterErrorBodyMultipleWritesTruncation(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	// Trigger error body capture
	rw.WriteHeader(http.StatusInternalServerError)

	// Write multiple chunks that together exceed 1KB
	chunk := make([]byte, 512)
	for i := range chunk {
		chunk[i] = 'y'
	}

	rw.Write(chunk) // 512 bytes
	rw.Write(chunk) // 1024 bytes total
	rw.Write(chunk) // Should not add more since already at limit

	if rw.errorBody.Len() != 1024 {
		t.Errorf("expected errorBody to be truncated to 1024 bytes, got %d", rw.errorBody.Len())
	}
}

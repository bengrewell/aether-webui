package _logging

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code, bytes written,
// and optionally the response body for error responses.
type responseWriter struct {
	http.ResponseWriter
	status    int
	bytes     int
	errorBody *bytes.Buffer // only populated for 4xx/5xx responses
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	// Capture body for error responses
	if code >= 400 {
		rw.errorBody = &bytes.Buffer{}
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Capture error response body (limited to 1KB to avoid memory issues)
	if rw.errorBody != nil && rw.errorBody.Len() < 1024 {
		remaining := 1024 - rw.errorBody.Len()
		if len(b) <= remaining {
			rw.errorBody.Write(b)
		} else {
			rw.errorBody.Write(b[:remaining])
		}
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// Unwrap returns the underlying ResponseWriter for http.ResponseController compatibility.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// RequestLogger returns Chi-compatible middleware that logs each HTTP request.
//
// Log level selection:
//   - /healthz requests -> DEBUG (reduces noise at INFO)
//   - 5xx responses -> ERROR
//   - 4xx responses -> WARN
//   - Everything else -> INFO
//
// When the default logger is at DEBUG level, extra fields (user_agent, proto,
// query, content_type) are included in each log line.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			level := requestLevel(r.URL.Path, rw.status)

			attrs := []slog.Attr{
				slog.String("component", "http"),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.String("duration", duration.Round(100*time.Microsecond).String()),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("bytes", rw.bytes),
			}

			// Include error body for 4xx/5xx responses
			if rw.errorBody != nil && rw.errorBody.Len() > 0 {
				attrs = append(attrs, slog.String("error", rw.errorBody.String()))
			}

			if slog.Default().Enabled(r.Context(), slog.LevelDebug) {
				attrs = append(attrs,
					slog.String("user_agent", r.UserAgent()),
					slog.String("proto", r.Proto),
					slog.String("query", r.URL.RawQuery),
					slog.String("content_type", r.Header.Get("Content-Type")),
				)
			}

			slog.LogAttrs(r.Context(), level, "request", attrs...)
		})
	}
}

func requestLevel(path string, status int) slog.Level {
	switch {
	case path == "/healthz":
		return slog.LevelDebug
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

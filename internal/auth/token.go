package auth

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

// TokenAuth returns middleware that validates a Bearer token from the
// Authorization header. If token is empty, the middleware is a no-op with
// zero overhead. The skip function exempts matching request paths from
// authentication.
func TokenAuth(token string, skip func(string) bool) func(http.Handler) http.Handler {
	if token == "" {
		return func(next http.Handler) http.Handler { return next }
	}

	tokenBytes := []byte(token)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skip != nil && skip(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			if auth == "" {
				writeAuthError(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}

			// Accept "Bearer <token>" (case-insensitive prefix).
			const prefix = "bearer "
			if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
				writeAuthError(w, http.StatusUnauthorized, "invalid Authorization header format")
				return
			}

			provided := []byte(auth[len(prefix):])
			if subtle.ConstantTimeCompare(provided, tokenBytes) != 1 {
				writeAuthError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// DefaultSkipPaths returns true for paths that should be exempt from token
// authentication:
//   - /healthz, /openapi.json — infrastructure endpoints
//   - /docs, /docs/* — API documentation
//   - Paths NOT starting with /api/ — frontend static files
func DefaultSkipPaths(path string) bool {
	switch {
	case path == "/healthz":
		return true
	case path == "/openapi.json":
		return true
	case path == "/docs" || strings.HasPrefix(path, "/docs/"):
		return true
	case !strings.HasPrefix(path, "/api/"):
		return true
	}
	return false
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"title":  http.StatusText(status),
		"detail": message,
	})
}

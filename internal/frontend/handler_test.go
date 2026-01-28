package frontend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// testFS returns a MapFS with typical SPA files.
func testFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":        {Data: []byte("<html><body>app</body></html>")},
		"app.js":            {Data: []byte("console.log('app')")},
		"style.css":         {Data: []byte("body{}")},
		"assets/image.png":  {Data: []byte("fakepng")},
		"assets/font.woff2": {Data: []byte("fakefont")},
	}
}

func TestServeIndexAtRoot(t *testing.T) {
	handler := NewHandler(testFS(), "")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html content type, got %q", ct)
	}
	if !strings.Contains(w.Body.String(), "<html>") {
		t.Error("expected index.html content in response body")
	}
}

func TestServeStaticFile(t *testing.T) {
	handler := NewHandler(testFS(), "")

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "javascript") {
		t.Errorf("expected javascript content type, got %q", ct)
	}
	if !strings.Contains(w.Body.String(), "console.log") {
		t.Error("expected app.js content in response body")
	}
}

func TestServeCSSFile(t *testing.T) {
	handler := NewHandler(testFS(), "")

	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/css") {
		t.Errorf("expected text/css content type, got %q", ct)
	}
}

func TestSPAFallback(t *testing.T) {
	handler := NewHandler(testFS(), "")

	req := httptest.NewRequest(http.MethodGet, "/some/unknown/route", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (SPA fallback), got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html for SPA fallback, got %q", ct)
	}
	if !strings.Contains(w.Body.String(), "<html>") {
		t.Error("expected index.html content for SPA fallback")
	}
}

func TestServeFileFromSubdirectory(t *testing.T) {
	handler := NewHandler(testFS(), "")

	req := httptest.NewRequest(http.MethodGet, "/assets/image.png", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "image/png") {
		t.Errorf("expected image/png content type, got %q", ct)
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/index.html", "text/html; charset=utf-8"},
		{"/style.css", "text/css; charset=utf-8"},
		{"/app.js", "application/javascript; charset=utf-8"},
		{"/app.mjs", "application/javascript; charset=utf-8"},
		{"/data.json", "application/json; charset=utf-8"},
		{"/logo.png", "image/png"},
		{"/photo.jpg", "image/jpeg"},
		{"/photo.jpeg", "image/jpeg"},
		{"/anim.gif", "image/gif"},
		{"/icon.svg", "image/svg+xml"},
		{"/favicon.ico", "image/x-icon"},
		{"/font.woff", "font/woff"},
		{"/font.woff2", "font/woff2"},
		{"/font.ttf", "font/ttf"},
		{"/font.eot", "application/vnd.ms-fontobject"},
		{"/image.webp", "image/webp"},
		{"/video.mp4", "video/mp4"},
		{"/video.webm", "video/webm"},
		{"/doc.pdf", "application/pdf"},
		{"/feed.xml", "application/xml"},
		{"/readme.txt", "text/plain; charset=utf-8"},
		{"/app.js.map", "application/json"},
		{"/unknown.xyz", ""},
	}

	for _, tc := range tests {
		got := getContentType(tc.path)
		if got != tc.want {
			t.Errorf("getContentType(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestHandlerWithFsPrefix(t *testing.T) {
	fs := fstest.MapFS{
		"dist/index.html": {Data: []byte("<html>dist</html>")},
		"dist/app.js":     {Data: []byte("distapp")},
	}
	handler := NewHandler(fs, "dist")

	// Root should serve dist/index.html.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "dist") {
		t.Error("expected dist/index.html content")
	}

	// Static file should resolve through prefix.
	req2 := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for /app.js, got %d", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "distapp") {
		t.Error("expected dist/app.js content")
	}
}

func TestServeDirectoryWithIndex(t *testing.T) {
	// Filesystem with a subdirectory that has its own index.html.
	fs := fstest.MapFS{
		"index.html":          {Data: []byte("<html>root</html>")},
		"sub/index.html":      {Data: []byte("<html>sub</html>")},
		"sub/other.js":        {Data: []byte("other")},
	}
	handler := NewHandler(fs, "")

	req := httptest.NewRequest(http.MethodGet, "/sub/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Should serve the sub directory's own index.html.
	if !strings.Contains(w.Body.String(), "sub") {
		t.Error("expected sub/index.html content")
	}
}

func TestServeDirectoryWithoutIndex(t *testing.T) {
	// Filesystem with a subdirectory that does NOT have index.html.
	fs := fstest.MapFS{
		"index.html":   {Data: []byte("<html>root</html>")},
		"sub/other.js": {Data: []byte("other")},
	}
	handler := NewHandler(fs, "")

	req := httptest.NewRequest(http.MethodGet, "/sub/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (SPA fallback), got %d", w.Code)
	}
	// Should fall back to root index.html for SPA routing.
	if !strings.Contains(w.Body.String(), "root") {
		t.Error("expected root index.html content for directory fallback")
	}
}

func TestServeFileWithNoExtension(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": {Data: []byte("<html>root</html>")},
		"LICENSE":     {Data: []byte("MIT License")},
	}
	handler := NewHandler(fs, "")

	req := httptest.NewRequest(http.MethodGet, "/LICENSE", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "MIT License") {
		t.Error("expected LICENSE content")
	}
}

func TestMissingIndex(t *testing.T) {
	// Filesystem with no index.html at all.
	fs := fstest.MapFS{
		"other.txt": {Data: []byte("hello")},
	}
	handler := NewHandler(fs, "")

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when index.html missing, got %d", w.Code)
	}
}

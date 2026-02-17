package _frontend

import (
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Handler serves static files from an fs.FS with SPA (Single Page Application) support.
// For paths that don't match a file, it serves index.html to support client-side routing.
type Handler struct {
	fs       fs.FS
	fsPrefix string // prefix to strip from fs paths (e.g., "dist" for embed.FS)
}

// NewHandler creates a new frontend handler using the provided filesystem.
// The fsPrefix is the directory prefix within the fs (e.g., "dist" for embedded files).
func NewHandler(filesystem fs.FS, fsPrefix string) *Handler {
	return &Handler{
		fs:       filesystem,
		fsPrefix: fsPrefix,
	}
}

// ServeHTTP implements http.Handler with SPA routing support.
// Static assets (JS, CSS, images, etc.) are served directly.
// All other paths serve index.html for React Router client-side routing.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the URL path
	urlPath := path.Clean(r.URL.Path)
	if urlPath == "/" {
		urlPath = "/index.html"
	}

	// Build the filesystem path
	fsPath := strings.TrimPrefix(urlPath, "/")
	if h.fsPrefix != "" {
		fsPath = path.Join(h.fsPrefix, fsPath)
	}

	// Try to open the requested file
	file, err := h.fs.Open(fsPath)
	if err != nil {
		// File not found - serve index.html for SPA routing
		h.serveIndex(w, r)
		return
	}
	defer file.Close()

	// Check if it's a directory
	stat, err := file.Stat()
	if err != nil {
		h.serveIndex(w, r)
		return
	}

	if stat.IsDir() {
		// Try to serve index.html from the directory
		indexPath := path.Join(fsPath, "index.html")
		indexFile, err := h.fs.Open(indexPath)
		if err != nil {
			h.serveIndex(w, r)
			return
		}
		defer indexFile.Close()
		file = indexFile
		stat, _ = indexFile.Stat()
	}

	// Set content type based on extension
	contentType := getContentType(urlPath)
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Serve the file
	if seeker, ok := file.(io.ReadSeeker); ok {
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), seeker)
	} else {
		w.Header().Set("Content-Length", string(rune(stat.Size())))
		io.Copy(w, file)
	}
}

// serveIndex serves the index.html file for SPA client-side routing.
func (h *Handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	indexPath := "index.html"
	if h.fsPrefix != "" {
		indexPath = path.Join(h.fsPrefix, indexPath)
	}

	file, err := h.fs.Open(indexPath)
	if err != nil {
		http.Error(w, "Frontend not available", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Frontend not available", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if seeker, ok := file.(io.ReadSeeker); ok {
		http.ServeContent(w, r, "index.html", stat.ModTime(), seeker)
	} else {
		io.Copy(w, file)
	}
}

// getContentType returns the content type for a file based on its extension.
func getContentType(filePath string) string {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".pdf":
		return "application/pdf"
	case ".xml":
		return "application/xml"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".map":
		return "application/json"
	default:
		return ""
	}
}

package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type StaticHandler struct {
	root string
}

func NewStaticHandler(root string) *StaticHandler {
	return &StaticHandler{
		root: root,
	}
}

func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Remove /static/ prefix and get the file path
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	if path == r.URL.Path {
		http.Error(w, "Invalid static path", http.StatusBadRequest)
		return
	}

	// Security: prevent directory traversal
	if strings.Contains(path, "..") || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Clean and join the path securely
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(h.root, cleanPath)

	// Verify the resolved path is within the root directory
	absRoot, err := filepath.Abs(h.root)
	if err != nil {
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(absFullPath, absRoot) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Check if file exists
	//#nosec G703 -- path is sanitized above with filepath.Clean and prefix check
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Set appropriate content type
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

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
	if strings.Contains(path, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(h.root, path)
	
	// Check if file exists
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
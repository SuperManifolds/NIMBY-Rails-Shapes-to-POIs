package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DownloadHandler struct {
	logger *slog.Logger
}

func NewDownloadHandler(logger *slog.Logger) *DownloadHandler {
	return &DownloadHandler{
		logger: logger,
	}
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "File not specified", http.StatusBadRequest)
		return
	}

	// Security: only allow downloading zip files from temp directory
	if !strings.HasSuffix(filename, ".zip") || strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}

	// Sanitize: use only the base name to prevent path traversal
	safeFilename := filepath.Base(filename)
	tempPath := filepath.Join(os.TempDir(), safeFilename)

	// Verify the resolved path is within temp directory
	tempDir := os.TempDir()
	if !strings.HasPrefix(tempPath, tempDir) {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	//#nosec G703 -- path is sanitized above with filepath.Base and prefix check
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", safeFilename))

	//#nosec G703 -- path is sanitized above with filepath.Base and prefix check
	file, err := os.Open(tempPath)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to open file for download", "error", err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to send file", "error", err)
	}

	// Clean up file after download
	go func() {
		time.Sleep(5 * time.Minute)
		os.Remove(tempPath)
	}()
}

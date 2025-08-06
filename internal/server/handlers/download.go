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
	if !strings.HasSuffix(filename, ".zip") || strings.Contains(filename, "..") {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}

	tempPath := filepath.Join(os.TempDir(), filename)
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

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

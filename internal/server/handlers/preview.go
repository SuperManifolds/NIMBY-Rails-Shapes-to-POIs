package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PreviewHandler struct {
	logger *slog.Logger
}

func NewPreviewHandler(logger *slog.Logger) *PreviewHandler {
	return &PreviewHandler{
		logger: logger,
	}
}

func (h *PreviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "File not specified", http.StatusBadRequest)
		return
	}

	// Security: only allow previewing PNG files from temp directory
	if !strings.HasSuffix(filename, ".png") || strings.Contains(filename, "..") {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}

	tempPath := filepath.Join(os.TempDir(), filename)
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	file, err := os.Open(tempPath)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to open preview file", "error", err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to send preview", "error", err)
	}

	// Clean up preview file after some time
	go func() {
		time.Sleep(10 * time.Minute)
		os.Remove(tempPath)
	}()
}
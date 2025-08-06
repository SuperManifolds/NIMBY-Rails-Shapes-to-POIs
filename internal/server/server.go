package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/supermanifolds/nimby_shapetopoi/internal/server/handlers"
	"github.com/supermanifolds/nimby_shapetopoi/internal/server/middleware"
)

type Server struct {
	logger *slog.Logger
	port   string
	server *http.Server
}

type Config struct {
	Port   string
	Logger *slog.Logger
}

func New(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return &Server{
		logger: cfg.Logger,
		port:   cfg.Port,
	}
}

func (s *Server) setupRoutes() {
	mux := http.NewServeMux()

	// Initialize handlers
	homeHandler := handlers.NewHomeHandler()
	uploadHandler := handlers.NewUploadHandler(s.logger)
	downloadHandler := handlers.NewDownloadHandler(s.logger)
	previewHandler := handlers.NewPreviewHandler(s.logger)
	healthHandler := handlers.NewHealthHandler()
	staticHandler := handlers.NewStaticHandler("static")

	// Setup middleware
	loggedMux := middleware.Logging(s.logger)(mux)

	// Routes using new Go 1.22 patterns
	mux.Handle("GET /", homeHandler)
	mux.Handle("POST /upload", uploadHandler)
	mux.Handle("GET /download/{filename}", downloadHandler)
	mux.Handle("GET /preview/{filename}", previewHandler)
	mux.Handle("GET /health", healthHandler)
	mux.Handle("GET /static/", staticHandler)

	s.server = &http.Server{
		Addr:              ":" + s.port,
		Handler:           loggedMux,
		ReadHeaderTimeout: 15 * time.Second,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.setupRoutes()

	s.logger.InfoContext(ctx, "Starting web server",
		"port", s.port,
		"url", "http://localhost:"+s.port)

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorContext(ctx, "Server failed to start", "error", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	s.logger.InfoContext(ctx, "Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(shutdownCtx)
}

package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			rw := &responseWriter{ResponseWriter: w, statusCode: 200}
			
			next.ServeHTTP(rw, r)
			
			duration := time.Since(start)
			logger.InfoContext(r.Context(), "HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration", duration,
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
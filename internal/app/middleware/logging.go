package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"internal-work-management-service/internal/infrastructure/logger"
)

var accessLogger = logger.New(logger.LogConfig{AppName: "http"})

// SetLogger overrides the default access logger
func SetLogger(l logger.Logger) {
	if l != nil {
		accessLogger = l
	}
}

// Logging writes access logs for every HTTP request
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		latency := time.Since(start)
		route := routePattern(r)
		reqID := RequestIDFromContext(r.Context())
		requestLogger := logger.Decorate(accessLogger, "component", "http_access")
		requestLogger.With(
			"req_id", reqID,
			"method", r.Method,
			"route", route,
			"path", r.URL.Path,
		).Info(
			"msg", "http_access",
			"status", rw.status,
			"latency", latency.String(),
			"latency_ms", latency.Milliseconds(),
			"bytes", rw.bytesWritten,
		)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += n
	return n, err
}

func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		if pattern := rc.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}

package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/signalridge/clinvoker/internal/metrics"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter creates a new response writer wrapper.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Metrics returns a middleware that records Prometheus metrics for HTTP requests.
// It tracks request count, duration, and status codes.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		rw := newResponseWriter(w)

		// Process request
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(rw.statusCode)

		metrics.RecordRequest(r.Method, path, status)
		metrics.RecordRequestDuration(r.Method, path, duration)
	})
}

// normalizePath normalizes URL paths for metric labels.
// This prevents high cardinality by grouping paths with IDs.
func normalizePath(path string) string {
	// Normalize common dynamic path patterns
	switch {
	case len(path) > 10 && path[:10] == "/sessions/":
		return "/sessions/:id"
	case len(path) > 10 && path[:10] == "/backends/":
		return "/backends/:name"
	default:
		return path
	}
}

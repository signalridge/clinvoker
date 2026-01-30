// Package middleware provides HTTP middleware components.
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// TraceIDHeader is the header name for trace IDs.
const TraceIDHeader = "X-Trace-ID"

// SpanIDHeader is the header name for span IDs.
const SpanIDHeader = "X-Span-ID"

// ParentSpanIDHeader is the header name for parent span IDs.
const ParentSpanIDHeader = "X-Parent-Span-ID"

// Context keys for tracing.
type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	spanIDKey  contextKey = "span_id"
	spanKey    contextKey = "span"
)

// Span represents a tracing span.
type Span struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
	Name         string
	StartTime    time.Time
	EndTime      time.Time
	Status       SpanStatus
	Attributes   map[string]any
}

// SpanStatus represents the status of a span.
type SpanStatus int

const (
	// SpanStatusUnset indicates the span status is not set.
	SpanStatusUnset SpanStatus = iota
	// SpanStatusOK indicates the span completed successfully.
	SpanStatusOK
	// SpanStatusError indicates the span encountered an error.
	SpanStatusError
)

// SpanExporter is an interface for exporting spans.
// Implement this interface to integrate with OpenTelemetry or other tracing systems.
type SpanExporter interface {
	// Export exports a completed span.
	Export(ctx context.Context, span *Span)
	// Shutdown gracefully shuts down the exporter.
	Shutdown(ctx context.Context) error
}

// NoopExporter is a span exporter that does nothing.
type NoopExporter struct{}

// Export does nothing.
func (NoopExporter) Export(ctx context.Context, span *Span) {}

// Shutdown does nothing.
func (NoopExporter) Shutdown(ctx context.Context) error { return nil }

// TracingConfig contains configuration for the tracing middleware.
type TracingConfig struct {
	// Enabled determines if tracing is enabled.
	Enabled bool

	// ServiceName is the name of the service for tracing.
	ServiceName string

	// Exporter is the span exporter to use.
	// If nil, a NoopExporter is used.
	Exporter SpanExporter

	// PropagateHeaders determines if trace context should be propagated in response headers.
	PropagateHeaders bool
}

// DefaultTracingConfig returns the default tracing configuration.
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		Enabled:          false,
		ServiceName:      "clinvoker",
		Exporter:         NoopExporter{},
		PropagateHeaders: true,
	}
}

// TracingMiddleware creates a middleware that adds tracing to requests.
func TracingMiddleware(cfg TracingConfig) func(http.Handler) http.Handler {
	if cfg.Exporter == nil {
		cfg.Exporter = NoopExporter{}
	}

	return func(next http.Handler) http.Handler {
		if !cfg.Enabled {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract or generate trace ID
			traceID := r.Header.Get(TraceIDHeader)
			if traceID == "" {
				traceID = generateID()
			}

			// Extract parent span ID if present
			parentSpanID := r.Header.Get(SpanIDHeader)

			// Generate new span ID
			spanID := generateID()

			// Create span
			span := &Span{
				TraceID:      traceID,
				SpanID:       spanID,
				ParentSpanID: parentSpanID,
				Name:         r.Method + " " + r.URL.Path,
				StartTime:    time.Now(),
				Attributes:   make(map[string]any),
			}

			// Add standard attributes
			span.Attributes["http.method"] = r.Method
			span.Attributes["http.url"] = r.URL.String()
			span.Attributes["http.host"] = r.Host
			span.Attributes["http.user_agent"] = r.UserAgent()
			span.Attributes["service.name"] = cfg.ServiceName

			// Add trace context to request context
			ctx := context.WithValue(r.Context(), traceIDKey, traceID)
			ctx = context.WithValue(ctx, spanIDKey, spanID)
			ctx = context.WithValue(ctx, spanKey, span)

			// Wrap response writer to capture status code
			wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			// Propagate headers if configured
			if cfg.PropagateHeaders {
				w.Header().Set(TraceIDHeader, traceID)
				w.Header().Set(SpanIDHeader, spanID)
				if parentSpanID != "" {
					w.Header().Set(ParentSpanIDHeader, parentSpanID)
				}
			}

			// Call next handler
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Complete span
			span.EndTime = time.Now()
			span.Attributes["http.status_code"] = wrapped.statusCode
			span.Attributes["http.response_size"] = wrapped.bytesWritten

			if wrapped.statusCode >= 400 {
				span.Status = SpanStatusError
			} else {
				span.Status = SpanStatusOK
			}

			// Export span
			cfg.Exporter.Export(ctx, span)
		})
	}
}

// statusRecorder wraps http.ResponseWriter to record the status code and bytes written.
type statusRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += n
	return n, err
}

// GetTraceID returns the trace ID from the context.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSpanID returns the span ID from the context.
func GetSpanID(ctx context.Context) string {
	if id, ok := ctx.Value(spanIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSpan returns the current span from the context.
func GetSpan(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanKey).(*Span); ok {
		return span
	}
	return nil
}

// AddSpanAttribute adds an attribute to the current span.
func AddSpanAttribute(ctx context.Context, key string, value any) {
	if span := GetSpan(ctx); span != nil {
		span.Attributes[key] = value
	}
}

// SetSpanStatus sets the status of the current span.
func SetSpanStatus(ctx context.Context, status SpanStatus) {
	if span := GetSpan(ctx); span != nil {
		span.Status = status
	}
}

// generateID generates a unique ID for traces and spans.
func generateID() string {
	return uuid.New().String()[:16]
}

// LoggingExporter is a span exporter that logs spans (for debugging).
type LoggingExporter struct {
	LogFunc func(format string, args ...any)
}

// Export logs the span.
func (e LoggingExporter) Export(ctx context.Context, span *Span) {
	if e.LogFunc == nil {
		return
	}
	duration := span.EndTime.Sub(span.StartTime)
	e.LogFunc("trace=%s span=%s parent=%s name=%s duration=%v status=%d attrs=%v",
		span.TraceID, span.SpanID, span.ParentSpanID, span.Name, duration, span.Status, span.Attributes)
}

// Shutdown does nothing for the logging exporter.
func (LoggingExporter) Shutdown(ctx context.Context) error { return nil }

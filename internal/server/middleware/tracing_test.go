package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestTracingMiddleware_Disabled(t *testing.T) {
	cfg := DefaultTracingConfig()
	cfg.Enabled = false

	middleware := TracingMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not have trace context when disabled
		if GetTraceID(r.Context()) != "" {
			t.Error("expected no trace ID when tracing is disabled")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestTracingMiddleware_Enabled(t *testing.T) {
	var exportedSpan *Span
	var mu sync.Mutex

	exporter := &testExporter{
		exportFunc: func(ctx context.Context, span *Span) {
			mu.Lock()
			exportedSpan = span
			mu.Unlock()
		},
	}

	cfg := DefaultTracingConfig()
	cfg.Enabled = true
	cfg.Exporter = exporter
	cfg.ServiceName = "test-service"

	middleware := TracingMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should have trace context
		traceID := GetTraceID(r.Context())
		if traceID == "" {
			t.Error("expected trace ID")
		}

		spanID := GetSpanID(r.Context())
		if spanID == "" {
			t.Error("expected span ID")
		}

		// Add custom attribute
		AddSpanAttribute(r.Context(), "custom.attr", "value")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test/path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Check exported span
	mu.Lock()
	defer mu.Unlock()

	if exportedSpan == nil {
		t.Fatal("expected span to be exported")
	}

	if exportedSpan.TraceID == "" {
		t.Error("expected trace ID in span")
	}
	if exportedSpan.SpanID == "" {
		t.Error("expected span ID in span")
	}
	if exportedSpan.Name != "GET /test/path" {
		t.Errorf("expected name 'GET /test/path', got %q", exportedSpan.Name)
	}
	if exportedSpan.Status != SpanStatusOK {
		t.Errorf("expected status OK, got %d", exportedSpan.Status)
	}
	if exportedSpan.Attributes["http.method"] != "GET" {
		t.Errorf("expected method GET, got %v", exportedSpan.Attributes["http.method"])
	}
	if exportedSpan.Attributes["http.status_code"] != http.StatusOK {
		t.Errorf("expected status code %d, got %v", http.StatusOK, exportedSpan.Attributes["http.status_code"])
	}
	if exportedSpan.Attributes["custom.attr"] != "value" {
		t.Errorf("expected custom.attr 'value', got %v", exportedSpan.Attributes["custom.attr"])
	}
	if exportedSpan.Attributes["service.name"] != "test-service" {
		t.Errorf("expected service.name 'test-service', got %v", exportedSpan.Attributes["service.name"])
	}
}

func TestTracingMiddleware_PropagatesTraceID(t *testing.T) {
	cfg := DefaultTracingConfig()
	cfg.Enabled = true
	cfg.PropagateHeaders = true

	middleware := TracingMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request with existing trace ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(TraceIDHeader, "existing-trace-id")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should propagate the trace ID in response
	if rec.Header().Get(TraceIDHeader) != "existing-trace-id" {
		t.Errorf("expected trace ID to be propagated, got %q", rec.Header().Get(TraceIDHeader))
	}

	// Should have a span ID
	if rec.Header().Get(SpanIDHeader) == "" {
		t.Error("expected span ID in response")
	}
}

func TestTracingMiddleware_ParentSpanID(t *testing.T) {
	var exportedSpan *Span
	var mu sync.Mutex

	exporter := &testExporter{
		exportFunc: func(ctx context.Context, span *Span) {
			mu.Lock()
			exportedSpan = span
			mu.Unlock()
		},
	}

	cfg := DefaultTracingConfig()
	cfg.Enabled = true
	cfg.Exporter = exporter

	middleware := TracingMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(TraceIDHeader, "trace-123")
	req.Header.Set(SpanIDHeader, "parent-span-456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	mu.Lock()
	defer mu.Unlock()

	if exportedSpan == nil {
		t.Fatal("expected span to be exported")
	}

	if exportedSpan.TraceID != "trace-123" {
		t.Errorf("expected trace ID 'trace-123', got %q", exportedSpan.TraceID)
	}
	if exportedSpan.ParentSpanID != "parent-span-456" {
		t.Errorf("expected parent span ID 'parent-span-456', got %q", exportedSpan.ParentSpanID)
	}
}

func TestTracingMiddleware_ErrorStatus(t *testing.T) {
	var exportedSpan *Span
	var mu sync.Mutex

	exporter := &testExporter{
		exportFunc: func(ctx context.Context, span *Span) {
			mu.Lock()
			exportedSpan = span
			mu.Unlock()
		},
	}

	cfg := DefaultTracingConfig()
	cfg.Enabled = true
	cfg.Exporter = exporter

	middleware := TracingMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	mu.Lock()
	defer mu.Unlock()

	if exportedSpan == nil {
		t.Fatal("expected span to be exported")
	}

	if exportedSpan.Status != SpanStatusError {
		t.Errorf("expected status Error, got %d", exportedSpan.Status)
	}
}

func TestSetSpanStatus(t *testing.T) {
	span := &Span{
		Status: SpanStatusUnset,
	}

	ctx := context.WithValue(context.Background(), spanKey, span)

	SetSpanStatus(ctx, SpanStatusError)

	if span.Status != SpanStatusError {
		t.Errorf("expected status Error, got %d", span.Status)
	}
}

func TestSetSpanStatus_NoSpan(t *testing.T) {
	// Should not panic when no span in context
	SetSpanStatus(context.Background(), SpanStatusError)
}

func TestAddSpanAttribute_NoSpan(t *testing.T) {
	// Should not panic when no span in context
	AddSpanAttribute(context.Background(), "key", "value")
}

func TestNoopExporter(t *testing.T) {
	exporter := NoopExporter{}

	// Should not panic
	exporter.Export(context.Background(), &Span{})

	err := exporter.Shutdown(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoggingExporter(t *testing.T) {
	var logged string
	exporter := LoggingExporter{
		LogFunc: func(format string, args ...any) {
			logged = format
		},
	}

	span := &Span{
		TraceID: "trace-123",
		SpanID:  "span-456",
		Name:    "test",
	}

	exporter.Export(context.Background(), span)

	if logged == "" {
		t.Error("expected log output")
	}
}

func TestLoggingExporter_NilLogFunc(t *testing.T) {
	exporter := LoggingExporter{}

	// Should not panic
	exporter.Export(context.Background(), &Span{})
}

func TestGetTraceID_NotSet(t *testing.T) {
	id := GetTraceID(context.Background())
	if id != "" {
		t.Errorf("expected empty trace ID, got %q", id)
	}
}

func TestGetSpanID_NotSet(t *testing.T) {
	id := GetSpanID(context.Background())
	if id != "" {
		t.Errorf("expected empty span ID, got %q", id)
	}
}

func TestGetSpan_NotSet(t *testing.T) {
	span := GetSpan(context.Background())
	if span != nil {
		t.Error("expected nil span")
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if len(id1) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id1))
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}

// testExporter is a test helper that captures exported spans.
type testExporter struct {
	exportFunc func(ctx context.Context, span *Span)
}

func (e *testExporter) Export(ctx context.Context, span *Span) {
	if e.exportFunc != nil {
		e.exportFunc(ctx, span)
	}
}

func (e *testExporter) Shutdown(ctx context.Context) error {
	return nil
}

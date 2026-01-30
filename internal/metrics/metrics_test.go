package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRequestMetrics(t *testing.T) {
	before := testutil.ToFloat64(RequestsTotal.WithLabelValues("GET", "/health", "200"))
	RecordRequest("GET", "/health", "200")
	after := testutil.ToFloat64(RequestsTotal.WithLabelValues("GET", "/health", "200"))
	if after != before+1 {
		t.Fatalf("RequestsTotal did not increment: before=%v after=%v", before, after)
	}

	// Histogram observation should not panic
	RecordRequestDuration("GET", "/health", 0.123)
}

func TestBackendMetrics(t *testing.T) {
	before := testutil.ToFloat64(BackendExecutions.WithLabelValues("claude", "success"))
	RecordBackendExecution("claude", "success")
	after := testutil.ToFloat64(BackendExecutions.WithLabelValues("claude", "success"))
	if after != before+1 {
		t.Fatalf("BackendExecutions did not increment: before=%v after=%v", before, after)
	}

	// Histogram observation should not panic
	RecordBackendExecutionDuration("claude", 0.42)
}

func TestSessionMetrics(t *testing.T) {
	SetActiveSessions(3)
	if got := testutil.ToFloat64(ActiveSessions); got != 3 {
		t.Fatalf("ActiveSessions = %v, want 3", got)
	}

	before := testutil.ToFloat64(SessionsCreated)
	IncrementSessionsCreated()
	after := testutil.ToFloat64(SessionsCreated)
	if after != before+1 {
		t.Fatalf("SessionsCreated did not increment: before=%v after=%v", before, after)
	}
}

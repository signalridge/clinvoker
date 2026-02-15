//revive:disable:var-naming
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

func TestExecutionModeMetrics(t *testing.T) {
	beforeChain := testutil.ToFloat64(ChainStepExecutions.WithLabelValues("claude", "ok"))
	RecordChainStepExecution("claude", "ok", 0.2)
	afterChain := testutil.ToFloat64(ChainStepExecutions.WithLabelValues("claude", "ok"))
	if afterChain != beforeChain+1 {
		t.Fatalf("ChainStepExecutions did not increment: before=%v after=%v", beforeChain, afterChain)
	}

	beforeCompare := testutil.ToFloat64(CompareBackendExecutions.WithLabelValues("gemini", "failed"))
	RecordCompareBackendExecution("gemini", "failed", 0.3)
	afterCompare := testutil.ToFloat64(CompareBackendExecutions.WithLabelValues("gemini", "failed"))
	if afterCompare != beforeCompare+1 {
		t.Fatalf("CompareBackendExecutions did not increment: before=%v after=%v", beforeCompare, afterCompare)
	}

	beforeParallel := testutil.ToFloat64(ParallelTaskExecutions.WithLabelValues("codex", "canceled"))
	RecordParallelTaskExecution("codex", "canceled", 0.1)
	afterParallel := testutil.ToFloat64(ParallelTaskExecutions.WithLabelValues("codex", "canceled"))
	if afterParallel != beforeParallel+1 {
		t.Fatalf("ParallelTaskExecutions did not increment: before=%v after=%v", beforeParallel, afterParallel)
	}
}

func TestExecutionModeMetrics_LabelNormalization(t *testing.T) {
	before := testutil.ToFloat64(ChainStepExecutions.WithLabelValues("unknown", "failed"))
	RecordChainStepExecution("", "not-a-real-status", 0.05)
	after := testutil.ToFloat64(ChainStepExecutions.WithLabelValues("unknown", "failed"))
	if after != before+1 {
		t.Fatalf("normalized ChainStepExecutions did not increment expected label set: before=%v after=%v", before, after)
	}
}

func TestPolicyMetrics(t *testing.T) {
	beforeDecision := testutil.ToFloat64(PolicyDecisions.WithLabelValues("allow", "shadow"))
	RecordPolicyDecision("allow", "shadow")
	afterDecision := testutil.ToFloat64(PolicyDecisions.WithLabelValues("allow", "shadow"))
	if afterDecision != beforeDecision+1 {
		t.Fatalf("PolicyDecisions did not increment: before=%v after=%v", beforeDecision, afterDecision)
	}

	RecordPolicyEvalDuration(0.003)

	beforeFallback := testutil.ToFloat64(PolicyFallbacks.WithLabelValues("fail-open", "engine_error"))
	RecordPolicyFallback("fail-open", "engine_error")
	afterFallback := testutil.ToFloat64(PolicyFallbacks.WithLabelValues("fail-open", "engine_error"))
	if afterFallback != beforeFallback+1 {
		t.Fatalf("PolicyFallbacks did not increment: before=%v after=%v", beforeFallback, afterFallback)
	}

	beforeQuota := testutil.ToFloat64(PolicyQuotaRejections.WithLabelValues("rate", "rate_exceeded"))
	RecordPolicyQuotaRejection("rate", "rate_exceeded")
	afterQuota := testutil.ToFloat64(PolicyQuotaRejections.WithLabelValues("rate", "rate_exceeded"))
	if afterQuota != beforeQuota+1 {
		t.Fatalf("PolicyQuotaRejections did not increment: before=%v after=%v", beforeQuota, afterQuota)
	}
}

func TestPolicyMetrics_LabelNormalization(t *testing.T) {
	beforeDecision := testutil.ToFloat64(PolicyDecisions.WithLabelValues("unknown", "unknown"))
	RecordPolicyDecision("not-real-decision", "not-real-mode")
	afterDecision := testutil.ToFloat64(PolicyDecisions.WithLabelValues("unknown", "unknown"))
	if afterDecision != beforeDecision+1 {
		t.Fatalf("normalized PolicyDecisions did not increment: before=%v after=%v", beforeDecision, afterDecision)
	}

	beforeFallback := testutil.ToFloat64(PolicyFallbacks.WithLabelValues("unknown", "other"))
	RecordPolicyFallback("not-real-failure-mode", "untrusted-reason-from-request")
	afterFallback := testutil.ToFloat64(PolicyFallbacks.WithLabelValues("unknown", "other"))
	if afterFallback != beforeFallback+1 {
		t.Fatalf("normalized PolicyFallbacks did not increment: before=%v after=%v", beforeFallback, afterFallback)
	}

	beforeQuota := testutil.ToFloat64(PolicyQuotaRejections.WithLabelValues("global", "other"))
	RecordPolicyQuotaRejection("user-input-scope", "another-untrusted-reason")
	afterQuota := testutil.ToFloat64(PolicyQuotaRejections.WithLabelValues("global", "other"))
	if afterQuota != beforeQuota+1 {
		t.Fatalf("normalized PolicyQuotaRejections did not increment: before=%v after=%v", beforeQuota, afterQuota)
	}
}

//revive:enable:var-naming

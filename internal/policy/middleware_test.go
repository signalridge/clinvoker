package policy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/requestctx"
)

func TestRedactExplainDecisionPath(t *testing.T) {
	input := []string{
		"matched:rule-1",
		"subject_key_id=raw-secret-key",
		"tenant_id:tenant-a",
		"token=abc123",
		"action:deny",
	}

	got := redactExplainDecisionPath(input)
	if got[0] != "matched:rule-1" {
		t.Fatalf("expected non-sensitive entry to remain unchanged, got %q", got[0])
	}
	if got[1] != "subject_key_id=<redacted>" {
		t.Fatalf("expected subject_key_id to be redacted, got %q", got[1])
	}
	if got[2] != "tenant_id:<redacted>" {
		t.Fatalf("expected tenant_id to be redacted, got %q", got[2])
	}
	if got[3] != "token=<redacted>" {
		t.Fatalf("expected token to be redacted, got %q", got[3])
	}
	if got[4] != "action:deny" {
		t.Fatalf("expected action entry to remain unchanged, got %q", got[4])
	}
}

type failingQuotaStore struct{}

func (f failingQuotaStore) Reserve(context.Context, QuotaRequest) (QuotaResult, error) {
	return QuotaResult{}, errors.New("quota store failure")
}

func (f failingQuotaStore) Consume(context.Context, QuotaRequest) (QuotaResult, error) {
	return QuotaResult{}, errors.New("quota store failure")
}

func (f failingQuotaStore) Release(context.Context, QuotaRequest) error {
	return nil
}

func (f failingQuotaStore) Peek(context.Context, QuotaRequest) (QuotaResult, error) {
	return QuotaResult{}, nil
}

func (f failingQuotaStore) Close() error {
	return nil
}

func installEngineForTest(t *testing.T, runtime RuntimeConfig, rules []Rule, store QuotaStore) {
	t.Helper()

	compiled, err := Compile(&RuleSource{
		Version: "v1",
		Rules:   rules,
	}, runtime.DefaultDecision)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if store == nil {
		store = NewMemoryQuotaStore()
	}

	currentEngine.Store(&Engine{
		compiled: compiled,
		quota:    store,
		runtime:  runtime,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	t.Cleanup(ResetForTest)
}

func TestMiddleware_EnforceDeny_ReturnsStableContract(t *testing.T) {
	installEngineForTest(t, RuntimeConfig{
		Mode:            ModeEnforce,
		FailureMode:     FailureModeOpen,
		DefaultDecision: DefaultDecisionAllow,
	}, []Rule{
		{
			ID:       "deny-blocked",
			Enabled:  true,
			Priority: 10,
			Selectors: RuleSelectors{
				PathPrefix: "/api/v1/blocked",
				Methods:    []string{"GET"},
			},
			Action: RuleAction{Type: ActionDeny},
		},
	}, nil)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)))(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/blocked", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("next handler should not be called for enforce deny")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	var payload responseError
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if payload.Code != "policy_denied" {
		t.Fatalf("code = %q, want %q", payload.Code, "policy_denied")
	}
	if payload.RequestID == "" {
		t.Fatal("request_id should not be empty")
	}
	if payload.DecisionID == "" {
		t.Fatal("decision_id should not be empty")
	}
	if payload.Reason != "policy_deny" {
		t.Fatalf("reason = %q, want %q", payload.Reason, "policy_deny")
	}
	if len(payload.MatchedRuleIDs) != 1 || payload.MatchedRuleIDs[0] != "deny-blocked" {
		t.Fatalf("matched_rule_ids = %v, want [deny-blocked]", payload.MatchedRuleIDs)
	}
}

func TestMiddleware_ShadowMode_NonBlockingWithExplainHeaders(t *testing.T) {
	installEngineForTest(t, RuntimeConfig{
		Mode:            ModeShadow,
		FailureMode:     FailureModeOpen,
		ExplainEnabled:  true,
		DefaultDecision: DefaultDecisionAllow,
	}, []Rule{
		{
			ID:       "deny-blocked",
			Enabled:  true,
			Priority: 10,
			Selectors: RuleSelectors{
				PathPrefix: "/api/v1/blocked",
				Methods:    []string{"GET"},
			},
			Action: RuleAction{Type: ActionDeny},
		},
	}, nil)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)))(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/blocked", http.NoBody)
	req.Header.Set("X-Policy-Explain", "true")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler should be called in shadow mode")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Header().Get("X-Policy-Decision") != "deny" {
		t.Fatalf("X-Policy-Decision = %q, want %q", rec.Header().Get("X-Policy-Decision"), "deny")
	}
	if rec.Header().Get("X-Policy-Matched-Rules") != "deny-blocked" {
		t.Fatalf("X-Policy-Matched-Rules = %q, want %q", rec.Header().Get("X-Policy-Matched-Rules"), "deny-blocked")
	}
	if path := rec.Header().Get("X-Policy-Decision-Path"); !strings.Contains(path, "matched:deny-blocked") {
		t.Fatalf("X-Policy-Decision-Path should include matched rule, got %q", path)
	}
}

func TestMiddleware_FallbackFailOpen_AllowsRequest(t *testing.T) {
	installEngineForTest(t, RuntimeConfig{
		Mode:            ModeEnforce,
		FailureMode:     FailureModeOpen,
		DefaultDecision: DefaultDecisionAllow,
	}, []Rule{
		{
			ID:       "quota-rate",
			Enabled:  true,
			Priority: 10,
			Action: RuleAction{
				Type: ActionQuota,
				Quota: &QuotaSpec{
					RatePerMinute: 1,
				},
			},
		},
	}, failingQuotaStore{})

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)))(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/prompt", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler should be called in fail-open fallback")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMiddleware_FallbackFailClosed_BlocksRequest(t *testing.T) {
	installEngineForTest(t, RuntimeConfig{
		Mode:            ModeEnforce,
		FailureMode:     FailureModeClosed,
		DefaultDecision: DefaultDecisionAllow,
	}, []Rule{
		{
			ID:       "quota-rate",
			Enabled:  true,
			Priority: 10,
			Action: RuleAction{
				Type: ActionQuota,
				Quota: &QuotaSpec{
					RatePerMinute: 1,
				},
			},
		},
	}, failingQuotaStore{})

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)))(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/prompt", http.NoBody)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("next handler should not be called in fail-closed fallback")
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var payload responseError
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if payload.Code != "policy_engine_unavailable" {
		t.Fatalf("code = %q, want %q", payload.Code, "policy_engine_unavailable")
	}
	if payload.RequestID == "" || payload.DecisionID == "" {
		t.Fatalf("fallback response must include request_id and decision_id, got request_id=%q decision_id=%q", payload.RequestID, payload.DecisionID)
	}
}

func TestMiddleware_ReleasesConcurrencyQuotaOnPanic(t *testing.T) {
	installEngineForTest(t, RuntimeConfig{
		Mode:            ModeEnforce,
		FailureMode:     FailureModeOpen,
		DefaultDecision: DefaultDecisionAllow,
	}, []Rule{
		{
			ID:       "quota-concurrency",
			Enabled:  true,
			Priority: 10,
			Action: RuleAction{
				Type: ActionQuota,
				Quota: &QuotaSpec{
					Concurrency: 1,
				},
			},
		},
	}, NewMemoryQuotaStore())

	handler := Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/prompt", http.NoBody)
	req.RemoteAddr = "127.0.0.1:8080"
	rec := httptest.NewRecorder()

	func() {
		defer func() {
			_ = recover()
		}()
		handler.ServeHTTP(rec, req)
	}()

	engine := Current()
	if engine == nil {
		t.Fatal("engine should not be nil")
	}
	nextReq := httptest.NewRequest(http.MethodGet, "/api/v1/prompt", http.NoBody)
	nextReq.RemoteAddr = "127.0.0.1:8081"
	identity := requestctx.Identity{}
	outcome, err := engine.EvaluateRequest(requestctx.WithIdentity(context.Background(), &identity), nextReq)
	if err != nil {
		t.Fatalf("EvaluateRequest() error = %v", err)
	}
	if outcome.Decision != DecisionAllow {
		t.Fatalf("decision after panic = %q, want %q (quota should be released)", outcome.Decision, DecisionAllow)
	}
	if outcome.Release != nil {
		outcome.Release()
	}
}

func TestMiddleware_ReleasesConcurrencyQuotaOnCancellation(t *testing.T) {
	installEngineForTest(t, RuntimeConfig{
		Mode:            ModeEnforce,
		FailureMode:     FailureModeOpen,
		DefaultDecision: DefaultDecisionAllow,
	}, []Rule{
		{
			ID:       "quota-concurrency",
			Enabled:  true,
			Priority: 10,
			Action: RuleAction{
				Type: ActionQuota,
				Quota: &QuotaSpec{
					Concurrency: 1,
				},
			},
		},
	}, NewMemoryQuotaStore())

	started := make(chan struct{})
	done := make(chan struct{})

	handler := Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		w.WriteHeader(http.StatusNoContent)
	}))

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/prompt", http.NoBody).WithContext(ctx)
	req.RemoteAddr = "127.0.0.1:8080"
	rec := httptest.NewRecorder()

	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	<-started
	cancel()
	<-done

	engine := Current()
	if engine == nil {
		t.Fatal("engine should not be nil")
	}
	nextReq := httptest.NewRequest(http.MethodGet, "/api/v1/prompt", http.NoBody)
	nextReq.RemoteAddr = "127.0.0.1:8081"
	identity := requestctx.Identity{}
	outcome, err := engine.EvaluateRequest(requestctx.WithIdentity(context.Background(), &identity), nextReq)
	if err != nil {
		t.Fatalf("EvaluateRequest() error = %v", err)
	}
	if outcome.Decision != DecisionAllow {
		t.Fatalf("decision after cancellation = %q, want %q (quota should be released)", outcome.Decision, DecisionAllow)
	}
	if outcome.Release != nil {
		outcome.Release()
	}
}

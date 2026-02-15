package policy

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/signalridge/clinvoker/internal/requestctx"
)

func TestEngineEvaluate_ConcurrencyReleaseAllowsSubsequentRequests(t *testing.T) {
	compiled, err := Compile(&RuleSource{
		Version: "v1",
		Rules: []Rule{
			{
				ID:       "quota-concurrency",
				Enabled:  true,
				Priority: 10,
				Action: RuleAction{
					Type: ActionQuota,
					Quota: &QuotaSpec{
						Concurrency: 1,
						Scopes:      []string{"key"},
					},
				},
			},
		},
	}, DefaultDecisionAllow)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	engine := &Engine{
		compiled: compiled,
		quota:    NewMemoryQuotaStore(),
		runtime: RuntimeConfig{
			Mode:            ModeEnforce,
			FailureMode:     FailureModeOpen,
			DefaultDecision: DefaultDecisionAllow,
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	t.Cleanup(func() {
		if engine.quota != nil {
			_ = engine.quota.Close()
		}
	})

	identity := requestctx.Identity{
		SubjectKeyID: "key_abc123",
	}
	ctx := requestctx.WithIdentity(context.Background(), &identity)

	req1 := httptest.NewRequest("GET", "/api/v1/prompt", http.NoBody)
	req1.RemoteAddr = "127.0.0.1:8080"
	first, err := engine.EvaluateRequest(ctx, req1)
	if err != nil {
		t.Fatalf("first EvaluateRequest() error = %v", err)
	}
	if first.Decision != DecisionAllow {
		t.Fatalf("first decision = %q, want %q", first.Decision, DecisionAllow)
	}
	if first.Release == nil {
		t.Fatal("first decision should provide release callback for concurrency quota")
	}

	req2 := httptest.NewRequest("GET", "/api/v1/prompt", http.NoBody)
	req2.RemoteAddr = "127.0.0.1:8081"
	second, err := engine.EvaluateRequest(ctx, req2)
	if err != nil {
		t.Fatalf("second EvaluateRequest() error = %v", err)
	}
	if second.Decision != DecisionQuotaReject {
		t.Fatalf("second decision = %q, want %q", second.Decision, DecisionQuotaReject)
	}
	if second.Reason != "concurrency_exceeded" {
		t.Fatalf("second reason = %q, want %q", second.Reason, "concurrency_exceeded")
	}

	first.Release()

	req3 := httptest.NewRequest("GET", "/api/v1/prompt", http.NoBody)
	req3.RemoteAddr = "127.0.0.1:8082"
	third, err := engine.EvaluateRequest(ctx, req3)
	if err != nil {
		t.Fatalf("third EvaluateRequest() error = %v", err)
	}
	if third.Decision != DecisionAllow {
		t.Fatalf("third decision = %q, want %q", third.Decision, DecisionAllow)
	}
}

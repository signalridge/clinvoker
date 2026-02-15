package policy

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/requestctx"
)

// Engine executes policy decisions using compiled rules.
type Engine struct {
	compiled *CompiledPolicy
	quota    QuotaStore
	runtime  RuntimeConfig
	logger   *slog.Logger
}

var currentEngine atomic.Pointer[Engine]

// ConfigureFromConfig compiles and installs the current policy engine from config.
func ConfigureFromConfig(cfg *config.Config, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg == nil || !cfg.Server.Policy.Enabled {
		if prev := currentEngine.Load(); prev != nil && prev.quota != nil {
			_ = prev.quota.Close()
		}
		currentEngine.Store(nil)
		return nil
	}

	runtime := RuntimeConfig{
		Mode:            EngineMode(cfg.Server.Policy.Mode),
		FailureMode:     FailureMode(cfg.Server.Policy.FailureMode),
		ExplainEnabled:  cfg.Server.Policy.ExplainEnabled,
		DefaultDecision: DefaultDecision(cfg.Server.Policy.DefaultDecision),
		QuotaStore:      strings.ToLower(strings.TrimSpace(cfg.Server.Policy.QuotaStore)),
	}
	if runtime.QuotaStore == "" {
		runtime.QuotaStore = "memory"
	}
	if err := validateRuntimeConfig(runtime); err != nil {
		return err
	}

	source, err := LoadRuleSource(cfg.Server.Policy.RulesFile)
	if err != nil {
		return err
	}
	compiled, err := Compile(source, runtime.DefaultDecision)
	if err != nil {
		return err
	}

	var quota QuotaStore
	switch runtime.QuotaStore {
	case "memory":
		quota = NewMemoryQuotaStore()
	case "shared":
		return fmt.Errorf("quota_store 'shared' is not implemented in v0.7 internal-first rollout")
	default:
		return fmt.Errorf("unsupported quota_store %q", runtime.QuotaStore)
	}

	engine := &Engine{
		compiled: compiled,
		quota:    quota,
		runtime:  runtime,
		logger:   logger,
	}
	if prev := currentEngine.Load(); prev != nil && prev.quota != nil {
		_ = prev.quota.Close()
	}
	currentEngine.Store(engine)
	return nil
}

// Current returns the installed policy engine if enabled.
func Current() *Engine {
	return currentEngine.Load()
}

// ResetForTest clears the global engine state.
func ResetForTest() {
	if prev := currentEngine.Load(); prev != nil && prev.quota != nil {
		_ = prev.quota.Close()
	}
	currentEngine.Store(nil)
}

// EvaluateRequest performs rule and quota checks for the request.
func (e *Engine) EvaluateRequest(ctx context.Context, r *http.Request) (DecisionOutcome, error) {
	if e == nil {
		return DecisionOutcome{Decision: DecisionAllow, Reason: "policy_disabled"}, nil
	}
	if shouldForceEvalError(r) {
		return DecisionOutcome{}, fmt.Errorf("forced policy evaluation failure")
	}

	decisionID := newDecisionID()
	requestID := chiMiddleware.GetReqID(ctx)
	if requestID == "" {
		requestID = newRequestID()
	}

	identity := requestctx.IdentityFromContext(ctx)
	input := EvalInput{
		Method:       r.Method,
		Path:         r.URL.Path,
		SourceIP:     splitHost(r.RemoteAddr),
		Backend:      strings.TrimSpace(strings.ToLower(r.Header.Get("X-Backend"))),
		Model:        strings.TrimSpace(strings.ToLower(r.Header.Get("X-Model"))),
		SubjectKeyID: strings.TrimSpace(strings.ToLower(identity.SubjectKeyID)),
		TenantID:     strings.TrimSpace(strings.ToLower(identity.TenantID)),
	}

	eval := e.compiled.Evaluate(&input)
	outcome := DecisionOutcome{
		Decision:       eval.Decision,
		Reason:         eval.Reason,
		DecisionID:     decisionID,
		RequestID:      requestID,
		MatchedRuleIDs: append([]string{}, eval.MatchedRuleIDs...),
		DecisionPath:   append([]string{}, eval.DecisionPath...),
	}

	if eval.Action != ActionQuota || eval.Quota == nil {
		return outcome, nil
	}

	window := windowDuration(eval.Quota)
	key := buildQuotaKey(eval.RuleID, &input, eval.Quota)

	if eval.Quota.RatePerMinute > 0 {
		result, err := e.quota.Consume(ctx, QuotaRequest{
			Operation: QuotaOperationConsume,
			Kind:      QuotaKindRate,
			Key:       key,
			Limit:     eval.Quota.RatePerMinute,
			Amount:    1,
			Window:    window,
		})
		if err != nil {
			return DecisionOutcome{}, fmt.Errorf("rate quota consume: %w", err)
		}
		if !result.Allowed {
			outcome.Decision = DecisionQuotaReject
			outcome.Reason = reasonRateExceeded
			return outcome, nil
		}
	}

	if eval.Quota.TokenBudget > 0 {
		tokenAmount := int64(1)
		if headerTokens := strings.TrimSpace(r.Header.Get("X-Policy-Token-Usage")); headerTokens != "" {
			if parsed, parseErr := parseInt64(headerTokens); parseErr == nil && parsed > 0 {
				tokenAmount = parsed
			}
		}
		result, err := e.quota.Consume(ctx, QuotaRequest{
			Operation: QuotaOperationConsume,
			Kind:      QuotaKindToken,
			Key:       key,
			Limit:     eval.Quota.TokenBudget,
			Amount:    tokenAmount,
			Window:    window,
		})
		if err != nil {
			return DecisionOutcome{}, fmt.Errorf("token quota consume: %w", err)
		}
		if !result.Allowed {
			outcome.Decision = DecisionQuotaReject
			outcome.Reason = reasonTokenExceeded
			return outcome, nil
		}
	}

	if eval.Quota.Concurrency > 0 {
		reserved, err := e.quota.Reserve(ctx, QuotaRequest{
			Operation: QuotaOperationReserve,
			Kind:      QuotaKindConcurrency,
			Key:       key,
			Limit:     eval.Quota.Concurrency,
			Amount:    1,
			Window:    window,
		})
		if err != nil {
			return DecisionOutcome{}, fmt.Errorf("concurrency quota reserve: %w", err)
		}
		if !reserved.Allowed {
			outcome.Decision = DecisionQuotaReject
			outcome.Reason = reasonConcurrency
			return outcome, nil
		}
		outcome.Release = func() {
			_ = e.quota.Release(context.Background(), QuotaRequest{
				Operation: QuotaOperationRelease,
				Kind:      QuotaKindConcurrency,
				Key:       key,
				Amount:    1,
			})
		}
	}

	outcome.Decision = DecisionAllow
	outcome.Reason = reasonQuotaChecksPassed
	return outcome, nil
}

func splitHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func newDecisionID() string {
	return newID("dec")
}

func newRequestID() string {
	return newID("req")
}

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}

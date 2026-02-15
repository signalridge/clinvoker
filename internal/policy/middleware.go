package policy

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/signalridge/clinvoker/internal/metrics"
)

type responseError struct {
	Code           string   `json:"code"`
	Message        string   `json:"message"`
	RequestID      string   `json:"request_id"`
	DecisionID     string   `json:"decision_id"`
	Reason         string   `json:"reason"`
	MatchedRuleIDs []string `json:"matched_rule_ids,omitempty"`
}

// Middleware returns policy governance middleware bound to the current engine.
func Middleware(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			engine := Current()
			if engine == nil {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			mode := policyModeOrDefault(engine.runtime.Mode)
			failureMode := failureModeOrDefault(engine.runtime.FailureMode)

			outcome, err := engine.EvaluateRequest(r.Context(), r)
			if err != nil {
				fallbackDecision := DecisionAllow
				status := "allowed"
				if failureMode == FailureModeClosed {
					fallbackDecision = DecisionDeny
					status = "blocked"
				}
				decisionID := newDecisionID()
				requestID := chiMiddleware.GetReqID(r.Context())
				if requestID == "" {
					requestID = newRequestID()
				}

				emitPolicyMetrics(DecisionFallback, mode, failureMode, reasonEngineError, quotaScopeGlobal, time.Since(start).Seconds())
				emitAudit(logger, r, &AuditEvent{
					RequestID:      requestID,
					DecisionID:     decisionID,
					Decision:       DecisionFallback,
					Mode:           string(mode),
					Reason:         reasonEngineError,
					FailureMode:    string(failureMode),
					MatchedRuleIDs: nil,
					Status:         status,
					Error:          err.Error(),
				})

				if fallbackDecision == DecisionAllow {
					next.ServeHTTP(w, r)
					return
				}

				writePolicyError(w, http.StatusServiceUnavailable, &responseError{
					Code:       "policy_engine_unavailable",
					Message:    "policy evaluation failed",
					RequestID:  requestID,
					DecisionID: decisionID,
					Reason:     reasonEngineError,
				})
				return
			}

			if outcome.Release != nil {
				defer outcome.Release()
			}

			emitExplainHeadersIfEnabled(w, r, engine.runtime.ExplainEnabled, &outcome)

			if mode == ModeShadow {
				emitPolicyMetrics(outcome.Decision, mode, "", outcome.Reason, inferQuotaScope(outcome.Reason), time.Since(start).Seconds())
				emitAudit(logger, r, &AuditEvent{
					RequestID:      outcome.RequestID,
					DecisionID:     outcome.DecisionID,
					Decision:       outcome.Decision,
					Mode:           string(mode),
					Reason:         outcome.Reason,
					MatchedRuleIDs: outcome.MatchedRuleIDs,
					Status:         "shadow",
				})
				next.ServeHTTP(w, r)
				return
			}

			switch outcome.Decision {
			case DecisionAllow:
				emitPolicyMetrics(outcome.Decision, mode, "", outcome.Reason, inferQuotaScope(outcome.Reason), time.Since(start).Seconds())
				emitAudit(logger, r, &AuditEvent{
					RequestID:      outcome.RequestID,
					DecisionID:     outcome.DecisionID,
					Decision:       outcome.Decision,
					Mode:           string(mode),
					Reason:         outcome.Reason,
					MatchedRuleIDs: outcome.MatchedRuleIDs,
					Status:         "allowed",
				})
				next.ServeHTTP(w, r)
			case DecisionDeny:
				emitPolicyMetrics(outcome.Decision, mode, "", outcome.Reason, inferQuotaScope(outcome.Reason), time.Since(start).Seconds())
				emitAudit(logger, r, &AuditEvent{
					RequestID:      outcome.RequestID,
					DecisionID:     outcome.DecisionID,
					Decision:       outcome.Decision,
					Mode:           string(mode),
					Reason:         outcome.Reason,
					MatchedRuleIDs: outcome.MatchedRuleIDs,
					Status:         "blocked",
				})
				writePolicyError(w, http.StatusForbidden, &responseError{
					Code:           "policy_denied",
					Message:        "request denied by policy",
					RequestID:      outcome.RequestID,
					DecisionID:     outcome.DecisionID,
					Reason:         normalizeDecisionReason(outcome.Reason),
					MatchedRuleIDs: outcome.MatchedRuleIDs,
				})
			case DecisionQuotaReject:
				emitPolicyMetrics(outcome.Decision, mode, "", outcome.Reason, inferQuotaScope(outcome.Reason), time.Since(start).Seconds())
				emitAudit(logger, r, &AuditEvent{
					RequestID:      outcome.RequestID,
					DecisionID:     outcome.DecisionID,
					Decision:       outcome.Decision,
					Mode:           string(mode),
					Reason:         outcome.Reason,
					MatchedRuleIDs: outcome.MatchedRuleIDs,
					Status:         "blocked",
				})
				writePolicyError(w, http.StatusTooManyRequests, &responseError{
					Code:           "policy_quota_exceeded",
					Message:        "request rejected by policy quota",
					RequestID:      outcome.RequestID,
					DecisionID:     outcome.DecisionID,
					Reason:         normalizeQuotaReason(outcome.Reason),
					MatchedRuleIDs: outcome.MatchedRuleIDs,
				})
			case DecisionFallback:
				// DecisionFallback is handled in the evaluation-error branch above.
				next.ServeHTTP(w, r)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

func writePolicyError(w http.ResponseWriter, statusCode int, payload *responseError) {
	if payload == nil {
		payload = &responseError{
			Code:    "policy_error",
			Message: "policy error",
			Reason:  "unspecified",
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if payload.RequestID != "" {
		w.Header().Set("X-Request-ID", payload.RequestID)
	}
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func inferQuotaScope(reason string) string {
	reason = strings.ToLower(strings.TrimSpace(reason))
	switch reason {
	case reasonRateExceeded:
		return "rate"
	case reasonConcurrency:
		return "concurrency"
	case reasonTokenExceeded:
		return "token"
	default:
		return quotaScopeGlobal
	}
}

func emitPolicyMetrics(decision Decision, mode EngineMode, failureMode FailureMode, reason, scope string, durationSeconds float64) {
	metrics.RecordPolicyDecision(string(decision), string(mode))
	metrics.RecordPolicyEvalDuration(durationSeconds)
	if decision == DecisionFallback {
		metrics.RecordPolicyFallback(string(failureMode), reason)
	}
	if decision == DecisionQuotaReject {
		metrics.RecordPolicyQuotaRejection(scope, reason)
	}
}

func emitExplainHeadersIfEnabled(w http.ResponseWriter, r *http.Request, explainEnabled bool, outcome *DecisionOutcome) {
	if !explainEnabled {
		return
	}
	if outcome == nil {
		return
	}
	if !strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Policy-Explain")), "true") {
		return
	}

	w.Header().Set("X-Policy-Decision", string(outcome.Decision))
	w.Header().Set("X-Policy-Reason", normalizeDecisionReason(outcome.Reason))
	if outcome.DecisionID != "" {
		w.Header().Set("X-Policy-Decision-ID", outcome.DecisionID)
	}
	if len(outcome.MatchedRuleIDs) > 0 {
		w.Header().Set("X-Policy-Matched-Rules", strings.Join(outcome.MatchedRuleIDs, ","))
	}
	if len(outcome.DecisionPath) > 0 {
		w.Header().Set("X-Policy-Decision-Path", strings.Join(redactExplainDecisionPath(outcome.DecisionPath), ","))
	}
}

func redactExplainDecisionPath(path []string) []string {
	if len(path) == 0 {
		return nil
	}

	redacted := make([]string, 0, len(path))
	for _, step := range path {
		redacted = append(redacted, redactExplainStep(step))
	}
	return redacted
}

func redactExplainStep(step string) string {
	sensitiveKeys := []string{
		"api_key",
		"subject_key_id",
		"tenant_id",
		"authorization",
		"token",
	}

	step = strings.TrimSpace(step)
	lower := strings.ToLower(step)
	for _, key := range sensitiveKeys {
		for _, sep := range []string{"=", ":"} {
			pattern := key + sep
			idx := strings.Index(lower, pattern)
			if idx >= 0 {
				return step[:idx+len(pattern)] + "<redacted>"
			}
		}
	}
	return step
}

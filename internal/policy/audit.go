package policy

import (
	"log/slog"
	"net/http"
)

// AuditEvent is a structured policy audit payload.
type AuditEvent struct {
	RequestID      string
	DecisionID     string
	Decision       Decision
	Mode           string
	Reason         string
	MatchedRuleIDs []string
	FailureMode    string
	Status         string
	Error          string
}

func emitAudit(logger *slog.Logger, r *http.Request, event *AuditEvent) {
	if logger == nil {
		return
	}
	if event == nil {
		return
	}
	attrs := []any{
		"request_id", event.RequestID,
		"decision_id", event.DecisionID,
		"decision", event.Decision,
		"mode", event.Mode,
		"reason", normalizeDecisionReason(event.Reason),
		"matched_rule_ids", event.MatchedRuleIDs,
		"status", event.Status,
		"path", r.URL.Path,
		"method", r.Method,
	}
	if event.FailureMode != "" {
		attrs = append(attrs, "failure_mode", event.FailureMode)
	}
	if event.Error != "" {
		attrs = append(attrs, "error", event.Error)
	}
	logger.Info("policy_decision", attrs...)
}

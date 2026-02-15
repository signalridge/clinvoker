// Package policy implements deterministic request governance for allow/deny/quota decisions.
package policy

import "time"

// EngineMode controls policy behavior mode.
type EngineMode string

// Engine mode values.
const (
	// ModeShadow evaluates policy and emits telemetry without blocking requests.
	ModeShadow EngineMode = "shadow"
	// ModeEnforce evaluates policy and enforces deny/quota decisions.
	ModeEnforce EngineMode = "enforce"
)

// FailureMode controls behavior when policy evaluation fails.
type FailureMode string

// Failure mode values.
const (
	// FailureModeOpen allows requests when policy evaluation fails.
	FailureModeOpen FailureMode = "fail-open"
	// FailureModeClosed rejects requests when policy evaluation fails.
	FailureModeClosed FailureMode = "fail-closed"
)

// Decision is the final policy outcome.
type Decision string

// Decision values.
const (
	// DecisionAllow allows request execution.
	DecisionAllow Decision = "allow"
	// DecisionDeny blocks request execution.
	DecisionDeny Decision = "deny"
	// DecisionQuotaReject blocks request execution due to quota limits.
	DecisionQuotaReject Decision = "quota_reject"
	// DecisionFallback marks runtime fallback behavior on policy evaluation failure.
	DecisionFallback Decision = "fallback"
)

// ActionType is the action configured in a rule.
type ActionType string

// Action type values.
const (
	// ActionAllow produces an allow decision.
	ActionAllow ActionType = "allow"
	// ActionDeny produces a deny decision.
	ActionDeny ActionType = "deny"
	// ActionQuota runs quota checks and then allows/rejects.
	ActionQuota ActionType = "quota"
)

// DefaultDecision controls behavior when no rule matches.
type DefaultDecision string

// Default decision values.
const (
	// DefaultDecisionAllow allows requests when no rule matches.
	DefaultDecisionAllow DefaultDecision = "allow"
	// DefaultDecisionDeny denies requests when no rule matches.
	DefaultDecisionDeny DefaultDecision = "deny"
)

// RuleSource represents rule file schema.
type RuleSource struct {
	Version string `json:"version" yaml:"version"`
	Rules   []Rule `json:"rules" yaml:"rules"`
}

// Rule defines a single policy rule.
type Rule struct {
	ID        string        `json:"id" yaml:"id"`
	Enabled   bool          `json:"enabled" yaml:"enabled"`
	Priority  int           `json:"priority" yaml:"priority"`
	Selectors RuleSelectors `json:"selectors" yaml:"selectors"`
	Action    RuleAction    `json:"action" yaml:"action"`
}

// RuleSelectors constrains matching dimensions.
type RuleSelectors struct {
	SourceCIDRs   []string `json:"source_cidrs,omitempty" yaml:"source_cidrs,omitempty"`
	PathPrefix    string   `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty"`
	Methods       []string `json:"methods,omitempty" yaml:"methods,omitempty"`
	Backends      []string `json:"backends,omitempty" yaml:"backends,omitempty"`
	Models        []string `json:"models,omitempty" yaml:"models,omitempty"`
	SubjectKeyIDs []string `json:"subject_key_ids,omitempty" yaml:"subject_key_ids,omitempty"`
	TenantIDs     []string `json:"tenant_ids,omitempty" yaml:"tenant_ids,omitempty"`
}

// RuleAction configures allow/deny/quota behavior.
type RuleAction struct {
	Type  ActionType `json:"type" yaml:"type"`
	Quota *QuotaSpec `json:"quota,omitempty" yaml:"quota,omitempty"`
}

// QuotaSpec defines scoped quota checks.
type QuotaSpec struct {
	RatePerMinute int64    `json:"rate_per_minute,omitempty" yaml:"rate_per_minute,omitempty"`
	Concurrency   int64    `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	TokenBudget   int64    `json:"token_budget,omitempty" yaml:"token_budget,omitempty"`
	WindowSeconds int64    `json:"window_seconds,omitempty" yaml:"window_seconds,omitempty"`
	Scopes        []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// RuntimeConfig controls middleware behavior at runtime.
type RuntimeConfig struct {
	Mode            EngineMode
	FailureMode     FailureMode
	ExplainEnabled  bool
	DefaultDecision DefaultDecision
	QuotaStore      string
}

// EvalInput is normalized request context used by evaluator.
type EvalInput struct {
	Method       string
	Path         string
	SourceIP     string
	Backend      string
	Model        string
	SubjectKeyID string
	TenantID     string
}

// EvalResult contains deterministic rule evaluation output.
type EvalResult struct {
	RuleID         string
	MatchedRuleIDs []string
	Action         ActionType
	Decision       Decision
	Reason         string
	DecisionPath   []string
	Quota          *QuotaSpec
}

// DecisionOutcome is returned by the middleware engine.
type DecisionOutcome struct {
	Decision       Decision
	Reason         string
	DecisionID     string
	RequestID      string
	MatchedRuleIDs []string
	DecisionPath   []string
	Release        func()
	Fallback       bool
}

// QuotaOperation identifies quota store operations.
type QuotaOperation string

// Quota operation values.
const (
	// QuotaOperationReserve reserves quota units for in-flight work.
	QuotaOperationReserve QuotaOperation = "reserve"
	// QuotaOperationConsume consumes quota units.
	QuotaOperationConsume QuotaOperation = "consume"
	// QuotaOperationRelease releases previously reserved quota units.
	QuotaOperationRelease QuotaOperation = "release"
)

// QuotaKind identifies the quota dimension.
type QuotaKind string

// Quota kind values.
const (
	// QuotaKindRate represents request-rate limits.
	QuotaKindRate QuotaKind = "rate"
	// QuotaKindConcurrency represents in-flight concurrency limits.
	QuotaKindConcurrency QuotaKind = "concurrency"
	// QuotaKindToken represents token-budget limits.
	QuotaKindToken QuotaKind = "token"
)

// Quota scope values.
const (
	quotaScopeKey     = "key"
	quotaScopeTenant  = "tenant"
	quotaScopeSource  = "source"
	quotaScopeBackend = "backend"
	quotaScopeModel   = "model"
	quotaScopeGlobal  = "global"
)

// Common policy reason values.
const (
	reasonPolicyAllow       = "policy_allow"
	reasonPolicyDeny        = "policy_deny"
	reasonPolicyQuotaCheck  = "policy_quota_check"
	reasonPolicyInvalid     = "policy_invalid_action"
	reasonNoMatchAllow      = "no_match_default_allow"
	reasonNoMatchDeny       = "no_match_default_deny"
	reasonQuotaChecksPassed = "quota_checks_passed" //nolint:gosec // policy reason code, not a credential
	reasonRateExceeded      = "rate_exceeded"
	reasonConcurrency       = "concurrency_exceeded"
	reasonTokenExceeded     = "token_budget_exceeded" //nolint:gosec // policy reason code, not a credential
	reasonEngineError       = "engine_error"
)

// QuotaRequest is a generic quota store request.
type QuotaRequest struct {
	Operation QuotaOperation
	Kind      QuotaKind
	Key       string
	Limit     int64
	Amount    int64
	Window    time.Duration
}

// QuotaResult is a generic quota store result.
type QuotaResult struct {
	Allowed   bool
	Remaining int64
	ResetAt   time.Time
}

package policy

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// Evaluate returns deterministic rule outcome for normalized request input.
func (cp *CompiledPolicy) Evaluate(input *EvalInput) EvalResult {
	if input == nil {
		input = &EvalInput{}
	}
	input.Method = strings.ToUpper(strings.TrimSpace(input.Method))
	input.Path = strings.TrimSpace(input.Path)
	input.Backend = strings.ToLower(strings.TrimSpace(input.Backend))
	input.Model = strings.ToLower(strings.TrimSpace(input.Model))
	input.SubjectKeyID = strings.ToLower(strings.TrimSpace(input.SubjectKeyID))
	input.TenantID = strings.ToLower(strings.TrimSpace(input.TenantID))
	input.SourceIP = strings.TrimSpace(input.SourceIP)

	for _, rule := range cp.rules {
		if !rule.enabled {
			continue
		}
		if !matchesRule(&rule.selector, input) {
			continue
		}

		result := EvalResult{
			RuleID:         rule.id,
			MatchedRuleIDs: []string{rule.id},
			Action:         rule.action.Type,
			DecisionPath: []string{
				"matched:" + rule.id,
				"action:" + string(rule.action.Type),
			},
			Quota: rule.action.Quota,
		}

		switch rule.action.Type {
		case ActionAllow:
			result.Decision = DecisionAllow
			result.Reason = reasonPolicyAllow
		case ActionDeny:
			result.Decision = DecisionDeny
			result.Reason = reasonPolicyDeny
		case ActionQuota:
			result.Decision = DecisionAllow
			result.Reason = reasonPolicyQuotaCheck
		default:
			result.Decision = DecisionDeny
			result.Reason = reasonPolicyInvalid
		}
		return result
	}

	if cp.defaultDecision == DefaultDecisionDeny {
		return EvalResult{
			Decision:       DecisionDeny,
			Reason:         reasonNoMatchDeny,
			DecisionPath:   []string{"no_match", "default:deny"},
			MatchedRuleIDs: []string{},
		}
	}

	return EvalResult{
		Decision:       DecisionAllow,
		Reason:         reasonNoMatchAllow,
		DecisionPath:   []string{"no_match", "default:allow"},
		MatchedRuleIDs: []string{},
	}
}

func matchesRule(selector *compiledSelector, input *EvalInput) bool {
	if selector == nil || input == nil {
		return false
	}
	if selector.pathPrefix != "" && !strings.HasPrefix(input.Path, selector.pathPrefix) {
		return false
	}
	if len(selector.methods) > 0 {
		if _, ok := selector.methods[input.Method]; !ok {
			return false
		}
	}
	if len(selector.backends) > 0 {
		if _, ok := selector.backends[input.Backend]; !ok {
			return false
		}
	}
	if len(selector.models) > 0 {
		if _, ok := selector.models[input.Model]; !ok {
			return false
		}
	}
	if len(selector.subjectIDs) > 0 {
		if _, ok := selector.subjectIDs[input.SubjectKeyID]; !ok {
			return false
		}
	}
	if len(selector.tenantIDs) > 0 {
		if _, ok := selector.tenantIDs[input.TenantID]; !ok {
			return false
		}
	}
	if len(selector.sourceCIDRs) > 0 {
		parsed := net.ParseIP(input.SourceIP)
		if parsed == nil {
			return false
		}
		matchedCIDR := false
		for _, cidr := range selector.sourceCIDRs {
			if cidr.Contains(parsed) {
				matchedCIDR = true
				break
			}
		}
		if !matchedCIDR {
			return false
		}
	}
	return true
}

func buildQuotaKey(ruleID string, input *EvalInput, spec *QuotaSpec) string {
	if input == nil {
		input = &EvalInput{}
	}
	scopes := spec.Scopes
	if len(scopes) == 0 {
		scopes = []string{quotaScopeKey, quotaScopeTenant, quotaScopeSource, quotaScopeBackend, quotaScopeModel}
	}

	parts := []string{"rule=" + ruleID}
	for _, scope := range scopes {
		scope = strings.ToLower(strings.TrimSpace(scope))
		switch scope {
		case quotaScopeKey:
			if input.SubjectKeyID != "" {
				parts = append(parts, quotaScopeKey+"="+input.SubjectKeyID)
			}
		case quotaScopeTenant:
			if input.TenantID != "" {
				parts = append(parts, quotaScopeTenant+"="+input.TenantID)
			}
		case quotaScopeSource:
			if input.SourceIP != "" {
				parts = append(parts, quotaScopeSource+"="+input.SourceIP)
			}
		case quotaScopeBackend:
			if input.Backend != "" {
				parts = append(parts, quotaScopeBackend+"="+input.Backend)
			}
		case quotaScopeModel:
			if input.Model != "" {
				parts = append(parts, quotaScopeModel+"="+input.Model)
			}
		}
	}
	if len(parts) == 1 {
		parts = append(parts, quotaScopeGlobal)
	}
	return strings.Join(parts, "|")
}

func windowDuration(spec *QuotaSpec) time.Duration {
	if spec == nil || spec.WindowSeconds <= 0 {
		return time.Minute
	}
	return time.Duration(spec.WindowSeconds) * time.Second
}

func policyModeOrDefault(mode EngineMode) EngineMode {
	if mode == ModeEnforce {
		return ModeEnforce
	}
	return ModeShadow
}

func failureModeOrDefault(mode FailureMode) FailureMode {
	if mode == FailureModeClosed {
		return FailureModeClosed
	}
	return FailureModeOpen
}

func normalizeQuotaReason(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	switch raw {
	case reasonRateExceeded, reasonConcurrency, reasonTokenExceeded:
		return raw
	default:
		return "quota_rejected"
	}
}

func normalizeDecisionReason(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "unspecified"
	}
	return raw
}

func validateRuntimeConfig(cfg RuntimeConfig) error {
	if cfg.Mode != ModeShadow && cfg.Mode != ModeEnforce {
		return fmt.Errorf("invalid policy mode %q", cfg.Mode)
	}
	if cfg.FailureMode != FailureModeOpen && cfg.FailureMode != FailureModeClosed {
		return fmt.Errorf("invalid failure mode %q", cfg.FailureMode)
	}
	if cfg.DefaultDecision != DefaultDecisionAllow && cfg.DefaultDecision != DefaultDecisionDeny {
		return fmt.Errorf("invalid default decision %q", cfg.DefaultDecision)
	}
	return nil
}

package policy

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

type compiledSelector struct {
	sourceCIDRs []*net.IPNet
	pathPrefix  string
	methods     map[string]struct{}
	backends    map[string]struct{}
	models      map[string]struct{}
	subjectIDs  map[string]struct{}
	tenantIDs   map[string]struct{}
}

type compiledRule struct {
	id          string
	enabled     bool
	priority    int
	specificity int
	selector    compiledSelector
	action      RuleAction
}

// CompiledPolicy is the immutable runtime representation of policy rules.
type CompiledPolicy struct {
	version         string
	defaultDecision DefaultDecision
	rules           []compiledRule
}

// Compile builds a deterministic immutable rule set.
func Compile(source *RuleSource, defaultDecision DefaultDecision) (*CompiledPolicy, error) {
	if source == nil {
		return nil, fmt.Errorf("nil policy source")
	}
	if defaultDecision != DefaultDecisionAllow && defaultDecision != DefaultDecisionDeny {
		return nil, fmt.Errorf("invalid default decision %q", defaultDecision)
	}

	seen := make(map[string]struct{}, len(source.Rules))
	compiled := make([]compiledRule, 0, len(source.Rules))

	for i, rule := range source.Rules {
		ruleID := strings.TrimSpace(rule.ID)
		if ruleID == "" {
			return nil, fmt.Errorf("rule[%d]: id is required", i)
		}
		if _, exists := seen[ruleID]; exists {
			return nil, fmt.Errorf("rule[%d]: duplicate id %q", i, ruleID)
		}
		seen[ruleID] = struct{}{}

		if rule.Action.Type != ActionAllow && rule.Action.Type != ActionDeny && rule.Action.Type != ActionQuota {
			return nil, fmt.Errorf("rule[%s]: invalid action type %q", ruleID, rule.Action.Type)
		}

		selector, specificity, err := compileSelector(&rule.Selectors)
		if err != nil {
			return nil, fmt.Errorf("rule[%s]: %w", ruleID, err)
		}

		if rule.Action.Type == ActionQuota {
			if err := validateQuotaSpec(rule.Action.Quota); err != nil {
				return nil, fmt.Errorf("rule[%s]: %w", ruleID, err)
			}
		}

		compiled = append(compiled, compiledRule{
			id:          ruleID,
			enabled:     rule.Enabled,
			priority:    rule.Priority,
			specificity: specificity,
			selector:    selector,
			action:      rule.Action,
		})
	}

	sort.Slice(compiled, func(i, j int) bool {
		if compiled[i].priority != compiled[j].priority {
			return compiled[i].priority > compiled[j].priority
		}
		if compiled[i].specificity != compiled[j].specificity {
			return compiled[i].specificity > compiled[j].specificity
		}
		return compiled[i].id < compiled[j].id
	})

	return &CompiledPolicy{
		version:         source.Version,
		defaultDecision: defaultDecision,
		rules:           compiled,
	}, nil
}

func compileSelector(sel *RuleSelectors) (compiledSelector, int, error) {
	if sel == nil {
		sel = &RuleSelectors{}
	}
	compiled := compiledSelector{
		pathPrefix: strings.TrimSpace(sel.PathPrefix),
		methods:    setFromSlice(sel.Methods, strings.ToUpper),
		backends:   setFromSlice(sel.Backends, strings.ToLower),
		models:     setFromSlice(sel.Models, strings.ToLower),
		subjectIDs: setFromSlice(sel.SubjectKeyIDs, strings.ToLower),
		tenantIDs:  setFromSlice(sel.TenantIDs, strings.ToLower),
	}

	var specificity int
	if compiled.pathPrefix != "" {
		specificity++
	}
	if len(compiled.methods) > 0 {
		specificity++
	}
	if len(compiled.backends) > 0 {
		specificity++
	}
	if len(compiled.models) > 0 {
		specificity++
	}
	if len(compiled.subjectIDs) > 0 {
		specificity++
	}
	if len(compiled.tenantIDs) > 0 {
		specificity++
	}

	for _, cidr := range sel.SourceCIDRs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return compiledSelector{}, 0, fmt.Errorf("invalid source cidr %q", cidr)
		}
		compiled.sourceCIDRs = append(compiled.sourceCIDRs, network)
	}
	if len(compiled.sourceCIDRs) > 0 {
		specificity++
	}

	return compiled, specificity, nil
}

func validateQuotaSpec(spec *QuotaSpec) error {
	if spec == nil {
		return fmt.Errorf("quota action requires quota config")
	}
	if spec.RatePerMinute <= 0 && spec.Concurrency <= 0 && spec.TokenBudget <= 0 {
		return fmt.Errorf("quota config must set at least one positive limit")
	}
	if spec.WindowSeconds < 0 {
		return fmt.Errorf("quota window_seconds must be non-negative")
	}
	for _, scope := range spec.Scopes {
		switch strings.ToLower(strings.TrimSpace(scope)) {
		case "", quotaScopeKey, quotaScopeTenant, quotaScopeSource, quotaScopeBackend, quotaScopeModel:
		default:
			return fmt.Errorf("unsupported quota scope %q", scope)
		}
	}
	return nil
}

func setFromSlice(input []string, normalize func(string) string) map[string]struct{} {
	set := make(map[string]struct{}, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if normalize != nil {
			item = normalize(item)
		}
		set[item] = struct{}{}
	}
	return set
}

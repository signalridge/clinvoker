package policy

import "testing"

func TestEvaluate_NoMatchDefault(t *testing.T) {
	source := &RuleSource{
		Version: "v1",
		Rules:   []Rule{},
	}

	compiledAllow, err := Compile(source, DefaultDecisionAllow)
	if err != nil {
		t.Fatalf("Compile(default allow) error = %v", err)
	}
	allowInput := EvalInput{Method: "GET", Path: "/no-match"}
	allowResult := compiledAllow.Evaluate(&allowInput)
	if allowResult.Decision != DecisionAllow || allowResult.Reason != "no_match_default_allow" {
		t.Fatalf("allow default result = %+v, want decision=allow reason=no_match_default_allow", allowResult)
	}

	compiledDeny, err := Compile(source, DefaultDecisionDeny)
	if err != nil {
		t.Fatalf("Compile(default deny) error = %v", err)
	}
	denyInput := EvalInput{Method: "GET", Path: "/no-match"}
	denyResult := compiledDeny.Evaluate(&denyInput)
	if denyResult.Decision != DecisionDeny || denyResult.Reason != "no_match_default_deny" {
		t.Fatalf("deny default result = %+v, want decision=deny reason=no_match_default_deny", denyResult)
	}
}

func TestEvaluate_ConflictTieBreakByID_IsDeterministic(t *testing.T) {
	source := &RuleSource{
		Version: "v1",
		Rules: []Rule{
			{
				ID:       "b-deny",
				Enabled:  true,
				Priority: 10,
				Selectors: RuleSelectors{
					PathPrefix: "/api",
					Methods:    []string{"GET"},
				},
				Action: RuleAction{Type: ActionDeny},
			},
			{
				ID:       "a-allow",
				Enabled:  true,
				Priority: 10,
				Selectors: RuleSelectors{
					PathPrefix: "/api",
					Methods:    []string{"GET"},
				},
				Action: RuleAction{Type: ActionAllow},
			},
		},
	}

	compiled, err := Compile(source, DefaultDecisionDeny)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	input := EvalInput{
		Method: "GET",
		Path:   "/api/v1/prompt",
	}

	for i := 0; i < 10; i++ {
		result := compiled.Evaluate(&input)
		if result.RuleID != "a-allow" {
			t.Fatalf("iteration %d: RuleID = %q, want %q", i, result.RuleID, "a-allow")
		}
		if result.Decision != DecisionAllow {
			t.Fatalf("iteration %d: Decision = %q, want %q", i, result.Decision, DecisionAllow)
		}
	}
}

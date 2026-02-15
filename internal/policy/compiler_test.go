package policy

import (
	"reflect"
	"testing"
)

func TestCompile_DeterministicOrdering(t *testing.T) {
	source := &RuleSource{
		Version: "v1",
		Rules: []Rule{
			{
				ID:       "z-low-specific",
				Enabled:  true,
				Priority: 10,
				Selectors: RuleSelectors{
					PathPrefix: "/api",
				},
				Action: RuleAction{Type: ActionAllow},
			},
			{
				ID:       "b-high-specific",
				Enabled:  true,
				Priority: 10,
				Selectors: RuleSelectors{
					PathPrefix: "/api",
					Methods:    []string{"GET"},
				},
				Action: RuleAction{Type: ActionAllow},
			},
			{
				ID:       "a-high-specific",
				Enabled:  true,
				Priority: 10,
				Selectors: RuleSelectors{
					PathPrefix: "/api",
					Methods:    []string{"POST"},
				},
				Action: RuleAction{Type: ActionAllow},
			},
			{
				ID:       "p-high-priority",
				Enabled:  true,
				Priority: 20,
				Action:   RuleAction{Type: ActionAllow},
			},
		},
	}

	compiled, err := Compile(source, DefaultDecisionAllow)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	got := make([]string, 0, len(compiled.rules))
	for _, rule := range compiled.rules {
		got = append(got, rule.id)
	}

	want := []string{
		"p-high-priority",
		"a-high-specific",
		"b-high-specific",
		"z-low-specific",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("compiled order = %v, want %v", got, want)
	}
}

func TestCompile_RejectsInvalidCIDR(t *testing.T) {
	source := &RuleSource{
		Version: "v1",
		Rules: []Rule{
			{
				ID:      "invalid-cidr",
				Enabled: true,
				Selectors: RuleSelectors{
					SourceCIDRs: []string{"not-a-cidr"},
				},
				Action: RuleAction{Type: ActionAllow},
			},
		},
	}

	_, err := Compile(source, DefaultDecisionAllow)
	if err == nil {
		t.Fatal("Compile() should fail for invalid CIDR selector")
	}
}

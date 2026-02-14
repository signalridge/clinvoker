package config

import "testing"

func TestValidateUnifiedFlags_CommandTimeoutNonNegative(t *testing.T) {
	errs := validateUnifiedFlags(&UnifiedFlagsConfig{
		CommandTimeoutSecs: -1,
	})
	if len(errs) == 0 {
		t.Fatal("expected validation error for negative command_timeout_secs")
	}

	found := false
	for _, err := range errs {
		verr, ok := err.(*ValidationError)
		if !ok {
			continue
		}
		if verr.Field == "unified_flags.command_timeout_secs" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected validation error for unified_flags.command_timeout_secs")
	}
}

func TestValidateUnifiedFlags_CommandTimeoutZeroAllowed(t *testing.T) {
	errs := validateUnifiedFlags(&UnifiedFlagsConfig{
		CommandTimeoutSecs: 0,
	})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
}

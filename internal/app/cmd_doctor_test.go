package app

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
)

func TestRunDoctorChecks_InvalidConfigIncludesFail(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	cfg := config.Get()
	cfg.DefaultBackend = "invalid-backend"

	report := runDoctorChecks(cfg)
	if report.Summary.Fail == 0 {
		t.Fatalf("expected at least one failing check, got summary=%+v", report.Summary)
	}
}

func TestRunDoctor_JSONOutput(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	doctorJSON = true
	t.Cleanup(func() { doctorJSON = false })

	output := captureStdout(t, func() {
		if err := runDoctor(doctorCmd, nil); err != nil {
			t.Fatalf("runDoctor failed: %v", err)
		}
	})

	var report doctorReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Fatalf("failed to parse doctor json output: %v\noutput=%s", err, output)
	}
	if report.SchemaVersion != doctorSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", report.SchemaVersion, doctorSchemaVersion)
	}
	if len(report.Checks) == 0 {
		t.Fatal("doctor report should contain checks")
	}
}

func TestRunDoctor_ReturnsValidationFailedWhenFailingChecksExist(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	cfg := config.Get()
	cfg.DefaultBackend = "invalid-backend"

	err := runDoctor(doctorCmd, nil)
	if err == nil {
		t.Fatal("expected runDoctor to return error for invalid configuration")
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
}

package config

import (
	"errors"
	"strings"
	"testing"

	apperrors "github.com/signalridge/clinvoker/internal/errors"
)

func testValidConfig() *Config {
	enabled := true
	return &Config{
		DefaultBackend: "claude",
		UnifiedFlags: UnifiedFlagsConfig{
			ApprovalMode:       "default",
			SandboxMode:        "default",
			MaxTurns:           1,
			MaxTokens:          1,
			CommandTimeoutSecs: 1,
		},
		Backends: map[string]BackendConfig{
			"claude": {
				Model:        "test-model",
				ApprovalMode: "default",
				SandboxMode:  "default",
				Enabled:      &enabled,
			},
		},
		Session: SessionConfig{
			RetentionDays: 7,
		},
		Output: OutputConfig{
			Format: "json",
		},
		Server: ServerConfig{
			Host:                 "127.0.0.1",
			Port:                 8080,
			RequestTimeoutSecs:   30,
			ReadTimeoutSecs:      10,
			WriteTimeoutSecs:     30,
			IdleTimeoutSecs:      60,
			RateLimitEnabled:     true,
			RateLimitRPS:         5,
			RateLimitBurst:       10,
			RateLimitCleanupSecs: 60,
			MaxRequestBodyBytes:  1024,
		},
		Parallel: ParallelConfig{
			MaxWorkers: 2,
		},
	}
}

func hasValidationField(errs []error, field string) bool {
	for _, err := range errs {
		var verr *ValidationError
		if errors.As(err, &verr) && verr.Field == field {
			return true
		}
	}
	return false
}

func TestValidationError_Error(t *testing.T) {
	err := (&ValidationError{Field: "server.port", Message: "must be positive"}).Error()
	want := "config validation: server.port: must be positive"
	if err != want {
		t.Fatalf("Error() = %q, want %q", err, want)
	}
}

func TestValidateDefaultBackend(t *testing.T) {
	t.Run("empty backend", func(t *testing.T) {
		err := validateDefaultBackend("")
		if err == nil {
			t.Fatal("expected error for empty default backend")
		}
		if !hasValidationField([]error{err}, "default_backend") {
			t.Fatal("expected default_backend field validation error")
		}
	})

	t.Run("invalid backend", func(t *testing.T) {
		err := validateDefaultBackend("unknown")
		if err == nil {
			t.Fatal("expected error for invalid default backend")
		}
		if !hasValidationField([]error{err}, "default_backend") {
			t.Fatal("expected default_backend field validation error")
		}
	})

	t.Run("valid backend", func(t *testing.T) {
		for _, backend := range []string{"claude", "codex", "gemini"} {
			if err := validateDefaultBackend(backend); err != nil {
				t.Fatalf("validateDefaultBackend(%q) returned error: %v", backend, err)
			}
		}
	})
}

func TestValidateUnifiedFlags(t *testing.T) {
	t.Run("all valid", func(t *testing.T) {
		errs := validateUnifiedFlags(&UnifiedFlagsConfig{
			ApprovalMode:       "auto",
			SandboxMode:        "workspace",
			MaxTurns:           1,
			MaxTokens:          10,
			CommandTimeoutSecs: 60,
		})
		if len(errs) != 0 {
			t.Fatalf("expected no errors, got %d", len(errs))
		}
	})

	t.Run("invalid values", func(t *testing.T) {
		errs := validateUnifiedFlags(&UnifiedFlagsConfig{
			ApprovalMode:       "bad",
			SandboxMode:        "bad",
			MaxTurns:           -1,
			MaxTokens:          -1,
			CommandTimeoutSecs: -1,
		})
		expected := []string{
			"unified_flags.approval_mode",
			"unified_flags.sandbox_mode",
			"unified_flags.max_turns",
			"unified_flags.max_tokens",
			"unified_flags.command_timeout_secs",
		}
		for _, field := range expected {
			if !hasValidationField(errs, field) {
				t.Fatalf("expected validation error for field %q, errs=%v", field, errs)
			}
		}
	})
}

func TestValidateOutputConfig(t *testing.T) {
	t.Run("nil output", func(t *testing.T) {
		if errs := validateOutputConfig(nil); len(errs) != 0 {
			t.Fatalf("expected no errors for nil output, got %d", len(errs))
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		errs := validateOutputConfig(&OutputConfig{Format: "yaml"})
		if !hasValidationField(errs, "output.format") {
			t.Fatalf("expected output.format validation error, errs=%v", errs)
		}
	})

	t.Run("valid format", func(t *testing.T) {
		for _, format := range []string{"default", "text", "json", "stream-json"} {
			if errs := validateOutputConfig(&OutputConfig{Format: format}); len(errs) != 0 {
				t.Fatalf("expected no errors for format %q, got %d", format, len(errs))
			}
		}
	})
}

func TestValidateBackendConfig(t *testing.T) {
	t.Run("invalid backend modes", func(t *testing.T) {
		errs := validateBackendConfig("claude", &BackendConfig{
			ApprovalMode: "bad",
			SandboxMode:  "bad",
		})
		if !hasValidationField(errs, "backends.claude.approval_mode") {
			t.Fatalf("expected backends.claude.approval_mode error, errs=%v", errs)
		}
		if !hasValidationField(errs, "backends.claude.sandbox_mode") {
			t.Fatalf("expected backends.claude.sandbox_mode error, errs=%v", errs)
		}
	})

	t.Run("valid backend modes", func(t *testing.T) {
		errs := validateBackendConfig("codex", &BackendConfig{
			ApprovalMode: "always",
			SandboxMode:  "full",
		})
		if len(errs) != 0 {
			t.Fatalf("expected no validation errors, got %d", len(errs))
		}
	})
}

func TestValidateSessionConfig(t *testing.T) {
	if errs := validateSessionConfig(&SessionConfig{RetentionDays: 1}); len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %d", len(errs))
	}

	errs := validateSessionConfig(&SessionConfig{RetentionDays: -1})
	if !hasValidationField(errs, "session.retention_days") {
		t.Fatalf("expected session.retention_days validation error, errs=%v", errs)
	}
}

func TestValidateServerConfig(t *testing.T) {
	cases := []struct {
		name   string
		server ServerConfig
		fields []string
	}{
		{
			name: "invalid host",
			server: ServerConfig{
				Host: "bad host!",
				Port: 8080,
			},
			fields: []string{"server.host"},
		},
		{
			name: "invalid port range",
			server: ServerConfig{
				Host: "localhost",
				Port: 70000,
			},
			fields: []string{"server.port"},
		},
		{
			name: "negative timeouts",
			server: ServerConfig{
				Host:               "localhost",
				Port:               8080,
				RequestTimeoutSecs: -1,
				ReadTimeoutSecs:    -1,
				WriteTimeoutSecs:   -1,
				IdleTimeoutSecs:    -1,
			},
			fields: []string{
				"server.request_timeout_secs",
				"server.read_timeout_secs",
				"server.write_timeout_secs",
				"server.idle_timeout_secs",
			},
		},
		{
			name: "invalid rate limits",
			server: ServerConfig{
				Host:                 "localhost",
				Port:                 8080,
				RateLimitEnabled:     true,
				RateLimitRPS:         0,
				RateLimitBurst:       -1,
				RateLimitCleanupSecs: -1,
			},
			fields: []string{
				"server.rate_limit_rps",
				"server.rate_limit_burst",
				"server.rate_limit_cleanup_secs",
			},
		},
		{
			name: "invalid max request body",
			server: ServerConfig{
				Host:                "localhost",
				Port:                8080,
				MaxRequestBodyBytes: -1,
			},
			fields: []string{"server.max_request_body_bytes"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := validateServerConfig(&tc.server)
			for _, field := range tc.fields {
				if !hasValidationField(errs, field) {
					t.Fatalf("expected validation error for %q, errs=%v", field, errs)
				}
			}
		})
	}

	t.Run("valid server config", func(t *testing.T) {
		errs := validateServerConfig(&ServerConfig{
			Host:                 "localhost",
			Port:                 8080,
			RequestTimeoutSecs:   1,
			ReadTimeoutSecs:      1,
			WriteTimeoutSecs:     1,
			IdleTimeoutSecs:      1,
			RateLimitEnabled:     true,
			RateLimitRPS:         1,
			RateLimitBurst:       1,
			RateLimitCleanupSecs: 1,
			MaxRequestBodyBytes:  1,
		})
		if len(errs) != 0 {
			t.Fatalf("expected no validation errors, got %d", len(errs))
		}
	})
}

func TestValidateServerConfigPolicy(t *testing.T) {
	t.Run("invalid policy enum values", func(t *testing.T) {
		errs := validateServerConfig(&ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
			Policy: PolicyConfig{
				Mode:            "bad-mode",
				FailureMode:     "bad-failure",
				DefaultDecision: "bad-default",
				QuotaStore:      "bad-store",
			},
		})

		expected := []string{
			"server.policy.mode",
			"server.policy.failure_mode",
			"server.policy.default_decision",
			"server.policy.quota_store",
		}
		for _, field := range expected {
			if !hasValidationField(errs, field) {
				t.Fatalf("expected validation error for %q, errs=%v", field, errs)
			}
		}
	})

	t.Run("enabled policy requires rules file", func(t *testing.T) {
		errs := validateServerConfig(&ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
			Policy: PolicyConfig{
				Enabled:         true,
				Mode:            "shadow",
				FailureMode:     "fail-open",
				DefaultDecision: "allow",
				QuotaStore:      "memory",
				RulesFile:       "",
			},
		})
		if !hasValidationField(errs, "server.policy.rules_file") {
			t.Fatalf("expected server.policy.rules_file validation error, errs=%v", errs)
		}
	})

	t.Run("valid policy config", func(t *testing.T) {
		errs := validateServerConfig(&ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
			Policy: PolicyConfig{
				Enabled:         true,
				Mode:            "enforce",
				FailureMode:     "fail-closed",
				DefaultDecision: "deny",
				QuotaStore:      "shared",
				RulesFile:       "/tmp/policy.yaml",
				ExplainEnabled:  true,
			},
		})
		if len(errs) != 0 {
			t.Fatalf("expected no validation errors, got %v", errs)
		}
	})
}

func TestValidateParallelConfig(t *testing.T) {
	if errs := validateParallelConfig(&ParallelConfig{MaxWorkers: 1}); len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %d", len(errs))
	}

	errs := validateParallelConfig(&ParallelConfig{MaxWorkers: 0})
	if !hasValidationField(errs, "parallel.max_workers") {
		t.Fatalf("expected parallel.max_workers validation error, errs=%v", errs)
	}
}

func TestIsValidHostname(t *testing.T) {
	tooLong := strings.Repeat("a", 254)
	tests := []struct {
		host string
		want bool
	}{
		{"localhost", true},
		{"example.com", true},
		{"1.2.3.4", true},
		{"", false},
		{"bad host", false},
		{"host!", false},
		{tooLong, false},
	}

	for _, tt := range tests {
		got := isValidHostname(tt.host)
		if got != tt.want {
			t.Fatalf("isValidHostname(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

func TestValidate_AggregatesErrorsAcrossSections(t *testing.T) {
	cfg := testValidConfig()
	cfg.DefaultBackend = ""
	cfg.UnifiedFlags.MaxTurns = -1
	cfg.Backends["claude"] = BackendConfig{ApprovalMode: "bad"}
	cfg.Session.RetentionDays = -1
	cfg.Output.Format = "bad"
	cfg.Server.Port = -1
	cfg.Parallel.MaxWorkers = 0

	errs := Validate(cfg)
	expectedFields := []string{
		"default_backend",
		"unified_flags.max_turns",
		"backends.claude.approval_mode",
		"session.retention_days",
		"output.format",
		"server.port",
		"parallel.max_workers",
	}

	for _, field := range expectedFields {
		if !hasValidationField(errs, field) {
			t.Fatalf("expected aggregated validation error for %q, errs=%v", field, errs)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	t.Run("returns nil for valid config", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)
		if err := Init(""); err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		cfg := Get()
		*cfg = *testValidConfig()

		if err := ValidateConfig(); err != nil {
			t.Fatalf("expected nil error for valid config, got %v", err)
		}
	})

	t.Run("returns AppError for invalid config", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)
		if err := Init(""); err != nil {
			t.Fatalf("Init failed: %v", err)
		}

		cfg := Get()
		*cfg = *testValidConfig()
		cfg.DefaultBackend = ""
		cfg.Server.Port = -1

		err := ValidateConfig()
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !apperrors.IsCode(err, apperrors.ErrCodeConfigInvalid) {
			t.Fatalf("expected ErrCodeConfigInvalid, got %v", apperrors.GetCode(err))
		}
		if !strings.Contains(err.Error(), "configuration has") {
			t.Fatalf("expected aggregated validation message, got %q", err.Error())
		}
	})
}

func TestValidateBackendConfigPublic(t *testing.T) {
	err := ValidateBackendConfig("claude", &BackendConfig{
		ApprovalMode: "bad",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !apperrors.IsCode(err, apperrors.ErrCodeConfigInvalid) {
		t.Fatalf("expected ErrCodeConfigInvalid, got %v", apperrors.GetCode(err))
	}
}

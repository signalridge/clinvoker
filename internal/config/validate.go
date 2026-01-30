package config

import (
	"fmt"
	"net"
	"strings"

	apperrors "github.com/signalridge/clinvoker/internal/errors"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation: %s: %s", e.Field, e.Message)
}

// Validate validates the entire configuration and returns all errors found.
func Validate(cfg *Config) []error {
	var errs []error

	// Validate default backend
	if err := validateDefaultBackend(cfg.DefaultBackend); err != nil {
		errs = append(errs, err)
	}

	// Validate unified flags
	errs = append(errs, validateUnifiedFlags(&cfg.UnifiedFlags)...)

	// Validate backend configs
	for name, bc := range cfg.Backends {
		errs = append(errs, validateBackendConfig(name, &bc)...)
	}

	// Validate session config
	errs = append(errs, validateSessionConfig(&cfg.Session)...)

	// Validate output config
	errs = append(errs, validateOutputConfig(&cfg.Output)...)

	// Validate server config
	errs = append(errs, validateServerConfig(&cfg.Server)...)

	// Validate parallel config
	errs = append(errs, validateParallelConfig(&cfg.Parallel)...)

	return errs
}

// ValidateConfig validates the current configuration.
// Returns an AppError with all validation errors.
func ValidateConfig() error {
	cfg := Get()
	errs := Validate(cfg)
	if len(errs) == 0 {
		return nil
	}

	// Combine all errors into one
	messages := make([]string, len(errs))
	for i, err := range errs {
		messages[i] = err.Error()
	}

	return apperrors.New(apperrors.ErrCodeConfigInvalid,
		fmt.Sprintf("configuration has %d error(s): %s", len(errs), strings.Join(messages, "; ")))
}

// validateDefaultBackend validates the default backend setting.
func validateDefaultBackend(backend string) error {
	validBackends := map[string]bool{
		"claude": true,
		"codex":  true,
		"gemini": true,
	}

	if backend == "" {
		return &ValidationError{
			Field:   "default_backend",
			Message: "must not be empty",
		}
	}

	if !validBackends[backend] {
		return &ValidationError{
			Field:   "default_backend",
			Message: fmt.Sprintf("invalid backend %q (valid: claude, codex, gemini)", backend),
		}
	}

	return nil
}

// validateUnifiedFlags validates unified flag settings.
func validateUnifiedFlags(flags *UnifiedFlagsConfig) []error {
	var errs []error

	// Validate approval mode
	if flags.ApprovalMode != "" {
		validModes := map[string]bool{
			"default": true,
			"auto":    true,
			"none":    true,
			"always":  true,
		}
		if !validModes[flags.ApprovalMode] {
			errs = append(errs, &ValidationError{
				Field:   "unified_flags.approval_mode",
				Message: fmt.Sprintf("invalid mode %q (valid: default, auto, none, always)", flags.ApprovalMode),
			})
		}
	}

	// Validate sandbox mode
	if flags.SandboxMode != "" {
		validModes := map[string]bool{
			"default":   true,
			"read-only": true,
			"workspace": true,
			"full":      true,
		}
		if !validModes[flags.SandboxMode] {
			errs = append(errs, &ValidationError{
				Field:   "unified_flags.sandbox_mode",
				Message: fmt.Sprintf("invalid mode %q (valid: default, read-only, workspace, full)", flags.SandboxMode),
			})
		}
	}

	// Validate max_turns
	if flags.MaxTurns < 0 {
		errs = append(errs, &ValidationError{
			Field:   "unified_flags.max_turns",
			Message: "must be non-negative",
		})
	}

	// Validate max_tokens
	if flags.MaxTokens < 0 {
		errs = append(errs, &ValidationError{
			Field:   "unified_flags.max_tokens",
			Message: "must be non-negative",
		})
	}

	return errs
}

// validateOutputConfig validates output configuration.
func validateOutputConfig(output *OutputConfig) []error {
	var errs []error

	if output == nil {
		return errs
	}

	if output.Format != "" {
		validFormats := map[string]bool{
			"default":     true,
			"text":        true,
			"json":        true,
			"stream-json": true,
		}
		if !validFormats[output.Format] {
			errs = append(errs, &ValidationError{
				Field:   "output.format",
				Message: fmt.Sprintf("invalid format %q (valid: default, text, json, stream-json)", output.Format),
			})
		}
	}

	return errs
}

// validateBackendConfig validates a backend-specific configuration.
func validateBackendConfig(name string, bc *BackendConfig) []error {
	var errs []error

	// Validate approval mode if set
	if bc.ApprovalMode != "" {
		validModes := map[string]bool{
			"default": true,
			"auto":    true,
			"none":    true,
			"always":  true,
		}
		if !validModes[bc.ApprovalMode] {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("backends.%s.approval_mode", name),
				Message: fmt.Sprintf("invalid mode %q", bc.ApprovalMode),
			})
		}
	}

	// Validate sandbox mode if set
	if bc.SandboxMode != "" {
		validModes := map[string]bool{
			"default":   true,
			"read-only": true,
			"workspace": true,
			"full":      true,
		}
		if !validModes[bc.SandboxMode] {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("backends.%s.sandbox_mode", name),
				Message: fmt.Sprintf("invalid mode %q", bc.SandboxMode),
			})
		}
	}

	return errs
}

// validateSessionConfig validates session configuration.
func validateSessionConfig(session *SessionConfig) []error {
	var errs []error

	if session.RetentionDays < 0 {
		errs = append(errs, &ValidationError{
			Field:   "session.retention_days",
			Message: "must be non-negative",
		})
	}

	return errs
}

// validateServerConfig validates server configuration.
func validateServerConfig(server *ServerConfig) []error {
	var errs []error

	// Validate host
	if server.Host != "" {
		if ip := net.ParseIP(server.Host); ip == nil && server.Host != "localhost" {
			// Check if it's a valid hostname
			if !isValidHostname(server.Host) {
				errs = append(errs, &ValidationError{
					Field:   "server.host",
					Message: fmt.Sprintf("invalid host %q", server.Host),
				})
			}
		}
	}

	// Validate port
	if server.Port < 0 || server.Port > 65535 {
		errs = append(errs, &ValidationError{
			Field:   "server.port",
			Message: fmt.Sprintf("port must be between 0 and 65535, got %d", server.Port),
		})
	}

	// Validate timeouts
	if server.RequestTimeoutSecs < 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.request_timeout_secs",
			Message: "must be non-negative",
		})
	}

	if server.ReadTimeoutSecs < 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.read_timeout_secs",
			Message: "must be non-negative",
		})
	}

	if server.WriteTimeoutSecs < 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.write_timeout_secs",
			Message: "must be non-negative",
		})
	}

	if server.IdleTimeoutSecs < 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.idle_timeout_secs",
			Message: "must be non-negative",
		})
	}

	// Validate rate limit settings
	if server.RateLimitRPS < 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.rate_limit_rps",
			Message: "must be non-negative",
		})
	}

	if server.RateLimitBurst < 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.rate_limit_burst",
			Message: "must be non-negative",
		})
	}

	if server.RateLimitCleanupSecs < 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.rate_limit_cleanup_secs",
			Message: "must be non-negative",
		})
	}

	if server.RateLimitEnabled && server.RateLimitRPS == 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.rate_limit_rps",
			Message: "must be positive when rate limiting is enabled",
		})
	}

	if server.MaxRequestBodyBytes < 0 {
		errs = append(errs, &ValidationError{
			Field:   "server.max_request_body_bytes",
			Message: "must be non-negative",
		})
	}

	return errs
}

// validateParallelConfig validates parallel execution configuration.
func validateParallelConfig(parallel *ParallelConfig) []error {
	var errs []error

	if parallel.MaxWorkers < 1 {
		errs = append(errs, &ValidationError{
			Field:   "parallel.max_workers",
			Message: "must be at least 1",
		})
	}

	return errs
}

// isValidHostname checks if a string is a valid hostname.
func isValidHostname(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}

	// Simple validation: check for invalid characters
	for _, c := range host {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '.') {
			return false
		}
	}

	return true
}

// ValidateBackendConfig validates a single backend configuration.
func ValidateBackendConfig(name string, bc *BackendConfig) error {
	errs := validateBackendConfig(name, bc)
	if len(errs) == 0 {
		return nil
	}

	messages := make([]string, len(errs))
	for i, err := range errs {
		messages[i] = err.Error()
	}

	return apperrors.New(apperrors.ErrCodeConfigInvalid,
		fmt.Sprintf("backend %q configuration invalid: %s", name, strings.Join(messages, "; ")))
}

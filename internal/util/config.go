package util

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
)

// Track which backends we've already warned about allowed_tools support.
// This prevents spamming the log with repeated warnings.
var (
	warnedBackends   = make(map[string]bool)
	warnedBackendsMu sync.Mutex
)

// ApplyUnifiedDefaults applies unified flag defaults from config to options.
// This ensures consistent behavior between CLI and server when config defaults are set.
func ApplyUnifiedDefaults(opts *backend.UnifiedOptions, cfg *config.Config, effectiveDryRun bool) {
	if opts == nil || cfg == nil {
		return
	}

	if opts.ApprovalMode == "" && cfg.UnifiedFlags.ApprovalMode != "" {
		opts.ApprovalMode = backend.ApprovalMode(cfg.UnifiedFlags.ApprovalMode)
	}
	if opts.SandboxMode == "" && cfg.UnifiedFlags.SandboxMode != "" {
		opts.SandboxMode = backend.SandboxMode(cfg.UnifiedFlags.SandboxMode)
	}
	if opts.MaxTurns == 0 && cfg.UnifiedFlags.MaxTurns > 0 {
		opts.MaxTurns = cfg.UnifiedFlags.MaxTurns
	}
	if opts.MaxTokens == 0 && cfg.UnifiedFlags.MaxTokens > 0 {
		opts.MaxTokens = cfg.UnifiedFlags.MaxTokens
	}
	if !opts.Verbose && cfg.UnifiedFlags.Verbose {
		opts.Verbose = true
	}
	if !opts.DryRun && effectiveDryRun {
		opts.DryRun = true
	}
}

// ApplyOutputFormatDefault applies output format defaults with the following priority:
// 1) Current value (if explicitly set and not "default")
// 2) output.format (if set and not "default")
// 3) built-in default ("json")
// Returns the effective output format.
func ApplyOutputFormatDefault(current string, cfg *config.Config) string {
	current = strings.ToLower(strings.TrimSpace(current))
	if current != "" && current != string(backend.OutputDefault) {
		return current
	}
	if cfg != nil {
		if cfg.Output.Format != "" && cfg.Output.Format != string(backend.OutputDefault) {
			return strings.ToLower(cfg.Output.Format)
		}
	}
	return string(backend.OutputJSON)
}

// InternalOutputFormat determines the internal output format to use.
// For text output, we use JSON internally to capture session ID and parse response.
func InternalOutputFormat(requested backend.OutputFormat) backend.OutputFormat {
	if requested == "" || requested == backend.OutputDefault || requested == backend.OutputText {
		return backend.OutputJSON
	}
	return requested
}

// ApplyBackendDefaults applies backend-specific config defaults to options.
// This should be called after ApplyUnifiedDefaults to allow backend config to override unified config.
func ApplyBackendDefaults(opts *backend.UnifiedOptions, backendName string, cfg *config.Config) {
	if opts == nil || cfg == nil || backendName == "" {
		return
	}

	bc, ok := cfg.Backends[backendName]
	if !ok {
		return
	}

	// Backend-specific approval mode overrides unified
	if opts.ApprovalMode == "" || opts.ApprovalMode == backend.ApprovalDefault {
		if bc.ApprovalMode != "" {
			opts.ApprovalMode = backend.ApprovalMode(bc.ApprovalMode)
		}
	}

	// Backend-specific sandbox mode overrides unified
	if opts.SandboxMode == "" || opts.SandboxMode == backend.SandboxDefault {
		if bc.SandboxMode != "" {
			opts.SandboxMode = backend.SandboxMode(bc.SandboxMode)
		}
	}

	// Backend-specific system prompt (if not already set)
	if opts.SystemPrompt == "" && bc.SystemPrompt != "" {
		opts.SystemPrompt = bc.SystemPrompt
	}

	// Backend-specific allowed tools (if not already set)
	if opts.AllowedTools == "" && bc.AllowedTools != "" {
		opts.AllowedTools = bc.AllowedTools
	}

	// Warn if allowed_tools is set for a backend that doesn't support it
	if opts.AllowedTools != "" && backendName != "claude" {
		warnedBackendsMu.Lock()
		if !warnedBackends[backendName] {
			warnedBackends[backendName] = true
			warnedBackendsMu.Unlock()
			slog.Warn("allowed_tools option is only supported by Claude backend; ignoring for this backend",
				"backend", backendName,
				"allowed_tools", opts.AllowedTools)
		} else {
			warnedBackendsMu.Unlock()
		}
	}

	// Backend-specific extra flags (append to existing)
	if len(bc.ExtraFlags) > 0 {
		opts.ExtraFlags = append(opts.ExtraFlags, bc.ExtraFlags...)
	}
}

package util

import (
	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
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

// ApplyOutputFormatDefault applies the output format default from config if not set.
// Returns the effective output format.
func ApplyOutputFormatDefault(current string, cfg *config.Config) string {
	if current != "" && current != string(backend.OutputDefault) {
		return current
	}
	if cfg != nil && cfg.UnifiedFlags.OutputFormat != "" && cfg.UnifiedFlags.OutputFormat != string(backend.OutputDefault) {
		return cfg.UnifiedFlags.OutputFormat
	}
	return current
}

// InternalOutputFormat determines the internal output format to use.
// For text output, we use JSON internally to capture session ID and parse response.
func InternalOutputFormat(requested backend.OutputFormat) backend.OutputFormat {
	if requested == "" || requested == backend.OutputDefault || requested == backend.OutputText {
		return backend.OutputJSON
	}
	return requested
}

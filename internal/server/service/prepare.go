package service

import (
	"fmt"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/util"
)

// validOutputFormats contains the set of recognized output format values.
var validOutputFormats = map[string]bool{
	"":            true, // Empty string is valid (uses default)
	"default":     true,
	"text":        true,
	"json":        true,
	"stream-json": true,
}

// validateOutputFormat returns an error if the format is not recognized.
func validateOutputFormat(format string) error {
	if !validOutputFormats[format] {
		return fmt.Errorf("invalid output_format %q: must be one of default, text, json, stream-json", format)
	}
	return nil
}

type preparedPrompt struct {
	backend backend.Backend
	model   string
	opts    *backend.UnifiedOptions
	// requestedFormat captures the requested output format (after config defaults).
	requestedFormat backend.OutputFormat
}

func preparePrompt(req *PromptRequest, forceStateless bool) (*preparedPrompt, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	// Validate output format before processing
	if err := validateOutputFormat(req.OutputFormat); err != nil {
		return nil, err
	}

	// Use ValidateWorkDirFromConfig to enforce allowed/blocked path restrictions from config
	if err := ValidateWorkDirFromConfig(req.WorkDir); err != nil {
		return nil, err
	}

	b, err := backend.Get(req.Backend)
	if err != nil {
		return nil, err
	}

	// Use backend-specific flag validation for stricter isolation
	// This prevents cross-backend flag injection (e.g., codex flags passed to claude)
	if err := backend.ValidateExtraFlagsForBackend(req.Backend, req.Extra); err != nil {
		return nil, err
	}
	if !b.IsAvailable() {
		return nil, fmt.Errorf("backend %q is not available", req.Backend)
	}

	cfg := config.Get()
	model := req.Model
	if model == "" {
		if bcfg, ok := cfg.Backends[req.Backend]; ok {
			model = bcfg.Model
		}
	}

	requestedFormat := backend.OutputFormat(util.ApplyOutputFormatDefault(req.OutputFormat, cfg))

	opts := &backend.UnifiedOptions{
		WorkDir:      req.WorkDir,
		Model:        model,
		ApprovalMode: backend.ApprovalMode(req.ApprovalMode),
		SandboxMode:  backend.SandboxMode(req.SandboxMode),
		MaxTokens:    req.MaxTokens,
		MaxTurns:     req.MaxTurns,
		SystemPrompt: req.SystemPrompt,
		Verbose:      req.Verbose,
		DryRun:       req.DryRun,
		Ephemeral:    req.Ephemeral,
		ExtraFlags:   req.Extra,
	}

	util.ApplyUnifiedDefaults(opts, cfg, cfg.UnifiedFlags.DryRun)
	util.ApplyBackendDefaults(opts, req.Backend, cfg)

	if forceStateless {
		opts.Ephemeral = true
	}

	return &preparedPrompt{
		backend:         b,
		model:           model,
		opts:            opts,
		requestedFormat: requestedFormat,
	}, nil
}

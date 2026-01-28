package service

import (
	"fmt"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/util"
)

type preparedPrompt struct {
	backend backend.Backend
	model   string
	opts    *backend.UnifiedOptions
}

func preparePrompt(req *PromptRequest, forceStateless bool) (*preparedPrompt, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	if err := validateWorkDir(req.WorkDir); err != nil {
		return nil, err
	}

	if err := backend.ValidateExtraFlags(req.Extra); err != nil {
		return nil, err
	}

	b, err := backend.Get(req.Backend)
	if err != nil {
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

	if forceStateless {
		opts.Ephemeral = true
	}

	return &preparedPrompt{
		backend: b,
		model:   model,
		opts:    opts,
	}, nil
}

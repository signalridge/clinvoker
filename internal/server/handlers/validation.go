package handlers

import (
	"errors"
	"fmt"
	"strings"
)

// ValidatePromptRequest applies the same validation rules as the REST handler.
//
//nolint:gocritic // request is intentionally passed by value to avoid accidental mutation
func ValidatePromptRequest(req PromptRequest) error {
	if req.Backend == "" {
		return errors.New("backend is required")
	}
	if req.Prompt == "" {
		return errors.New("prompt is required")
	}
	return nil
}

// ValidateParallelRequest applies the same validation rules as the REST handler.
func ValidateParallelRequest(req ParallelRequest) error {
	if len(req.Tasks) == 0 {
		return errors.New("tasks are required")
	}
	return nil
}

// ValidateChainRequest applies the same validation rules as the REST handler.
func ValidateChainRequest(req ChainRequest) error {
	if len(req.Steps) == 0 {
		return errors.New("steps are required")
	}
	if req.PassSessionID || req.PersistSessions {
		return errors.New("chain is always ephemeral; pass_session_id and persist_sessions are not supported")
	}
	for i, step := range req.Steps {
		if strings.Contains(step.Prompt, "{{session}}") {
			return fmt.Errorf("chain step %d uses {{session}} but sessions are not persisted", i+1)
		}
	}
	return nil
}

// ValidateCompareRequest applies the same validation rules as the REST handler.
//
//nolint:gocritic // request is intentionally passed by value to avoid accidental mutation
func ValidateCompareRequest(req CompareRequest) error {
	if len(req.Backends) == 0 {
		return errors.New("backends are required")
	}
	if req.Prompt == "" {
		return errors.New("prompt is required")
	}
	return nil
}

package backend

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// Gemini implements the Backend interface for Gemini CLI.
type Gemini struct{}

// Name returns the backend identifier.
func (g *Gemini) Name() string {
	return "gemini"
}

// IsAvailable checks if Gemini CLI is installed.
func (g *Gemini) IsAvailable() bool {
	_, err := exec.LookPath("gemini")
	return err == nil
}

// BuildCommand creates an exec.Cmd for running a prompt with Gemini CLI.
// Note: --output-format should be added via opts.ExtraFlags if needed.
func (g *Gemini) BuildCommand(prompt string, opts *Options) *exec.Cmd {
	var args []string

	// Check if --output-format is already in ExtraFlags
	hasOutputFormat := false
	if opts != nil {
		for _, f := range opts.ExtraFlags {
			if f == "--output-format" || strings.HasPrefix(f, "--output-format=") {
				hasOutputFormat = true
				break
			}
		}
	}

	// Add --output-format text by default (unless already specified)
	if !hasOutputFormat {
		args = append(args, "--output-format", "text")
	}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	args = append(args, prompt)

	cmd := exec.Command("gemini", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// ResumeCommand creates an exec.Cmd for resuming a Gemini session.
// Note: --output-format should be added via opts.ExtraFlags if needed.
func (g *Gemini) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
	var args []string

	// Check if --output-format is already in ExtraFlags
	hasOutputFormat := false
	if opts != nil {
		for _, f := range opts.ExtraFlags {
			if f == "--output-format" || strings.HasPrefix(f, "--output-format=") {
				hasOutputFormat = true
				break
			}
		}
	}

	// Add --output-format text by default (unless already specified)
	if !hasOutputFormat {
		args = append(args, "--output-format", "text")
	}

	args = append(args, "--resume", sessionID)

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	if prompt != "" {
		args = append(args, prompt)
	}

	cmd := exec.Command("gemini", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// BuildCommandUnified creates an exec.Cmd using unified options.
func (g *Gemini) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
	return g.BuildCommand(prompt, MapFromUnified(g.Name(), opts))
}

// ResumeCommandUnified creates a resume exec.Cmd using unified options.
func (g *Gemini) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
	return g.ResumeCommand(sessionID, prompt, MapFromUnified(g.Name(), opts))
}

// ParseOutput returns the output as-is since Gemini with --output-format text produces clean output.
func (g *Gemini) ParseOutput(rawOutput string) string {
	return rawOutput
}

// geminiJSONResponse represents Gemini's JSON output format.
type geminiJSONResponse struct {
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
	Stats     struct {
		Models map[string]struct {
			Tokens struct {
				Input      int `json:"input"`
				Candidates int `json:"candidates"`
				Total      int `json:"total"`
			} `json:"tokens"`
		} `json:"models"`
	} `json:"stats"`
}

// ParseJSONResponse parses Gemini's JSON output into a unified response.
func (g *Gemini) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
	// Gemini may prepend "Loaded cached credentials." message, skip it
	cleanOutput := rawOutput
	if idx := strings.Index(rawOutput, "{"); idx > 0 {
		cleanOutput = rawOutput[idx:]
	}

	var resp geminiJSONResponse
	if err := json.Unmarshal([]byte(cleanOutput), &resp); err != nil {
		return nil, err
	}

	// Calculate total tokens from all models
	var usage TokenUsage
	for _, modelStats := range resp.Stats.Models {
		usage.InputTokens += modelStats.Tokens.Input
		usage.OutputTokens += modelStats.Tokens.Candidates
		usage.TotalTokens += modelStats.Tokens.Total
	}

	// Store raw response
	var raw map[string]any
	_ = json.Unmarshal([]byte(cleanOutput), &raw)

	return &UnifiedResponse{
		Content:   resp.Response,
		SessionID: resp.SessionID,
		Usage:     &usage,
		Raw:       raw,
	}, nil
}

// SeparateStderr returns true since Gemini outputs credential messages to stderr
// that should be filtered out for clean output.
func (g *Gemini) SeparateStderr() bool {
	return true
}

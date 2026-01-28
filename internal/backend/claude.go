package backend

import (
	"encoding/json"
	"os/exec"
)

// Claude implements the Backend interface for Claude Code CLI.
type Claude struct{}

// Name returns the backend identifier.
func (c *Claude) Name() string {
	return "claude"
}

// IsAvailable checks if Claude Code CLI is installed.
func (c *Claude) IsAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// BuildCommand creates an exec.Cmd for running a prompt with Claude Code.
func (c *Claude) BuildCommand(prompt string, opts *Options) *exec.Cmd {
	args := []string{"--print"}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		if opts.AllowedTools != "" {
			args = append(args, "--allowedTools", opts.AllowedTools)
		}
		for _, dir := range opts.AllowedDirs {
			args = append(args, "--add-dir", dir)
		}
		args = append(args, opts.ExtraFlags...)
	}

	args = append(args, prompt)

	cmd := exec.Command("claude", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// ResumeCommand creates an exec.Cmd for resuming a Claude Code session.
func (c *Claude) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
	args := []string{"--resume", sessionID, "--print"}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	if prompt != "" {
		args = append(args, prompt)
	}

	cmd := exec.Command("claude", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// BuildCommandUnified creates an exec.Cmd using unified options.
func (c *Claude) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
	return c.BuildCommand(prompt, MapFromUnified(c.Name(), opts))
}

// ResumeCommandUnified creates a resume exec.Cmd using unified options.
func (c *Claude) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
	return c.ResumeCommand(sessionID, prompt, MapFromUnified(c.Name(), opts))
}

// ParseOutput returns the output as-is since Claude with --print already produces clean output.
func (c *Claude) ParseOutput(rawOutput string) string {
	return rawOutput
}

// claudeJSONResponse represents Claude's JSON output format.
type claudeJSONResponse struct {
	Type       string `json:"type"`
	Result     string `json:"result"`
	SessionID  string `json:"session_id"`
	DurationMs int64  `json:"duration_ms"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// claudeErrorResponse represents Claude's error output format.
type claudeErrorResponse struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Error     string `json:"error"`
	SessionID string `json:"session_id,omitempty"`
}

// ParseJSONResponse parses Claude's JSON output into a unified response.
func (c *Claude) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
	// First try to parse as error response
	var errResp claudeErrorResponse
	if err := json.Unmarshal([]byte(rawOutput), &errResp); err == nil {
		// Check if it's an error response (has error or message field with error type)
		if errResp.Error != "" {
			return &UnifiedResponse{
				SessionID: errResp.SessionID,
				Error:     errResp.Error,
			}, nil
		}
		if errResp.Type == "error" && errResp.Message != "" {
			return &UnifiedResponse{
				SessionID: errResp.SessionID,
				Error:     errResp.Message,
			}, nil
		}
	}

	var resp claudeJSONResponse
	if err := json.Unmarshal([]byte(rawOutput), &resp); err != nil {
		// JSON parsing failed - this might be text output, let caller handle it
		return nil, err
	}

	// Store raw response
	var raw map[string]any
	_ = json.Unmarshal([]byte(rawOutput), &raw)

	return &UnifiedResponse{
		Content:    resp.Result,
		SessionID:  resp.SessionID,
		DurationMs: resp.DurationMs,
		Usage: &TokenUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
		Raw: raw,
	}, nil
}

// SeparateStderr returns false since Claude's stderr doesn't need filtering.
func (c *Claude) SeparateStderr() bool {
	return false
}

package backend

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// Codex implements the Backend interface for Codex CLI.
type Codex struct{}

// Name returns the backend identifier.
func (c *Codex) Name() string {
	return "codex"
}

// IsAvailable checks if Codex CLI is installed.
func (c *Codex) IsAvailable() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

// BuildCommand creates an exec.Cmd for running a prompt with Codex CLI.
// Uses 'codex exec' for non-interactive execution.
// Note: --json flag should be added via opts.ExtraFlags for JSON output.
func (c *Codex) BuildCommand(prompt string, opts *Options) *exec.Cmd {
	// Use 'exec' subcommand for non-interactive mode
	args := []string{"exec"}

	// Check if --json is already in ExtraFlags to avoid duplication
	hasJSON := false
	if opts != nil {
		for _, f := range opts.ExtraFlags {
			if f == "--json" {
				hasJSON = true
				break
			}
		}
	}

	// Add --json by default for parseable output (unless already present)
	if !hasJSON {
		args = append(args, "--json")
	}

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	args = append(args, prompt)

	cmd := exec.Command("codex", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// ResumeCommand creates an exec.Cmd for resuming a Codex session.
// Uses 'codex exec resume' for non-interactive execution.
func (c *Codex) ResumeCommand(sessionID, prompt string, opts *Options) *exec.Cmd {
	// Use 'exec resume' for non-interactive mode
	args := []string{"exec"}

	// Check if --json is already in ExtraFlags to avoid duplication
	hasJSON := false
	if opts != nil {
		for _, f := range opts.ExtraFlags {
			if f == "--json" {
				hasJSON = true
				break
			}
		}
	}

	// Add --json by default for parseable output (unless already present)
	if !hasJSON {
		args = append(args, "--json")
	}

	args = append(args, "resume", sessionID)

	if opts != nil {
		if opts.Model != "" {
			args = append(args, "--model", opts.Model)
		}
		args = append(args, opts.ExtraFlags...)
	}

	if prompt != "" {
		args = append(args, prompt)
	}

	cmd := exec.Command("codex", args...)
	if opts != nil && opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	return cmd
}

// BuildCommandUnified creates an exec.Cmd using unified options.
func (c *Codex) BuildCommandUnified(prompt string, opts *UnifiedOptions) *exec.Cmd {
	return c.BuildCommand(prompt, MapFromUnified(c.Name(), opts))
}

// ResumeCommandUnified creates a resume exec.Cmd using unified options.
func (c *Codex) ResumeCommandUnified(sessionID, prompt string, opts *UnifiedOptions) *exec.Cmd {
	return c.ResumeCommand(sessionID, prompt, MapFromUnified(c.Name(), opts))
}

// codexEvent represents a JSONL event from Codex CLI.
type codexEvent struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id,omitempty"`
	Item     struct {
		ID   string `json:"id,omitempty"`
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"item,omitempty"`
	Usage struct {
		InputTokens       int `json:"input_tokens"`
		CachedInputTokens int `json:"cached_input_tokens"`
		OutputTokens      int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

// ParseOutput extracts the agent message text from Codex JSONL output.
// Codex outputs events like: {"type":"item.completed","item":{"type":"agent_message","text":"..."}}
func (c *Codex) ParseOutput(rawOutput string) string {
	var messages []string

	lines := strings.Split(rawOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Extract text from item.completed events with agent_message type
		if event.Type == "item.completed" && event.Item.Type == "agent_message" && event.Item.Text != "" {
			messages = append(messages, event.Item.Text)
		}
	}

	return strings.Join(messages, "\n")
}

// ParseJSONResponse parses Codex's JSONL output into a unified response.
func (c *Codex) ParseJSONResponse(rawOutput string) (*UnifiedResponse, error) {
	var messages []string
	var sessionID string
	var usage TokenUsage
	var rawEvents []map[string]any

	lines := strings.Split(rawOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Store raw event
		var rawEvent map[string]any
		_ = json.Unmarshal([]byte(line), &rawEvent)
		rawEvents = append(rawEvents, rawEvent)

		switch event.Type {
		case "thread.started":
			sessionID = event.ThreadID
		case "item.completed":
			if event.Item.Type == "agent_message" && event.Item.Text != "" {
				messages = append(messages, event.Item.Text)
			}
		case "turn.completed":
			usage.InputTokens = event.Usage.InputTokens + event.Usage.CachedInputTokens
			usage.OutputTokens = event.Usage.OutputTokens
			usage.TotalTokens = usage.InputTokens + usage.OutputTokens
		}
	}

	return &UnifiedResponse{
		Content:   strings.Join(messages, "\n"),
		SessionID: sessionID,
		Usage:     &usage,
		Raw:       map[string]any{"events": rawEvents},
	}, nil
}

// SeparateStderr returns true to filter out Codex's internal debug messages
// (e.g., "failed to refresh available models" errors).
func (c *Codex) SeparateStderr() bool {
	return true
}

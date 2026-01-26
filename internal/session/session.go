// Package session provides session management for AI CLI interactions.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// SessionStatus represents the current status of a session.
type SessionStatus string

const (
	StatusActive    SessionStatus = "active"
	StatusCompleted SessionStatus = "completed"
	StatusError     SessionStatus = "error"
	StatusPaused    SessionStatus = "paused"
)

// TokenUsage tracks token consumption for a session.
type TokenUsage struct {
	InputTokens     int64 `json:"input_tokens"`
	OutputTokens    int64 `json:"output_tokens"`
	CachedTokens    int64 `json:"cached_tokens,omitempty"`
	ReasoningTokens int64 `json:"reasoning_tokens,omitempty"`
}

// Total returns the total tokens used.
func (t *TokenUsage) Total() int64 {
	return t.InputTokens + t.OutputTokens
}

// Session represents a CLI interaction session.
type Session struct {
	// ID is the unique session identifier.
	ID string `json:"id"`

	// Backend is the AI backend used (claude, codex, gemini).
	Backend string `json:"backend"`

	// CreatedAt is when the session was created.
	CreatedAt time.Time `json:"created_at"`

	// LastUsed is when the session was last used.
	LastUsed time.Time `json:"last_used"`

	// WorkingDir is the working directory for the session.
	WorkingDir string `json:"working_dir"`

	// BackendSessionID is the internal session ID from the backend (if any).
	BackendSessionID string `json:"backend_session_id,omitempty"`

	// Model is the AI model used for this session.
	Model string `json:"model,omitempty"`

	// InitialPrompt is the first prompt that started this session.
	InitialPrompt string `json:"initial_prompt,omitempty"`

	// Status is the current session status.
	Status SessionStatus `json:"status,omitempty"`

	// TurnCount is the number of conversation turns.
	TurnCount int `json:"turn_count,omitempty"`

	// TokenUsage tracks token consumption.
	TokenUsage *TokenUsage `json:"token_usage,omitempty"`

	// Tags are user-defined labels for organizing sessions.
	Tags []string `json:"tags,omitempty"`

	// Title is an optional human-readable title for the session.
	Title string `json:"title,omitempty"`

	// ParentID links to a parent session (for forked sessions).
	ParentID string `json:"parent_id,omitempty"`

	// ErrorMessage stores the last error if status is error.
	ErrorMessage string `json:"error_message,omitempty"`

	// Metadata contains additional session metadata.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SessionOptions provides optional parameters for creating a session.
type SessionOptions struct {
	Model         string
	InitialPrompt string
	Title         string
	Tags          []string
	ParentID      string
}

// NewSession creates a new session with a generated ID.
func NewSession(backend, workDir string) (*Session, error) {
	return NewSessionWithOptions(backend, workDir, nil)
}

// NewSessionWithOptions creates a new session with additional options.
func NewSessionWithOptions(backend, workDir string, opts *SessionOptions) (*Session, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	sess := &Session{
		ID:         id,
		Backend:    backend,
		CreatedAt:  now,
		LastUsed:   now,
		WorkingDir: workDir,
		Status:     StatusActive,
		TokenUsage: &TokenUsage{},
		Metadata:   make(map[string]string),
	}

	if opts != nil {
		sess.Model = opts.Model
		sess.InitialPrompt = opts.InitialPrompt
		sess.Title = opts.Title
		sess.Tags = opts.Tags
		sess.ParentID = opts.ParentID
	}

	return sess, nil
}

// MarkUsed updates the last used timestamp.
func (s *Session) MarkUsed() {
	s.LastUsed = time.Now()
}

// Age returns the duration since the session was created.
func (s *Session) Age() time.Duration {
	return time.Since(s.CreatedAt)
}

// IdleDuration returns the duration since the session was last used.
func (s *Session) IdleDuration() time.Duration {
	return time.Since(s.LastUsed)
}

// SetBackendSessionID sets the backend's internal session ID.
func (s *Session) SetBackendSessionID(id string) {
	s.BackendSessionID = id
}

// SetMetadata sets a metadata key-value pair.
func (s *Session) SetMetadata(key, value string) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}
	s.Metadata[key] = value
}

// AddTokens adds token usage to the session.
func (s *Session) AddTokens(input, output int64) {
	if s.TokenUsage == nil {
		s.TokenUsage = &TokenUsage{}
	}
	s.TokenUsage.InputTokens += input
	s.TokenUsage.OutputTokens += output
}

// AddCachedTokens adds cached token count.
func (s *Session) AddCachedTokens(cached int64) {
	if s.TokenUsage == nil {
		s.TokenUsage = &TokenUsage{}
	}
	s.TokenUsage.CachedTokens += cached
}

// AddReasoningTokens adds reasoning token count (for o3/o4 models).
func (s *Session) AddReasoningTokens(reasoning int64) {
	if s.TokenUsage == nil {
		s.TokenUsage = &TokenUsage{}
	}
	s.TokenUsage.ReasoningTokens += reasoning
}

// IncrementTurn increments the turn count.
func (s *Session) IncrementTurn() {
	s.TurnCount++
}

// SetStatus sets the session status.
func (s *Session) SetStatus(status SessionStatus) {
	s.Status = status
}

// SetError sets the session to error status with a message.
func (s *Session) SetError(msg string) {
	s.Status = StatusError
	s.ErrorMessage = msg
}

// Complete marks the session as completed.
func (s *Session) Complete() {
	s.Status = StatusCompleted
}

// AddTag adds a tag to the session.
func (s *Session) AddTag(tag string) {
	for _, t := range s.Tags {
		if t == tag {
			return // Already exists
		}
	}
	s.Tags = append(s.Tags, tag)
}

// RemoveTag removes a tag from the session.
func (s *Session) RemoveTag(tag string) {
	for i, t := range s.Tags {
		if t == tag {
			s.Tags = append(s.Tags[:i], s.Tags[i+1:]...)
			return
		}
	}
}

// HasTag checks if the session has a specific tag.
func (s *Session) HasTag(tag string) bool {
	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// SetTitle sets the session title.
func (s *Session) SetTitle(title string) {
	s.Title = title
}

// SetModel sets the model used for this session.
func (s *Session) SetModel(model string) {
	s.Model = model
}

// Fork creates a new session based on this one.
func (s *Session) Fork() (*Session, error) {
	newSess, err := NewSessionWithOptions(s.Backend, s.WorkingDir, &SessionOptions{
		Model:    s.Model,
		ParentID: s.ID,
		Tags:     append([]string{}, s.Tags...),
	})
	if err != nil {
		return nil, err
	}

	// Copy metadata
	for k, v := range s.Metadata {
		newSess.SetMetadata(k, v)
	}

	return newSess, nil
}

// DisplayName returns a human-readable name for the session.
func (s *Session) DisplayName() string {
	if s.Title != "" {
		return s.Title
	}
	if s.InitialPrompt != "" {
		// Truncate to first 50 chars
		prompt := s.InitialPrompt
		if len(prompt) > 50 {
			prompt = prompt[:47] + "..."
		}
		return prompt
	}
	return s.ID[:8] // Short ID
}

// generateID generates a random session ID.
func generateID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

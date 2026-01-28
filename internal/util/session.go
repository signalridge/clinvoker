package util

import (
	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/session"
)

// UpdateSessionFromResponse updates a session with execution results.
// It increments turn count, records token usage, and sets appropriate status.
func UpdateSessionFromResponse(sess *session.Session, exitCode int, errMsg string, resp *backend.UnifiedResponse) {
	if sess == nil {
		return
	}

	sess.IncrementTurn()

	if resp != nil && resp.Usage != nil {
		sess.AddTokens(int64(resp.Usage.InputTokens), int64(resp.Usage.OutputTokens))
	}

	if resp != nil && resp.Error != "" {
		sess.SetError(resp.Error)
		return
	}

	if exitCode == 0 {
		sess.Complete()
		return
	}

	if errMsg != "" {
		sess.SetError(errMsg)
		return
	}

	sess.SetError("backend execution failed")
}

// TokenUsageFromBackend converts backend.TokenUsage to session.TokenUsage.
// Returns nil if input is nil.
func TokenUsageFromBackend(usage *backend.TokenUsage) *session.TokenUsage {
	if usage == nil {
		return nil
	}
	return &session.TokenUsage{
		InputTokens:  int64(usage.InputTokens),
		OutputTokens: int64(usage.OutputTokens),
	}
}

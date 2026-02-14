package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
)

// staleSessionPatterns contains error message patterns indicating a stale session.
var staleSessionPatterns = []string{
	"no conversation found with session id",
	"session not found",
	"conversation not found",
	"invalid session",
	"session has expired",
}

// isStaleSessionError checks if an error message indicates a stale backend session.
func isStaleSessionError(errMsg string) bool {
	if errMsg == "" {
		return false
	}
	errLower := strings.ToLower(errMsg)
	for _, pattern := range staleSessionPatterns {
		if strings.Contains(errLower, pattern) {
			return true
		}
	}
	return false
}

// StaleSessionChoice represents user's choice when handling a stale session.
type StaleSessionChoice int

// StaleSessionChoice values for user selection.
const (
	StaleSessionNew    StaleSessionChoice = iota // Start a new session
	StaleSessionSelect                           // Select another session to resume
	StaleSessionCancel                           // Cancel operation
)

// promptStaleSessionChoice prompts user for action when session is stale.
func promptStaleSessionChoice(sess *session.Session) StaleSessionChoice {
	fmt.Printf("\nSession %s has expired (backend session no longer exists).\n",
		shortSessionID(sess.ID))
	fmt.Println("\nOptions:")
	fmt.Println("  [N] New    - Start a new session")
	fmt.Println("  [S] Select - Choose another session to resume")
	fmt.Println("  [C] Cancel - Abort operation")
	fmt.Print("\nEnter choice (N/S/C): ")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		// Treat empty input as cancel
		return StaleSessionCancel
	}

	switch strings.ToLower(strings.TrimSpace(input)) {
	case "n", "new":
		return StaleSessionNew
	case "s", "select":
		return StaleSessionSelect
	default:
		return StaleSessionCancel
	}
}

// staleSessionContext holds context for stale session handling.
type staleSessionContext struct {
	store   *session.Store
	sess    *session.Session
	prompt  string
	flags   *normalizedFlags
	backend backend.Backend
	opts    *backend.UnifiedOptions
}

// handleStaleSession handles the interactive recovery when a session is stale.
func handleStaleSession(ctx *staleSessionContext) error {
	// Mark current session as expired
	ctx.sess.MarkExpired()
	if err := ctx.store.Save(ctx.sess); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", err)
	}

	// Loop to handle user choices
	for {
		choice := promptStaleSessionChoice(ctx.sess)

		switch choice {
		case StaleSessionNew:
			return startNewSession(ctx)

		case StaleSessionSelect:
			if selectedSess := selectAndResumeSession(ctx); selectedSess != nil {
				// Successfully resumed another session
				return nil
			}
			// selectedSess == nil means user went back or no sessions available
			// Continue loop to show main options

		case StaleSessionCancel:
			return fmt.Errorf("operation canceled by user")
		}
	}
}

// selectAndResumeSession loops letting user select a session until success or cancel.
// Returns the successfully resumed session, or nil if user cancels/goes back.
func selectAndResumeSession(ctx *staleSessionContext) *session.Session {
	for {
		// Get resumable sessions excluding the current expired one
		sessions := getResumableSessionsExcluding(ctx.store, ctx.sess.ID)

		if len(sessions) == 0 {
			fmt.Println("\nNo other resumable sessions available.")
			return nil // Let caller handle (show main options)
		}

		// Display session list
		displaySessionList(sessions)

		// Get user selection
		idx := promptSessionSelection(len(sessions))
		if idx == 0 {
			return nil // User chose to go back
		}

		selectedSess := sessions[idx-1]
		fmt.Printf("\nAttempting to resume session %s...\n", shortSessionID(selectedSess.ID))

		// Try to resume the selected session
		result, err := attemptResumeSession(ctx, selectedSess)

		if isResumeSuccess(result, err) {
			// Success
			return selectedSess
		}

		if isStaleResumeFailure(result, err) {
			// Failed due to stale backend session - mark as expired and continue loop.
			fmt.Printf("Session %s also expired. Marked as expired.\n",
				shortSessionID(selectedSess.ID))
			selectedSess.MarkExpired()
			if saveErr := ctx.store.Save(selectedSess); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update session: %v\n", saveErr)
			}
			continue
		}

		// Non-stale failure: keep session state unchanged and let user choose again.
		fmt.Printf("Session %s failed to resume: %s\n",
			shortSessionID(selectedSess.ID), resumeFailureReason(result, err))
	}
}

// getResumableSessionsExcluding returns resumable sessions excluding the given ID.
func getResumableSessionsExcluding(store *session.Store, excludeID string) []*session.Session {
	sessions, _ := store.List()
	resumable := filterResumableSessions(sessions)

	result := make([]*session.Session, 0, len(resumable))
	for _, s := range resumable {
		if s.ID != excludeID {
			result = append(result, s)
		}
	}
	return result
}

// displaySessionList shows available sessions in a formatted table.
func displaySessionList(sessions []*session.Session) {
	fmt.Printf("\nAvailable sessions (%d):\n", len(sessions))
	fmt.Printf("  %-3s %-8s %-8s %-12s %s\n", "#", "ID", "BACKEND", "LAST USED", "TITLE/PROMPT")
	fmt.Println("  " + strings.Repeat("-", 60))

	for i, s := range sessions {
		if i >= maxSessionsDisplay {
			fmt.Printf("  ... and %d more\n", len(sessions)-maxSessionsDisplay)
			break
		}

		title := s.DisplayName()
		if len(title) > maxSessionTitleLen {
			title = title[:maxSessionTitleLen-3] + "..."
		}

		fmt.Printf("  %-3d %-8s %-8s %-12s %s\n",
			i+1,
			shortSessionID(s.ID),
			s.Backend,
			formatTimeAgo(s.LastUsed),
			title,
		)
	}
}

// promptSessionSelection prompts user to select a session number.
// Returns 0 if user chooses to go back, or a valid 1-based index.
func promptSessionSelection(maxIdx int) int {
	fmt.Print("\nEnter session number (or 0 to go back): ")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return 0
	}

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 0 || idx > maxIdx {
		return 0
	}
	return idx
}

// attemptResumeSession tries to resume a session and returns the result.
func attemptResumeSession(ctx *staleSessionContext, sess *session.Session) (*ExecutionResult, error) {
	targetBackend, targetOpts, err := buildResumeTarget(ctx, sess)
	if err != nil {
		return nil, err
	}

	// Build resume command
	execCmd := targetBackend.ResumeCommandUnified(sess.BackendSessionID, ctx.prompt, targetOpts)

	// Update session usage
	sess.MarkUsed()
	if err := ctx.store.Save(sess); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
	}

	// Execute
	userFormat := backend.OutputFormat(ctx.flags.outputFormat)
	execCfg := &ExecutionConfig{
		Backend:    targetBackend,
		Session:    sess,
		OutputMode: DetermineOutputMode(userFormat),
		Stdin:      true,
		Timeout:    GetCommandTimeout(),
	}

	result, err := ExecuteCommand(execCfg, execCmd)

	// Update session with result
	if result != nil {
		sess.MarkUsed()
		if result.SessionID != "" {
			sess.BackendSessionID = result.SessionID
		}
		if saveErr := ctx.store.Save(sess); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", saveErr)
		}
	}

	return result, err
}

func buildResumeTarget(ctx *staleSessionContext, sess *session.Session) (backend.Backend, *backend.UnifiedOptions, error) {
	if !config.IsBackendEnabled(sess.Backend) {
		return nil, nil, fmt.Errorf("backend %q is disabled in config", sess.Backend)
	}

	targetBackend, err := backend.Get(sess.Backend)
	if err != nil {
		return nil, nil, fmt.Errorf("backend error: %w", err)
	}

	if !ctx.flags.dryRun && !targetBackend.IsAvailable() {
		return nil, nil, fmt.Errorf("backend %q is not available", sess.Backend)
	}

	userFormat := backend.OutputFormat(ctx.flags.outputFormat)
	internalFormat := DetermineInternalFormat(userFormat)
	cfg := config.Get()
	opts := &backend.UnifiedOptions{
		WorkDir:      sess.WorkingDir,
		Model:        modelName,
		OutputFormat: internalFormat,
	}
	applyUnifiedDefaults(opts, cfg, ctx.flags.dryRun)
	applyBackendDefaults(opts, sess.Backend, cfg)

	if opts.Model == "" {
		if bcfg, ok := cfg.Backends[sess.Backend]; ok {
			opts.Model = bcfg.Model
		}
	}

	return targetBackend, opts, nil
}

func isResumeSuccess(result *ExecutionResult, err error) bool {
	if err != nil {
		return false
	}
	if result == nil {
		return true
	}
	return result.ExitCode == 0 && result.Error == ""
}

func isStaleResumeFailure(result *ExecutionResult, err error) bool {
	if result != nil && isStaleSessionError(result.Error) {
		return true
	}
	if err != nil && isStaleSessionError(err.Error()) {
		return true
	}
	return false
}

func resumeFailureReason(result *ExecutionResult, err error) string {
	if err != nil {
		return err.Error()
	}
	if result != nil {
		if result.Error != "" {
			return result.Error
		}
		return fmt.Sprintf("exit code %d", result.ExitCode)
	}
	return "unknown error"
}

// startNewSession creates and executes a new session.
func startNewSession(ctx *staleSessionContext) error {
	cfg := config.Get()

	// Create new session with parent link
	sessOpts := &session.SessionOptions{
		Model:         ctx.opts.Model,
		InitialPrompt: ctx.prompt,
		Tags:          cfg.Session.DefaultTags,
		ParentID:      ctx.sess.ID, // Link to expired session
	}

	newSess, err := ctx.store.CreateWithOptions(ctx.sess.Backend, ctx.sess.WorkingDir, sessOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create session: %v\n", err)
	}

	// Build command for new session
	execCmd := ctx.backend.BuildCommandUnified(ctx.prompt, ctx.opts)

	userFormat := backend.OutputFormat(ctx.flags.outputFormat)
	execCfg := &ExecutionConfig{
		Backend:    ctx.backend,
		Session:    newSess,
		OutputMode: DetermineOutputMode(userFormat),
		Stdin:      true,
		Timeout:    GetCommandTimeout(),
	}

	result, err := ExecuteCommand(execCfg, execCmd)

	// Update new session
	if newSess != nil {
		newSess.MarkUsed()
		if result != nil && result.SessionID != "" {
			newSess.BackendSessionID = result.SessionID
		}
		if saveErr := ctx.store.Save(newSess); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", saveErr)
		}
	}

	if err != nil {
		return err
	}

	if newSess != nil {
		fmt.Printf("\nNew session started: %s\n", shortSessionID(newSess.ID))
	}

	if result != nil && result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	return nil
}

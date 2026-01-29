package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
	"github.com/signalridge/clinvoker/internal/util"
)

// resumeCmd resumes a previous session.
var resumeCmd = &cobra.Command{
	Use:   "resume [session-id] [prompt]",
	Short: "Resume a previous session",
	Long: `Resume a previous AI CLI session.

Examples:
  clinvk resume abc123 "continue working"
  clinvk resume --last "follow up"
  clinvk resume --last
  clinvk resume --backend claude
  clinvk resume (interactive picker)`,
	Args: cobra.MaximumNArgs(2),
	RunE: runResume,
}

var (
	resumeLast        bool
	resumeBackend     string
	resumeWorkDir     bool
	resumeInteractive bool
)

func init() {
	resumeCmd.Flags().BoolVar(&resumeLast, "last", false, "resume the most recent session")
	resumeCmd.Flags().StringVarP(&resumeBackend, "backend", "b", "", "filter sessions by backend")
	resumeCmd.Flags().BoolVar(&resumeWorkDir, "here", false, "filter sessions by current working directory")
	resumeCmd.Flags().BoolVarP(&resumeInteractive, "interactive", "i", false, "show interactive session picker")
}

func runResume(cmd *cobra.Command, args []string) error {
	store := session.NewStore()
	cfg := config.Get()

	// Apply config default output format if flag not explicitly set
	if !cmd.Flags().Changed("output-format") {
		outputFormat = util.ApplyOutputFormatDefault(outputFormat, cfg)
	}

	var sess *session.Session
	var prompt string
	var err error

	// Build filter based on flags
	filter := &session.ListFilter{}
	if resumeBackend != "" {
		filter.Backend = resumeBackend
	}
	if resumeWorkDir {
		wd, err := os.Getwd()
		if err == nil {
			filter.WorkDir = wd
		}
	}

	if resumeLast {
		// Get the most recent session matching the filter
		sessions, err := store.ListWithFilter(filter)
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
		if len(sessions) == 0 {
			return fmt.Errorf("no sessions found matching criteria")
		}
		sess = sessions[0]
		if len(args) > 0 {
			prompt = args[0]
		}
	} else if len(args) > 0 {
		// Try to find session by ID or prefix
		sess, err = store.GetByPrefix(args[0])
		if err != nil {
			// Fall back to exact match
			sess, err = store.Get(args[0])
			if err != nil {
				return err
			}
		}
		if len(args) > 1 {
			prompt = args[1]
		}
	} else if resumeInteractive || (len(args) == 0 && !resumeLast) {
		// Interactive picker
		sess, err = interactiveSessionPicker(store, filter)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("session ID required (or use --last, --interactive)")
	}

	// Get backend
	b, err := backend.Get(sess.Backend)
	if err != nil {
		return fmt.Errorf("backend error: %w", err)
	}

	if !dryRun && !b.IsAvailable() {
		return fmt.Errorf("backend %q is not available", sess.Backend)
	}

	// Determine internal output format (use JSON internally for text to capture session ID)
	userOutputFormat := backend.OutputFormat(outputFormat)
	internalOutputFormat := userOutputFormat
	if userOutputFormat == backend.OutputText || userOutputFormat == backend.OutputDefault || userOutputFormat == "" {
		internalOutputFormat = backend.OutputJSON
	}

	// Build unified options
	opts := &backend.UnifiedOptions{
		WorkDir:      sess.WorkingDir,
		Model:        modelName,
		OutputFormat: internalOutputFormat,
	}

	if bcfg, ok := cfg.Backends[sess.Backend]; ok {
		if opts.Model == "" {
			opts.Model = bcfg.Model
		}
	}

	// Build resume command
	backendSessionID := sess.BackendSessionID
	if backendSessionID == "" {
		backendSessionID = sess.ID
	}
	execCmd := b.ResumeCommandUnified(backendSessionID, prompt, opts)

	if dryRun {
		fmt.Printf("Would resume session %s (%s)\n", shortSessionID(sess.ID), sess.Backend)
		fmt.Printf("Command: %s %v\n", execCmd.Path, execCmd.Args[1:])
		return nil
	}

	// Update session
	sess.MarkUsed()
	if err := store.Save(sess); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
	}

	// Execute based on output format
	var exitCode int
	switch userOutputFormat {
	case backend.OutputJSON:
		exitCode, _, err = executeWithJSONOutputAndCapture(b, execCmd, sess)
	case backend.OutputStreamJSON:
		exitCode, err = executeWithStreamOutput(b, execCmd)
	default:
		// Text output: use JSON internally, extract content for display
		exitCode, _, err = executeTextViaJSON(b, execCmd, sess)
	}

	if err != nil {
		return err
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// interactiveSessionPicker displays an interactive session picker.
func interactiveSessionPicker(store *session.Store, filter *session.ListFilter) (*session.Session, error) {
	sessions, err := store.ListWithFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	// Display sessions with numbers
	fmt.Println("Available sessions:")
	fmt.Println()
	fmt.Printf("  %-3s %-8s %-8s %-20s %s\n", "#", "ID", "BACKEND", "LAST USED", "TITLE/PROMPT")
	fmt.Println("  " + strings.Repeat("-", 70))

	for i, s := range sessions {
		// Limit display to 20 sessions
		if i >= maxSessionsDisplay {
			fmt.Printf("  ... and %d more sessions\n", len(sessions)-maxSessionsDisplay)
			break
		}

		title := s.DisplayName()
		if len(title) > maxTitleDisplayLen {
			title = title[:maxTitleDisplayLen-3] + "..."
		}

		fmt.Printf("  %-3d %-8s %-8s %-20s %s\n",
			i+1,
			shortSessionID(s.ID),
			s.Backend,
			formatTimeAgo(s.LastUsed),
			title,
		)
	}

	fmt.Println()
	fmt.Print("Enter session number (or q to quit): ")

	var input string
	fmt.Scanln(&input)

	if input == "q" || input == "" {
		return nil, fmt.Errorf("canceled")
	}

	var idx int
	_, err = fmt.Sscanf(input, "%d", &idx)
	if err != nil || idx < 1 || idx > len(sessions) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	return sessions[idx-1], nil
}

// formatTimeAgo returns a human-readable time ago string.
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("2006-01-02")
	}
}

package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
)

// sessionsCmd manages sessions.
var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage sessions",
	Long:  "List, show, delete, or clean up sessions.",
}

var (
	listBackendFilter string
	listStatusFilter  string
	listLimit         int
)

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := session.NewStore()

		filter := &session.ListFilter{
			Backend: listBackendFilter,
			Limit:   listLimit,
		}
		if listStatusFilter != "" {
			filter.Status = session.SessionStatus(listStatusFilter)
		}

		sessions, err := store.ListWithFilter(filter)
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			return nil
		}

		fmt.Printf("%-8s %-8s %-10s %-15s %-12s %s\n", "ID", "BACKEND", "STATUS", "LAST USED", "TOKENS", "TITLE/PROMPT")
		fmt.Println(strings.Repeat("-", 90))
		for _, s := range sessions {
			status := string(s.Status)
			if status == "" {
				status = "unknown"
			}

			tokens := "-"
			if s.TokenUsage != nil && s.TokenUsage.Total() > 0 {
				tokens = fmt.Sprintf("%d", s.TokenUsage.Total())
			}

			title := s.DisplayName()
			if len(title) > maxSessionTitleLen {
				title = title[:maxSessionTitleLen-3] + "..."
			}

			fmt.Printf("%-8s %-8s %-10s %-15s %-12s %s\n",
				shortSessionID(s.ID),
				s.Backend,
				status,
				formatTimeAgo(s.LastUsed),
				tokens,
				title,
			)
		}

		return nil
	},
}

func init() {
	sessionsListCmd.Flags().StringVarP(&listBackendFilter, "backend", "b", "", "filter by backend")
	sessionsListCmd.Flags().StringVar(&listStatusFilter, "status", "", "filter by status (active, completed, error, paused)")
	sessionsListCmd.Flags().IntVarP(&listLimit, "limit", "n", 0, "limit number of sessions shown")
}

var sessionsShowCmd = &cobra.Command{
	Use:   "show <session-id>",
	Short: "Show session details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := session.NewStore()
		// Try prefix match first
		sess, err := store.GetByPrefix(args[0])
		if err != nil {
			// Fall back to exact match
			sess, err = store.Get(args[0])
			if err != nil {
				return err
			}
		}

		fmt.Printf("ID:                %s\n", sess.ID)
		fmt.Printf("Backend:           %s\n", sess.Backend)
		if sess.Model != "" {
			fmt.Printf("Model:             %s\n", sess.Model)
		}
		fmt.Printf("Status:            %s\n", sess.Status)
		fmt.Printf("Created:           %s\n", sess.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Last Used:         %s (%s)\n", sess.LastUsed.Format(time.RFC3339), formatTimeAgo(sess.LastUsed))
		fmt.Printf("Working Directory: %s\n", sess.WorkingDir)
		if sess.BackendSessionID != "" {
			fmt.Printf("Backend Session:   %s\n", sess.BackendSessionID)
		}
		if sess.Title != "" {
			fmt.Printf("Title:             %s\n", sess.Title)
		}
		if sess.InitialPrompt != "" {
			prompt := sess.InitialPrompt
			if len(prompt) > maxPromptDisplayLen {
				prompt = prompt[:maxPromptDisplayLen-3] + "..."
			}
			fmt.Printf("Initial Prompt:    %s\n", prompt)
		}
		fmt.Printf("Turns:             %d\n", sess.TurnCount)
		if sess.TokenUsage != nil {
			fmt.Printf("Token Usage:\n")
			fmt.Printf("  Input:           %d\n", sess.TokenUsage.InputTokens)
			fmt.Printf("  Output:          %d\n", sess.TokenUsage.OutputTokens)
			if sess.TokenUsage.CachedTokens > 0 {
				fmt.Printf("  Cached:          %d\n", sess.TokenUsage.CachedTokens)
			}
			if sess.TokenUsage.ReasoningTokens > 0 {
				fmt.Printf("  Reasoning:       %d\n", sess.TokenUsage.ReasoningTokens)
			}
			fmt.Printf("  Total:           %d\n", sess.TokenUsage.Total())
		}
		if len(sess.Tags) > 0 {
			fmt.Printf("Tags:              %s\n", strings.Join(sess.Tags, ", "))
		}
		if sess.ParentID != "" {
			fmt.Printf("Parent Session:    %s\n", sess.ParentID)
		}
		if sess.ErrorMessage != "" {
			fmt.Printf("Error:             %s\n", sess.ErrorMessage)
		}
		if len(sess.Metadata) > 0 {
			fmt.Println("Metadata:")
			for k, v := range sess.Metadata {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		return nil
	},
}

var sessionsDeleteCmd = &cobra.Command{
	Use:   "delete <session-id>",
	Short: "Delete a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := session.NewStore()
		sess, err := store.GetByPrefix(args[0])
		if err != nil {
			// Fall back to exact match
			sess, err = store.Get(args[0])
			if err != nil {
				return err
			}
		}
		if err := store.Delete(sess.ID); err != nil {
			return err
		}
		fmt.Printf("Session %s deleted.\n", sess.ID)
		return nil
	},
}

var cleanOlderThan string

var sessionsCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up old sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		var days int
		if cleanOlderThan != "" {
			var err error
			if strings.HasSuffix(cleanOlderThan, "d") {
				_, err = fmt.Sscanf(cleanOlderThan, "%dd", &days)
			} else {
				_, err = fmt.Sscanf(cleanOlderThan, "%d", &days)
			}
			if err != nil || days < 0 {
				return fmt.Errorf("invalid --older-than value: %q", cleanOlderThan)
			}
		}
		if days == 0 {
			days = config.Get().Session.RetentionDays
		}

		store := session.NewStore()
		deleted, err := store.CleanByDays(days)
		if err != nil {
			return err
		}

		fmt.Printf("Deleted %d session(s) older than %d days.\n", deleted, days)
		return nil
	},
}

func init() {
	sessionsCleanCmd.Flags().StringVar(&cleanOlderThan, "older-than", "", "delete sessions older than (e.g., 30d)")
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsShowCmd)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
	sessionsCmd.AddCommand(sessionsCleanCmd)
}

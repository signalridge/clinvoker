package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
	listTagFilter     string
	listQueryFilter   string
	listLimit         int
	listOffset        int
	listJSON          bool
	showJSON          bool
)

const cleanPreviewSampleLimit = 10

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := session.NewStore()

		filter := &session.ListFilter{
			Backend: listBackendFilter,
			Tag:     listTagFilter,
			Limit:   listLimit,
			Offset:  listOffset,
		}
		if listStatusFilter != "" {
			filter.Status = session.SessionStatus(listStatusFilter)
		}

		var result *session.ListResult
		if listQueryFilter != "" {
			matched, err := store.Search(listQueryFilter)
			if err != nil {
				return err
			}
			filtered := filterSearchedSessions(matched, filter)
			total := len(filtered)
			paged := paginateSessions(filtered, listLimit, listOffset)
			result = &session.ListResult{
				Sessions: paged,
				Total:    total,
				Limit:    listLimit,
				Offset:   listOffset,
			}
		} else {
			var err error
			result, err = store.ListPaginated(filter)
			if err != nil {
				return err
			}
		}
		sessions := result.Sessions

		if listJSON {
			output := buildSessionsListJSON(result, listBackendFilter, listStatusFilter, listTagFilter, listQueryFilter)
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(output)
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
	sessionsListCmd.Flags().StringVar(&listTagFilter, "tag", "", "filter by tag")
	sessionsListCmd.Flags().StringVar(&listQueryFilter, "search", "", "search sessions by id/title/initial_prompt")
	sessionsListCmd.Flags().IntVarP(&listLimit, "limit", "n", 0, "limit number of sessions shown")
	sessionsListCmd.Flags().IntVar(&listOffset, "offset", 0, "number of sessions to skip")
	sessionsListCmd.Flags().BoolVar(&listJSON, "json", false, "output sessions as JSON")
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

		if showJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(sess)
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

var tagDryRun bool

var sessionsTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage session tags",
	Long:  "Add or remove tags for a session by ID or ID prefix.",
}

var sessionsTagAddCmd = &cobra.Command{
	Use:   "add <session-id> <tag> [tag...]",
	Short: "Add one or more tags to a session",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSessionTagMutation(true, args[0], args[1:])
	},
}

var sessionsTagRemoveCmd = &cobra.Command{
	Use:   "rm <session-id> <tag> [tag...]",
	Short: "Remove one or more tags from a session",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSessionTagMutation(false, args[0], args[1:])
	},
}

var cleanOlderThan string
var cleanDryRun bool

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
		if cleanDryRun {
			count, samples, err := previewCleanByDays(store, days, cleanPreviewSampleLimit)
			if err != nil {
				return err
			}

			fmt.Printf("Dry run: would delete %d session(s) older than %d days.\n", count, days)
			if len(samples) > 0 {
				fmt.Printf("Sample session IDs: %s\n", strings.Join(samples, ", "))
			}
			fmt.Println("No sessions were deleted. Re-run without --dry-run to apply.")
			if count > 0 {
				fmt.Println("Note: candidate sessions may change between dry-run and actual cleanup.")
			}
			return nil
		}

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
	sessionsCleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "preview sessions to be deleted without making changes")
	sessionsShowCmd.Flags().BoolVar(&showJSON, "json", false, "output session details as JSON")
	sessionsTagCmd.PersistentFlags().BoolVar(&tagDryRun, "dry-run", false, "preview tag changes without saving")
	sessionsTagCmd.AddCommand(sessionsTagAddCmd)
	sessionsTagCmd.AddCommand(sessionsTagRemoveCmd)
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsShowCmd)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
	sessionsCmd.AddCommand(sessionsCleanCmd)
	sessionsCmd.AddCommand(sessionsTagCmd)
}

type sessionsListJSONOutput struct {
	Items   []sessionsListJSONItem  `json:"items"`
	Total   int                     `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	Filters sessionsListJSONFilters `json:"filters"`
	Sort    sessionsListJSONSort    `json:"sort"`
}

type sessionsListJSONItem struct {
	ID            string   `json:"id"`
	Backend       string   `json:"backend"`
	Status        string   `json:"status"`
	LastUsed      string   `json:"last_used"`
	Model         string   `json:"model"`
	Tags          []string `json:"tags"`
	Title         string   `json:"title"`
	PromptPreview string   `json:"prompt_preview"`
}

type sessionsListJSONFilters struct {
	Backend string `json:"backend"`
	Status  string `json:"status"`
	Tag     string `json:"tag,omitempty"`
	Query   string `json:"query,omitempty"`
}

type sessionsListJSONSort struct {
	By    string `json:"by"`
	Order string `json:"order"`
}

func buildSessionsListJSON(result *session.ListResult, backendFilter, statusFilter, tagFilter, queryFilter string) *sessionsListJSONOutput {
	items := make([]sessionsListJSONItem, 0, len(result.Sessions))
	for _, sess := range result.Sessions {
		status := string(sess.Status)
		if status == "" {
			status = "unknown"
		}

		promptPreview := strings.TrimSpace(strings.ReplaceAll(sess.InitialPrompt, "\n", " "))
		if len(promptPreview) > maxPromptDisplayLen {
			promptPreview = promptPreview[:maxPromptDisplayLen-3] + "..."
		}

		tags := make([]string, len(sess.Tags))
		copy(tags, sess.Tags)

		items = append(items, sessionsListJSONItem{
			ID:            sess.ID,
			Backend:       sess.Backend,
			Status:        status,
			LastUsed:      sess.LastUsed.UTC().Format(time.RFC3339),
			Model:         sess.Model,
			Tags:          tags,
			Title:         sess.Title,
			PromptPreview: promptPreview,
		})
	}

	return &sessionsListJSONOutput{
		Items:  items,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
		Filters: sessionsListJSONFilters{
			Backend: backendFilter,
			Status:  statusFilter,
			Tag:     tagFilter,
			Query:   queryFilter,
		},
		Sort: sessionsListJSONSort{
			By:    "last_used",
			Order: "desc",
		},
	}
}

func filterSearchedSessions(sessions []*session.Session, filter *session.ListFilter) []*session.Session {
	if filter == nil {
		return sessions
	}

	filtered := make([]*session.Session, 0, len(sessions))
	for _, sess := range sessions {
		if filter.Backend != "" && sess.Backend != filter.Backend {
			continue
		}
		if filter.Status != "" && sess.Status != filter.Status {
			continue
		}
		if filter.Tag != "" && !sess.HasTag(filter.Tag) {
			continue
		}
		filtered = append(filtered, sess)
	}
	return filtered
}

func paginateSessions(sessions []*session.Session, limit, offset int) []*session.Session {
	if offset > 0 {
		if offset >= len(sessions) {
			return []*session.Session{}
		}
		sessions = sessions[offset:]
	}
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
	}
	return sessions
}

func normalizeSessionTag(tag string) (string, error) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return "", fmt.Errorf("tag cannot be empty")
	}
	if len(tag) > 32 {
		return "", fmt.Errorf("tag too long (max 32 chars): %s", tag)
	}
	for _, r := range tag {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !valid {
			return "", fmt.Errorf("invalid tag %q: only [a-z0-9_-] allowed", tag)
		}
	}
	return tag, nil
}

func runSessionTagMutation(add bool, sessionRef string, rawTags []string) error {
	store := session.NewStore()
	sess, err := store.GetByPrefix(sessionRef)
	if err != nil {
		sess, err = store.Get(sessionRef)
		if err != nil {
			return err
		}
	}

	normalizedTags := make([]string, 0, len(rawTags))
	for _, rawTag := range rawTags {
		tag, err := normalizeSessionTag(rawTag)
		if err != nil {
			return err
		}
		normalizedTags = append(normalizedTags, tag)
	}

	changedCount := 0
	for _, tag := range normalizedTags {
		hasTag := sess.HasTag(tag)
		if add {
			if hasTag {
				continue
			}
			changedCount++
			if !tagDryRun {
				sess.AddTag(tag)
			}
			continue
		}

		if !hasTag {
			continue
		}
		changedCount++
		if !tagDryRun {
			sess.RemoveTag(tag)
		}
	}

	op := "add"
	if !add {
		op = "remove"
	}
	if tagDryRun {
		fmt.Printf("Dry run: would %s %d tag change(s) on session %s.\n", op, changedCount, sess.ID)
		fmt.Printf("Requested tags: %s\n", strings.Join(normalizedTags, ", "))
		fmt.Println("No session data was modified.")
		return nil
	}

	if changedCount == 0 {
		fmt.Printf("No tag changes needed for session %s.\n", sess.ID)
		return nil
	}

	if err := store.Save(sess); err != nil {
		return err
	}

	fmt.Printf("Applied %d tag change(s) on session %s.\n", changedCount, sess.ID)
	if len(sess.Tags) > 0 {
		fmt.Printf("Current tags: %s\n", strings.Join(sess.Tags, ", "))
	} else {
		fmt.Println("Current tags: (none)")
	}
	return nil
}

func previewCleanByDays(store *session.Store, days, sampleLimit int) (int, []string, error) {
	metas, err := store.ListMeta()
	if err != nil {
		return 0, nil, err
	}

	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	type candidate struct {
		id       string
		lastUsed time.Time
	}
	candidates := make([]candidate, 0)

	for _, meta := range metas {
		if meta.LastUsed.Before(cutoff) {
			candidates = append(candidates, candidate{id: meta.ID, lastUsed: meta.LastUsed})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].lastUsed.Equal(candidates[j].lastUsed) {
			return candidates[i].id < candidates[j].id
		}
		return candidates[i].lastUsed.Before(candidates[j].lastUsed)
	})

	count := len(candidates)
	if sampleLimit < 0 {
		sampleLimit = 0
	}
	if sampleLimit > count {
		sampleLimit = count
	}

	samples := make([]string, 0, sampleLimit)
	for i := 0; i < sampleLimit; i++ {
		samples = append(samples, candidates[i].id)
	}

	return count, samples, nil
}

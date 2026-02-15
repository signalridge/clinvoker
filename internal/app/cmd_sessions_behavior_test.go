package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/session"
)

func TestSessionsListCmd_JSONOutput_WithPaginationAndFilters(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	store := session.NewStore()

	older, err := store.CreateWithOptions("claude", "", &session.SessionOptions{InitialPrompt: "older prompt"})
	if err != nil {
		t.Fatalf("create older session: %v", err)
	}
	newer, err := store.CreateWithOptions("claude", "", &session.SessionOptions{InitialPrompt: "newer prompt"})
	if err != nil {
		t.Fatalf("create newer session: %v", err)
	}
	otherBackend, err := store.CreateWithOptions("codex", "", &session.SessionOptions{InitialPrompt: "other backend"})
	if err != nil {
		t.Fatalf("create other backend session: %v", err)
	}

	older.LastUsed = time.Now().Add(-2 * time.Hour)
	older.Status = session.StatusCompleted
	if err := store.Save(older); err != nil {
		t.Fatalf("save older session: %v", err)
	}

	newer.LastUsed = time.Now().Add(-1 * time.Hour)
	newer.Status = session.StatusActive
	if err := store.Save(newer); err != nil {
		t.Fatalf("save newer session: %v", err)
	}

	otherBackend.LastUsed = time.Now()
	otherBackend.Status = session.StatusActive
	if err := store.Save(otherBackend); err != nil {
		t.Fatalf("save other backend session: %v", err)
	}

	listBackendFilter = "claude"
	listLimit = 1
	listOffset = 1
	listJSON = true

	output := captureStdout(t, func() {
		if err := sessionsListCmd.RunE(sessionsListCmd, nil); err != nil {
			t.Fatalf("sessions list run failed: %v", err)
		}
	})

	var parsed sessionsListJSONOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("unmarshal sessions list json output: %v\noutput=%s", err, output)
	}

	if parsed.Total != 2 {
		t.Fatalf("total = %d, want 2", parsed.Total)
	}
	if parsed.Limit != 1 {
		t.Fatalf("limit = %d, want 1", parsed.Limit)
	}
	if parsed.Offset != 1 {
		t.Fatalf("offset = %d, want 1", parsed.Offset)
	}
	if parsed.Filters.Backend != "claude" {
		t.Fatalf("filters.backend = %q, want %q", parsed.Filters.Backend, "claude")
	}
	if parsed.Sort.By != "last_used" || parsed.Sort.Order != "desc" {
		t.Fatalf("unexpected sort semantics: %+v", parsed.Sort)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("items length = %d, want 1", len(parsed.Items))
	}
	if parsed.Items[0].ID != older.ID {
		t.Fatalf("items[0].id = %q, want %q", parsed.Items[0].ID, older.ID)
	}
	if parsed.Items[0].Status != string(session.StatusCompleted) {
		t.Fatalf("items[0].status = %q, want %q", parsed.Items[0].Status, session.StatusCompleted)
	}
}

func TestSessionsListCmd_JSONOutput_IncludesFixedFieldsWhenEmpty(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	store := session.NewStore()
	sess, err := store.CreateWithOptions("claude", "", &session.SessionOptions{})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := store.Save(sess); err != nil {
		t.Fatalf("save session: %v", err)
	}

	listBackendFilter = "claude"
	listLimit = 1
	listOffset = 0
	listJSON = true

	output := captureStdout(t, func() {
		if err := sessionsListCmd.RunE(sessionsListCmd, nil); err != nil {
			t.Fatalf("sessions list run failed: %v", err)
		}
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("unmarshal sessions list json output: %v\noutput=%s", err, output)
	}

	items, ok := parsed["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items should contain one session, got: %#v", parsed["items"])
	}

	item, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("item should be an object, got: %#v", items[0])
	}

	title, exists := item["title"]
	if !exists {
		t.Fatalf("title key should exist in JSON output: %#v", item)
	}
	if titleStr, ok := title.(string); !ok || titleStr != "" {
		t.Fatalf("title should be an empty string, got: %#v", title)
	}

	promptPreview, exists := item["prompt_preview"]
	if !exists {
		t.Fatalf("prompt_preview key should exist in JSON output: %#v", item)
	}
	if previewStr, ok := promptPreview.(string); !ok || previewStr != "" {
		t.Fatalf("prompt_preview should be an empty string, got: %#v", promptPreview)
	}
}

func TestSessionsShowCmd_JSONOutput(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	store := session.NewStore()
	sess, err := store.CreateWithOptions("claude", "", &session.SessionOptions{
		InitialPrompt: "show json prompt",
		Title:         "show json title",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := store.Save(sess); err != nil {
		t.Fatalf("save session: %v", err)
	}

	showJSON = true
	output := captureStdout(t, func() {
		if err := sessionsShowCmd.RunE(sessionsShowCmd, []string{sess.ID}); err != nil {
			t.Fatalf("sessions show run failed: %v", err)
		}
	})

	var parsed session.Session
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("unmarshal sessions show json output: %v\noutput=%s", err, output)
	}
	if parsed.ID != sess.ID {
		t.Fatalf("id = %q, want %q", parsed.ID, sess.ID)
	}
	if parsed.Backend != "claude" {
		t.Fatalf("backend = %q, want %q", parsed.Backend, "claude")
	}
}

func TestSessionsCleanCmd_DryRun_DoesNotDelete(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	store := session.NewStore()

	oldSession, err := store.CreateWithOptions("claude", "", &session.SessionOptions{InitialPrompt: "old session"})
	if err != nil {
		t.Fatalf("create old session: %v", err)
	}
	oldSession.LastUsed = time.Now().Add(-45 * 24 * time.Hour)
	if err := store.Save(oldSession); err != nil {
		t.Fatalf("save old session: %v", err)
	}

	newSession, err := store.CreateWithOptions("claude", "", &session.SessionOptions{InitialPrompt: "new session"})
	if err != nil {
		t.Fatalf("create new session: %v", err)
	}
	newSession.LastUsed = time.Now().Add(-2 * 24 * time.Hour)
	if err := store.Save(newSession); err != nil {
		t.Fatalf("save new session: %v", err)
	}

	cleanOlderThan = "30d"
	cleanDryRun = true

	output := captureStdout(t, func() {
		if err := sessionsCleanCmd.RunE(sessionsCleanCmd, nil); err != nil {
			t.Fatalf("sessions clean run failed: %v", err)
		}
	})

	if !strings.Contains(output, "Dry run: would delete 1 session(s) older than 30 days.") {
		t.Fatalf("unexpected dry-run output: %s", output)
	}
	if !strings.Contains(output, "No sessions were deleted.") {
		t.Fatalf("dry-run output should confirm no deletion: %s", output)
	}
	if !strings.Contains(output, oldSession.ID) {
		t.Fatalf("dry-run output should include sample old session id: %s", output)
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("store.Count failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("count after dry-run = %d, want 2", count)
	}

	if _, err := store.Get(oldSession.ID); err != nil {
		t.Fatalf("old session should still exist after dry-run: %v", err)
	}
}

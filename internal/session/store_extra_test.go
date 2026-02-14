package session

import (
	"strings"
	"testing"
	"time"
)

func TestStore_ListPaginated_Filtering(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	now := time.Now()

	sess1, err := store.Create("claude", "/dir1")
	if err != nil {
		t.Fatalf("create session 1: %v", err)
	}
	sess1.AddTag("alpha")
	sess1.Status = StatusActive
	sess1.LastUsed = now.Add(-2 * time.Hour)
	if err := store.Save(sess1); err != nil {
		t.Fatalf("save session 1: %v", err)
	}

	sess2, err := store.Create("claude", "/dir2")
	if err != nil {
		t.Fatalf("create session 2: %v", err)
	}
	sess2.AddTag("alpha")
	sess2.Status = StatusError
	sess2.LastUsed = now
	if err := store.Save(sess2); err != nil {
		t.Fatalf("save session 2: %v", err)
	}

	sess3, err := store.Create("codex", "/dir3")
	if err != nil {
		t.Fatalf("create session 3: %v", err)
	}
	sess3.AddTag("beta")
	sess3.Status = StatusCompleted
	sess3.LastUsed = now.Add(-time.Hour)
	if err := store.Save(sess3); err != nil {
		t.Fatalf("save session 3: %v", err)
	}

	result, err := store.ListPaginated(&ListFilter{
		Backend: "claude",
		Tag:     "alpha",
		Limit:   1,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("ListPaginated failed: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("Total = %d, want %d", result.Total, 2)
	}
	if len(result.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want %d", len(result.Sessions), 1)
	}
	if result.Sessions[0].ID != sess2.ID {
		t.Fatalf("expected most recent session %q, got %q", sess2.ID, result.Sessions[0].ID)
	}

	result, err = store.ListPaginated(&ListFilter{
		Backend: "claude",
		Tag:     "alpha",
		Limit:   1,
		Offset:  1,
	})
	if err != nil {
		t.Fatalf("ListPaginated offset failed: %v", err)
	}
	if len(result.Sessions) != 1 || result.Sessions[0].ID != sess1.ID {
		t.Fatalf("expected second session %q, got %v", sess1.ID, result.Sessions)
	}
}

func TestStore_GetByPrefix(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess1, err := NewSessionWithOptions("claude", "/tmp", nil)
	if err != nil {
		t.Fatalf("new session 1: %v", err)
	}
	sess1.ID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := store.Save(sess1); err != nil {
		t.Fatalf("save session 1: %v", err)
	}

	sess2, err := NewSessionWithOptions("claude", "/tmp", nil)
	if err != nil {
		t.Fatalf("new session 2: %v", err)
	}
	sess2.ID = "aaaa11111111111111111111111111"
	if err := store.Save(sess2); err != nil {
		t.Fatalf("save session 2: %v", err)
	}

	if _, err := store.GetByPrefix("aaaa"); err == nil {
		t.Fatal("expected ambiguous prefix error")
	}

	got, err := store.GetByPrefix("aaaaaaaa")
	if err != nil {
		t.Fatalf("GetByPrefix failed: %v", err)
	}
	if got.ID != sess1.ID {
		t.Fatalf("got ID %q, want %q", got.ID, sess1.ID)
	}

	if _, err := store.GetByPrefix("ZZZ"); err == nil {
		t.Fatal("expected error for invalid prefix")
	}
}

func TestStore_Fork(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess, err := store.Create("claude", "/tmp")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	sess.Model = "model-x"
	sess.AddTag("alpha")
	sess.SetMetadata("k", "v")
	if err := store.Save(sess); err != nil {
		t.Fatalf("save session: %v", err)
	}

	forked, err := store.Fork(sess.ID)
	if err != nil {
		t.Fatalf("fork session: %v", err)
	}
	if forked.ID == sess.ID {
		t.Fatal("forked session should have a new ID")
	}
	if forked.ParentID != sess.ID {
		t.Fatalf("ParentID = %q, want %q", forked.ParentID, sess.ID)
	}
	if forked.Backend != sess.Backend {
		t.Fatalf("Backend = %q, want %q", forked.Backend, sess.Backend)
	}
	if !forked.HasTag("alpha") {
		t.Fatal("expected forked session to copy tags")
	}
	if forked.Metadata["k"] != "v" {
		t.Fatal("expected forked session to copy metadata")
	}
}

func TestStore_StatsAndCount(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess1, _ := store.Create("claude", "/tmp")
	sess1.Status = StatusActive
	sess1.TokenUsage = &TokenUsage{InputTokens: 10, OutputTokens: 5}
	if err := store.Save(sess1); err != nil {
		t.Fatalf("save session 1: %v", err)
	}

	sess2, _ := store.Create("codex", "/tmp")
	sess2.Status = StatusError
	sess2.TokenUsage = &TokenUsage{InputTokens: 3, OutputTokens: 7}
	if err := store.Save(sess2); err != nil {
		t.Fatalf("save session 2: %v", err)
	}

	sess3, _ := store.Create("claude", "/tmp")
	sess3.Status = StatusCompleted
	if err := store.Save(sess3); err != nil {
		t.Fatalf("save session 3: %v", err)
	}

	stats, err := store.Stats()
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	if stats.TotalSessions != 3 {
		t.Fatalf("TotalSessions = %d, want %d", stats.TotalSessions, 3)
	}
	if stats.SessionsByBackend["claude"] != 2 {
		t.Fatalf("SessionsByBackend[claude] = %d, want %d", stats.SessionsByBackend["claude"], 2)
	}
	if stats.SessionsByStatus[StatusActive] != 1 {
		t.Fatalf("SessionsByStatus[active] = %d, want %d", stats.SessionsByStatus[StatusActive], 1)
	}

	statsWithTokens, err := store.StatsWithTokens()
	if err != nil {
		t.Fatalf("StatsWithTokens failed: %v", err)
	}
	if statsWithTokens.TotalInputTokens != 13 {
		t.Fatalf("TotalInputTokens = %d, want %d", statsWithTokens.TotalInputTokens, 13)
	}
	if statsWithTokens.TotalOutputTokens != 12 {
		t.Fatalf("TotalOutputTokens = %d, want %d", statsWithTokens.TotalOutputTokens, 12)
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("Count = %d, want %d", count, 3)
	}
}

func TestStore_Search(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess, _ := store.Create("claude", "/tmp")
	sess.Title = "Fix auth bug"
	sess.InitialPrompt = "Investigate auth regression"
	if err := store.Save(sess); err != nil {
		t.Fatalf("save session: %v", err)
	}

	matches, err := store.Search("auth")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected matches for query")
	}
	found := false
	for _, m := range matches {
		if m.ID == sess.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("expected session to appear in search results")
	}

	matches, err = store.Search("missing")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	for _, m := range matches {
		if strings.Contains(strings.ToLower(m.Title), "missing") {
			t.Fatal("unexpected match for missing query")
		}
	}
}

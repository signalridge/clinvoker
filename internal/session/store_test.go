package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "clinvoker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	store := &Store{dir: tmpDir}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestStore_CreateAndGet(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess, err := store.Create("claude", "/test/dir")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	retrieved, err := store.Get(sess.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}

	if retrieved.ID != sess.ID {
		t.Errorf("ID mismatch: expected %q, got %q", sess.ID, retrieved.ID)
	}
	if retrieved.Backend != "claude" {
		t.Errorf("Backend mismatch: expected 'claude', got %q", retrieved.Backend)
	}
	if retrieved.WorkingDir != "/test/dir" {
		t.Errorf("WorkingDir mismatch: expected '/test/dir', got %q", retrieved.WorkingDir)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestStore_Save(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess, _ := store.Create("claude", "/tmp")
	sess.SetMetadata("key", "value")

	err := store.Save(sess)
	if err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	retrieved, _ := store.Get(sess.ID)
	if retrieved.Metadata["key"] != "value" {
		t.Error("metadata not persisted")
	}
}

func TestStore_Delete(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess, _ := store.Create("claude", "/tmp")

	err := store.Delete(sess.ID)
	if err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	_, err = store.Get(sess.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestStore_DeleteNotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	err := store.Delete("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestStore_List(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create multiple sessions
	sess1, _ := store.Create("claude", "/dir1")
	time.Sleep(10 * time.Millisecond)
	sess2, _ := store.Create("codex", "/dir2")
	time.Sleep(10 * time.Millisecond)
	sess3, _ := store.Create("gemini", "/dir3")

	sessions, err := store.List()
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}

	// Should be sorted by LastUsed, most recent first
	if sessions[0].ID != sess3.ID {
		t.Error("sessions not sorted by LastUsed")
	}
	if sessions[1].ID != sess2.ID {
		t.Error("sessions not sorted by LastUsed")
	}
	if sessions[2].ID != sess1.ID {
		t.Error("sessions not sorted by LastUsed")
	}
}

func TestStore_Last(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	store.Create("claude", "/dir1")
	time.Sleep(10 * time.Millisecond)
	sess2, _ := store.Create("codex", "/dir2")

	last, err := store.Last()
	if err != nil {
		t.Fatalf("failed to get last session: %v", err)
	}

	if last.ID != sess2.ID {
		t.Errorf("expected last session ID %q, got %q", sess2.ID, last.ID)
	}
}

func TestStore_LastEmpty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	_, err := store.Last()
	if err == nil {
		t.Error("expected error for empty store")
	}
}

func TestStore_LastForBackend(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	store.Create("claude", "/dir1")
	time.Sleep(10 * time.Millisecond)
	store.Create("codex", "/dir2")
	time.Sleep(10 * time.Millisecond)
	sess3, _ := store.Create("claude", "/dir3")

	last, err := store.LastForBackend("claude")
	if err != nil {
		t.Fatalf("failed to get last claude session: %v", err)
	}

	if last.ID != sess3.ID {
		t.Errorf("expected session ID %q, got %q", sess3.ID, last.ID)
	}
}

func TestStore_LastForBackendNotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	store.Create("claude", "/tmp")

	_, err := store.LastForBackend("gemini")
	if err == nil {
		t.Error("expected error for backend with no sessions")
	}
}

func TestStore_Clean(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create sessions with different ages
	sess1, _ := store.Create("claude", "/dir1")
	sess1.LastUsed = time.Now().Add(-48 * time.Hour)
	store.Save(sess1)

	sess2, _ := store.Create("codex", "/dir2")
	sess2.LastUsed = time.Now().Add(-12 * time.Hour)
	store.Save(sess2)

	store.Create("gemini", "/dir3") // Recent

	deleted, err := store.Clean(24 * time.Hour)
	if err != nil {
		t.Fatalf("failed to clean sessions: %v", err)
	}

	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	sessions, _ := store.List()
	if len(sessions) != 2 {
		t.Errorf("expected 2 remaining sessions, got %d", len(sessions))
	}
}

func TestStore_CleanByDays(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess, _ := store.Create("claude", "/tmp")
	sess.LastUsed = time.Now().Add(-40 * 24 * time.Hour) // 40 days old
	store.Save(sess)

	store.Create("codex", "/tmp") // Recent

	deleted, err := store.CleanByDays(30)
	if err != nil {
		t.Fatalf("failed to clean sessions: %v", err)
	}

	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}

func TestStore_SessionPath(t *testing.T) {
	store := &Store{dir: "/test/sessions"}

	path := store.sessionPath("abc123")
	expected := filepath.Join("/test/sessions", "abc123.json")

	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

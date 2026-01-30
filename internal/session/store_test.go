package session

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "clinvoker-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	store := NewStoreWithDir(tmpDir)

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

	sess, err := store.Create("claude", "/tmp")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	sess.SetMetadata("key", "value")

	if err := store.Save(sess); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	retrieved, err := store.Get(sess.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if retrieved.Metadata["key"] != "value" {
		t.Error("metadata not persisted")
	}
}

func TestStore_Delete(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	sess, err := store.Create("claude", "/tmp")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if err := store.Delete(sess.ID); err != nil {
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

	// Create multiple sessions with explicit timestamps for deterministic ordering
	now := time.Now()

	sess1, err := store.Create("claude", "/dir1")
	if err != nil {
		t.Fatalf("failed to create session 1: %v", err)
	}
	sess1.LastUsed = now.Add(-2 * time.Hour) // Oldest
	if err := store.Save(sess1); err != nil {
		t.Fatalf("failed to save session 1: %v", err)
	}

	sess2, err := store.Create("codex", "/dir2")
	if err != nil {
		t.Fatalf("failed to create session 2: %v", err)
	}
	sess2.LastUsed = now.Add(-1 * time.Hour) // Middle
	if err := store.Save(sess2); err != nil {
		t.Fatalf("failed to save session 2: %v", err)
	}

	sess3, err := store.Create("gemini", "/dir3")
	if err != nil {
		t.Fatalf("failed to create session 3: %v", err)
	}
	sess3.LastUsed = now // Most recent
	if err := store.Save(sess3); err != nil {
		t.Fatalf("failed to save session 3: %v", err)
	}

	sessions, err := store.List()
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}

	// Should be sorted by LastUsed, most recent first
	if sessions[0].ID != sess3.ID {
		t.Errorf("expected most recent session %s first, got %s", sess3.ID, sessions[0].ID)
	}
	if sessions[1].ID != sess2.ID {
		t.Errorf("expected middle session %s second, got %s", sess2.ID, sessions[1].ID)
	}
	if sessions[2].ID != sess1.ID {
		t.Errorf("expected oldest session %s last, got %s", sess1.ID, sessions[2].ID)
	}
}

func TestStore_Last(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	now := time.Now()

	sess1, err := store.Create("claude", "/dir1")
	if err != nil {
		t.Fatalf("failed to create session 1: %v", err)
	}
	sess1.LastUsed = now.Add(-1 * time.Hour) // Older
	if err := store.Save(sess1); err != nil {
		t.Fatalf("failed to save session 1: %v", err)
	}

	sess2, err := store.Create("codex", "/dir2")
	if err != nil {
		t.Fatalf("failed to create session 2: %v", err)
	}
	sess2.LastUsed = now // More recent
	if err := store.Save(sess2); err != nil {
		t.Fatalf("failed to save session 2: %v", err)
	}

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

	now := time.Now()

	sess1, err := store.Create("claude", "/dir1")
	if err != nil {
		t.Fatalf("failed to create session 1: %v", err)
	}
	sess1.LastUsed = now.Add(-2 * time.Hour) // Oldest claude
	if err := store.Save(sess1); err != nil {
		t.Fatalf("failed to save session 1: %v", err)
	}

	sess2, err := store.Create("codex", "/dir2")
	if err != nil {
		t.Fatalf("failed to create session 2: %v", err)
	}
	sess2.LastUsed = now.Add(-1 * time.Hour) // Codex (different backend)
	if err := store.Save(sess2); err != nil {
		t.Fatalf("failed to save session 2: %v", err)
	}

	sess3, err := store.Create("claude", "/dir3")
	if err != nil {
		t.Fatalf("failed to create session 3: %v", err)
	}
	sess3.LastUsed = now // Most recent claude
	if err := store.Save(sess3); err != nil {
		t.Fatalf("failed to save session 3: %v", err)
	}

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

	_, err := store.Create("claude", "/tmp")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	_, err = store.LastForBackend("gemini")
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

// Concurrent access tests

func TestStore_ConcurrentSave(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create multiple sessions for concurrent save
	const goroutines = 10
	sessions := make([]*Session, goroutines)
	for i := 0; i < goroutines; i++ {
		sess, err := store.Create("claude", "/tmp")
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}
		sessions[i] = sess
	}

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)

	// Each goroutine saves its own session concurrently
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sess := sessions[i]
			sess.SetMetadata("key", "value")
			if err := store.Save(sess); err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent save error: %v", err)
	}

	// Verify all sessions are still readable
	for _, sess := range sessions {
		_, err := store.Get(sess.ID)
		if err != nil {
			t.Errorf("failed to get session after concurrent saves: %v", err)
		}
	}
}

func TestStore_ConcurrentCreate(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	const goroutines = 10
	var wg sync.WaitGroup
	sessions := make(chan *Session, goroutines)
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sess, err := store.Create("claude", "/tmp")
			if err != nil {
				errCh <- err
				return
			}
			sessions <- sess
		}()
	}

	wg.Wait()
	close(sessions)
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent create error: %v", err)
	}

	// Verify all sessions were created
	var count int
	for range sessions {
		count++
	}
	if count != goroutines {
		t.Errorf("expected %d sessions, got %d", goroutines, count)
	}

	// Verify list returns all sessions
	list, err := store.List()
	if err != nil {
		t.Errorf("failed to list sessions: %v", err)
	}
	if len(list) != goroutines {
		t.Errorf("expected %d sessions in list, got %d", goroutines, len(list))
	}
}

func TestStore_ConcurrentReadWrite(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create initial sessions
	const initialSessions = 5
	for i := 0; i < initialSessions; i++ {
		_, err := store.Create("claude", "/tmp")
		if err != nil {
			t.Fatalf("failed to create initial session: %v", err)
		}
	}

	const goroutines = 10
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*2)

	// Half the goroutines read
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				_, err := store.List()
				if err != nil {
					errCh <- err
				}
			}
		}()
	}

	// Half the goroutines write
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				_, err := store.Create("gemini", "/tmp")
				if err != nil {
					errCh <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent read/write error: %v", err)
	}
}

func TestStore_ConcurrentDelete(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create sessions to delete
	const numSessions = 10
	sessionIDs := make([]string, numSessions)
	for i := 0; i < numSessions; i++ {
		sess, err := store.Create("claude", "/tmp")
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}
		sessionIDs[i] = sess.ID
	}

	var wg sync.WaitGroup
	errCh := make(chan error, numSessions)

	for _, id := range sessionIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := store.Delete(id); err != nil {
				errCh <- err
			}
		}(id)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent delete error: %v", err)
	}

	// Verify all sessions were deleted
	list, err := store.List()
	if err != nil {
		t.Errorf("failed to list sessions: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 sessions after delete, got %d", len(list))
	}
}

func TestStore_RaceDetection(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	const goroutines = 10
	var wg sync.WaitGroup

	// Mixed operations: create, save, get, list, delete
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Create
			sess, err := store.Create("backend-"+string(rune('a'+i%3)), "/tmp")
			if err != nil {
				return
			}

			// Save updates
			sess.SetMetadata("iteration", "value")
			_ = store.Save(sess)

			// Get
			_, _ = store.Get(sess.ID)

			// List
			_, _ = store.List()

			// Delete (some of them)
			if i%2 == 0 {
				_ = store.Delete(sess.ID)
			}
		}(i)
	}

	wg.Wait()

	// Final list should not panic
	_, err := store.List()
	if err != nil {
		t.Errorf("final list failed: %v", err)
	}
}

func TestMetaMatchesFilter_NilFilter(t *testing.T) {
	store := NewStore()
	meta := &SessionMeta{
		Backend: "claude",
		Model:   "model-x",
		Tags:    []string{"tag-a"},
	}

	if !store.metaMatchesFilter(meta, nil) {
		t.Fatal("expected nil filter to match any session meta")
	}
}

package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/signalridge/clinvoker/internal/config"
)

// sessionIDPattern validates session IDs (hex string, 16 characters).
var sessionIDPattern = regexp.MustCompile(`^[a-f0-9]{16}$`)

// indexFileName is the name of the persisted index file.
const indexFileName = "index.json"

// SessionMeta holds lightweight metadata for indexing without loading full session.
type SessionMeta struct {
	ID        string
	Backend   string
	Status    SessionStatus
	LastUsed  time.Time
	Model     string
	WorkDir   string
	Title     string
	Tags      []string
	CreatedAt time.Time
}

// Store handles session persistence with in-memory index for performance.
type Store struct {
	mu    sync.RWMutex
	dir   string
	index map[string]*SessionMeta // Lightweight metadata cache
	dirty bool                    // True if index needs refresh from disk
}

// validateSessionID checks if the session ID is valid and safe.
func validateSessionID(id string) error {
	if id == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	// Check for path traversal attempts
	if strings.Contains(id, "/") || strings.Contains(id, "\\") || strings.Contains(id, "..") {
		return fmt.Errorf("invalid session ID: contains path characters")
	}
	// For full session IDs, validate format
	if len(id) == 16 && !sessionIDPattern.MatchString(id) {
		return fmt.Errorf("invalid session ID format")
	}
	return nil
}

// validateSessionPrefix checks if a session ID prefix is valid and safe.
// This is a relaxed validation for prefix lookups that allows shorter IDs.
func validateSessionPrefix(prefix string) error {
	if prefix == "" {
		return fmt.Errorf("session prefix cannot be empty")
	}
	// Check for path traversal attempts
	if strings.Contains(prefix, "/") || strings.Contains(prefix, "\\") || strings.Contains(prefix, "..") {
		return fmt.Errorf("invalid session prefix: contains path characters")
	}
	// Prefix must only contain valid hex characters
	for _, c := range prefix {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return fmt.Errorf("invalid session prefix: must contain only hex characters (0-9, a-f)")
		}
	}
	return nil
}

// NewStore creates a new session store.
func NewStore() *Store {
	return &Store{
		dir:   config.SessionsDir(),
		index: make(map[string]*SessionMeta),
		dirty: true, // Index needs to be loaded on first use
	}
}

// NewStoreWithDir creates a new session store with a custom directory.
func NewStoreWithDir(dir string) *Store {
	return &Store{
		dir:   dir,
		index: make(map[string]*SessionMeta),
		dirty: true,
	}
}

// updateIndex updates the index entry for a session.
func (s *Store) updateIndex(sess *Session) {
	s.index[sess.ID] = &SessionMeta{
		ID:        sess.ID,
		Backend:   sess.Backend,
		Status:    sess.Status,
		LastUsed:  sess.LastUsed,
		Model:     sess.Model,
		WorkDir:   sess.WorkingDir,
		Title:     sess.Title,
		Tags:      sess.Tags,
		CreatedAt: sess.CreatedAt,
	}
}

// removeFromIndex removes a session from the index.
func (s *Store) removeFromIndex(id string) {
	delete(s.index, id)
}

// persistedIndex is the JSON structure for the persisted index file.
type persistedIndex struct {
	Version int                     `json:"version"`
	Index   map[string]*SessionMeta `json:"index"`
}

// persistIndex saves the index to disk for fast startup.
// Caller must hold write lock.
func (s *Store) persistIndex() error {
	if err := s.ensureStoreDirLocked(); err != nil {
		return err
	}

	data, err := json.Marshal(&persistedIndex{
		Version: 1,
		Index:   s.index,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	indexPath := filepath.Join(s.dir, indexFileName)
	if err := os.WriteFile(indexPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// loadPersistedIndex loads the index from disk if it exists.
// Returns true if index was successfully loaded, false otherwise.
// Caller must hold write lock.
func (s *Store) loadPersistedIndex() bool {
	indexPath := filepath.Join(s.dir, indexFileName)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return false // File doesn't exist or can't be read
	}

	var pi persistedIndex
	if err := json.Unmarshal(data, &pi); err != nil {
		return false // Invalid JSON
	}

	// Validate index version
	if pi.Version != 1 {
		return false // Unknown version
	}

	s.index = pi.Index
	if s.index == nil {
		s.index = make(map[string]*SessionMeta)
	}
	s.dirty = false
	return true
}

// ensureIndexLoaded loads the index from disk if needed.
// NOTE: This must be called with at least a read lock held.
// If the index needs rebuilding, caller must upgrade to write lock first.
func (s *Store) ensureIndexLoaded() error {
	// Fast path: index is already loaded
	if !s.dirty {
		return nil
	}
	// Slow path: index needs to be rebuilt
	// This should only be called when write lock is held
	return s.rebuildIndex()
}

// ensureIndexLoadedForRead prepares the index for read operations.
// It releases the read lock, takes a write lock to rebuild if needed,
// then returns with the read lock held again.
//
// CONCURRENCY NOTE: There is a brief window between RUnlock and Lock where
// another goroutine could modify state. This is safe because:
// 1. Double-check pattern: we re-check dirty flag after acquiring write lock
// 2. Index rebuild is idempotent: multiple rebuilds produce same result
// 3. Write operations (Save, Delete, Create) hold write lock and update index directly
// 4. The index is only marked dirty on initialization or explicit InvalidateIndex()
func (s *Store) ensureIndexLoadedForRead() error {
	// Fast path: index is already loaded (most common case)
	if !s.dirty {
		return nil
	}

	// Slow path: need to rebuild index
	// Release read lock and acquire write lock for rebuild
	s.mu.RUnlock()
	s.mu.Lock()

	// Double-check after acquiring write lock - another goroutine may have
	// already rebuilt the index while we were waiting for the lock
	var err error
	if s.dirty {
		err = s.rebuildIndex()
	}

	// Downgrade back to read lock for caller's read operation
	s.mu.Unlock()
	s.mu.RLock()

	return err
}

// rebuildIndex rebuilds the in-memory index from disk.
// First tries to load from persisted index file for fast startup.
// Falls back to scanning session files if index file is missing/invalid.
func (s *Store) rebuildIndex() error {
	if err := s.ensureStoreDirLocked(); err != nil {
		return err
	}

	// Fast path: try to load persisted index
	if s.loadPersistedIndex() {
		return nil
	}

	// Slow path: scan session files
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			s.index = make(map[string]*SessionMeta)
			s.dirty = false
			return nil
		}
		return fmt.Errorf("failed to read sessions dir: %w", err)
	}

	newIndex := make(map[string]*SessionMeta, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		// Skip the index file itself
		if entry.Name() == indexFileName {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // remove .json
		sess, err := s.getLocked(id)
		if err != nil {
			continue
		}

		newIndex[id] = &SessionMeta{
			ID:        sess.ID,
			Backend:   sess.Backend,
			Status:    sess.Status,
			LastUsed:  sess.LastUsed,
			Model:     sess.Model,
			WorkDir:   sess.WorkingDir,
			Title:     sess.Title,
			Tags:      sess.Tags,
			CreatedAt: sess.CreatedAt,
		}
	}

	s.index = newIndex
	s.dirty = false

	// Persist the newly built index for next startup
	_ = s.persistIndex()

	return nil
}

// InvalidateIndex marks the index as needing refresh.
// Call this if files are modified externally.
func (s *Store) InvalidateIndex() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = true
}

// ValidateIndex checks all index entries against actual files and removes stale entries.
// Returns the number of stale entries removed and any error encountered.
// This should be called periodically or when ghost sessions are suspected.
func (s *Store) ValidateIndex() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.dirty {
		// Index not loaded yet - rebuild will validate automatically
		if err := s.rebuildIndex(); err != nil {
			return 0, err
		}
		return 0, nil // rebuildIndex scans files, no stale entries possible
	}

	// Check each index entry against actual files
	var staleIDs []string
	for id := range s.index {
		path := s.sessionPath(id)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			staleIDs = append(staleIDs, id)
		}
	}

	// Remove stale entries
	for _, id := range staleIDs {
		delete(s.index, id)
	}

	// Persist cleaned index if any entries were removed
	if len(staleIDs) > 0 {
		_ = s.persistIndex()
	}

	return len(staleIDs), nil
}

// CountValidated returns the count of sessions that actually exist on disk.
// Unlike Count(), this validates each index entry against the file system.
// Use Count() for fast lookups when accuracy is less critical.
func (s *Store) CountValidated() (int, error) {
	removed, err := s.ValidateIndex()
	if err != nil {
		return 0, err
	}
	if removed > 0 {
		// Log would go here if logger available
	}
	return s.Count()
}

// Create creates a new session and saves it.
func (s *Store) Create(backend, workDir string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureStoreDirLocked(); err != nil {
		return nil, fmt.Errorf("failed to create sessions dir: %w", err)
	}

	sess, err := NewSession(backend, workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if err := s.saveLocked(sess); err != nil {
		return nil, err
	}

	s.updateIndex(sess)

	// Persist index synchronously for Create to ensure consistency
	// (new sessions should be visible immediately after restart)
	if err := s.persistIndex(); err != nil {
		// Log but don't fail - session is already saved to disk
		// Index can be rebuilt on next startup
		_ = err // Silent ignore, consider adding logging in future
	}

	return sess, nil
}

// Save persists a session to disk.
func (s *Store) Save(sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.saveLocked(sess); err != nil {
		return err
	}

	s.updateIndex(sess)

	// Persist index asynchronously (don't block on index persistence errors)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		_ = s.persistIndex()
	}()

	return nil
}

// saveLocked persists a session to disk. Caller must hold s.mu.
func (s *Store) saveLocked(sess *Session) error {
	if err := s.ensureStoreDirLocked(); err != nil {
		return fmt.Errorf("failed to create sessions dir: %w", err)
	}

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	path := s.sessionPath(sess.ID)
	// Use 0600 to protect potentially sensitive prompt data
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Get retrieves a session by ID.
func (s *Store) Get(id string) (*Session, error) {
	if err := validateSessionID(id); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getLocked(id)
}

// getLocked retrieves a session by ID. Caller must hold s.mu.
func (s *Store) getLocked(id string) (*Session, error) {
	path := s.sessionPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read session: %w", err)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	return &sess, nil
}

// Delete removes a session.
func (s *Store) Delete(id string) error {
	if err := validateSessionID(id); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.deleteLocked(id); err != nil {
		return err
	}

	s.removeFromIndex(id)

	// Persist index asynchronously (don't block on index persistence errors)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		_ = s.persistIndex()
	}()

	return nil
}

// deleteLocked removes a session. Caller must hold s.mu.
func (s *Store) deleteLocked(id string) error {
	path := s.sessionPath(id)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session not found: %s", id)
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Also remove session artifacts directory if it exists
	artifactsDir := filepath.Join(s.dir, id)
	_ = os.RemoveAll(artifactsDir)

	return nil
}

// List returns all sessions, sorted by last used (most recent first).
func (s *Store) List() ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listLocked()
}

// listLocked returns all sessions. Caller must hold s.mu (read lock).
func (s *Store) listLocked() ([]*Session, error) {
	if err := s.ensureIndexLoadedForRead(); err != nil {
		return nil, err
	}

	// Load full sessions from index
	sessions := make([]*Session, 0, len(s.index))
	for id := range s.index {
		sess, err := s.getLocked(id)
		if err != nil {
			continue
		}
		sessions = append(sessions, sess)
	}

	// Sort by last used, most recent first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastUsed.After(sessions[j].LastUsed)
	})

	return sessions, nil
}

// ListMeta returns lightweight metadata for all sessions (faster than List).
// Use this when you don't need the full session data.
func (s *Store) ListMeta() ([]*SessionMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureIndexLoadedForRead(); err != nil {
		return nil, err
	}

	metas := make([]*SessionMeta, 0, len(s.index))
	for _, meta := range s.index {
		metas = append(metas, meta)
	}

	// Sort by last used, most recent first
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].LastUsed.After(metas[j].LastUsed)
	})

	return metas, nil
}

// Last returns the most recently used session.
func (s *Store) Last() (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureIndexLoadedForRead(); err != nil {
		return nil, err
	}

	if len(s.index) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	// Find most recent from index
	var latest *SessionMeta
	for _, meta := range s.index {
		if latest == nil || meta.LastUsed.After(latest.LastUsed) {
			latest = meta
		}
	}

	return s.getLocked(latest.ID)
}

// LastForBackend returns the most recently used session for a backend.
func (s *Store) LastForBackend(backend string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureIndexLoadedForRead(); err != nil {
		return nil, err
	}

	// Find most recent for backend from index
	var latest *SessionMeta
	for _, meta := range s.index {
		if meta.Backend == backend {
			if latest == nil || meta.LastUsed.After(latest.LastUsed) {
				latest = meta
			}
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no sessions found for backend: %s", backend)
	}

	return s.getLocked(latest.ID)
}

// Clean removes sessions older than the specified duration.
func (s *Store) Clean(maxAge time.Duration) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureIndexLoaded(); err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-maxAge)
	var deleted int
	var toDelete []string

	for id, meta := range s.index {
		if meta.LastUsed.Before(cutoff) {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		if err := s.deleteLocked(id); err == nil {
			s.removeFromIndex(id)
			deleted++
		}
	}

	return deleted, nil
}

// CleanByDays removes sessions older than the specified number of days.
func (s *Store) CleanByDays(days int) (int, error) {
	return s.Clean(time.Duration(days) * 24 * time.Hour)
}

// ListFilter provides filtering options for listing sessions.
type ListFilter struct {
	Backend string
	Status  SessionStatus
	Tag     string
	Model   string
	WorkDir string
	Limit   int
	Offset  int // Number of sessions to skip (for pagination)
}

// ListResult contains paginated session results.
type ListResult struct {
	Sessions []*Session
	Total    int // Total number of matching sessions (before pagination)
	Limit    int // Limit used
	Offset   int // Offset used
}

// ListWithFilter returns sessions matching the filter criteria.
// This uses the index for efficient filtering when possible.
func (s *Store) ListWithFilter(filter *ListFilter) ([]*Session, error) {
	result, err := s.ListPaginated(filter)
	if err != nil {
		return nil, err
	}
	return result.Sessions, nil
}

// ListPaginated returns sessions with pagination metadata.
// This is the recommended method for paginated queries.
func (s *Store) ListPaginated(filter *ListFilter) (*ListResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureIndexLoadedForRead(); err != nil {
		return nil, err
	}

	if filter == nil {
		sessions, err := s.listLocked()
		if err != nil {
			return nil, err
		}
		return &ListResult{
			Sessions: sessions,
			Total:    len(sessions),
			Limit:    0,
			Offset:   0,
		}, nil
	}

	// First filter using index metadata
	var matchingIDs []string
	for id, meta := range s.index {
		if !s.metaMatchesFilter(meta, filter) {
			continue
		}
		matchingIDs = append(matchingIDs, id)
	}

	// Sort IDs by last used (from index)
	sort.Slice(matchingIDs, func(i, j int) bool {
		return s.index[matchingIDs[i]].LastUsed.After(s.index[matchingIDs[j]].LastUsed)
	})

	total := len(matchingIDs)

	// Apply offset
	if filter.Offset > 0 {
		if filter.Offset >= len(matchingIDs) {
			matchingIDs = nil
		} else {
			matchingIDs = matchingIDs[filter.Offset:]
		}
	}

	// Apply limit
	if filter.Limit > 0 && len(matchingIDs) > filter.Limit {
		matchingIDs = matchingIDs[:filter.Limit]
	}

	// Load full sessions only for matches
	sessions := make([]*Session, 0, len(matchingIDs))
	for _, id := range matchingIDs {
		sess, err := s.getLocked(id)
		if err != nil {
			continue
		}
		sessions = append(sessions, sess)
	}

	return &ListResult{
		Sessions: sessions,
		Total:    total,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	}, nil
}

// metaMatchesFilter checks if metadata matches the filter criteria.
func (s *Store) metaMatchesFilter(meta *SessionMeta, filter *ListFilter) bool {
	if filter == nil {
		return true
	}
	if filter.Backend != "" && meta.Backend != filter.Backend {
		return false
	}
	if filter.Status != "" && meta.Status != filter.Status {
		return false
	}
	if filter.Model != "" && meta.Model != filter.Model {
		return false
	}
	if filter.WorkDir != "" && meta.WorkDir != filter.WorkDir {
		return false
	}
	if filter.Tag != "" && !slices.Contains(meta.Tags, filter.Tag) {
		return false
	}
	return true
}

// ListByBackend returns all sessions for a specific backend.
func (s *Store) ListByBackend(backend string) ([]*Session, error) {
	return s.ListWithFilter(&ListFilter{Backend: backend})
}

// ListByTag returns all sessions with a specific tag.
func (s *Store) ListByTag(tag string) ([]*Session, error) {
	return s.ListWithFilter(&ListFilter{Tag: tag})
}

// ListByStatus returns all sessions with a specific status.
func (s *Store) ListByStatus(status SessionStatus) ([]*Session, error) {
	return s.ListWithFilter(&ListFilter{Status: status})
}

// ListActive returns all active sessions.
func (s *Store) ListActive() ([]*Session, error) {
	return s.ListByStatus(StatusActive)
}

// ListForWorkDir returns all sessions for the current working directory.
func (s *Store) ListForWorkDir(workDir string) ([]*Session, error) {
	return s.ListWithFilter(&ListFilter{WorkDir: workDir})
}

// Search searches sessions by ID prefix, title, or initial prompt.
func (s *Store) Search(query string) ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureIndexLoadedForRead(); err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matchingIDs []string

	// First pass: filter using index (ID prefix and title)
	for id, meta := range s.index {
		if strings.HasPrefix(strings.ToLower(id), query) {
			matchingIDs = append(matchingIDs, id)
			continue
		}
		if meta.Title != "" && strings.Contains(strings.ToLower(meta.Title), query) {
			matchingIDs = append(matchingIDs, id)
			continue
		}
	}

	// Load sessions and also check initial prompt (not in index)
	sessions := make([]*Session, 0, len(matchingIDs))
	checkedIDs := make(map[string]bool)

	for _, id := range matchingIDs {
		sess, err := s.getLocked(id)
		if err != nil {
			continue
		}
		sessions = append(sessions, sess)
		checkedIDs[id] = true
	}

	// Check remaining sessions for prompt match
	for id := range s.index {
		if checkedIDs[id] {
			continue
		}
		sess, err := s.getLocked(id)
		if err != nil {
			continue
		}
		if sess.InitialPrompt != "" && strings.Contains(strings.ToLower(sess.InitialPrompt), query) {
			sessions = append(sessions, sess)
		}
	}

	// Sort by last used
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastUsed.After(sessions[j].LastUsed)
	})

	return sessions, nil
}

// GetByPrefix returns a session by ID prefix (for short ID lookup).
func (s *Store) GetByPrefix(prefix string) (*Session, error) {
	// Validate prefix before any locking
	if err := validateSessionPrefix(prefix); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureIndexLoadedForRead(); err != nil {
		return nil, err
	}

	var matches []string
	for id := range s.index {
		if strings.HasPrefix(id, prefix) {
			matches = append(matches, id)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no session found with prefix: %s", prefix)
	case 1:
		return s.getLocked(matches[0])
	default:
		return nil, fmt.Errorf("ambiguous prefix %s: matches %d sessions", prefix, len(matches))
	}
}

// Fork creates a new session based on an existing one.
func (s *Store) Fork(sessionID string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	original, err := s.getLocked(sessionID)
	if err != nil {
		return nil, err
	}

	forked, err := original.Fork()
	if err != nil {
		return nil, err
	}

	if err := s.saveLocked(forked); err != nil {
		return nil, err
	}

	s.updateIndex(forked)
	return forked, nil
}

// CreateWithOptions creates a new session with options and saves it.
func (s *Store) CreateWithOptions(backend, workDir string, opts *SessionOptions) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureStoreDirLocked(); err != nil {
		return nil, fmt.Errorf("failed to create sessions dir: %w", err)
	}

	sess, err := NewSessionWithOptions(backend, workDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if err := s.saveLocked(sess); err != nil {
		return nil, err
	}

	s.updateIndex(sess)
	return sess, nil
}

// Stats returns statistics about all sessions.
// Uses the index for efficient calculation without loading full sessions.
func (s *Store) Stats() (*StoreStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureIndexLoadedForRead(); err != nil {
		return nil, err
	}

	stats := &StoreStats{
		TotalSessions:     len(s.index),
		SessionsByBackend: make(map[string]int),
		SessionsByStatus:  make(map[SessionStatus]int),
	}

	for _, meta := range s.index {
		stats.SessionsByBackend[meta.Backend]++
		if meta.Status != "" {
			stats.SessionsByStatus[meta.Status]++
		}
	}

	// Token usage requires loading full sessions (expensive)
	// Only calculate if needed - for now we skip it in the fast path
	// Callers can use List() + iterate if they need token stats

	return stats, nil
}

// StatsWithTokens returns full statistics including token usage.
// This is slower as it loads all sessions.
func (s *Store) StatsWithTokens() (*StoreStats, error) {
	sessions, err := s.List()
	if err != nil {
		return nil, err
	}

	stats := &StoreStats{
		TotalSessions:     len(sessions),
		SessionsByBackend: make(map[string]int),
		SessionsByStatus:  make(map[SessionStatus]int),
	}

	for _, sess := range sessions {
		stats.SessionsByBackend[sess.Backend]++
		if sess.Status != "" {
			stats.SessionsByStatus[sess.Status]++
		}
		if sess.TokenUsage != nil {
			stats.TotalInputTokens += sess.TokenUsage.InputTokens
			stats.TotalOutputTokens += sess.TokenUsage.OutputTokens
		}
	}

	return stats, nil
}

// StoreStats provides statistics about the session store.
type StoreStats struct {
	TotalSessions     int
	SessionsByBackend map[string]int
	SessionsByStatus  map[SessionStatus]int
	TotalInputTokens  int64
	TotalOutputTokens int64
}

// Count returns the number of sessions in the store.
// This is very fast as it only checks the index size.
func (s *Store) Count() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureIndexLoadedForRead(); err != nil {
		return 0, err
	}

	return len(s.index), nil
}

func (s *Store) sessionPath(id string) string {
	return filepath.Join(s.dir, id+".json")
}

// ensureStoreDirLocked creates the store directory. Caller must hold s.mu.
func (s *Store) ensureStoreDirLocked() error {
	if s.dir == "" {
		s.dir = config.SessionsDir()
	}
	// Use 0700 for security - only owner can access session data
	return os.MkdirAll(s.dir, 0700)
}

package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/signalridge/clinvoker/internal/config"
)

// Store handles session persistence.
type Store struct {
	mu  sync.Mutex
	dir string
}

// NewStore creates a new session store.
func NewStore() *Store {
	return &Store{
		dir: config.SessionsDir(),
	}
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

	return sess, nil
}

// Save persists a session to disk.
func (s *Store) Save(sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveLocked(sess)
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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.deleteLocked(id)
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
	os.RemoveAll(artifactsDir)

	return nil
}

// List returns all sessions, sorted by last used (most recent first).
func (s *Store) List() ([]*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.listLocked()
}

// listLocked returns all sessions. Caller must hold s.mu.
func (s *Store) listLocked() ([]*Session, error) {
	if err := s.ensureStoreDirLocked(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read sessions dir: %w", err)
	}

	var sessions []*Session
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // remove .json
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

// Last returns the most recently used session.
func (s *Store) Last() (*Session, error) {
	sessions, err := s.List()
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}
	return sessions[0], nil
}

// LastForBackend returns the most recently used session for a backend.
func (s *Store) LastForBackend(backend string) (*Session, error) {
	sessions, err := s.List()
	if err != nil {
		return nil, err
	}

	for _, sess := range sessions {
		if sess.Backend == backend {
			return sess, nil
		}
	}

	return nil, fmt.Errorf("no sessions found for backend: %s", backend)
}

// Clean removes sessions older than the specified duration.
func (s *Store) Clean(maxAge time.Duration) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessions, err := s.listLocked()
	if err != nil {
		return 0, err
	}

	var deleted int
	for _, sess := range sessions {
		if sess.IdleDuration() > maxAge {
			if err := s.deleteLocked(sess.ID); err == nil {
				deleted++
			}
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
}

// ListWithFilter returns sessions matching the filter criteria.
func (s *Store) ListWithFilter(filter *ListFilter) ([]*Session, error) {
	sessions, err := s.List()
	if err != nil {
		return nil, err
	}

	if filter == nil {
		return sessions, nil
	}

	var filtered []*Session
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
		if filter.Model != "" && sess.Model != filter.Model {
			continue
		}
		if filter.WorkDir != "" && sess.WorkingDir != filter.WorkDir {
			continue
		}

		filtered = append(filtered, sess)

		if filter.Limit > 0 && len(filtered) >= filter.Limit {
			break
		}
	}

	return filtered, nil
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
	sessions, err := s.List()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matches []*Session
	for _, sess := range sessions {
		if strings.HasPrefix(strings.ToLower(sess.ID), query) {
			matches = append(matches, sess)
			continue
		}
		if sess.Title != "" && strings.Contains(strings.ToLower(sess.Title), query) {
			matches = append(matches, sess)
			continue
		}
		if sess.InitialPrompt != "" && strings.Contains(strings.ToLower(sess.InitialPrompt), query) {
			matches = append(matches, sess)
			continue
		}
	}

	return matches, nil
}

// GetByPrefix returns a session by ID prefix (for short ID lookup).
func (s *Store) GetByPrefix(prefix string) (*Session, error) {
	sessions, err := s.List()
	if err != nil {
		return nil, err
	}

	var matches []*Session
	for _, sess := range sessions {
		if strings.HasPrefix(sess.ID, prefix) {
			matches = append(matches, sess)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no session found with prefix: %s", prefix)
	case 1:
		return matches[0], nil
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

	return sess, nil
}

// Stats returns statistics about all sessions.
func (s *Store) Stats() (*StoreStats, error) {
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

func (s *Store) sessionPath(id string) string {
	return filepath.Join(s.dir, id+".json")
}

// ensureStoreDirLocked creates the store directory. Caller must hold s.mu.
func (s *Store) ensureStoreDirLocked() error {
	if s.dir == "" {
		s.dir = config.SessionsDir()
	}
	return os.MkdirAll(s.dir, 0755)
}

package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"
)

// KeyStatus represents lifecycle status for an API key.
type KeyStatus string

// Key status values.
const (
	KeyStatusActive  KeyStatus = "active"
	KeyStatusGrace   KeyStatus = "grace"
	KeyStatusRevoked KeyStatus = "revoked"
)

const lifecycleEventResultOK = "ok"

// KeyMetadata describes non-secret lifecycle metadata for a key.
type KeyMetadata struct {
	KeyID      string    `json:"key_id"`
	Source     string    `json:"source"`
	Status     KeyStatus `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	RotatedAt  time.Time `json:"rotated_at,omitempty"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
	LastUsedAt time.Time `json:"last_used_at,omitempty"`
	Note       string    `json:"note,omitempty"`
}

// LifecycleEvent records auditable lifecycle operations.
type LifecycleEvent struct {
	RequestID   string    `json:"request_id,omitempty"`
	Operator    string    `json:"operator"`
	Operation   string    `json:"operation"`
	TargetKeyID string    `json:"target_key_id"`
	Timestamp   time.Time `json:"timestamp"`
	Result      string    `json:"result"`
	Reason      string    `json:"reason,omitempty"`
}

type managedKey struct {
	secret string
	meta   KeyMetadata
}

// LifecycleManager keeps in-memory key lifecycle state and audit records.
type LifecycleManager struct {
	mu sync.RWMutex

	source string

	keys map[string]*managedKey

	// previousActive stores a rollback snapshot for transition windows.
	previousActiveIDs []string
	transitionUntil   time.Time

	events []LifecycleEvent
}

// NewLifecycleManager creates a new lifecycle manager.
func NewLifecycleManager(source string) *LifecycleManager {
	source = strings.TrimSpace(source)
	if source == "" {
		source = "managed"
	}
	return &LifecycleManager{
		source: source,
		keys:   make(map[string]*managedKey),
		events: make([]LifecycleEvent, 0, 16),
	}
}

// SanitizeKeyID returns a deterministic masked key identifier.
func SanitizeKeyID(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return "key_" + hex.EncodeToString(h[:])[:12]
}

func detectSourceForLoadedKeys() string {
	if len(loadFromEnv()) > 0 {
		return "env"
	}
	if len(loadFromGopass()) > 0 {
		return "gopass"
	}
	return "managed"
}

// LoadAPIKeyMetadata returns metadata for currently loaded keys without exposing plaintext keys.
func LoadAPIKeyMetadata() []KeyMetadata {
	keys := LoadAPIKeys()
	now := time.Now().UTC()
	source := detectSourceForLoadedKeys()

	metas := make([]KeyMetadata, 0, len(keys))
	for _, key := range keys {
		metas = append(metas, KeyMetadata{
			KeyID:     SanitizeKeyID(key),
			Source:    source,
			Status:    KeyStatusActive,
			CreatedAt: now,
		})
	}
	return metas
}

func (m *LifecycleManager) appendEvent(operator, operation, keyID, reason string) {
	if strings.TrimSpace(operator) == "" {
		operator = "unknown"
	}
	m.events = append(m.events, LifecycleEvent{
		Operator:    operator,
		Operation:   operation,
		TargetKeyID: keyID,
		Timestamp:   time.Now().UTC(),
		Result:      lifecycleEventResultOK,
		Reason:      reason,
	})
}

// SeedActiveKeys initializes active keys with metadata if not already present.
func (m *LifecycleManager) SeedActiveKeys(keys []string, operator, note string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		keyID := SanitizeKeyID(key)
		if _, exists := m.keys[keyID]; exists {
			continue
		}
		m.keys[keyID] = &managedKey{
			secret: key,
			meta: KeyMetadata{
				KeyID:     keyID,
				Source:    m.source,
				Status:    KeyStatusActive,
				CreatedAt: now,
				Note:      note,
			},
		}
		m.appendEvent(operator, "create", keyID, "")
	}
}

// ListMetadata returns all known key metadata sorted by key id.
func (m *LifecycleManager) ListMetadata() []KeyMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metas := make([]KeyMetadata, 0, len(m.keys))
	for _, k := range m.keys {
		metas = append(metas, k.meta)
	}
	sort.Slice(metas, func(i, j int) bool { return metas[i].KeyID < metas[j].KeyID })
	return metas
}

// Events returns lifecycle audit events in append order.
func (m *LifecycleManager) Events() []LifecycleEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]LifecycleEvent, len(m.events))
	copy(out, m.events)
	return out
}

// Rotate adds new active keys and puts previous active keys into grace period.
func (m *LifecycleManager) Rotate(newKeys []string, transitionWindow time.Duration, operator, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	if transitionWindow < 0 {
		transitionWindow = 0
	}

	// Capture previous active set for rollback.
	m.previousActiveIDs = m.previousActiveIDs[:0]
	for keyID, entry := range m.keys {
		if entry.meta.Status == KeyStatusActive {
			m.previousActiveIDs = append(m.previousActiveIDs, keyID)
			entry.meta.Status = KeyStatusGrace
			entry.meta.RotatedAt = now
		}
	}
	sort.Strings(m.previousActiveIDs)
	m.transitionUntil = now.Add(transitionWindow)

	for _, key := range newKeys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		keyID := SanitizeKeyID(key)
		entry, exists := m.keys[keyID]
		if !exists {
			entry = &managedKey{secret: key}
			m.keys[keyID] = entry
		}
		entry.secret = key
		entry.meta.KeyID = keyID
		entry.meta.Source = m.source
		entry.meta.Status = KeyStatusActive
		if entry.meta.CreatedAt.IsZero() {
			entry.meta.CreatedAt = now
		}
		entry.meta.RotatedAt = now
		m.appendEvent(operator, "rotate", keyID, reason)
	}
}

// RevokeGrace revokes keys currently in grace state.
func (m *LifecycleManager) RevokeGrace(operator, reason string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	revoked := 0
	for keyID, entry := range m.keys {
		if entry.meta.Status != KeyStatusGrace {
			continue
		}
		entry.meta.Status = KeyStatusRevoked
		revoked++
		m.appendEvent(operator, "revoke", keyID, reason)
	}
	return revoked
}

// Rollback restores previous active keys when transition window has not expired.
func (m *LifecycleManager) Rollback(operator, reason string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.previousActiveIDs) == 0 {
		return false
	}
	if !m.transitionUntil.IsZero() && time.Now().UTC().After(m.transitionUntil) {
		return false
	}

	for keyID, entry := range m.keys {
		if entry.meta.Status == KeyStatusActive {
			entry.meta.Status = KeyStatusRevoked
			m.appendEvent(operator, "rollback-revoke", keyID, reason)
		}
	}

	for _, keyID := range m.previousActiveIDs {
		if entry, ok := m.keys[keyID]; ok {
			entry.meta.Status = KeyStatusActive
			m.appendEvent(operator, "rollback-restore", keyID, reason)
		}
	}

	m.previousActiveIDs = nil
	m.transitionUntil = time.Time{}
	return true
}

// ActiveKeys returns effective active key secrets.
// During transition windows, grace keys remain accepted.
func (m *LifecycleManager) ActiveKeys(now time.Time) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if now.IsZero() {
		now = time.Now().UTC()
	}

	keys := make([]string, 0, len(m.keys))
	for _, entry := range m.keys {
		switch entry.meta.Status {
		case KeyStatusActive:
			keys = append(keys, entry.secret)
		case KeyStatusGrace:
			if m.transitionUntil.IsZero() || now.Before(m.transitionUntil) || now.Equal(m.transitionUntil) {
				keys = append(keys, entry.secret)
			}
		case KeyStatusRevoked:
			// Revoked keys must never be returned as active.
		}
	}
	return keys
}

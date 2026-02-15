package auth

import (
	"strings"
	"testing"
	"time"
)

func TestSanitizeKeyID(t *testing.T) {
	id1 := SanitizeKeyID("secret-a")
	id2 := SanitizeKeyID("secret-a")
	id3 := SanitizeKeyID("secret-b")

	if id1 != id2 {
		t.Fatalf("expected deterministic key id, got %q and %q", id1, id2)
	}
	if id1 == id3 {
		t.Fatalf("expected different ids for different secrets, got %q", id1)
	}
	if !strings.HasPrefix(id1, "key_") {
		t.Fatalf("expected key id prefix 'key_', got %q", id1)
	}
}

func TestLifecycleManager_RotateRollbackAndRevoke(t *testing.T) {
	mgr := NewLifecycleManager("managed")
	mgr.SeedActiveKeys([]string{"old-key"}, "tester", "initial")

	before := mgr.ActiveKeys(time.Now())
	if len(before) != 1 || before[0] != "old-key" {
		t.Fatalf("active keys before rotate = %v, want [old-key]", before)
	}

	mgr.Rotate([]string{"new-key"}, 10*time.Minute, "tester", "rotate")
	activeDuringTransition := mgr.ActiveKeys(time.Now())
	if len(activeDuringTransition) != 2 {
		t.Fatalf("active keys during transition = %v, want 2 keys", activeDuringTransition)
	}

	if !mgr.Rollback("tester", "bad rollout") {
		t.Fatal("expected rollback to succeed during transition window")
	}

	afterRollback := mgr.ActiveKeys(time.Now())
	if len(afterRollback) != 1 || afterRollback[0] != "old-key" {
		t.Fatalf("active keys after rollback = %v, want [old-key]", afterRollback)
	}

	mgr.Rotate([]string{"new-key"}, 0, "tester", "rotate-again")
	revoked := mgr.RevokeGrace("tester", "transition complete")
	if revoked < 1 {
		t.Fatalf("expected at least one grace key revoked, got %d", revoked)
	}

	events := mgr.Events()
	if len(events) == 0 {
		t.Fatal("expected lifecycle audit events")
	}
}

func TestLoadAPIKeyMetadata(t *testing.T) {
	ResetCache()
	t.Cleanup(ResetCache)
	t.Setenv(EnvAPIKeys, "k1,k2")

	metas := LoadAPIKeyMetadata()
	if len(metas) != 2 {
		t.Fatalf("metadata length = %d, want 2", len(metas))
	}
	for _, meta := range metas {
		if meta.KeyID == "" {
			t.Fatal("metadata key_id should not be empty")
		}
		if meta.Status != KeyStatusActive {
			t.Fatalf("metadata status = %q, want %q", meta.Status, KeyStatusActive)
		}
		if meta.Source != "env" {
			t.Fatalf("metadata source = %q, want env", meta.Source)
		}
	}
}

func TestLifecycleManager_TransitionWindowExpiry(t *testing.T) {
	mgr := NewLifecycleManager("managed")
	mgr.SeedActiveKeys([]string{"old-key"}, "tester", "initial")

	mgr.Rotate([]string{"new-key"}, 20*time.Millisecond, "tester", "rotation")

	during := mgr.ActiveKeys(time.Now().UTC())
	if len(during) != 2 {
		t.Fatalf("active keys during transition = %v, want 2 keys", during)
	}

	time.Sleep(50 * time.Millisecond)

	after := mgr.ActiveKeys(time.Now().UTC())
	if len(after) != 1 || after[0] != "new-key" {
		t.Fatalf("active keys after transition expiry = %v, want [new-key]", after)
	}

	if mgr.Rollback("tester", "expired transition window") {
		t.Fatal("rollback should fail after transition window expiration")
	}
}

func TestLifecycleManager_EventsContainRequiredFields(t *testing.T) {
	mgr := NewLifecycleManager("managed")
	mgr.SeedActiveKeys([]string{"old-key"}, "ops-user", "bootstrap")
	mgr.Rotate([]string{"new-key"}, 0, "ops-user", "routine rotation")
	_ = mgr.RevokeGrace("ops-user", "grace window ended")

	events := mgr.Events()
	if len(events) == 0 {
		t.Fatal("expected lifecycle audit events")
	}

	for i, event := range events {
		if strings.TrimSpace(event.Operator) == "" {
			t.Fatalf("events[%d].Operator should not be empty", i)
		}
		if strings.TrimSpace(event.Operation) == "" {
			t.Fatalf("events[%d].Operation should not be empty", i)
		}
		if strings.TrimSpace(event.TargetKeyID) == "" {
			t.Fatalf("events[%d].TargetKeyID should not be empty", i)
		}
		if event.Timestamp.IsZero() {
			t.Fatalf("events[%d].Timestamp should not be zero", i)
		}
		if strings.TrimSpace(event.Result) == "" {
			t.Fatalf("events[%d].Result should not be empty", i)
		}
	}
}

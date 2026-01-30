package session

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFileLock_Basic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lock := NewFileLock(filepath.Join(tmpDir, "test"))

	// Test lock/unlock
	if err := lock.Lock(); err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	if err := lock.Unlock(); err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}
}

func TestFileLock_TryLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lock := NewFileLock(filepath.Join(tmpDir, "test"))

	// First TryLock should succeed
	if err := lock.TryLock(); err != nil {
		t.Fatalf("first TryLock should succeed: %v", err)
	}
	defer lock.Unlock()

	// Second TryLock on same lock should fail (same process, different lock instance)
	lock2 := NewFileLock(filepath.Join(tmpDir, "test"))
	err = lock2.TryLock()
	if err == nil {
		lock2.Unlock()
		t.Fatal("second TryLock should fail when lock is held")
	}
	if err != ErrLockBusy {
		t.Fatalf("expected ErrLockBusy, got: %v", err)
	}
}

func TestFileLock_SharedLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lock1 := NewFileLock(filepath.Join(tmpDir, "test"))
	lock2 := NewFileLock(filepath.Join(tmpDir, "test"))

	// First shared lock
	if err := lock1.LockShared(); err != nil {
		t.Fatalf("first shared lock should succeed: %v", err)
	}
	defer lock1.Unlock()

	// Second shared lock should also succeed
	if err := lock2.TryLockShared(); err != nil {
		t.Fatalf("second shared lock should succeed: %v", err)
	}
	if err := lock2.Unlock(); err != nil {
		t.Fatalf("failed to release shared lock: %v", err)
	}
}

func TestFileLock_ExclusiveBlocksShared(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lock1 := NewFileLock(filepath.Join(tmpDir, "test"))
	lock2 := NewFileLock(filepath.Join(tmpDir, "test"))

	// Acquire exclusive lock
	if err := lock1.Lock(); err != nil {
		t.Fatalf("exclusive lock should succeed: %v", err)
	}
	defer lock1.Unlock()

	// Shared lock should fail
	err = lock2.TryLockShared()
	if err == nil {
		lock2.Unlock()
		t.Fatal("shared lock should fail when exclusive lock is held")
	}
}

func TestFileLock_WithLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lock := NewFileLock(filepath.Join(tmpDir, "test"))

	executed := false
	err = lock.WithLock(func() error {
		executed = true
		return nil
	})

	if err != nil {
		t.Fatalf("WithLock failed: %v", err)
	}
	if !executed {
		t.Fatal("function was not executed")
	}
}

func TestFileLock_LockWithTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lock1 := NewFileLock(filepath.Join(tmpDir, "test"))
	lock2 := NewFileLock(filepath.Join(tmpDir, "test"))

	// Acquire first lock
	if err := lock1.Lock(); err != nil {
		t.Fatalf("first lock should succeed: %v", err)
	}
	defer lock1.Unlock()

	// Second lock with timeout should fail
	start := time.Now()
	err = lock2.LockWithTimeout(50 * time.Millisecond)
	elapsed := time.Since(start)

	if err != ErrLockTimeout {
		t.Fatalf("expected ErrLockTimeout, got: %v", err)
	}

	// Should have waited approximately the timeout duration
	if elapsed < 40*time.Millisecond {
		t.Fatalf("should have waited for timeout, elapsed: %v", elapsed)
	}
}

func TestFileLock_Concurrent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test")
	var counter atomic.Int32
	var wg sync.WaitGroup
	iterations := 10

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lock := NewFileLock(lockPath)
			if err := lock.Lock(); err != nil {
				t.Errorf("failed to acquire lock: %v", err)
				return
			}
			defer lock.Unlock()
			counter.Add(1)
		}()
	}

	wg.Wait()

	if counter.Load() != int32(iterations) {
		t.Errorf("expected counter %d, got %d", iterations, counter.Load())
	}
}

func TestFileLock_UnlockWithoutLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lock := NewFileLock(filepath.Join(tmpDir, "test"))

	// Unlock without lock should not error
	if err := lock.Unlock(); err != nil {
		t.Fatalf("unlock without lock should not error: %v", err)
	}
}

func TestFileLock_CreatesDirIfNeeded(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filelock-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Lock in a subdirectory that doesn't exist
	lockPath := filepath.Join(tmpDir, "subdir", "test")
	lock := NewFileLock(lockPath)

	if err := lock.Lock(); err != nil {
		t.Fatalf("should create directory and acquire lock: %v", err)
	}
	defer lock.Unlock()

	// Verify directory was created
	if _, err := os.Stat(filepath.Join(tmpDir, "subdir")); os.IsNotExist(err) {
		t.Fatal("directory was not created")
	}
}

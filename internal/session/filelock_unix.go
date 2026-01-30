//go:build !windows

package session

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// FileLock provides cross-process file locking using flock.
// This ensures multiple processes (CLI instances, server instances) can safely
// access the session store without corrupting data.
type FileLock struct {
	path   string
	file   *os.File
	mu     sync.Mutex  // Held between Lock() and Unlock() to serialize access
	locked atomic.Bool // True when lock is held (both mu and flock)
}

// NewFileLock creates a new file lock for the given path.
// The lock file will be created at path + ".lock".
func NewFileLock(path string) *FileLock {
	return &FileLock{
		path: path + ".lock",
	}
}

// Lock acquires an exclusive lock, blocking until available.
// Returns an error if the lock cannot be acquired.
func (l *FileLock) Lock() error {
	return l.lockWithTimeout(0, false)
}

// LockShared acquires a shared (read) lock, blocking until available.
// Multiple processes can hold shared locks simultaneously.
func (l *FileLock) LockShared() error {
	return l.lockWithTimeout(0, true)
}

// TryLock attempts to acquire an exclusive lock without blocking.
// Returns ErrLockBusy if the lock is held by another process.
func (l *FileLock) TryLock() error {
	return l.tryLockInternal(false)
}

// TryLockShared attempts to acquire a shared lock without blocking.
// Returns ErrLockBusy if an exclusive lock is held by another process.
func (l *FileLock) TryLockShared() error {
	return l.tryLockInternal(true)
}

// LockWithTimeout attempts to acquire an exclusive lock with a timeout.
// Returns ErrLockTimeout if the lock cannot be acquired within the timeout.
func (l *FileLock) LockWithTimeout(timeout time.Duration) error {
	return l.lockWithTimeout(timeout, false)
}

// Unlock releases the lock.
// It is safe to call Unlock on an unlocked FileLock.
func (l *FileLock) Unlock() error {
	// Check if we're actually locked (using atomic to avoid race)
	if !l.locked.Load() {
		// Not locked - this is safe to call without holding the lock
		return nil
	}

	// We're locked, so we hold l.mu from Lock() call
	// Release flock first
	if l.file != nil {
		if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
			return fmt.Errorf("failed to unlock file: %w", err)
		}

		// Close and remove the lock file
		if err := l.file.Close(); err != nil {
			return fmt.Errorf("failed to close lock file: %w", err)
		}
		l.file = nil

		// Try to remove the lock file (best effort, may fail if another process holds it)
		_ = os.Remove(l.path)
	}

	// Mark as unlocked and release in-process mutex
	l.locked.Store(false)
	l.mu.Unlock()

	return nil
}

// lockWithTimeout acquires a lock with optional timeout.
// If timeout is 0, blocks indefinitely.
// NOTE: l.mu is held after successful return and must be released by Unlock().
func (l *FileLock) lockWithTimeout(timeout time.Duration, shared bool) error {
	l.mu.Lock()
	// NOTE: Do NOT defer l.mu.Unlock() here - it's released in Unlock()

	// Ensure the lock file's directory exists
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		l.mu.Unlock()
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Open or create the lock file
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		l.mu.Unlock()
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	// Determine lock type
	lockType := syscall.LOCK_EX
	if shared {
		lockType = syscall.LOCK_SH
	}

	// If no timeout, use blocking lock
	if timeout == 0 {
		if err := syscall.Flock(int(file.Fd()), lockType); err != nil {
			_ = file.Close()
			l.mu.Unlock()
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		l.file = file
		l.locked.Store(true)
		return nil
	}

	// With timeout, use polling with non-blocking attempts
	deadline := time.Now().Add(timeout)
	pollInterval := 10 * time.Millisecond

	for {
		err := syscall.Flock(int(file.Fd()), lockType|syscall.LOCK_NB)
		if err == nil {
			l.file = file
			l.locked.Store(true)
			return nil
		}

		if err != syscall.EWOULDBLOCK {
			_ = file.Close()
			l.mu.Unlock()
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if time.Now().After(deadline) {
			_ = file.Close()
			l.mu.Unlock()
			return ErrLockTimeout
		}

		time.Sleep(pollInterval)
		// Exponential backoff up to 100ms
		if pollInterval < 100*time.Millisecond {
			pollInterval *= 2
		}
	}
}

// tryLockInternal attempts to acquire a lock without blocking.
// NOTE: l.mu is held after successful return and must be released by Unlock().
func (l *FileLock) tryLockInternal(shared bool) error {
	l.mu.Lock()
	// NOTE: Do NOT defer l.mu.Unlock() here - it's released in Unlock()

	// Ensure the lock file's directory exists
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		l.mu.Unlock()
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Open or create the lock file
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		l.mu.Unlock()
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	// Determine lock type
	lockType := syscall.LOCK_EX | syscall.LOCK_NB
	if shared {
		lockType = syscall.LOCK_SH | syscall.LOCK_NB
	}

	if err := syscall.Flock(int(file.Fd()), lockType); err != nil {
		_ = file.Close()
		l.mu.Unlock()
		if err == syscall.EWOULDBLOCK {
			return ErrLockBusy
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	l.file = file
	l.locked.Store(true)
	return nil
}

// WithLock executes fn while holding an exclusive lock.
// The lock is automatically released after fn returns.
func (l *FileLock) WithLock(fn func() error) error {
	if err := l.Lock(); err != nil {
		return err
	}
	defer func() {
		_ = l.Unlock()
	}()
	return fn()
}

// WithLockShared executes fn while holding a shared lock.
// The lock is automatically released after fn returns.
func (l *FileLock) WithLockShared(fn func() error) error {
	if err := l.LockShared(); err != nil {
		return err
	}
	defer func() {
		_ = l.Unlock()
	}()
	return fn()
}

//go:build windows

package session

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

// Windows-specific constants for LockFileEx
const (
	lockfileExclusiveLock   = 0x00000002
	lockfileFailImmediately = 0x00000001
)

// FileLock provides cross-process file locking using LockFileEx on Windows.
// This ensures multiple processes (CLI instances, server instances) can safely
// access the session store without corrupting data.
type FileLock struct {
	path   string
	file   *os.File
	mu     sync.Mutex  // Held between Lock() and Unlock() to serialize access
	locked atomic.Bool // True when lock is held (both mu and file lock)
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
	// Release file lock first
	if l.file != nil {
		if err := unlockFile(syscall.Handle(l.file.Fd())); err != nil {
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

	handle := syscall.Handle(file.Fd())

	// If no timeout, use blocking lock
	if timeout == 0 {
		if err := lockFile(handle, shared, false); err != nil {
			file.Close()
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
		err := lockFile(handle, shared, true)
		if err == nil {
			l.file = file
			l.locked.Store(true)
			return nil
		}

		// Check if it's a "lock busy" error
		if !isLockBusyError(err) {
			file.Close()
			l.mu.Unlock()
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if time.Now().After(deadline) {
			file.Close()
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

	handle := syscall.Handle(file.Fd())

	if err := lockFile(handle, shared, true); err != nil {
		file.Close()
		l.mu.Unlock()
		if isLockBusyError(err) {
			return ErrLockBusy
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	l.file = file
	l.locked.Store(true)
	return nil
}

// lockFile locks the file using Windows LockFileEx
func lockFile(handle syscall.Handle, shared bool, nonBlocking bool) error {
	var flags uint32
	if !shared {
		flags |= lockfileExclusiveLock
	}
	if nonBlocking {
		flags |= lockfileFailImmediately
	}

	// Lock the entire file (0 to max)
	ol := new(syscall.Overlapped)
	err := lockFileEx(handle, flags, 0, 1, 0, ol)
	if err != nil {
		return err
	}
	return nil
}

// unlockFile unlocks the file using Windows UnlockFileEx
func unlockFile(handle syscall.Handle) error {
	ol := new(syscall.Overlapped)
	return unlockFileEx(handle, 0, 1, 0, ol)
}

// isLockBusyError checks if the error indicates the lock is busy
func isLockBusyError(err error) bool {
	if err == nil {
		return false
	}
	// ERROR_LOCK_VIOLATION = 33
	// ERROR_IO_PENDING = 997
	if errno, ok := err.(syscall.Errno); ok {
		return errno == 33 || errno == 997
	}
	return false
}

// Windows API wrappers
var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

func lockFileEx(handle syscall.Handle, flags, reserved, bytesLow, bytesHigh uint32, ol *syscall.Overlapped) error {
	r1, _, e1 := syscall.Syscall6(
		procLockFileEx.Addr(),
		6,
		uintptr(handle),
		uintptr(flags),
		uintptr(reserved),
		uintptr(bytesLow),
		uintptr(bytesHigh),
		uintptr(unsafe.Pointer(ol)),
	)
	if r1 == 0 {
		if e1 != 0 {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

func unlockFileEx(handle syscall.Handle, reserved, bytesLow, bytesHigh uint32, ol *syscall.Overlapped) error {
	r1, _, e1 := syscall.Syscall6(
		procUnlockFileEx.Addr(),
		5,
		uintptr(handle),
		uintptr(reserved),
		uintptr(bytesLow),
		uintptr(bytesHigh),
		uintptr(unsafe.Pointer(ol)),
		0,
	)
	if r1 == 0 {
		if e1 != 0 {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// WithLock executes fn while holding an exclusive lock.
// The lock is automatically released after fn returns.
func (l *FileLock) WithLock(fn func() error) error {
	if err := l.Lock(); err != nil {
		return err
	}
	defer l.Unlock()
	return fn()
}

// WithLockShared executes fn while holding a shared lock.
// The lock is automatically released after fn returns.
func (l *FileLock) WithLockShared(fn func() error) error {
	if err := l.LockShared(); err != nil {
		return err
	}
	defer l.Unlock()
	return fn()
}

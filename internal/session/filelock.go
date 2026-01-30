package session

import (
	"fmt"
	"time"
)

// Common lock errors.
var (
	ErrLockBusy    = fmt.Errorf("lock is held by another process")
	ErrLockTimeout = fmt.Errorf("timeout waiting for lock")
)

// FileLockInterface defines the file locking interface.
// Implementations are platform-specific.
type FileLockInterface interface {
	Lock() error
	LockShared() error
	TryLock() error
	TryLockShared() error
	LockWithTimeout(timeout time.Duration) error
	Unlock() error
	WithLock(fn func() error) error
	WithLockShared(fn func() error) error
}

package backend

import (
	"fmt"
	"sync"
	"time"
)

// registry holds all registered backends.
// Access is protected by mu for thread-safety.
var (
	registry = make(map[string]Backend)
	mu       sync.RWMutex
)

// availabilityCache caches backend availability check results
// to avoid repeated exec.LookPath calls during health checks.
var (
	availabilityCache   = make(map[string]*cachedAvailability)
	availabilityCacheMu sync.RWMutex
	// AvailabilityCacheTTL controls how long availability results are cached.
	// Exported for testing purposes.
	AvailabilityCacheTTL = 30 * time.Second
)

// cachedAvailability holds cached availability check result.
type cachedAvailability struct {
	available bool
	checkedAt time.Time
}

// Register adds a backend to the registry.
// This function is thread-safe and can be called from multiple goroutines.
// If a backend with the same name already exists, it will be replaced.
func Register(b Backend) {
	mu.Lock()
	defer mu.Unlock()
	registry[b.Name()] = b

	// Invalidate availability cache for this backend
	InvalidateAvailabilityCache(b.Name())
}

// Unregister removes a backend from the registry by name.
// This is primarily used for testing to clean up temporary registrations.
func Unregister(name string) {
	mu.Lock()
	defer mu.Unlock()
	delete(registry, name)

	// Invalidate availability cache for this backend
	InvalidateAvailabilityCache(name)
}

// Get returns a backend by name.
// Returns an error if a backend is not found in the registry.
// This function is thread-safe and can be called from multiple goroutines.
func Get(name string) (Backend, error) {
	mu.RLock()
	defer mu.RUnlock()

	b, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown backend: %s", name)
	}
	return b, nil
}

// List returns all registered backend names.
// The order of names is not guaranteed to be consistent.
// This function is thread-safe and can be called from multiple goroutines.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// Available returns all available (installed) backends.
// A backend is considered available if its CLI tool is installed
// and accessible in the system PATH.
// This function is thread-safe and can be called from multiple goroutines.
// Uses cached availability checks for performance.
func Available() []Backend {
	mu.RLock()
	defer mu.RUnlock()

	var backends []Backend
	for _, b := range registry {
		if IsAvailableCached(b.Name()) {
			backends = append(backends, b)
		}
	}
	return backends
}

// IsAvailableCached returns cached availability status for a backend.
// If cache is expired or missing, performs actual check and caches result.
// This is significantly faster than calling IsAvailable() directly for
// frequent health checks.
func IsAvailableCached(name string) bool {
	// Fast path: check cache with read lock
	availabilityCacheMu.RLock()
	if cached, ok := availabilityCache[name]; ok && time.Since(cached.checkedAt) < AvailabilityCacheTTL {
		result := cached.available
		availabilityCacheMu.RUnlock()
		return result
	}
	availabilityCacheMu.RUnlock()

	// Cache miss or expired: perform actual check
	mu.RLock()
	b, ok := registry[name]
	mu.RUnlock()

	available := ok && b != nil && b.IsAvailable()

	// Update cache with write lock
	availabilityCacheMu.Lock()
	availabilityCache[name] = &cachedAvailability{
		available: available,
		checkedAt: time.Now(),
	}
	availabilityCacheMu.Unlock()

	return available
}

// InvalidateAvailabilityCache clears the availability cache for a specific backend.
// Call this if you know a backend's availability has changed (e.g., CLI installed/uninstalled).
func InvalidateAvailabilityCache(name string) {
	availabilityCacheMu.Lock()
	delete(availabilityCache, name)
	availabilityCacheMu.Unlock()
}

// InvalidateAllAvailabilityCache clears the entire availability cache.
func InvalidateAllAvailabilityCache() {
	availabilityCacheMu.Lock()
	availabilityCache = make(map[string]*cachedAvailability)
	availabilityCacheMu.Unlock()
}

// UnregisterAll removes all backends from the registry.
// This is primarily used for testing to reset the registry state.
func UnregisterAll() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]Backend)
}

func init() {
	// Register all backends
	Register(&Claude{})
	Register(&Codex{})
	Register(&Gemini{})
}

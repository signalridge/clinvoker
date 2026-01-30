package backend

import (
	"fmt"
	"sync"
	"time"
)

// Registry manages backend registration and lookup.
// It is thread-safe and can be used concurrently from multiple goroutines.
type Registry struct {
	mu                   sync.RWMutex
	backends             map[string]Backend
	availabilityCache    map[string]*cachedAvailability
	availabilityCacheTTL time.Duration
}

// cachedAvailability holds cached availability check result.
type cachedAvailability struct {
	available bool
	checkedAt time.Time
}

// NewRegistry creates a new empty Registry.
// Use this for dependency injection in tests or when you need an isolated registry.
func NewRegistry() *Registry {
	return &Registry{
		backends:             make(map[string]Backend),
		availabilityCache:    make(map[string]*cachedAvailability),
		availabilityCacheTTL: 30 * time.Second,
	}
}

// NewRegistryWithDefaults creates a new Registry pre-populated with default backends.
// This is the recommended way to create a production registry.
func NewRegistryWithDefaults() *Registry {
	r := NewRegistry()
	r.Register(&Claude{})
	r.Register(&Codex{})
	r.Register(&Gemini{})
	return r
}

// Register adds a backend to the registry.
func (r *Registry) Register(b Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends[b.Name()] = b
	delete(r.availabilityCache, b.Name())
}

// Unregister removes a backend from the registry by name.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.backends, name)
	delete(r.availabilityCache, name)
}

// Get returns a backend by name.
func (r *Registry) Get(name string) (Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.backends[name]
	if !ok {
		available := make([]string, 0, len(r.backends))
		for n := range r.backends {
			available = append(available, n)
		}
		return nil, fmt.Errorf("unknown backend %q (available: %v)", name, available)
	}
	return b, nil
}

// List returns all registered backend names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.backends))
	for name := range r.backends {
		names = append(names, name)
	}
	return names
}

// Available returns all available (installed) backends.
func (r *Registry) Available() []Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var backends []Backend
	for _, b := range r.backends {
		if r.isAvailableCachedLocked(b) {
			backends = append(backends, b)
		}
	}
	return backends
}

// isAvailableCachedLocked checks availability with caching.
// Caller must hold at least a read lock on r.mu.
func (r *Registry) isAvailableCachedLocked(b Backend) bool {
	name := b.Name()
	if cached, ok := r.availabilityCache[name]; ok && time.Since(cached.checkedAt) < r.availabilityCacheTTL {
		return cached.available
	}

	available := b.IsAvailable()
	r.availabilityCache[name] = &cachedAvailability{
		available: available,
		checkedAt: time.Now(),
	}
	return available
}

// IsAvailableCached returns cached availability status for a backend.
func (r *Registry) IsAvailableCached(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.backends[name]
	if !ok || b == nil {
		return false
	}
	return r.isAvailableCachedLocked(b)
}

// InvalidateAvailabilityCache clears the availability cache for a specific backend.
func (r *Registry) InvalidateAvailabilityCache(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.availabilityCache, name)
}

// InvalidateAllAvailabilityCache clears the entire availability cache.
func (r *Registry) InvalidateAllAvailabilityCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.availabilityCache = make(map[string]*cachedAvailability)
}

// SetAvailabilityCacheTTL sets the cache TTL for availability checks.
func (r *Registry) SetAvailabilityCacheTTL(ttl time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.availabilityCacheTTL = ttl
}

// UnregisterAll removes all backends from the registry.
func (r *Registry) UnregisterAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends = make(map[string]Backend)
	r.availabilityCache = make(map[string]*cachedAvailability)
}

// ============================================================================
// Global registry (backward compatibility)
// ============================================================================

// globalRegistry is the default registry used by package-level functions.
var globalRegistry = NewRegistry()

// AvailabilityCacheTTL controls how long availability results are cached.
// Exported for testing purposes. Deprecated: use Registry.SetAvailabilityCacheTTL instead.
var AvailabilityCacheTTL = 30 * time.Second

// Register adds a backend to the global registry.
// This function is thread-safe and can be called from multiple goroutines.
// If a backend with the same name already exists, it will be replaced.
func Register(b Backend) {
	globalRegistry.Register(b)
}

// Unregister removes a backend from the global registry by name.
// This is primarily used for testing to clean up temporary registrations.
func Unregister(name string) {
	globalRegistry.Unregister(name)
}

// Get returns a backend by name from the global registry.
// Returns an error if a backend is not found in the registry.
// This function is thread-safe and can be called from multiple goroutines.
func Get(name string) (Backend, error) {
	return globalRegistry.Get(name)
}

// List returns all registered backend names from the global registry.
// The order of names is not guaranteed to be consistent.
// This function is thread-safe and can be called from multiple goroutines.
func List() []string {
	return globalRegistry.List()
}

// Available returns all available (installed) backends from the global registry.
// A backend is considered available if its CLI tool is installed
// and accessible in the system PATH.
// This function is thread-safe and can be called from multiple goroutines.
// Uses cached availability checks for performance.
func Available() []Backend {
	return globalRegistry.Available()
}

// IsAvailableCached returns cached availability status for a backend from the global registry.
// If cache is expired or missing, performs actual check and caches result.
// This is significantly faster than calling IsAvailable() directly for
// frequent health checks.
func IsAvailableCached(name string) bool {
	return globalRegistry.IsAvailableCached(name)
}

// InvalidateAvailabilityCache clears the availability cache for a specific backend.
// Call this if you know a backend's availability has changed (e.g., CLI installed/uninstalled).
func InvalidateAvailabilityCache(name string) {
	globalRegistry.InvalidateAvailabilityCache(name)
}

// InvalidateAllAvailabilityCache clears the entire availability cache.
func InvalidateAllAvailabilityCache() {
	globalRegistry.InvalidateAllAvailabilityCache()
}

// UnregisterAll removes all backends from the global registry.
// This is primarily used for testing to reset the registry state.
func UnregisterAll() {
	globalRegistry.UnregisterAll()
}

// DefaultRegistry returns the global registry instance.
// This can be used to access registry methods directly or for dependency injection.
func DefaultRegistry() *Registry {
	return globalRegistry
}

func init() {
	// Register all default backends with the global registry
	globalRegistry.Register(&Claude{})
	globalRegistry.Register(&Codex{})
	globalRegistry.Register(&Gemini{})
}

package backend

import (
	"fmt"
	"sync"
)

// registry holds all registered backends.
// Access is protected by mu for thread-safety.
var (
	registry = make(map[string]Backend)
	mu       sync.RWMutex
)

// Register adds a backend to the registry.
// This function is thread-safe and can be called from multiple goroutines.
// If a backend with the same name already exists, it will be replaced.
func Register(b Backend) {
	mu.Lock()
	defer mu.Unlock()
	registry[b.Name()] = b
}

// Get returns a backend by name.
// Returns an error if the backend is not found in the registry.
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
func Available() []Backend {
	mu.RLock()
	defer mu.RUnlock()

	var backends []Backend
	for _, b := range registry {
		if b.IsAvailable() {
			backends = append(backends, b)
		}
	}
	return backends
}

// register is an internal function used during testing to add backends.
// For normal use, backends are registered in init().
func register(b Backend) {
	Register(b)
}

// unregisterAll removes all backends from the registry.
// This is primarily used for testing to reset the registry state.
func unregisterAll() {
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

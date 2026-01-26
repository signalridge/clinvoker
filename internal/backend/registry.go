package backend

import (
	"fmt"
	"sync"
)

var (
	registry = make(map[string]Backend)
	mu       sync.RWMutex
)

// Register adds a backend to the registry.
func Register(b Backend) {
	mu.Lock()
	defer mu.Unlock()
	registry[b.Name()] = b
}

// Get returns a backend by name.
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

func init() {
	// Register all backends
	Register(&Claude{})
	Register(&Codex{})
	Register(&Gemini{})
}

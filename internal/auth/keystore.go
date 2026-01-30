// Package auth provides authentication utilities for the server.
package auth

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/signalridge/clinvoker/internal/config"
)

const (
	// EnvAPIKeys is the environment variable name for API keys.
	EnvAPIKeys = "CLINVK_API_KEYS" //nolint:gosec // Not a credential, just env var name

	// GopassPath is the gopass path for API keys.
	GopassPath = "clinvk/server/api-keys" //nolint:gosec // Not a credential, just path
)

var (
	cachedKeys []string
	cacheOnce  sync.Once
)

// LoadAPIKeys loads API keys using a layered approach:
// 1. Environment variable CLINVK_API_KEYS (comma-separated)
// 2. gopass clinvk/server/api-keys (if gopass is available)
// 3. Config file api_keys field
//
// The first source with valid keys wins. Empty result means auth is disabled.
func LoadAPIKeys() []string {
	cacheOnce.Do(func() {
		cachedKeys = loadAPIKeysInternal()
	})
	return cachedKeys
}

// loadAPIKeysInternal performs the actual loading without caching.
func loadAPIKeysInternal() []string {
	// 1. Environment variable (highest priority)
	if keys := loadFromEnv(); len(keys) > 0 {
		return keys
	}

	// 2. gopass (if available)
	if keys := loadFromGopass(); len(keys) > 0 {
		return keys
	}

	// 3. Config file (lowest priority)
	return loadFromConfig()
}

// loadFromEnv loads API keys from the CLINVK_API_KEYS environment variable.
func loadFromEnv() []string {
	envKeys := os.Getenv(EnvAPIKeys)
	if envKeys == "" {
		return nil
	}
	return parseKeys(envKeys)
}

// loadFromGopass attempts to load API keys from gopass.
// Returns nil if gopass is not available or the secret doesn't exist.
func loadFromGopass() []string {
	// Check if gopass is available
	if _, err := exec.LookPath("gopass"); err != nil {
		return nil
	}

	// Use a timeout context for the gopass command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gopass", "show", "--password", GopassPath)
	out, err := cmd.Output()
	if err != nil {
		// Silent fallback - gopass might not have this secret
		return nil
	}

	return parseKeys(string(out))
}

// loadFromConfig loads API keys from the configuration file.
func loadFromConfig() []string {
	cfg := config.Get()
	if cfg == nil {
		return nil
	}
	return cfg.Server.APIKeys
}

// parseKeys parses a comma-separated string of API keys.
// Empty and whitespace-only keys are filtered out.
func parseKeys(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

// ResetCache clears the cached API keys (mainly for testing).
func ResetCache() {
	cacheOnce = sync.Once{}
	cachedKeys = nil
}

// HasAPIKeys returns true if any API keys are configured.
func HasAPIKeys() bool {
	return len(LoadAPIKeys()) > 0
}

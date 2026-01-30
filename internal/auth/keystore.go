// Package auth provides authentication utilities for the server.
package auth

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/signalridge/clinvoker/internal/config"
)

// gopassPathPattern validates gopass paths to prevent injection attacks.
// Allows alphanumeric characters, slashes, hyphens, underscores, and dots.
var gopassPathPattern = regexp.MustCompile(`^[a-zA-Z0-9/_.\-]+$`)

const (
	// EnvAPIKeys is the environment variable name for API keys.
	EnvAPIKeys = "CLINVK_API_KEYS" //nolint:gosec // Not a credential, just env var name

	// EnvAPIKeysGopassPath is the environment variable for gopass path.
	EnvAPIKeysGopassPath = "CLINVK_API_KEYS_GOPASS_PATH" //nolint:gosec // Not a credential, just env var name

	// DefaultKeyReloadInterval is the default interval for reloading API keys.
	// Set to 0 to disable automatic reloading.
	DefaultKeyReloadInterval = 5 * time.Minute
)

var (
	cachedKeys     []string
	cacheMu        sync.RWMutex
	cacheLoadedAt  time.Time
	reloadInterval = DefaultKeyReloadInterval
)

// LoadAPIKeys loads API keys using a layered approach:
// 1. Environment variable CLINVK_API_KEYS (comma-separated)
// 2. gopass (if api_keys_gopass_path is configured)
//
// The first source with valid keys wins. Empty result means auth is disabled.
// Note: Config file storage is intentionally NOT supported for security reasons.
// Keys are automatically reloaded after the reload interval (default: 5 minutes).
func LoadAPIKeys() []string {
	cacheMu.RLock()
	// Check if cache is valid
	if cachedKeys != nil && (reloadInterval == 0 || time.Since(cacheLoadedAt) < reloadInterval) {
		keys := cachedKeys
		cacheMu.RUnlock()
		return keys
	}
	cacheMu.RUnlock()

	// Cache miss or expired - reload with write lock
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double-check after acquiring write lock
	if cachedKeys != nil && (reloadInterval == 0 || time.Since(cacheLoadedAt) < reloadInterval) {
		return cachedKeys
	}

	cachedKeys = loadAPIKeysInternal()
	cacheLoadedAt = time.Now()
	return cachedKeys
}

// loadAPIKeysInternal performs the actual loading without caching.
func loadAPIKeysInternal() []string {
	// 1. Environment variable (highest priority)
	if keys := loadFromEnv(); len(keys) > 0 {
		return keys
	}

	// 2. gopass (if configured)
	return loadFromGopass()
}

// loadFromEnv loads API keys from the CLINVK_API_KEYS environment variable.
func loadFromEnv() []string {
	envKeys := os.Getenv(EnvAPIKeys)
	if envKeys == "" {
		return nil
	}
	return parseKeys(envKeys)
}

// getGopassPath returns the gopass path from env or config.
// Returns empty string if not configured.
func getGopassPath() string {
	// 1. Check environment variable first
	if path := os.Getenv(EnvAPIKeysGopassPath); path != "" {
		return path
	}

	// 2. Check config file
	cfg := config.Get()
	if cfg != nil && cfg.Server.APIKeysGopassPath != "" {
		return cfg.Server.APIKeysGopassPath
	}

	return ""
}

// loadFromGopass attempts to load API keys from gopass.
// Returns nil if gopass path is not configured, gopass is not available,
// or the secret doesn't exist.
func loadFromGopass() []string {
	// Get the gopass path from config
	gopassPath := getGopassPath()
	if gopassPath == "" {
		return nil // Not configured, skip gopass
	}

	// Validate gopass path to prevent injection attacks
	if !isValidGopassPath(gopassPath) {
		return nil // Invalid path format
	}

	// Check if gopass is available
	if _, err := exec.LookPath("gopass"); err != nil {
		return nil
	}

	// Use a timeout context for the gopass command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gopass", "show", "--password", gopassPath)
	out, err := cmd.Output()
	if err != nil {
		// Silent fallback - gopass might not have this secret
		return nil
	}

	return parseKeys(string(out))
}

// isValidGopassPath validates that the gopass path contains only safe characters.
// Allowed: alphanumeric, slashes, hyphens, underscores, dots.
func isValidGopassPath(path string) bool {
	if path == "" {
		return false
	}
	return gopassPathPattern.MatchString(path)
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

// ResetCache clears the cached API keys, forcing a reload on next access.
// This can be used for testing or to manually trigger a key reload.
func ResetCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cachedKeys = nil
	cacheLoadedAt = time.Time{}
}

// SetReloadInterval sets the interval for automatic key reloading.
// Set to 0 to disable automatic reloading (keys loaded once).
// This should be called before the server starts.
func SetReloadInterval(interval time.Duration) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	reloadInterval = interval
}

// ForceReload immediately reloads API keys from all sources.
// Returns the newly loaded keys.
func ForceReload() []string {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cachedKeys = loadAPIKeysInternal()
	cacheLoadedAt = time.Now()
	return cachedKeys
}

// HasAPIKeys returns true if any API keys are configured.
func HasAPIKeys() bool {
	return len(LoadAPIKeys()) > 0
}

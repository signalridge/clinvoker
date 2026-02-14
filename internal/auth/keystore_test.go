package auth

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/config"
)

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
}

func TestParseKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "single key",
			input:    "key1",
			expected: []string{"key1"},
		},
		{
			name:     "multiple keys",
			input:    "key1,key2,key3",
			expected: []string{"key1", "key2", "key3"},
		},
		{
			name:     "keys with whitespace",
			input:    " key1 , key2 , key3 ",
			expected: []string{"key1", "key2", "key3"},
		},
		{
			name:     "empty keys filtered",
			input:    "key1,,key2,  ,key3",
			expected: []string{"key1", "key2", "key3"},
		},
		{
			name:     "trailing newline",
			input:    "key1,key2\n",
			expected: []string{"key1", "key2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseKeys(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseKeys(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i, key := range result {
				if key != tt.expected[i] {
					t.Errorf("parseKeys(%q)[%d] = %q, want %q", tt.input, i, key, tt.expected[i])
				}
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save and restore env
	original := os.Getenv(EnvAPIKeys)
	defer os.Setenv(EnvAPIKeys, original)

	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "no env var",
			envValue: "",
			expected: nil,
		},
		{
			name:     "single key",
			envValue: "test-key",
			expected: []string{"test-key"},
		},
		{
			name:     "multiple keys",
			envValue: "key1,key2",
			expected: []string{"key1", "key2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(EnvAPIKeys, tt.envValue)
			result := loadFromEnv()
			if len(result) != len(tt.expected) {
				t.Errorf("loadFromEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadAPIKeys_EnvPriority(t *testing.T) {
	// Reset cache
	ResetCache()

	// Save and restore
	original := os.Getenv(EnvAPIKeys)
	defer os.Setenv(EnvAPIKeys, original)
	defer ResetCache()

	// Initialize config with keys
	config.Reset()
	_ = config.Init("")

	// Set env var
	os.Setenv(EnvAPIKeys, "env-key")

	keys := LoadAPIKeys()
	if len(keys) != 1 || keys[0] != "env-key" {
		t.Errorf("LoadAPIKeys() = %v, want [env-key]", keys)
	}
}

func TestHasAPIKeys(t *testing.T) {
	// Reset cache
	ResetCache()

	// Save and restore
	original := os.Getenv(EnvAPIKeys)
	defer os.Setenv(EnvAPIKeys, original)
	defer ResetCache()

	// Test with no keys
	os.Setenv(EnvAPIKeys, "")
	config.Reset()
	_ = config.Init("")

	if HasAPIKeys() {
		t.Error("HasAPIKeys() = true, want false when no keys configured")
	}

	// Reset and test with keys
	ResetCache()
	os.Setenv(EnvAPIKeys, "test-key")

	if !HasAPIKeys() {
		t.Error("HasAPIKeys() = false, want true when keys are configured")
	}
}

func TestResetCache(t *testing.T) {
	// Reset to clean state
	ResetCache()

	// Save and restore
	original := os.Getenv(EnvAPIKeys)
	defer os.Setenv(EnvAPIKeys, original)
	defer ResetCache()

	// Set initial key
	os.Setenv(EnvAPIKeys, "key1")
	keys1 := LoadAPIKeys()

	// Change env
	os.Setenv(EnvAPIKeys, "key2")

	// Should still return cached value
	keys2 := LoadAPIKeys()
	if len(keys2) != len(keys1) || (len(keys1) > 0 && keys1[0] != keys2[0]) {
		t.Error("LoadAPIKeys() should return cached value")
	}

	// Reset cache
	ResetCache()

	// Should now return new value
	keys3 := LoadAPIKeys()
	if len(keys3) != 1 || keys3[0] != "key2" {
		t.Errorf("LoadAPIKeys() after ResetCache() = %v, want [key2]", keys3)
	}
}

func TestIsValidGopassPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "simple path", path: "project/api-keys", want: true},
		{name: "dots hyphens underscores", path: "my.project/api_keys-prod", want: true},
		{name: "uppercase", path: "TeamA/Service/API_KEYS", want: true},
		{name: "empty", path: "", want: false},
		{name: "contains space", path: "project/api keys", want: false},
		{name: "contains semicolon", path: "project;rm", want: false},
		{name: "contains dollar", path: "project/$HOME", want: false},
		{name: "contains newline", path: "project/\nkeys", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidGopassPath(tt.path)
			if got != tt.want {
				t.Errorf("isValidGopassPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetGopassPath_Priority(t *testing.T) {
	t.Cleanup(config.Reset)
	config.Reset()
	if err := config.Init(""); err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	// Config fallback
	config.Get().Server.APIKeysGopassPath = "config/path"
	t.Setenv(EnvAPIKeysGopassPath, "")
	if got := getGopassPath(); got != "config/path" {
		t.Errorf("getGopassPath() = %q, want %q", got, "config/path")
	}

	// Env takes precedence
	t.Setenv(EnvAPIKeysGopassPath, "env/path")
	if got := getGopassPath(); got != "env/path" {
		t.Errorf("getGopassPath() with env = %q, want %q", got, "env/path")
	}

	// Empty when both unset
	t.Setenv(EnvAPIKeysGopassPath, "")
	config.Get().Server.APIKeysGopassPath = ""
	if got := getGopassPath(); got != "" {
		t.Errorf("getGopassPath() = %q, want empty", got)
	}
}

func TestLoadFromGopass_EarlyReturnPaths(t *testing.T) {
	t.Cleanup(config.Reset)
	config.Reset()
	if err := config.Init(""); err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	t.Run("not configured", func(t *testing.T) {
		t.Setenv(EnvAPIKeysGopassPath, "")
		if got := loadFromGopass(); got != nil {
			t.Errorf("loadFromGopass() = %v, want nil", got)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		t.Setenv(EnvAPIKeysGopassPath, "invalid path")
		if got := loadFromGopass(); got != nil {
			t.Errorf("loadFromGopass() = %v, want nil for invalid path", got)
		}
	})

	t.Run("gopass missing", func(t *testing.T) {
		t.Setenv(EnvAPIKeysGopassPath, "valid/path")
		t.Setenv("PATH", "")
		if got := loadFromGopass(); got != nil {
			t.Errorf("loadFromGopass() = %v, want nil when gopass not available", got)
		}
	})
}

func TestSetReloadInterval_CacheExpiry(t *testing.T) {
	originalInterval := reloadInterval
	t.Cleanup(func() {
		SetReloadInterval(originalInterval)
		ResetCache()
	})

	ResetCache()
	SetReloadInterval(50 * time.Millisecond)

	t.Setenv(EnvAPIKeys, "key1")
	assertStringSliceEqual(t, LoadAPIKeys(), []string{"key1"})

	t.Setenv(EnvAPIKeys, "key2")
	// Before expiry, should still use cached key1.
	assertStringSliceEqual(t, LoadAPIKeys(), []string{"key1"})

	time.Sleep(80 * time.Millisecond)
	assertStringSliceEqual(t, LoadAPIKeys(), []string{"key2"})
}

func TestSetReloadInterval_ZeroDisablesAutoReload(t *testing.T) {
	originalInterval := reloadInterval
	t.Cleanup(func() {
		SetReloadInterval(originalInterval)
		ResetCache()
	})

	ResetCache()
	SetReloadInterval(0)

	t.Setenv(EnvAPIKeys, "key1")
	assertStringSliceEqual(t, LoadAPIKeys(), []string{"key1"})

	t.Setenv(EnvAPIKeys, "key2")
	// Reload disabled, cached key1 should be returned.
	assertStringSliceEqual(t, LoadAPIKeys(), []string{"key1"})
}

func TestForceReload(t *testing.T) {
	originalInterval := reloadInterval
	t.Cleanup(func() {
		SetReloadInterval(originalInterval)
		ResetCache()
	})

	ResetCache()
	SetReloadInterval(0)

	t.Setenv(EnvAPIKeys, "old-key")
	assertStringSliceEqual(t, LoadAPIKeys(), []string{"old-key"})

	t.Setenv(EnvAPIKeys, "new-key")
	assertStringSliceEqual(t, ForceReload(), []string{"new-key"})
	assertStringSliceEqual(t, LoadAPIKeys(), []string{"new-key"})
}

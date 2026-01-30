package auth

import (
	"os"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
)

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

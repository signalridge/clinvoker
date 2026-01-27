// Package config provides configuration management using viper.
package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	DefaultBackend string                   `mapstructure:"default_backend"`
	UnifiedFlags   UnifiedFlagsConfig       `mapstructure:"unified_flags"`
	Backends       map[string]BackendConfig `mapstructure:"backends"`
	Session        SessionConfig            `mapstructure:"session"`
	Output         OutputConfig             `mapstructure:"output"`
	Parallel       ParallelConfig           `mapstructure:"parallel"`
	Server         ServerConfig             `mapstructure:"server"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	// Host is the address to bind to (e.g., "127.0.0.1", "0.0.0.0").
	Host string `mapstructure:"host"`

	// Port is the port to listen on.
	Port int `mapstructure:"port"`
}

// UnifiedFlagsConfig contains unified flag settings that apply across backends.
type UnifiedFlagsConfig struct {
	// ApprovalMode controls how backends ask for approval (default, auto, none, always).
	ApprovalMode string `mapstructure:"approval_mode"`

	// SandboxMode controls file/network access (default, read-only, workspace, full).
	SandboxMode string `mapstructure:"sandbox_mode"`

	// OutputFormat controls output format (default, text, json, stream-json).
	OutputFormat string `mapstructure:"output_format"`

	// Verbose enables verbose output.
	Verbose bool `mapstructure:"verbose"`

	// DryRun simulates execution without changes.
	DryRun bool `mapstructure:"dry_run"`

	// MaxTurns limits the number of agentic turns.
	MaxTurns int `mapstructure:"max_turns"`

	// MaxTokens limits response tokens.
	MaxTokens int `mapstructure:"max_tokens"`
}

// BackendConfig contains backend-specific configuration.
type BackendConfig struct {
	// Model specifies the default model for this backend.
	Model string `mapstructure:"model"`

	// AllowedTools specifies allowed tools (backend-specific format).
	AllowedTools string `mapstructure:"allowed_tools"`

	// ApprovalMode overrides the unified approval mode for this backend.
	ApprovalMode string `mapstructure:"approval_mode"`

	// SandboxMode overrides the unified sandbox mode for this backend.
	SandboxMode string `mapstructure:"sandbox_mode"`

	// ExtraFlags contains additional backend-specific flags.
	ExtraFlags []string `mapstructure:"extra_flags"`

	// Enabled indicates if this backend is enabled.
	Enabled *bool `mapstructure:"enabled"`

	// SystemPrompt provides a default system prompt for this backend.
	SystemPrompt string `mapstructure:"system_prompt"`
}

// SessionConfig contains session management configuration.
type SessionConfig struct {
	// AutoResume enables automatic session resumption.
	AutoResume bool `mapstructure:"auto_resume"`

	// RetentionDays specifies how long to keep sessions.
	RetentionDays int `mapstructure:"retention_days"`

	// DefaultTags are tags automatically added to new sessions.
	DefaultTags []string `mapstructure:"default_tags"`

	// StoreTokenUsage enables token usage tracking.
	StoreTokenUsage bool `mapstructure:"store_token_usage"`
}

// OutputConfig contains output settings.
type OutputConfig struct {
	// Format is the default output format.
	Format string `mapstructure:"format"`

	// ShowTokens shows token usage after each response.
	ShowTokens bool `mapstructure:"show_tokens"`

	// ShowTiming shows execution timing.
	ShowTiming bool `mapstructure:"show_timing"`

	// Color enables colored output.
	Color bool `mapstructure:"color"`
}

// ParallelConfig contains parallel execution settings.
type ParallelConfig struct {
	// MaxWorkers is the maximum number of parallel workers.
	MaxWorkers int `mapstructure:"max_workers"`

	// FailFast stops all tasks on first failure.
	FailFast bool `mapstructure:"fail_fast"`

	// AggregateOutput combines output from all tasks.
	AggregateOutput bool `mapstructure:"aggregate_output"`
}

// IsBackendEnabled checks if a backend is enabled (defaults to true).
func (c *BackendConfig) IsBackendEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

var (
	cfg  *Config
	once sync.Once
)

// Init initializes the configuration.
func Init(cfgFile string) error {
	var initErr error

	once.Do(func() {
		cfg = &Config{
			DefaultBackend: "claude",
			UnifiedFlags: UnifiedFlagsConfig{
				ApprovalMode: "default",
				SandboxMode:  "default",
				OutputFormat: "default",
			},
			Backends: make(map[string]BackendConfig),
			Session: SessionConfig{
				AutoResume:      true,
				RetentionDays:   30,
				StoreTokenUsage: true,
			},
			Output: OutputConfig{
				Format:     "text",
				ShowTokens: false,
				ShowTiming: false,
				Color:      true,
			},
			Parallel: ParallelConfig{
				MaxWorkers:      3,
				FailFast:        false,
				AggregateOutput: true,
			},
			Server: ServerConfig{
				Host: "127.0.0.1",
				Port: 8080,
			},
		}

		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				initErr = err
				return
			}

			configDir := filepath.Join(home, ".clinvk")
			viper.AddConfigPath(configDir)
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
		}

		// Environment variables
		viper.SetEnvPrefix("CLINVK")
		viper.AutomaticEnv()

		// Bind environment variables
		viper.BindEnv("default_backend", "CLINVK_BACKEND")
		viper.BindEnv("backends.claude.model", "CLINVK_CLAUDE_MODEL")
		viper.BindEnv("backends.codex.model", "CLINVK_CODEX_MODEL")
		viper.BindEnv("backends.gemini.model", "CLINVK_GEMINI_MODEL")

		// Read config file (ignore if not found)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				initErr = err
				return
			}
		}

		if err := viper.Unmarshal(cfg); err != nil {
			initErr = err
			return
		}
	})

	return initErr
}

// Get returns the current configuration.
// This is safe to call from multiple goroutines concurrently.
func Get() *Config {
	// Always call Init to ensure initialization happens via sync.Once
	// This avoids race conditions from checking cfg == nil directly
	_ = Init("")
	return cfg
}

// ConfigDir returns the configuration directory path.
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".clinvk"
	}
	return filepath.Join(home, ".clinvk")
}

// SessionsDir returns the sessions directory path.
func SessionsDir() string {
	return filepath.Join(ConfigDir(), "sessions")
}

// EnsureConfigDir creates the configuration directory if it doesn't exist.
func EnsureConfigDir() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(SessionsDir(), 0755)
}

// Set sets a configuration value.
func Set(key string, value interface{}) error {
	viper.Set(key, value)
	return WriteConfig()
}

// WriteConfig writes the current configuration to disk.
func WriteConfig() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath := filepath.Join(ConfigDir(), "config.yaml")
	return viper.WriteConfigAs(configPath)
}

// GetBackendConfig returns the configuration for a specific backend.
// Returns a new empty BackendConfig if the backend is not configured.
func GetBackendConfig(backend string) BackendConfig {
	c := Get()
	if bc, ok := c.Backends[backend]; ok {
		return bc
	}
	return BackendConfig{}
}

// GetEffectiveApprovalMode returns the approval mode for a backend.
// Backend-specific setting takes precedence over unified setting.
func GetEffectiveApprovalMode(backend string) string {
	bc := GetBackendConfig(backend)
	if bc.ApprovalMode != "" {
		return bc.ApprovalMode
	}
	return Get().UnifiedFlags.ApprovalMode
}

// GetEffectiveSandboxMode returns the sandbox mode for a backend.
// Backend-specific setting takes precedence over unified setting.
func GetEffectiveSandboxMode(backend string) string {
	bc := GetBackendConfig(backend)
	if bc.SandboxMode != "" {
		return bc.SandboxMode
	}
	return Get().UnifiedFlags.SandboxMode
}

// GetEffectiveModel returns the model for a backend.
func GetEffectiveModel(backend string) string {
	bc := GetBackendConfig(backend)
	return bc.Model
}

// EnabledBackends returns a list of enabled backend names.
func EnabledBackends() []string {
	c := Get()
	// Default all backends enabled
	backends := []string{"claude", "codex", "gemini"}

	var enabled []string
	for _, name := range backends {
		if bc, ok := c.Backends[name]; ok {
			if !bc.IsBackendEnabled() {
				continue
			}
		}
		enabled = append(enabled, name)
	}
	return enabled
}

// Reset resets the configuration (mainly for testing).
func Reset() {
	once = sync.Once{}
	cfg = nil
}

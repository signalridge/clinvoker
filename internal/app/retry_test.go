package app

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
)

type retryTestBackend struct{}

func (b *retryTestBackend) Name() string      { return "retry-test" }
func (b *retryTestBackend) IsAvailable() bool { return true }
func (b *retryTestBackend) BuildCommand(prompt string, _ *backend.Options) *exec.Cmd {
	return exec.Command("echo", prompt)
}
func (b *retryTestBackend) ResumeCommand(sessionID, prompt string, _ *backend.Options) *exec.Cmd {
	return exec.Command("echo", sessionID, prompt)
}
func (b *retryTestBackend) BuildCommandUnified(prompt string, _ *backend.UnifiedOptions) *exec.Cmd {
	return exec.Command("sh", "-c", fmt.Sprintf("echo %q", prompt))
}
func (b *retryTestBackend) ResumeCommandUnified(sessionID, prompt string, _ *backend.UnifiedOptions) *exec.Cmd {
	return exec.Command("echo", sessionID, prompt)
}
func (b *retryTestBackend) ParseOutput(rawOutput string) string { return rawOutput }
func (b *retryTestBackend) ParseJSONResponse(rawOutput string) (*backend.UnifiedResponse, error) {
	if strings.Contains(rawOutput, "rate-limit") {
		return &backend.UnifiedResponse{Error: "rate limit exceeded"}, nil
	}
	return &backend.UnifiedResponse{Content: strings.TrimSpace(rawOutput)}, nil
}
func (b *retryTestBackend) SeparateStderr() bool { return false }

func TestResolveRetryPolicyPrecedence(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	cfg := config.Get()
	cfg.Retry.Global = config.RetryPolicyConfig{Enabled: true, MaxAttempts: 2}
	cfg.Retry.ByBackend = map[string]config.RetryPolicyConfig{"claude": {Enabled: true, MaxAttempts: 3}}
	cfg.Retry.ByCommand = map[string]config.RetryPolicyConfig{"compare": {Enabled: true, MaxAttempts: 4}}

	p := resolveRetryPolicy(cfg, "claude", "compare")
	if p.MaxAttempts != 4 {
		t.Fatalf("MaxAttempts = %d, want 4", p.MaxAttempts)
	}

	p = resolveRetryPolicy(cfg, "claude", "parallel")
	if p.MaxAttempts != 3 {
		t.Fatalf("backend precedence MaxAttempts = %d, want 3", p.MaxAttempts)
	}
}

func TestExecuteWithRetryJSON_RateLimitThenSuccess(t *testing.T) {
	config.Reset()
	if err := config.Init(""); err != nil {
		t.Fatalf("config.Init failed: %v", err)
	}
	t.Cleanup(config.Reset)

	cfg := config.Get()
	cfg.Retry.Global = config.RetryPolicyConfig{Enabled: true, MaxAttempts: 3, BackoffInitialMS: 1, BackoffMaxMS: 1, BackoffMultiplier: 1.0}
	cfg.Retry.ByCommand = map[string]config.RetryPolicyConfig{
		"compare": {
			Enabled:           true,
			MaxAttempts:       3,
			BackoffInitialMS:  1,
			BackoffMaxMS:      1,
			BackoffMultiplier: 1.0,
			JitterRatio:       0,
			RetryableErrors:   []string{"rate limit"},
		},
	}

	b := &retryTestBackend{}
	attempt := 0
	buildCmd := func() *exec.Cmd {
		attempt++
		if attempt == 1 {
			return b.BuildCommandUnified("rate-limit", nil)
		}
		return b.BuildCommandUnified("success", nil)
	}

	result, retry, err := executeWithRetryJSON(b, buildCmd, cfg, "claude", "compare", 2*time.Second, true)
	if err != nil {
		t.Fatalf("executeWithRetryJSON returned error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("result.ExitCode = %d, want 0", result.ExitCode)
	}
	if retry.AttemptsUsed != 2 {
		t.Fatalf("retry.AttemptsUsed = %d, want 2", retry.AttemptsUsed)
	}
	if retry.TerminationReason != retryReasonSuccess {
		t.Fatalf("retry.TerminationReason = %q, want %q", retry.TerminationReason, retryReasonSuccess)
	}
}

func TestExecuteWithRetryJSON_NonIdempotentGuard(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	cfg := config.Get()
	cfg.Retry.ByCommand = map[string]config.RetryPolicyConfig{
		"parallel": {
			Enabled:            true,
			MaxAttempts:        3,
			BackoffInitialMS:   1,
			BackoffMaxMS:       1,
			BackoffMultiplier:  1,
			AllowNonIdempotent: false,
		},
	}

	b := &retryTestBackend{}
	attempt := 0
	buildCmd := func() *exec.Cmd {
		attempt++
		return b.BuildCommandUnified("rate-limit", nil)
	}

	_, retry, _ := executeWithRetryJSON(b, buildCmd, cfg, "claude", "parallel", 2*time.Second, false)
	if retry.AttemptsUsed != 1 {
		t.Fatalf("retry.AttemptsUsed = %d, want 1 when non-idempotent retries are disallowed", retry.AttemptsUsed)
	}
	if retry.TerminationReason != retryReasonNonRetryableError {
		t.Fatalf("retry.TerminationReason = %q, want %q", retry.TerminationReason, retryReasonNonRetryableError)
	}
}

func TestExecuteWithRetryJSON_RetryBudgetExhausted(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	cfg := config.Get()
	cfg.Retry.ByCommand = map[string]config.RetryPolicyConfig{
		"compare": {
			Enabled:           true,
			MaxAttempts:       2,
			BackoffInitialMS:  1,
			BackoffMaxMS:      1,
			BackoffMultiplier: 1.0,
			JitterRatio:       0,
			RetryableErrors:   []string{"rate limit"},
		},
	}

	b := &retryTestBackend{}
	buildCmd := func() *exec.Cmd {
		return b.BuildCommandUnified("rate-limit", nil)
	}

	result, retry, err := executeWithRetryJSON(b, buildCmd, cfg, "claude", "compare", 2*time.Second, true)
	if err != nil {
		t.Fatalf("executeWithRetryJSON returned error: %v", err)
	}
	if result.ExitCode == 0 {
		t.Fatalf("result.ExitCode = %d, want non-zero", result.ExitCode)
	}
	if retry.AttemptsUsed != 2 {
		t.Fatalf("retry.AttemptsUsed = %d, want 2", retry.AttemptsUsed)
	}
	if retry.TerminationReason != retryReasonRetryBudgetExhausted {
		t.Fatalf("retry.TerminationReason = %q, want %q", retry.TerminationReason, retryReasonRetryBudgetExhausted)
	}
	if retry.LastErrorCategory != "rate_limit" {
		t.Fatalf("retry.LastErrorCategory = %q, want %q", retry.LastErrorCategory, "rate_limit")
	}
}

func TestExecuteWithRetryJSON_TimeoutBudgetExhausted(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	cfg := config.Get()
	cfg.Retry.ByCommand = map[string]config.RetryPolicyConfig{
		"compare": {
			Enabled:           true,
			MaxAttempts:       3,
			BackoffInitialMS:  50,
			BackoffMaxMS:      50,
			BackoffMultiplier: 1.0,
			JitterRatio:       0,
			RetryableErrors:   []string{"rate limit"},
		},
	}

	b := &retryTestBackend{}
	buildCmd := func() *exec.Cmd {
		return b.BuildCommandUnified("rate-limit", nil)
	}

	_, retry, err := executeWithRetryJSON(b, buildCmd, cfg, "claude", "compare", 20*time.Millisecond, true)
	if !errors.Is(err, ErrCommandTimeout) {
		t.Fatalf("error = %v, want %v", err, ErrCommandTimeout)
	}
	if retry.AttemptsUsed != 1 {
		t.Fatalf("retry.AttemptsUsed = %d, want 1", retry.AttemptsUsed)
	}
	if retry.TerminationReason != retryReasonTimeoutExhausted {
		t.Fatalf("retry.TerminationReason = %q, want %q", retry.TerminationReason, retryReasonTimeoutExhausted)
	}
	if retry.LastErrorCategory != retryCategoryRateLimit && retry.LastErrorCategory != retryCategoryTimeout {
		t.Fatalf(
			"retry.LastErrorCategory = %q, want one of [%q, %q]",
			retry.LastErrorCategory,
			retryCategoryRateLimit,
			retryCategoryTimeout,
		)
	}
}

func TestExecuteWithRetryJSON_NonRetryableStopsImmediately(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	cfg := config.Get()
	cfg.Retry.ByCommand = map[string]config.RetryPolicyConfig{
		"compare": {
			Enabled:           true,
			MaxAttempts:       3,
			BackoffInitialMS:  1,
			BackoffMaxMS:      1,
			BackoffMultiplier: 1.0,
			JitterRatio:       0,
			RetryableErrors:   []string{"rate limit"},
		},
	}

	b := &retryTestBackend{}
	buildCmd := func() *exec.Cmd {
		// Non-zero process exit without retryable error markers.
		return exec.Command("sh", "-c", "echo fatal validation failed >&2; exit 2")
	}

	_, retry, _ := executeWithRetryJSON(b, buildCmd, cfg, "claude", "compare", 2*time.Second, true)
	if retry.AttemptsUsed != 1 {
		t.Fatalf("retry.AttemptsUsed = %d, want 1", retry.AttemptsUsed)
	}
	if retry.TerminationReason != retryReasonNonRetryableError {
		t.Fatalf("retry.TerminationReason = %q, want %q", retry.TerminationReason, retryReasonNonRetryableError)
	}
	if retry.LastErrorCategory != retryCategoryUnknown {
		t.Fatalf("retry.LastErrorCategory = %q, want %q", retry.LastErrorCategory, retryCategoryUnknown)
	}
}

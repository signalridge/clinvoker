package app

import (
	cryptorand "crypto/rand"
	"errors"
	"math"
	"math/big"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/signalridge/clinvoker/internal/backend"
	"github.com/signalridge/clinvoker/internal/config"
)

const (
	retryReasonSuccess              = "success"
	retryReasonNonRetryableError    = "non_retryable_error"
	retryReasonRetryBudgetExhausted = "retry_budget_exhausted"
	retryReasonTimeoutExhausted     = "timeout_exhausted"

	retryCategoryTimeout   = "timeout"
	retryCategoryRateLimit = "rate_limit"
	retryCategoryNone      = "none"
	retryCategoryUnknown   = "unknown"
)

var defaultRetryableErrorSubstrings = []string{
	"timeout",
	"timed out",
	"rate limit",
	"temporarily unavailable",
	"connection reset",
	"connection refused",
	"connection aborted",
	"eof",
	"502",
	"503",
	"504",
}

var jitterFallbackCounter uint64

// RetryTelemetry captures retry execution details for observability output.
type RetryTelemetry struct {
	Enabled           bool   `json:"enabled"`
	AttemptsUsed      int    `json:"attempts_used,omitempty"`
	AttemptsTotal     int    `json:"attempts_total,omitempty"`
	TerminationReason string `json:"termination_reason,omitempty"`
	LastErrorCategory string `json:"last_error_category,omitempty"`
}

func resolveRetryPolicy(cfg *config.Config, backendName, commandName string) config.RetryPolicyConfig {
	if cfg == nil {
		cfg = config.Get()
	}

	policy := cfg.Retry.Global
	if p, ok := cfg.Retry.ByBackend[backendName]; ok {
		policy = p
	}
	if p, ok := cfg.Retry.ByCommand[commandName]; ok {
		policy = p
	}

	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
	}
	if policy.BackoffInitialMS <= 0 {
		policy.BackoffInitialMS = 250
	}
	if policy.BackoffMaxMS <= 0 {
		policy.BackoffMaxMS = 5000
	}
	if policy.BackoffInitialMS > policy.BackoffMaxMS {
		policy.BackoffInitialMS = policy.BackoffMaxMS
	}
	if policy.BackoffMultiplier < 1.0 {
		policy.BackoffMultiplier = 2.0
	}
	if policy.JitterRatio < 0 || policy.JitterRatio > 1 {
		policy.JitterRatio = 0.2
	}
	if len(policy.RetryableErrors) == 0 {
		policy.RetryableErrors = append([]string{}, defaultRetryableErrorSubstrings...)
	}

	return policy
}

func classifyRetryable(capture *CaptureResult, execErr error, policy *config.RetryPolicyConfig) (string, bool) {
	if errors.Is(execErr, ErrCommandTimeout) {
		return retryCategoryTimeout, true
	}

	msg := ""
	if capture != nil && capture.Error != "" {
		msg = capture.Error
	}
	if msg == "" && execErr != nil {
		msg = execErr.Error()
	}
	msg = strings.ToLower(msg)

	if strings.Contains(msg, "context canceled") || strings.Contains(msg, "killed") {
		return "canceled", false
	}

	if strings.Contains(msg, "timeout") || strings.Contains(msg, "timed out") {
		return retryCategoryTimeout, true
	}

	if strings.Contains(msg, "rate limit") || strings.Contains(msg, "429") {
		return retryCategoryRateLimit, true
	}

	for _, needle := range policy.RetryableErrors {
		if needle == "" {
			continue
		}
		if strings.Contains(msg, strings.ToLower(needle)) {
			return "transient", true
		}
	}

	if capture != nil && capture.ExitCode == 0 && capture.Error == "" {
		return retryCategoryNone, false
	}

	return retryCategoryUnknown, false
}

func secureSignedUnitFloat64() float64 {
	const scale = int64(1_000_000)
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(scale*2+1))
	if err != nil {
		// Fallback to a deterministic-but-varying value to avoid synchronized retries.
		seed := atomic.AddUint64(&jitterFallbackCounter, 0x9e3779b97f4a7c15)
		seed ^= seed >> 12
		seed ^= seed << 25
		seed ^= seed >> 27
		return float64(seed%2_000_001)/1_000_000 - 1.0
	}
	return float64(n.Int64()-scale) / float64(scale)
}

func jitteredBackoff(base time.Duration, ratio float64) time.Duration {
	if ratio <= 0 {
		return base
	}

	span := float64(base) * ratio
	if span <= 0 {
		return base
	}

	jitter := secureSignedUnitFloat64() * span
	d := float64(base) + jitter
	if d < 0 {
		d = 0
	}
	return time.Duration(d)
}

func nextBackoff(current time.Duration, multiplier float64, maxDelay time.Duration) time.Duration {
	next := time.Duration(math.Round(float64(current) * multiplier))
	if next <= 0 {
		next = current
	}
	if next > maxDelay {
		return maxDelay
	}
	return next
}

func timeoutExhaustedResult() *CaptureResult {
	return &CaptureResult{
		ExitCode: 124,
		Error:    ErrCommandTimeout.Error(),
	}
}

func selectExecutionError(execErr error, captureErr string) string {
	captureMsg := strings.TrimSpace(captureErr)
	if execErr == nil {
		return captureMsg
	}

	execMsg := strings.TrimSpace(execErr.Error())
	if captureMsg == "" {
		return execMsg
	}
	if execMsg == "" || captureMsg == execMsg || strings.Contains(captureMsg, execMsg) {
		return captureMsg
	}
	if strings.Contains(execMsg, captureMsg) {
		return execMsg
	}
	return captureMsg + "; cause: " + execMsg
}

func executeWithRetryJSON(
	b backend.Backend,
	buildCmd func() *exec.Cmd,
	cfg *config.Config,
	backendName string,
	commandName string,
	overallTimeout time.Duration,
	isIdempotent bool,
) (*CaptureResult, RetryTelemetry, error) {
	policy := resolveRetryPolicy(cfg, backendName, commandName)
	telemetry := RetryTelemetry{
		Enabled:       policy.Enabled,
		AttemptsTotal: policy.MaxAttempts,
	}

	// Retry is disabled either by policy or by non-idempotent safety guard.
	retryActive := policy.Enabled && policy.MaxAttempts > 1
	if !isIdempotent && !policy.AllowNonIdempotent {
		retryActive = false
	}

	start := time.Now()
	backoff := time.Duration(policy.BackoffInitialMS) * time.Millisecond
	maxBackoff := time.Duration(policy.BackoffMaxMS) * time.Millisecond

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		telemetry.AttemptsUsed = attempt

		attemptTimeout := overallTimeout
		if overallTimeout > 0 {
			remaining := overallTimeout - time.Since(start)
			if remaining <= 0 {
				telemetry.TerminationReason = retryReasonTimeoutExhausted
				return timeoutExhaustedResult(), telemetry, ErrCommandTimeout
			}
			attemptTimeout = remaining
		}

		capture, execErr := ExecuteAndCaptureWithJSONTimeout(b, buildCmd(), attemptTimeout)
		if capture == nil {
			capture = &CaptureResult{ExitCode: 1}
		}

		if execErr == nil && capture.ExitCode == 0 && capture.Error == "" {
			telemetry.TerminationReason = retryReasonSuccess
			telemetry.LastErrorCategory = retryCategoryNone
			return capture, telemetry, nil
		}

		category, retryable := classifyRetryable(capture, execErr, &policy)
		telemetry.LastErrorCategory = category

		if !retryActive {
			telemetry.TerminationReason = retryReasonNonRetryableError
			return capture, telemetry, execErr
		}

		if !retryable {
			telemetry.TerminationReason = retryReasonNonRetryableError
			return capture, telemetry, execErr
		}

		if attempt >= policy.MaxAttempts {
			telemetry.TerminationReason = retryReasonRetryBudgetExhausted
			return capture, telemetry, execErr
		}

		sleepDur := jitteredBackoff(backoff, policy.JitterRatio)
		if overallTimeout > 0 {
			remaining := overallTimeout - time.Since(start)
			if remaining <= sleepDur {
				telemetry.TerminationReason = retryReasonTimeoutExhausted
				return capture, telemetry, ErrCommandTimeout
			}
		}
		time.Sleep(sleepDur)
		backoff = nextBackoff(backoff, policy.BackoffMultiplier, maxBackoff)
	}

	telemetry.TerminationReason = retryReasonRetryBudgetExhausted
	return &CaptureResult{ExitCode: 1}, telemetry, nil
}

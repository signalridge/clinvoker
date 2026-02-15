package app

import (
	"strings"
	"testing"

	"github.com/signalridge/clinvoker/internal/auth"
	"github.com/signalridge/clinvoker/internal/config"
)

func TestBuildServeSecurityState_Warnings(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	t.Setenv(auth.EnvAPIKeys, "")
	auth.ResetCache()

	cfg := config.Get()
	cfg.Server.RateLimitEnabled = false
	cfg.Server.CORSAllowedOrigins = []string{"*"}
	cfg.Server.CORSAllowCredentials = true

	state := buildServeSecurityState(cfg, "0.0.0.0")
	if !state.BindPublic {
		t.Fatal("expected bind public to be true")
	}
	if len(state.HighRiskWarnings) < 2 {
		t.Fatalf("expected multiple high risk warnings, got %v", state.HighRiskWarnings)
	}
}

func TestBuildServeSecurityState_DefaultLocalhostOriginsAreNotWildcard(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	t.Setenv(auth.EnvAPIKeys, "")
	auth.ResetCache()

	cfg := config.Get()
	cfg.Server.CORSAllowedOrigins = nil // use default localhost origins
	cfg.Server.CORSAllowCredentials = false

	state := buildServeSecurityState(cfg, "127.0.0.1")
	if state.CORSWildcard {
		t.Fatalf("expected localhost defaults to be treated as non-wildcard, origins=%v", state.CORSOrigins)
	}
	for _, warning := range state.HighRiskWarnings {
		if warning == "auth disabled with wildcard CORS" {
			t.Fatalf("unexpected wildcard warning for localhost default origins: %v", state.HighRiskWarnings)
		}
	}
}

func TestPrintServeSecuritySummary_IncludesWarningRemediation(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	t.Setenv(auth.EnvAPIKeys, "")
	auth.ResetCache()

	cfg := config.Get()
	cfg.Server.CORSAllowedOrigins = []string{"*"}
	cfg.Server.CORSAllowCredentials = true

	output := captureStdout(t, func() {
		printServeSecuritySummary(cfg, "0.0.0.0")
	})

	if !strings.Contains(output, "Security warnings:") {
		t.Fatalf("expected security warnings section, got output: %s", output)
	}
	if !strings.Contains(output, "suggestion:") {
		t.Fatalf("expected remediation suggestions in warnings, got output: %s", output)
	}
}

func TestBuildServeSecurityState_WithAuthAndPrivateBind(t *testing.T) {
	setupTestConfig(t)
	defer config.Reset()
	resetAppGlobals()
	t.Cleanup(resetAppGlobals)

	t.Setenv(auth.EnvAPIKeys, "k1")
	auth.ResetCache()

	cfg := config.Get()
	cfg.Server.RateLimitEnabled = true
	cfg.Server.RateLimitRPS = 5
	cfg.Server.CORSAllowedOrigins = []string{"http://localhost:3000"}
	cfg.Server.CORSAllowCredentials = false
	cfg.Server.TrustedProxies = []string{"127.0.0.1"}

	state := buildServeSecurityState(cfg, "127.0.0.1")
	if !state.AuthEnabled {
		t.Fatal("auth should be enabled when API keys exist")
	}
	if !state.RateLimitEnabled {
		t.Fatal("rate limit should be enabled")
	}
	if !state.TrustedProxies {
		t.Fatal("trusted proxies should be enabled")
	}
	if len(state.HighRiskWarnings) != 0 {
		t.Fatalf("expected no high risk warnings, got %v", state.HighRiskWarnings)
	}
}

func TestEnabledText(t *testing.T) {
	if got := enabledText(true); got != "enabled" {
		t.Fatalf("enabledText(true) = %q, want %q", got, "enabled")
	}
	if got := enabledText(false); got != "disabled" {
		t.Fatalf("enabledText(false) = %q, want %q", got, "disabled")
	}
}

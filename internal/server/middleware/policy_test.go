package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"log/slog"

	"github.com/signalridge/clinvoker/internal/auth"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/policy"
	"github.com/signalridge/clinvoker/internal/requestctx"
)

func TestPolicyMiddleware_RedactsRawKeyFromContextLogsAndExplain(t *testing.T) {
	t.Setenv(auth.EnvAPIKeys, "raw-secret-key-123")
	auth.ResetCache()
	t.Cleanup(auth.ResetCache)

	policy.ResetForTest()
	t.Cleanup(policy.ResetForTest)

	rulesPath := filepath.Join(t.TempDir(), "policy.yaml")
	rulesYAML := `version: v1
rules:
  - id: deny-sensitive
    enabled: true
    priority: 10
    selectors:
      path_prefix: /blocked
      methods: [GET]
    action:
      type: deny
`
	if err := os.WriteFile(rulesPath, []byte(rulesYAML), 0o600); err != nil {
		t.Fatalf("write rules file error = %v", err)
	}

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(io.Discard, &logBuf), nil))
	cfg := &config.Config{
		Server: config.ServerConfig{
			Policy: config.PolicyConfig{
				Enabled:         true,
				Mode:            "shadow",
				FailureMode:     "fail-open",
				ExplainEnabled:  true,
				DefaultDecision: "allow",
				QuotaStore:      "memory",
				RulesFile:       rulesPath,
			},
		},
	}
	if err := policy.ConfigureFromConfig(cfg, logger); err != nil {
		t.Fatalf("ConfigureFromConfig() error = %v", err)
	}

	var gotIdentity requestctx.Identity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIdentity = requestctx.IdentityFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	handler := APIKeyAuth()(Policy(logger)(next))
	req := httptest.NewRequest(http.MethodGet, "/blocked", http.NoBody)
	req.Header.Set("X-Api-Key", "raw-secret-key-123")
	req.Header.Set("X-Policy-Explain", "true")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if gotIdentity.SubjectKeyID == "" {
		t.Fatal("expected sanitized SubjectKeyID in request context")
	}
	if gotIdentity.SubjectKeyID == "raw-secret-key-123" {
		t.Fatal("raw API key must not appear in context identity")
	}

	for key, values := range rec.Header() {
		for _, value := range values {
			if strings.Contains(value, "raw-secret-key-123") {
				t.Fatalf("raw API key leaked in header %q=%q", key, value)
			}
		}
	}

	if strings.Contains(logBuf.String(), "raw-secret-key-123") {
		t.Fatalf("raw API key leaked in policy logs: %s", logBuf.String())
	}
}

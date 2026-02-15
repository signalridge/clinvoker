package policy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShouldForceEvalError(t *testing.T) {
	t.Setenv(forceEvalErrorEnv, "false")
	req := httptest.NewRequest("GET", "/api/v1/prompt", http.NoBody)
	req.Header.Set(forceEvalErrorHeader, "true")
	if shouldForceEvalError(req) {
		t.Fatal("shouldForceEvalError should be false when env is disabled")
	}

	t.Setenv(forceEvalErrorEnv, "true")
	req = httptest.NewRequest("GET", "/api/v1/prompt", http.NoBody)
	req.Header.Set(forceEvalErrorHeader, "false")
	if shouldForceEvalError(req) {
		t.Fatal("shouldForceEvalError should be false when header is disabled")
	}

	req = httptest.NewRequest("GET", "/api/v1/prompt", http.NoBody)
	req.Header.Set(forceEvalErrorHeader, "true")
	if !shouldForceEvalError(req) {
		t.Fatal("shouldForceEvalError should be true when env and header are enabled")
	}
}

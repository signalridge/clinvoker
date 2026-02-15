package policy

import (
	"net/http"
	"os"
	"strings"
)

const (
	forceEvalErrorEnv    = "CLINVK_POLICY_TEST_FORCE_EVAL_ERROR"
	forceEvalErrorHeader = "X-Policy-Test-Force-Eval-Error"
)

func shouldForceEvalError(r *http.Request) bool {
	if r == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv(forceEvalErrorEnv)), "true") {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get(forceEvalErrorHeader)), "true")
}

package middleware

import (
	"log/slog"
	"net/http"

	apppolicy "github.com/signalridge/clinvoker/internal/policy"
)

// Policy enforces centralized request governance controls.
func Policy(logger *slog.Logger) func(http.Handler) http.Handler {
	return apppolicy.Middleware(logger)
}

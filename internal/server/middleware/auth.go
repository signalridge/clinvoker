// Package middleware provides HTTP middleware for the server.
package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/signalridge/clinvoker/internal/auth"
)

// APIKeyAuth returns a middleware that validates API keys.
// If no API keys are configured, the middleware passes through all requests (backward compatible).
//
// Supported authentication methods:
// - Header: X-Api-Key: <key>
// - Header: Authorization: Bearer <key>
func APIKeyAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			keys := auth.LoadAPIKeys()

			// No keys configured = auth disabled (backward compatible)
			if len(keys) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Extract key from request
			apiKey := extractAPIKey(r)
			if apiKey == "" {
				writeUnauthorized(w, "missing API key")
				return
			}

			// Validate key
			if !validateKey(apiKey, keys) {
				writeUnauthorized(w, "invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts the API key from the request.
// Supports X-Api-Key header and Authorization: Bearer header.
func extractAPIKey(r *http.Request) string {
	// Try X-Api-Key header first
	if key := r.Header.Get("X-Api-Key"); key != "" {
		return key
	}

	// Try Authorization: Bearer header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	return ""
}

// validateKey checks if the provided key matches any valid key.
// Uses constant-time comparison to prevent timing attacks.
func validateKey(key string, validKeys []string) bool {
	keyBytes := []byte(key)
	for _, validKey := range validKeys {
		if subtle.ConstantTimeCompare(keyBytes, []byte(validKey)) == 1 {
			return true
		}
	}
	return false
}

// errorResponse represents an error response structure.
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// writeUnauthorized writes a 401 Unauthorized response.
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.WriteHeader(http.StatusUnauthorized)

	resp := errorResponse{
		Error:   "unauthorized",
		Message: message,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// SkipAuthPaths returns a middleware that skips API key auth for specified paths.
// Useful for health checks and other public endpoints.
func SkipAuthPaths(skipPaths ...string) func(http.Handler) http.Handler {
	skipSet := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skipSet[path] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		authMiddleware := APIKeyAuth()
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for specified paths
			if _, skip := skipSet[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			// Apply auth middleware
			authMiddleware(next).ServeHTTP(w, r)
		})
	}
}

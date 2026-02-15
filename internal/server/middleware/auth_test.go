package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/signalridge/clinvoker/internal/auth"
	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/requestctx"
)

func setupTest(t *testing.T) func() {
	t.Helper()

	// Save original env
	originalEnv := os.Getenv(auth.EnvAPIKeys)

	// Clear env
	os.Setenv(auth.EnvAPIKeys, "")

	// Reset caches
	auth.ResetCache()
	config.Reset()
	_ = config.Init("")

	return func() {
		os.Setenv(auth.EnvAPIKeys, originalEnv)
		auth.ResetCache()
		config.Reset()
	}
}

func TestAPIKeyAuth_NoKeysConfigured(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Create test handler
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware
	middleware := APIKeyAuth()
	handler := middleware(next)

	// Test request without API key
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("Expected next handler to be called when no keys configured")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestAPIKeyAuth_ValidKey_XApiKeyHeader(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Set API key
	os.Setenv(auth.EnvAPIKeys, "test-key-123")
	auth.ResetCache()

	// Create test handler
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := APIKeyAuth()
	handler := middleware(next)

	// Test request with valid X-Api-Key
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Api-Key", "test-key-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("Expected next handler to be called with valid key")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestAPIKeyAuth_ValidKey_BearerToken(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Set API key
	os.Setenv(auth.EnvAPIKeys, "test-key-123")
	auth.ResetCache()

	// Create test handler
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := APIKeyAuth()
	handler := middleware(next)

	// Test request with valid Bearer token
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-key-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("Expected next handler to be called with valid Bearer token")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Set API key
	os.Setenv(auth.EnvAPIKeys, "test-key-123")
	auth.ResetCache()

	// Create test handler
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := APIKeyAuth()
	handler := middleware(next)

	// Test request with invalid key
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Api-Key", "wrong-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if nextCalled {
		t.Error("Expected next handler NOT to be called with invalid key")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse unauthorized body: %v", err)
	}
	if resp.Code != "unauthorized" {
		t.Fatalf("code = %q, want %q", resp.Code, "unauthorized")
	}
	if resp.Message != "invalid API key" {
		t.Fatalf("message = %q, want %q", resp.Message, "invalid API key")
	}
	if resp.RequestID == "" {
		t.Fatal("request_id should not be empty")
	}
	if rr.Header().Get("X-Request-ID") != resp.RequestID {
		t.Fatalf("X-Request-ID header should match body request_id")
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Set API key
	os.Setenv(auth.EnvAPIKeys, "test-key-123")
	auth.ResetCache()

	// Create test handler
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := APIKeyAuth()
	handler := middleware(next)

	// Test request without key
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if nextCalled {
		t.Error("Expected next handler NOT to be called without key")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse unauthorized body: %v", err)
	}
	if resp.Code != "unauthorized" {
		t.Fatalf("code = %q, want %q", resp.Code, "unauthorized")
	}
	if resp.Message != "missing API key" {
		t.Fatalf("message = %q, want %q", resp.Message, "missing API key")
	}
	if resp.RequestID == "" {
		t.Fatal("request_id should not be empty")
	}
	if rr.Header().Get("X-Request-ID") != resp.RequestID {
		t.Fatalf("X-Request-ID header should match body request_id")
	}
}

func TestAPIKeyAuth_MultipleValidKeys(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Set multiple API keys
	os.Setenv(auth.EnvAPIKeys, "key1,key2,key3")
	auth.ResetCache()

	// Create test handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := APIKeyAuth()
	handler := middleware(next)

	// Test each key
	for _, key := range []string{"key1", "key2", "key3"} {
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("X-Api-Key", key)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200 for key %q, got %d", key, rr.Code)
		}
	}
}

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		expectedKey string
	}{
		{
			name:        "no headers",
			headers:     nil,
			expectedKey: "",
		},
		{
			name:        "X-Api-Key header",
			headers:     map[string]string{"X-Api-Key": "my-key"},
			expectedKey: "my-key",
		},
		{
			name:        "Bearer token",
			headers:     map[string]string{"Authorization": "Bearer my-key"},
			expectedKey: "my-key",
		},
		{
			name:        "Bearer token lowercase",
			headers:     map[string]string{"Authorization": "bearer my-key"},
			expectedKey: "my-key",
		},
		{
			name:        "X-Api-Key takes precedence",
			headers:     map[string]string{"X-Api-Key": "key1", "Authorization": "Bearer key2"},
			expectedKey: "key1",
		},
		{
			name:        "non-Bearer auth",
			headers:     map[string]string{"Authorization": "Basic abc123"},
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			key := extractAPIKey(req)
			if key != tt.expectedKey {
				t.Errorf("extractAPIKey() = %q, want %q", key, tt.expectedKey)
			}
		})
	}
}

func TestSkipAuthPaths(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Set API key
	os.Setenv(auth.EnvAPIKeys, "test-key")
	auth.ResetCache()

	// Create test handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware that skips /health
	middleware := SkipAuthPaths("/health", "/ready")
	handler := middleware(next)

	// Test skipped path without key
	req := httptest.NewRequest("GET", "/health", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for skipped path, got %d", rr.Code)
	}

	// Test non-skipped path without key
	req = httptest.NewRequest("GET", "/api/test", http.NoBody)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for non-skipped path without key, got %d", rr.Code)
	}

	// Test non-skipped path with key
	req = httptest.NewRequest("GET", "/api/test", http.NoBody)
	req.Header.Set("X-Api-Key", "test-key")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for non-skipped path with key, got %d", rr.Code)
	}
}

func TestAPIKeyAuth_RequestIDPropagation(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	os.Setenv(auth.EnvAPIKeys, "test-key")
	auth.ResetCache()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := APIKeyAuth()(next)
	handler := chiMiddleware.RequestID(authHandler)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Api-Key", "wrong-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	var resp errorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resp.RequestID == "" {
		t.Fatal("response request_id should not be empty")
	}

	headerReqID := rr.Header().Get("X-Request-ID")
	if headerReqID == "" {
		t.Fatal("X-Request-ID header should not be empty")
	}
	if headerReqID != resp.RequestID {
		t.Fatalf("header request id %q != body request id %q", headerReqID, resp.RequestID)
	}
}

func TestAPIKeyAuth_InjectsSanitizedIdentityOnSuccess(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	rawKey := "sensitive-key-123"
	os.Setenv(auth.EnvAPIKeys, rawKey)
	auth.ResetCache()

	var gotIdentity requestctx.Identity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIdentity = requestctx.IdentityFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := APIKeyAuth()(next)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Api-Key", rawKey)
	req.Header.Set("X-Tenant-ID", "tenant-a")
	req.Header.Set("X-Subject-Tags", "internal,agent")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if gotIdentity.Authenticated != true {
		t.Fatal("Authenticated should be true for valid key")
	}
	if gotIdentity.SubjectKeyID == "" {
		t.Fatal("SubjectKeyID should not be empty on authenticated request")
	}
	if gotIdentity.SubjectKeyID != auth.SanitizeKeyID(rawKey) {
		t.Fatalf("SubjectKeyID = %q, want %q", gotIdentity.SubjectKeyID, auth.SanitizeKeyID(rawKey))
	}
	if gotIdentity.SubjectKeyID == rawKey {
		t.Fatal("SubjectKeyID must not contain raw key")
	}
	if gotIdentity.TenantID != "tenant-a" {
		t.Fatalf("TenantID = %q, want %q", gotIdentity.TenantID, "tenant-a")
	}
	if len(gotIdentity.SubjectTags) != 2 || gotIdentity.SubjectTags[0] != "internal" || gotIdentity.SubjectTags[1] != "agent" {
		t.Fatalf("SubjectTags = %v, want [internal agent]", gotIdentity.SubjectTags)
	}
}

func TestAPIKeyAuth_NoConfiguredKeys_DoesNotMarkAuthenticated(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	os.Setenv(auth.EnvAPIKeys, "")
	auth.ResetCache()

	var gotIdentity requestctx.Identity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIdentity = requestctx.IdentityFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := APIKeyAuth()(next)

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Api-Key", "caller-provided-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if gotIdentity.Authenticated {
		t.Fatal("Authenticated should be false when auth is disabled")
	}
	if gotIdentity.SubjectKeyID != "" {
		t.Fatalf("SubjectKeyID = %q, want empty when unauthenticated", gotIdentity.SubjectKeyID)
	}
}

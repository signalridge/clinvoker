package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(2, 2) // 2 RPS, burst of 2
	defer limiter.Stop()

	// First 2 requests should be allowed (burst)
	if !limiter.Allow("192.168.1.1") {
		t.Error("First request should be allowed")
	}
	if !limiter.Allow("192.168.1.1") {
		t.Error("Second request should be allowed (burst)")
	}

	// Third immediate request should be blocked
	if limiter.Allow("192.168.1.1") {
		t.Error("Third immediate request should be blocked")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(1, 1) // 1 RPS, burst of 1
	defer limiter.Stop()

	// First request from IP1 should be allowed
	if !limiter.Allow("192.168.1.1") {
		t.Error("First request from IP1 should be allowed")
	}

	// Second request from IP1 should be blocked
	if limiter.Allow("192.168.1.1") {
		t.Error("Second request from IP1 should be blocked")
	}

	// First request from IP2 should be allowed (separate limiter)
	if !limiter.Allow("192.168.1.2") {
		t.Error("First request from IP2 should be allowed")
	}
}

func TestRateLimit_Middleware(t *testing.T) {
	// Create test handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware with low limit for testing
	middleware, limiter := RateLimitWithLimiter(1, 1, 0)
	defer limiter.Stop()
	handler := middleware(next)

	// First request should succeed
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("First request: expected status 200, got %d", rr.Code)
	}

	// Second immediate request should be rate limited
	req = httptest.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.1:12345"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: expected status 429, got %d", rr.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	// NOTE: With trusted proxies feature, X-Forwarded-For and X-Real-IP
	// are only trusted when the request comes from a configured trusted proxy.
	// By default (no trusted proxies configured), direct RemoteAddr is always used.
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expectedIP string
	}{
		{
			name:       "remote addr only",
			remoteAddr: "192.168.1.1:12345",
			headers:    nil,
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For ignored without trusted proxy",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195"},
			expectedIP: "10.0.0.1", // Direct IP used (not trusted proxy)
		},
		{
			name:       "X-Forwarded-For multiple ignored without trusted proxy",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178"},
			expectedIP: "10.0.0.1", // Direct IP used (not trusted proxy)
		},
		{
			name:       "X-Real-IP ignored without trusted proxy",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "203.0.113.195"},
			expectedIP: "10.0.0.1", // Direct IP used (not trusted proxy)
		},
		{
			name:       "both headers ignored without trusted proxy",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
				"X-Real-IP":       "198.51.100.178",
			},
			expectedIP: "10.0.0.1", // Direct IP used (not trusted proxy)
		},
		{
			name:       "remote addr without port",
			remoteAddr: "192.168.1.1",
			headers:    nil,
			expectedIP: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getClientIP() = %q, want %q", ip, tt.expectedIP)
			}
		})
	}
}

func TestGetClientIP_WithTrustedProxy(t *testing.T) {
	// Configure trusted proxies for this test
	cfg := config.Get()
	originalProxies := cfg.Server.TrustedProxies
	cfg.Server.TrustedProxies = []string{"10.0.0.1", "10.0.0.0/8"}
	defer func() {
		cfg.Server.TrustedProxies = originalProxies
	}()

	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For honored from trusted proxy",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195"},
			expectedIP: "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For multiple from trusted proxy",
			remoteAddr: "10.0.0.50:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195, 70.41.3.18"},
			expectedIP: "203.0.113.195",
		},
		{
			name:       "X-Real-IP honored from trusted proxy",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "203.0.113.195"},
			expectedIP: "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For precedence over X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
				"X-Real-IP":       "198.51.100.178",
			},
			expectedIP: "203.0.113.195",
		},
		{
			name:       "untrusted proxy still uses direct IP",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195"},
			expectedIP: "192.168.1.1", // Not in trusted list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("getClientIP() = %q, want %q", ip, tt.expectedIP)
			}
		})
	}
}

func TestRateLimit_BurstCapacity(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware with burst of 5
	middleware := RateLimit(1, 5)
	handler := middleware(next)

	// First 5 requests should succeed (burst capacity)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, rr.Code)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Request 6: expected status 429, got %d", rr.Code)
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"  hello", "hello"},
		{"hello  ", "hello"},
		{"  hello  ", "hello"},
		{"\thello\t", "hello"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		result := trimSpace(tt.input)
		if result != tt.expected {
			t.Errorf("trimSpace(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(2, 2) // 2 RPS, burst of 2

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
	middleware := RateLimit(1, 1)
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
			name:       "X-Forwarded-For single",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195"},
			expectedIP: "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For multiple",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178"},
			expectedIP: "203.0.113.195",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "203.0.113.195"},
			expectedIP: "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
				"X-Real-IP":       "198.51.100.178",
			},
			expectedIP: "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For with whitespace",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "  203.0.113.195  "},
			expectedIP: "203.0.113.195",
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

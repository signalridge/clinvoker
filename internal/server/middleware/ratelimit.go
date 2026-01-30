package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/signalridge/clinvoker/internal/config"
)

// RateLimiter provides per-IP rate limiting with LRU eviction.
type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*limiterEntry
	rps      rate.Limit
	burst    int
	cleanup  time.Duration
	maxSize  int
	stopCh   chan struct{}
	stopOnce sync.Once
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

const (
	// DefaultRateLimiterCleanup is the default cleanup interval for stale entries.
	DefaultRateLimiterCleanup = 3 * time.Minute
	// DefaultRateLimiterMaxSize is the default maximum number of IP entries to track.
	// This prevents memory exhaustion from IP spoofing attacks.
	DefaultRateLimiterMaxSize = 10000
)

// NewRateLimiter creates a new rate limiter using the default cleanup interval.
// rps is requests per second, burst is the maximum burst size.
func NewRateLimiter(rps, burst int) *RateLimiter {
	return NewRateLimiterWithCleanup(rps, burst, DefaultRateLimiterCleanup)
}

// NewRateLimiterWithCleanup creates a new rate limiter with a custom cleanup interval.
// If cleanup <= 0, DefaultRateLimiterCleanup is used.
func NewRateLimiterWithCleanup(rps, burst int, cleanup time.Duration) *RateLimiter {
	return NewRateLimiterWithOptions(rps, burst, cleanup, DefaultRateLimiterMaxSize)
}

// NewRateLimiterWithOptions creates a new rate limiter with full configuration.
// If cleanup <= 0, DefaultRateLimiterCleanup is used.
// If maxSize <= 0, DefaultRateLimiterMaxSize is used.
func NewRateLimiterWithOptions(rps, burst int, cleanup time.Duration, maxSize int) *RateLimiter {
	if cleanup <= 0 {
		cleanup = DefaultRateLimiterCleanup
	}
	if maxSize <= 0 {
		maxSize = DefaultRateLimiterMaxSize
	}

	rl := &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		rps:      rate.Limit(rps),
		burst:    burst,
		cleanup:  cleanup,
		maxSize:  maxSize,
		stopCh:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from the given IP is allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		// Evict oldest entries if at capacity (LRU eviction)
		if len(rl.limiters) >= rl.maxSize {
			rl.evictOldestLocked()
		}

		entry = &limiterEntry{
			limiter:  rate.NewLimiter(rl.rps, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[ip] = entry
	} else {
		entry.lastSeen = time.Now()
	}

	return entry.limiter.Allow()
}

// evictOldestLocked removes the oldest entry from the map.
// Caller must hold the write lock.
func (rl *RateLimiter) evictOldestLocked() {
	var oldestIP string
	var oldestTime time.Time

	for ip, entry := range rl.limiters {
		if oldestIP == "" || entry.lastSeen.Before(oldestTime) {
			oldestIP = ip
			oldestTime = entry.lastSeen
		}
	}

	if oldestIP != "" {
		delete(rl.limiters, oldestIP)
	}
}

// cleanupLoop removes stale limiter entries periodically.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupStale()
		case <-rl.stopCh:
			return
		}
	}
}

// cleanupStale removes entries that haven't been seen recently.
func (rl *RateLimiter) cleanupStale() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.cleanup)
	for ip, entry := range rl.limiters {
		if entry.lastSeen.Before(cutoff) {
			delete(rl.limiters, ip)
		}
	}
}

// Stop terminates the cleanup goroutine.
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		close(rl.stopCh)
	})
}

// Middleware returns a middleware that enforces rate limiting using this limiter.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !rl.Allow(ip) {
				writeTooManyRequests(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit returns a middleware that limits requests per IP.
// Returns 429 Too Many Requests when limit is exceeded.
func RateLimit(rps, burst int) func(http.Handler) http.Handler {
	middleware, _ := RateLimitWithLimiter(rps, burst, 0)
	return middleware
}

// RateLimitWithLimiter returns a rate limiting middleware and the underlying limiter.
// The caller is responsible for calling Stop() on the limiter when no longer needed.
func RateLimitWithLimiter(rps, burst int, cleanup time.Duration) (func(http.Handler) http.Handler, *RateLimiter) {
	limiter := NewRateLimiterWithCleanup(rps, burst, cleanup)
	mw := limiter.Middleware()
	return mw, limiter
}

// getClientIP extracts the client IP from the request.
// Only trusts X-Forwarded-For and X-Real-IP headers if the request
// comes from a trusted proxy (configured via trusted_proxies setting).
// If trusted_proxies is empty, proxy headers are ignored (RemoteAddr only).
func getClientIP(r *http.Request) string {
	// Get the direct connection IP
	directIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		directIP = r.RemoteAddr
	}

	// Only trust proxy headers if request comes from a trusted proxy
	if !isTrustedProxy(directIP) {
		return directIP
	}

	// Try X-Forwarded-For first (leftmost is original client)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := indexByte(xff, ','); idx != -1 {
			xff = xff[:idx]
		}
		ip := trimSpace(xff)
		// Validate that the extracted value is a valid IP
		if ip != "" && isValidIP(ip) {
			return ip
		}
	}

	// Try X-Real-IP
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		ip := trimSpace(xrip)
		// Validate that the extracted value is a valid IP
		if isValidIP(ip) {
			return ip
		}
	}

	// Fall back to direct IP
	return directIP
}

// isTrustedProxy checks if the given IP is in the trusted proxies list.
// Returns false if no trusted proxies are configured (secure by default).
func isTrustedProxy(ip string) bool {
	cfg := config.Get()
	if cfg == nil || len(cfg.Server.TrustedProxies) == 0 {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, trusted := range cfg.Server.TrustedProxies {
		// Try parsing as CIDR
		if _, cidr, err := net.ParseCIDR(trusted); err == nil {
			if cidr.Contains(parsedIP) {
				return true
			}
			continue
		}

		// Try parsing as single IP
		if trustedIP := net.ParseIP(trusted); trustedIP != nil {
			if trustedIP.Equal(parsedIP) {
				return true
			}
		}
	}

	return false
}

// indexByte returns the index of the first occurrence of c in s, or -1 if not present.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// trimSpace trims leading and trailing whitespace.
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// isValidIP checks if the given string is a valid IPv4 or IPv6 address.
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// writeTooManyRequests writes a 429 Too Many Requests response.
func writeTooManyRequests(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "1")
	w.WriteHeader(http.StatusTooManyRequests)

	resp := errorResponse{
		Error:   "too_many_requests",
		Message: "rate limit exceeded, please retry later",
	}
	_ = json.NewEncoder(w).Encode(resp)
}

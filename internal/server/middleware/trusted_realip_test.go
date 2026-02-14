package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
)

func TestTrustedRealIP(t *testing.T) {
	cfg := config.Get()
	originalProxies := append([]string(nil), cfg.Server.TrustedProxies...)
	defer func() {
		cfg.Server.TrustedProxies = originalProxies
	}()

	tests := []struct {
		name       string
		proxies    []string
		remoteAddr string
		headers    map[string]string
		wantRemote string
	}{
		{
			name:       "ignores headers when no trusted proxies configured",
			proxies:    nil,
			remoteAddr: "203.0.113.10:4567",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.8",
			},
			wantRemote: "203.0.113.10:4567",
		},
		{
			name:       "ignores headers from untrusted proxy",
			proxies:    []string{"10.0.0.0/8"},
			remoteAddr: "203.0.113.10:4567",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.8",
			},
			wantRemote: "203.0.113.10:4567",
		},
		{
			name:       "uses first XFF IP from trusted proxy",
			proxies:    []string{"10.0.0.0/8"},
			remoteAddr: "10.0.0.5:4567",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.8, 10.0.0.5",
			},
			wantRemote: "198.51.100.8:4567",
		},
		{
			name:       "falls back to X-Real-IP when XFF invalid",
			proxies:    []string{"10.0.0.0/8"},
			remoteAddr: "10.0.0.5:4567",
			headers: map[string]string{
				"X-Forwarded-For": "not-an-ip",
				"X-Real-IP":       "198.51.100.20",
			},
			wantRemote: "198.51.100.20:4567",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.Server.TrustedProxies = append([]string(nil), tt.proxies...)

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			handler := TrustedRealIP(http.HandlerFunc(func(_ http.ResponseWriter, gotReq *http.Request) {
				if gotReq.RemoteAddr != tt.wantRemote {
					t.Fatalf("RemoteAddr = %q, want %q", gotReq.RemoteAddr, tt.wantRemote)
				}
			}))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		})
	}
}

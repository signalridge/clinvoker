package middleware

import (
	"net"
	"net/http"
)

// TrustedRealIP updates r.RemoteAddr using proxy headers only when the direct peer
// is in the trusted_proxies list.
//
// This prevents spoofed X-Forwarded-For / X-Real-IP headers from untrusted clients.
func TrustedRealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		directIP, directPort := splitRemoteAddr(r.RemoteAddr)

		if isTrustedProxy(directIP) {
			if clientIP := extractClientIPFromHeaders(r); clientIP != "" {
				if directPort != "" {
					r.RemoteAddr = net.JoinHostPort(clientIP, directPort)
				} else {
					r.RemoteAddr = clientIP
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

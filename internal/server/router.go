package server

import (
	"log/slog"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/server/middleware"
)

var defaultSkipAuthPaths = []string{"/health", "/docs", "/openapi.json", "/schemas", "/metrics"}

// NewRouter creates a chi router configured with the shared middleware stack.
func NewRouter(logger *slog.Logger) (chi.Router, *middleware.RateLimiter) {
	return NewRouterWithSkipAuthPaths(logger, defaultSkipAuthPaths...)
}

// NewRouterWithSkipAuthPaths creates a chi router with configurable auth skip paths.
func NewRouterWithSkipAuthPaths(logger *slog.Logger, skipPaths ...string) (chi.Router, *middleware.RateLimiter) {
	if logger == nil {
		logger = slog.Default()
	}

	router := chi.NewRouter()
	if len(skipPaths) == 0 {
		skipPaths = defaultSkipAuthPaths
	}

	// Get app config for middleware configuration
	appCfg := config.Get()

	// Add middleware in order:
	// RequestID → TrustedRealIP → Recoverer → RequestLogger → RequestSize → RateLimit → APIKeyAuth → Timeout → CORS
	router.Use(chiMiddleware.RequestID)
	router.Use(middleware.TrustedRealIP)
	router.Use(chiMiddleware.Recoverer)

	// Add request logging
	router.Use(RequestLogger(logger))

	// Add request size limits (optional)
	if appCfg.Server.MaxRequestBodyBytes > 0 {
		router.Use(middleware.RequestSize(appCfg.Server.MaxRequestBodyBytes))
	}

	// Add rate limiting if enabled
	var limiter *middleware.RateLimiter
	if appCfg.Server.RateLimitEnabled && appCfg.Server.RateLimitRPS > 0 {
		rps := appCfg.Server.RateLimitRPS
		burst := appCfg.Server.RateLimitBurst
		if burst <= 0 {
			burst = rps * 2 // Default burst to 2x RPS
		}
		cleanup := time.Duration(appCfg.Server.RateLimitCleanupSecs) * time.Second
		limiter = middleware.NewRateLimiterWithCleanup(rps, burst, cleanup)
		router.Use(limiter.Middleware())
		logger.Info("Rate limiting enabled", "rps", rps, "burst", burst)
	}

	// Add metrics middleware if enabled
	if appCfg.Server.MetricsEnabled {
		router.Use(middleware.Metrics)
	}

	// Add API key authentication (skips health, docs, and metrics endpoints)
	router.Use(middleware.SkipAuthPaths(skipPaths...))

	// Get timeout from config, with fallback to 5 minutes
	requestTimeout := time.Duration(appCfg.Server.RequestTimeoutSecs) * time.Second
	if requestTimeout <= 0 {
		requestTimeout = 5 * time.Minute
	}
	router.Use(chiMiddleware.Timeout(requestTimeout))

	// Add CORS - configurable via config, defaults to localhost for security
	corsOrigins := appCfg.Server.CORSAllowedOrigins
	if len(corsOrigins) == 0 {
		// Default to localhost only if no origins configured
		corsOrigins = []string{"http://localhost:*", "http://127.0.0.1:*"}
	}
	corsMaxAge := appCfg.Server.CORSMaxAge
	if corsMaxAge <= 0 {
		corsMaxAge = 300 // Default 5 minutes
	}

	// Warn about insecure CORS configuration
	// AllowCredentials + wildcard origins is a security risk and browsers may reject it
	if appCfg.Server.CORSAllowCredentials {
		for _, origin := range corsOrigins {
			if origin == "*" || strings.Contains(origin, "*") {
				logger.Warn("CORS configuration warning: AllowCredentials=true with wildcard origin may be rejected by browsers or create security risks",
					"origin", origin,
					"credentials", true)
				break
			}
		}
	}

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Api-Key", "anthropic-version"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: appCfg.Server.CORSAllowCredentials,
		MaxAge:           corsMaxAge,
	}))

	return router, limiter
}

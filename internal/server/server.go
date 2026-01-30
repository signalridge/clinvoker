// Package server provides HTTP server functionality for exposing AI backends as APIs.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/server/middleware"
	"github.com/signalridge/clinvoker/internal/server/service"
)

// Config holds server configuration.
type Config struct {
	Host string
	Port int
}

// Version is the server version.
const Version = "1.0.0"

// Server is the HTTP server for the AI backend APIs.
type Server struct {
	config    Config
	router    chi.Router
	api       huma.API
	executor  *service.Executor
	logger    *slog.Logger
	server    *http.Server
	limiter   *middleware.RateLimiter
	startTime time.Time
}

// New creates a new server instance.
func New(cfg Config, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	router := chi.NewRouter()

	// Get app config for middleware configuration
	appCfg := config.Get()

	// Add middleware in order:
	// RequestID → RealIP → Recoverer → RequestLogger → RequestSize → RateLimit → APIKeyAuth → Timeout → CORS
	router.Use(chiMiddleware.RequestID)
	router.Use(chiMiddleware.RealIP)
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
		logger.Info("Prometheus metrics enabled at /metrics")
	}

	// Add API key authentication (skips health, docs, and metrics endpoints)
	router.Use(middleware.SkipAuthPaths("/health", "/docs", "/openapi.json", "/schemas", "/metrics"))

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

	// Add /metrics endpoint for Prometheus if enabled
	if appCfg.Server.MetricsEnabled {
		router.Handle("/metrics", promhttp.Handler())
	}

	// Create huma API
	humaConfig := huma.DefaultConfig("clinvoker API", "1.0.0")
	humaConfig.Info.Description = "Unified AI CLI wrapper API for multiple backends"
	api := humachi.New(router, humaConfig)

	srv := &Server{
		config:    cfg,
		router:    router,
		api:       api,
		executor:  service.NewExecutor(),
		logger:    logger,
		limiter:   limiter,
		startTime: time.Now(),
	}
	return srv
}

// StartTime returns the server's start time.
func (s *Server) StartTime() time.Time {
	return s.startTime
}

// Uptime returns the server's uptime duration.
func (s *Server) Uptime() time.Duration {
	return time.Since(s.startTime)
}

// API returns the huma API for route registration.
func (s *Server) API() huma.API {
	return s.api
}

// Router returns the chi router.
func (s *Server) Router() chi.Router {
	return s.router
}

// Executor returns the service executor.
func (s *Server) Executor() *service.Executor {
	return s.executor
}

// Logger returns the server logger.
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Get timeouts from config with sensible defaults
	cfg := config.Get()
	readTimeout := time.Duration(cfg.Server.ReadTimeoutSecs) * time.Second
	if readTimeout <= 0 {
		readTimeout = 30 * time.Second
	}
	writeTimeout := time.Duration(cfg.Server.WriteTimeoutSecs) * time.Second
	if writeTimeout <= 0 {
		writeTimeout = 5 * time.Minute
	}
	idleTimeout := time.Duration(cfg.Server.IdleTimeoutSecs) * time.Second
	if idleTimeout <= 0 {
		idleTimeout = 120 * time.Second
	}

	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: 10 * time.Second, // Keep header timeout fixed for security
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	s.logger.Info("Starting server", "addr", addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	s.logger.Info("Shutting down server")
	if s.limiter != nil {
		s.limiter.Stop()
	}
	return s.server.Shutdown(ctx)
}

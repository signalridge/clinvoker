// Package server provides HTTP server functionality for exposing AI backends as APIs.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/signalridge/clinvoker/internal/server/service"
)

// Config holds server configuration.
type Config struct {
	Host string
	Port int
}

// Server is the HTTP server for the AI backend APIs.
type Server struct {
	config   Config
	router   chi.Router
	api      huma.API
	executor *service.Executor
	logger   *slog.Logger
	server   *http.Server
}

// New creates a new server instance.
func New(cfg Config, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	router := chi.NewRouter()

	// Add middleware
	router.Use(chiMiddleware.RequestID)
	router.Use(chiMiddleware.RealIP)
	router.Use(chiMiddleware.Recoverer)
	router.Use(chiMiddleware.Timeout(5 * time.Minute))

	// Add CORS - configured for local development
	// Note: AllowCredentials removed to work safely with permissive origins
	// For production with credentials, specify explicit allowed origins
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Api-Key", "anthropic-version"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Create huma API
	humaConfig := huma.DefaultConfig("clinvoker API", "1.0.0")
	humaConfig.Info.Description = "Unified AI CLI wrapper API for multiple backends"
	api := humachi.New(router, humaConfig)

	return &Server{
		config:   cfg,
		router:   router,
		api:      api,
		executor: service.NewExecutor(),
		logger:   logger,
	}
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

	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       120 * time.Second,
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
	return s.server.Shutdown(ctx)
}

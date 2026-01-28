package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/server"
	"github.com/signalridge/clinvoker/internal/server/handlers"
	"github.com/signalridge/clinvoker/internal/server/service"
)

// serveCmd starts the HTTP server.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	Long: `Start an HTTP server that exposes AI backends as APIs.

The server provides three distinct API styles:

  1. Custom RESTful API (/api/v1/*)
     Full-featured API with all clinvk capabilities:
     - POST /api/v1/prompt     - Execute single prompt
     - POST /api/v1/parallel   - Execute multiple prompts in parallel
     - POST /api/v1/chain      - Execute prompts in sequence
     - POST /api/v1/compare    - Compare responses across backends
     - GET  /api/v1/backends   - List available backends
     - GET  /api/v1/sessions   - List sessions

  2. OpenAI Compatible API (/openai/v1/*)
     Drop-in replacement for OpenAI API:
     - GET  /openai/v1/models           - List available models
     - POST /openai/v1/chat/completions - Create chat completion

  3. Anthropic Compatible API (/anthropic/v1/*)
     Drop-in replacement for Anthropic API:
     - POST /anthropic/v1/messages      - Create message

Configuration (in ~/.clinvk/config.yaml):
  server:
    host: "0.0.0.0"    # Bind to all interfaces
    port: 8080         # Listen port

Examples:
  clinvk serve
  clinvk serve --port 8080
  clinvk serve --host 0.0.0.0 --port 3000`,
	RunE: runServe,
}

var (
	serveHost string
	servePort int
)

func init() {
	serveCmd.Flags().StringVar(&serveHost, "host", "", "host to bind to (default from config or 127.0.0.1)")
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 0, "port to listen on (default from config or 8080)")

	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Set up logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Get config defaults
	appCfg := config.Get()

	// Use CLI flags if provided, otherwise use config
	host := serveHost
	if host == "" {
		host = appCfg.Server.Host
		if host == "" {
			host = "127.0.0.1"
		}
	}

	port := servePort
	if port == 0 {
		port = appCfg.Server.Port
		if port == 0 {
			port = 8080
		}
	}

	// Create server config
	cfg := server.Config{
		Host: host,
		Port: port,
	}

	// Create server
	srv := server.New(cfg, logger)

	// Register routes
	customHandlers := handlers.NewCustomHandlers(srv.Executor())
	customHandlers.Register(srv.API())

	// Register OpenAI compatible handlers (stateless)
	openaiHandlers := handlers.NewOpenAIHandlers(service.NewStatelessRunner(srv.Logger()), srv.Logger())
	openaiHandlers.Register(srv.API())

	// Register Anthropic compatible handlers (stateless)
	anthropicHandlers := handlers.NewAnthropicHandlers(service.NewStatelessRunner(srv.Logger()), srv.Logger())
	anthropicHandlers.Register(srv.API())

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	// Print startup message
	fmt.Printf("clinvk API server starting on http://%s:%d\n", host, port)
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  Custom API:     /api/v1/prompt, /api/v1/parallel, /api/v1/chain, /api/v1/compare")
	fmt.Println("  OpenAI:         /openai/v1/models, /openai/v1/chat/completions")
	fmt.Println("  Anthropic:      /anthropic/v1/messages")
	fmt.Println("  Docs:           /openapi.json")
	fmt.Println("  Health:         /health")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal or error
	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("Received shutdown signal")
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	logger.Info("Server stopped")
	return nil
}

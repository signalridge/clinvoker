package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/signalridge/clinvoker/internal/config"
	"github.com/signalridge/clinvoker/internal/server"
	"github.com/signalridge/clinvoker/internal/server/handlers"
	"github.com/signalridge/clinvoker/internal/server/mcp"
	"github.com/signalridge/clinvoker/internal/server/service"
)

// mcpCmd starts the MCP server.
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP (Model Context Protocol) server",
	Long: `Start an MCP server that exposes clinvoker capabilities as MCP tools.

Supported transports:
  stdio   - JSON-RPC over stdin/stdout (default, for local/CLI use)
  http    - JSON-RPC over HTTP POST with SSE notifications (for server deployments)

Available MCP tools:
  clinvk_prompt          - Execute a prompt on an AI backend
  clinvk_parallel        - Execute multiple prompts in parallel
  clinvk_chain           - Execute prompts in sequence (chain)
  clinvk_compare         - Compare responses across backends
  clinvk_backends        - List available backends
  clinvk_sessions        - List sessions
  clinvk_session_get     - Get a session by ID
  clinvk_session_delete  - Delete a session
  clinvk_health          - Check health status

Examples:
  clinvk mcp
  clinvk mcp --transport stdio
  clinvk mcp --transport http --port 8081
  clinvk mcp --transport http --host 0.0.0.0 --port 3000`,
	RunE: runMCP,
}

var (
	mcpTransport string
	mcpHost      string
	mcpPort      int
	mcpPath      string
	mcpExpose    bool
)

func init() {
	mcpCmd.Flags().StringVar(&mcpTransport, "transport", "", "transport type (stdio, http) (default from config or stdio)")
	mcpCmd.Flags().StringVar(&mcpHost, "host", "", "host to bind to for HTTP transport (default from config or 127.0.0.1)")
	mcpCmd.Flags().IntVarP(&mcpPort, "port", "p", 0, "port for HTTP transport (default from config or 8081)")
	mcpCmd.Flags().StringVar(&mcpPath, "path", "", "HTTP endpoint path (default from config or /mcp)")
	mcpCmd.Flags().BoolVar(&mcpExpose, "expose-health", false, "expose /health endpoint in MCP mode (default from config)")

	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, _ []string) error {
	// MCP uses stderr for logging to keep stdout clean for JSON-RPC (stdio)
	// or to separate server logs from HTTP responses.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	appCfg := config.Get()

	transport := mcpTransport
	if transport == "" {
		transport = appCfg.MCP.Transport
		if transport == "" {
			transport = "stdio"
		}
	}

	switch transport {
	case "stdio":
		return runMCPStdio(logger)
	case "http":
		exposeHealth := appCfg.MCP.ExposeHealth
		if cmd.Flags().Changed("expose-health") {
			exposeHealth = mcpExpose
		}
		return runMCPHTTP(logger, exposeHealth)
	default:
		return fmt.Errorf("unsupported transport: %s (supported: stdio, http)", transport)
	}
}

func newMCPDispatcher(logger *slog.Logger) (*mcp.Dispatcher, *service.Executor) {
	executor := service.NewExecutor()
	registry := mcp.NewRegistry()
	mcp.RegisterAllTools(registry, executor)
	return mcp.NewDispatcher(registry, logger), executor
}

func runMCPStdio(logger *slog.Logger) error {
	logger.Info("starting MCP server", "transport", "stdio")

	dispatcher, _ := newMCPDispatcher(logger)
	transport := mcp.NewStdioTransport(dispatcher, logger, os.Stdin, os.Stdout)

	// Set up graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := transport.Run(ctx); err != nil && ctx.Err() == nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	logger.Info("MCP server stopped")
	return nil
}

func runMCPHTTP(logger *slog.Logger, exposeHealth bool) error {
	appCfg := config.Get()

	// Resolve host.
	host := mcpHost
	if host == "" {
		host = appCfg.MCP.Host
	}
	if host == "" {
		host = "127.0.0.1"
	}

	// Resolve port (use a different default than the main server).
	port := mcpPort
	if port == 0 {
		port = appCfg.MCP.Port
		if port == 0 {
			port = 8081
		}
	}

	// Resolve path.
	path := mcpPath
	if path == "" {
		path = appCfg.MCP.HTTPPath
	}
	if path == "" {
		path = "/mcp"
	}

	dispatcher, executor := newMCPDispatcher(logger)
	transport := mcp.NewHTTPTransport(dispatcher, logger)

	router, limiter := server.NewRouter(logger)
	if limiter != nil {
		defer limiter.Stop()
	}
	router.Handle(path, transport.Handler())

	// Health check endpoint.
	if exposeHealth {
		startTime := time.Now()
		router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			body := mcpHealthResponse(r.Context(), executor, startTime)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(body)
		})
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	readTimeout := time.Duration(appCfg.Server.ReadTimeoutSecs) * time.Second
	if readTimeout <= 0 {
		readTimeout = 30 * time.Second
	}
	writeTimeout := time.Duration(appCfg.Server.WriteTimeoutSecs) * time.Second
	if writeTimeout <= 0 {
		writeTimeout = 5 * time.Minute
	}
	idleTimeout := time.Duration(appCfg.Server.IdleTimeoutSecs) * time.Second
	if idleTimeout <= 0 {
		idleTimeout = 120 * time.Second
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	// Set up graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	fmt.Fprintf(os.Stderr, "MCP HTTP server starting on http://%s%s\n", addr, path)
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n")

	select {
	case err := <-errCh:
		return fmt.Errorf("MCP HTTP server error: %w", err)
	case <-ctx.Done():
		logger.Info("received shutdown signal")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("MCP HTTP shutdown error: %w", err)
	}

	logger.Info("MCP HTTP server stopped")
	return nil
}

func mcpHealthResponse(ctx context.Context, executor *service.Executor, startTime time.Time) handlers.HealthResponseBody {
	backends := executor.ListBackends(ctx)
	storeHealth := executor.GetSessionStoreHealth(ctx)

	backendStatus := make([]handlers.BackendHealthStatus, len(backends))
	allBackendsAvailable := true
	for i, b := range backends {
		backendStatus[i] = handlers.BackendHealthStatus{
			Name:      b.Name,
			Available: b.Available,
		}
		if !b.Available {
			allBackendsAvailable = false
		}
	}

	sessionStoreStatus := handlers.SessionStoreStatus{
		Available:    storeHealth.Available,
		SessionCount: storeHealth.SessionCount,
		Error:        storeHealth.Error,
	}

	status := "ok"
	if !allBackendsAvailable {
		status = "degraded"
	}
	if !storeHealth.Available {
		status = "unhealthy"
	}

	uptime := time.Since(startTime)
	return handlers.HealthResponseBody{
		Status:       status,
		Version:      mcp.ServerVersion,
		Uptime:       formatDuration(uptime),
		UptimeMillis: uptime.Milliseconds(),
		Backends:     backendStatus,
		SessionStore: sessionStoreStatus,
	}
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

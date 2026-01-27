package server

import (
	"github.com/signalridge/clinvoker/internal/server/handlers"
)

// RegisterRoutes registers all API routes on the server.
func (s *Server) RegisterRoutes() {
	// Register custom RESTful API handlers
	customHandlers := handlers.NewCustomHandlers(s.executor)
	customHandlers.Register(s.api)

	// Register OpenAI-compatible API handlers
	openaiHandlers := handlers.NewOpenAIHandlers(s.executor)
	openaiHandlers.Register(s.api)

	// Register Anthropic-compatible API handlers
	anthropicHandlers := handlers.NewAnthropicHandlers(s.executor)
	anthropicHandlers.Register(s.api)
}

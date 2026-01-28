package server

import (
	"github.com/signalridge/clinvoker/internal/server/handlers"
	"github.com/signalridge/clinvoker/internal/server/service"
)

// RegisterRoutes registers all API routes on the server.
func (s *Server) RegisterRoutes() {
	// Register custom RESTful API handlers
	customHandlers := handlers.NewCustomHandlers(s.executor)
	customHandlers.Register(s.api)

	// Register OpenAI-compatible API handlers
	openaiHandlers := handlers.NewOpenAIHandlers(service.NewStatelessRunner(s.logger), s.logger)
	openaiHandlers.Register(s.api)

	// Register Anthropic-compatible API handlers
	anthropicHandlers := handlers.NewAnthropicHandlers(service.NewStatelessRunner(s.logger), s.logger)
	anthropicHandlers.Register(s.api)
}

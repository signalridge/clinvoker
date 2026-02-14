package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
)

// Dispatcher routes JSON-RPC requests to the appropriate handler.
type Dispatcher struct {
	registry *Registry
	logger   *slog.Logger
}

// NewDispatcher creates a new MCP dispatcher with the given tool registry.
func NewDispatcher(registry *Registry, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		registry: registry,
		logger:   logger,
	}
}

// Dispatch processes a JSON-RPC request and returns a response.
// For notifications (no ID), it returns nil.
func (d *Dispatcher) Dispatch(ctx context.Context, req *Request) *Response {
	d.logger.Debug("dispatching MCP request", "method", req.Method, "id", string(req.ID))

	if req.Method == MethodToolsCall && req.IsNotification() {
		// tools/call notifications are ignored to avoid side effects.
		return nil
	}

	var resp *Response
	switch req.Method {
	case MethodInitialize:
		resp = d.handleInitialize(req)
	case MethodPing:
		resp = d.handlePing(req)
	case MethodToolsList:
		resp = d.handleToolsList(req)
	case MethodToolsCall:
		resp = d.handleToolsCall(ctx, req)
	case "notifications/initialized":
		// Client acknowledgement after initialize; no response needed.
		return nil
	default:
		if req.IsNotification() {
			// Unknown notifications are silently ignored per MCP spec.
			return nil
		}
		return NewErrorResponse(req.ID, CodeMethodNotFound,
			fmt.Sprintf("method not found: %s", req.Method), nil)
	}

	if req.IsNotification() {
		return nil
	}

	return resp
}

func (d *Dispatcher) handleInitialize(req *Request) *Response {
	var params InitializeParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, "invalid initialize params", nil)
		}
	}

	if params.ProtocolVersion != "" && params.ProtocolVersion != ProtocolVersion {
		return NewErrorResponse(
			req.ID,
			CodeInvalidParams,
			fmt.Sprintf("unsupported protocol version: %s", params.ProtocolVersion),
			nil,
		)
	}

	d.logger.Info("MCP client initialized",
		"client", params.ClientInfo.Name,
		"clientVersion", params.ClientInfo.Version,
		"protocolVersion", params.ProtocolVersion,
	)

	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: Implementation{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}

	return NewResponse(req.ID, result)
}

func (d *Dispatcher) handlePing(req *Request) *Response {
	return NewResponse(req.ID, map[string]any{})
}

func (d *Dispatcher) handleToolsList(req *Request) *Response {
	tools := d.registry.ListTools()
	result := ToolsListResult{Tools: tools}
	return NewResponse(req.ID, result)
}

func (d *Dispatcher) handleToolsCall(ctx context.Context, req *Request) *Response {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, CodeInvalidParams, "invalid tools/call params", nil)
	}

	if params.Name == "" {
		return NewErrorResponse(req.ID, CodeInvalidParams, "missing tool name", nil)
	}

	def, ok := d.registry.Get(params.Name)
	if !ok {
		return NewErrorResponse(req.ID, CodeToolNotFound,
			fmt.Sprintf("unknown tool: %s", params.Name), nil)
	}

	d.logger.Info("calling MCP tool", "tool", params.Name)

	result, err := def.Handler(ctx, params.Arguments)
	if err != nil {
		var rpcErr *RPCErrorDetail
		if errors.As(err, &rpcErr) {
			return NewErrorResponse(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		}
		code, msg := MapExecutorError(err)
		d.logger.Error("tool call failed", "tool", params.Name, "error", err)
		return NewErrorResponse(req.ID, code, msg, nil)
	}

	return NewResponse(req.ID, result)
}

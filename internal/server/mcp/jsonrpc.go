// Package mcp provides a Model Context Protocol (MCP) server implementation.
package mcp

import "encoding/json"

// JSON-RPC 2.0 version constant.
const jsonRPCVersion = "2.0"

// Standard JSON-RPC 2.0 error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsNotification returns true if this request is a notification (no ID).
func (r *Request) IsNotification() bool {
	return r.ID == nil || string(r.ID) == "null"
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Notification represents a JSON-RPC 2.0 notification (no ID, no response expected).
type Notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// NewResponse creates a successful response for a given request ID.
func NewResponse(id json.RawMessage, result any) *Response {
	return &Response{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates an error response for a given request ID.
func NewErrorResponse(id json.RawMessage, code int, message string, data any) *Response {
	return &Response{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// NewNotification creates a JSON-RPC notification.
func NewNotification(method string, params any) *Notification {
	return &Notification{
		JSONRPC: jsonRPCVersion,
		Method:  method,
		Params:  params,
	}
}

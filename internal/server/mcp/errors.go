package mcp

import "strings"

// RPCErrorDetail represents a JSON-RPC error that should be surfaced to clients.
type RPCErrorDetail struct {
	Code    int
	Message string
	Data    any
}

func (e *RPCErrorDetail) Error() string {
	return e.Message
}

// NewRPCErrorDetail creates a new coded RPC error.
func NewRPCErrorDetail(code int, message string, data any) *RPCErrorDetail {
	return &RPCErrorDetail{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Application-level error codes (beyond JSON-RPC standard codes).
const (
	// CodeToolNotFound indicates the requested tool does not exist.
	CodeToolNotFound = -32001

	// CodeToolExecutionError indicates a tool execution failure.
	CodeToolExecutionError = -32002

	// CodeBackendUnavailable indicates the requested backend is not available.
	CodeBackendUnavailable = -32003

	// CodeTimeout indicates the operation timed out.
	CodeTimeout = -32004
)

// MapExecutorError maps executor/backend error messages to appropriate JSON-RPC error codes.
func MapExecutorError(err error) (int, string) {
	if err == nil {
		return CodeInternalError, "unknown error"
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "not available"), strings.Contains(lower, "disabled"):
		return CodeBackendUnavailable, msg
	case strings.Contains(lower, "unknown backend"):
		return CodeInvalidParams, msg
	case strings.Contains(lower, "not found"):
		return CodeInvalidParams, msg
	case strings.Contains(lower, "timed out"), strings.Contains(lower, "deadline exceeded"):
		return CodeTimeout, msg
	case strings.Contains(lower, "invalid"),
		strings.Contains(lower, "required"),
		strings.Contains(lower, "not allowed"),
		strings.Contains(lower, "blocked path"),
		strings.Contains(lower, "must be"):
		return CodeInvalidParams, msg
	default:
		return CodeToolExecutionError, msg
	}
}

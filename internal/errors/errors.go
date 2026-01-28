// Package errors provides structured error types for clinvoker.
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents a machine-readable error code.
type ErrorCode string

// Error codes for different error categories.
const (
	// Backend errors
	ErrCodeBackendUnavailable ErrorCode = "backend_unavailable"
	ErrCodeBackendNotFound    ErrorCode = "backend_not_found"
	ErrCodeBackendTimeout     ErrorCode = "backend_timeout"
	ErrCodeBackendExecution   ErrorCode = "backend_execution_error"

	// Request errors
	ErrCodeInvalidRequest  ErrorCode = "invalid_request"
	ErrCodeMissingRequired ErrorCode = "missing_required_field"
	ErrCodeValidation      ErrorCode = "validation_error"

	// Session errors
	ErrCodeSessionNotFound ErrorCode = "session_not_found"
	ErrCodeSessionExpired  ErrorCode = "session_expired"
	ErrCodeSessionConflict ErrorCode = "session_conflict"

	// Configuration errors
	ErrCodeConfigNotFound ErrorCode = "config_not_found"
	ErrCodeConfigInvalid  ErrorCode = "config_invalid"
	ErrCodeConfigParse    ErrorCode = "config_parse_error"

	// I/O errors
	ErrCodeIOError    ErrorCode = "io_error"
	ErrCodePermission ErrorCode = "permission_denied"

	// Internal errors
	ErrCodeInternal ErrorCode = "internal_error"
	ErrCodeUnknown  ErrorCode = "unknown_error"
)

// AppError is a structured error with code, message, and context.
type AppError struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Cause   error          `json:"-"`
	Context map[string]any `json:"context,omitempty"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithCause sets the underlying cause.
func (e *AppError) WithCause(err error) *AppError {
	e.Cause = err
	return e
}

// WithContext adds context to the error.
func (e *AppError) WithContext(key string, value any) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// New creates a new AppError with the given code and message.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an error with an AppError.
func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Sentinel errors for common cases.
var (
	// Backend errors
	ErrBackendUnavailable = New(ErrCodeBackendUnavailable, "backend is not available")
	ErrBackendNotFound    = New(ErrCodeBackendNotFound, "backend not found")
	ErrBackendTimeout     = New(ErrCodeBackendTimeout, "backend operation timed out")

	// Session errors
	ErrSessionNotFound = New(ErrCodeSessionNotFound, "session not found")
	ErrSessionExpired  = New(ErrCodeSessionExpired, "session has expired")

	// Configuration errors
	ErrConfigNotFound = New(ErrCodeConfigNotFound, "configuration file not found")
	ErrConfigInvalid  = New(ErrCodeConfigInvalid, "configuration is invalid")

	// Request errors
	ErrInvalidRequest = New(ErrCodeInvalidRequest, "invalid request")
	ErrMissingField   = New(ErrCodeMissingRequired, "required field is missing")
)

// IsCode checks if an error has the given error code.
func IsCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetCode returns the error code from an error, or ErrCodeUnknown if not an AppError.
func GetCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrCodeUnknown
}

// GetContext returns the context from an error, or nil if not an AppError.
func GetContext(err error) map[string]any {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Context
	}
	return nil
}

// BackendError creates a backend-related error.
func BackendError(backend string, err error) *AppError {
	return Wrap(ErrCodeBackendExecution, "backend execution failed", err).
		WithContext("backend", backend)
}

// ValidationError creates a validation error.
func ValidationError(field, reason string) *AppError {
	return New(ErrCodeValidation, reason).
		WithContext("field", field)
}

// ConfigError creates a configuration error.
func ConfigError(message string, err error) *AppError {
	return Wrap(ErrCodeConfigInvalid, message, err)
}

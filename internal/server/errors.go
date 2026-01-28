package server

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "github.com/signalridge/clinvoker/internal/errors"
)

// APIError represents a structured API error response.
type APIError struct {
	Error   string         `json:"error"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// NewAPIError creates a new API error from an AppError.
func NewAPIError(err *apperrors.AppError) *APIError {
	return &APIError{
		Error:   string(err.Code),
		Code:    string(err.Code),
		Message: err.Message,
		Details: err.Context,
	}
}

// HTTPStatusCode returns the appropriate HTTP status code for an error code.
func HTTPStatusCode(code apperrors.ErrorCode) int {
	switch code {
	case apperrors.ErrCodeBackendUnavailable,
		apperrors.ErrCodeBackendNotFound:
		return http.StatusServiceUnavailable

	case apperrors.ErrCodeBackendTimeout:
		return http.StatusGatewayTimeout

	case apperrors.ErrCodeInvalidRequest,
		apperrors.ErrCodeMissingRequired,
		apperrors.ErrCodeValidation:
		return http.StatusBadRequest

	case apperrors.ErrCodeSessionNotFound,
		apperrors.ErrCodeConfigNotFound:
		return http.StatusNotFound

	case apperrors.ErrCodeSessionConflict:
		return http.StatusConflict

	case apperrors.ErrCodePermission:
		return http.StatusForbidden

	case apperrors.ErrCodeSessionExpired:
		return http.StatusGone

	case apperrors.ErrCodeConfigInvalid,
		apperrors.ErrCodeConfigParse:
		return http.StatusUnprocessableEntity

	case apperrors.ErrCodeIOError,
		apperrors.ErrCodeBackendExecution,
		apperrors.ErrCodeInternal,
		apperrors.ErrCodeUnknown:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}

// WriteError writes an error response to the HTTP response writer.
func WriteError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		// Wrap unknown errors
		appErr = apperrors.Wrap(apperrors.ErrCodeUnknown, err.Error(), err)
	}

	apiErr := NewAPIError(appErr)
	statusCode := HTTPStatusCode(appErr.Code)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(apiErr)
}

// ErrorHandler is a middleware that handles errors in a structured way.
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response writer wrapper to catch errors
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Recover from panics
		defer func() {
			if rec := recover(); rec != nil {
				var err error
				switch v := rec.(type) {
				case error:
					err = v
				case string:
					err = errors.New(v)
				default:
					err = errors.New("unknown panic")
				}
				appErr := apperrors.Wrap(apperrors.ErrCodeInternal, "internal server error", err)
				WriteError(w, appErr)
			}
		}()

		next.ServeHTTP(rw, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// StatusCode returns the captured status code.
func (rw *responseWriter) StatusCode() int {
	return rw.statusCode
}

// BadRequest creates a 400 Bad Request error.
func BadRequest(message string) *apperrors.AppError {
	return apperrors.New(apperrors.ErrCodeInvalidRequest, message)
}

// NotFound creates a 404 Not Found error.
func NotFound(resource, id string) *apperrors.AppError {
	return apperrors.New(apperrors.ErrCodeSessionNotFound, resource+" not found").
		WithContext("id", id)
}

// InternalError creates a 500 Internal Server Error.
func InternalError(message string, cause error) *apperrors.AppError {
	return apperrors.Wrap(apperrors.ErrCodeInternal, message, cause)
}

// ValidationError creates a 400 Validation Error.
func ValidationError(field, reason string) *apperrors.AppError {
	return apperrors.ValidationError(field, reason)
}

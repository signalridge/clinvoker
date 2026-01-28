package errors

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name: "without cause",
			err: &AppError{
				Code:    ErrCodeBackendNotFound,
				Message: "test backend not found",
			},
			expected: "backend_not_found: test backend not found",
		},
		{
			name: "with cause",
			err: &AppError{
				Code:    ErrCodeBackendExecution,
				Message: "execution failed",
				Cause:   errors.New("connection refused"),
			},
			expected: "backend_execution_error: execution failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &AppError{
		Code:    ErrCodeInternal,
		Message: "internal error",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestAppError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		target   error
		expected bool
	}{
		{
			name:     "same code matches",
			err:      New(ErrCodeBackendNotFound, "test"),
			target:   New(ErrCodeBackendNotFound, "different message"),
			expected: true,
		},
		{
			name:     "different code does not match",
			err:      New(ErrCodeBackendNotFound, "test"),
			target:   New(ErrCodeSessionNotFound, "test"),
			expected: false,
		},
		{
			name:     "non-AppError does not match",
			err:      New(ErrCodeBackendNotFound, "test"),
			target:   errors.New("regular error"),
			expected: false,
		},
		{
			name:     "sentinel error matches",
			err:      New(ErrCodeBackendNotFound, "custom message"),
			target:   ErrBackendNotFound,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Is(tt.target); got != tt.expected {
				t.Errorf("Is() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_WithCause(t *testing.T) {
	cause := errors.New("underlying cause")
	err := New(ErrCodeInternal, "internal error").WithCause(cause)

	if err.Cause != cause {
		t.Errorf("WithCause() did not set Cause correctly")
	}
}

func TestAppError_WithContext(t *testing.T) {
	err := New(ErrCodeValidation, "validation failed").
		WithContext("field", "email").
		WithContext("reason", "invalid format")

	if err.Context == nil {
		t.Fatal("WithContext() did not initialize Context")
	}

	if err.Context["field"] != "email" {
		t.Errorf("WithContext() field = %v, want 'email'", err.Context["field"])
	}

	if err.Context["reason"] != "invalid format" {
		t.Errorf("WithContext() reason = %v, want 'invalid format'", err.Context["reason"])
	}
}

func TestNew(t *testing.T) {
	err := New(ErrCodeSessionNotFound, "session abc123 not found")

	if err.Code != ErrCodeSessionNotFound {
		t.Errorf("New() Code = %v, want %v", err.Code, ErrCodeSessionNotFound)
	}

	if err.Message != "session abc123 not found" {
		t.Errorf("New() Message = %v, want 'session abc123 not found'", err.Message)
	}

	if err.Cause != nil {
		t.Errorf("New() Cause = %v, want nil", err.Cause)
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("file not found")
	err := Wrap(ErrCodeIOError, "failed to read config", cause)

	if err.Code != ErrCodeIOError {
		t.Errorf("Wrap() Code = %v, want %v", err.Code, ErrCodeIOError)
	}

	if err.Message != "failed to read config" {
		t.Errorf("Wrap() Message = %v, want 'failed to read config'", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("Wrap() Cause = %v, want %v", err.Cause, cause)
	}
}

func TestIsCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     ErrorCode
		expected bool
	}{
		{
			name:     "AppError with matching code",
			err:      New(ErrCodeBackendNotFound, "test"),
			code:     ErrCodeBackendNotFound,
			expected: true,
		},
		{
			name:     "AppError with different code",
			err:      New(ErrCodeBackendNotFound, "test"),
			code:     ErrCodeSessionNotFound,
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			code:     ErrCodeBackendNotFound,
			expected: false,
		},
		{
			name:     "wrapped AppError",
			err:      Wrap(ErrCodeIOError, "wrapped", errors.New("cause")),
			code:     ErrCodeIOError,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCode(tt.err, tt.code); got != tt.expected {
				t.Errorf("IsCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name:     "AppError",
			err:      New(ErrCodeBackendNotFound, "test"),
			expected: ErrCodeBackendNotFound,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: ErrCodeUnknown,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: ErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCode(tt.err); got != tt.expected {
				t.Errorf("GetCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetContext(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectNil   bool
		expectField string
	}{
		{
			name:        "AppError with context",
			err:         New(ErrCodeValidation, "test").WithContext("field", "email"),
			expectNil:   false,
			expectField: "email",
		},
		{
			name:      "AppError without context",
			err:       New(ErrCodeValidation, "test"),
			expectNil: true,
		},
		{
			name:      "regular error",
			err:       errors.New("regular error"),
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetContext(tt.err)
			if tt.expectNil && got != nil {
				t.Errorf("GetContext() = %v, want nil", got)
			}
			if !tt.expectNil && got == nil {
				t.Error("GetContext() = nil, want non-nil")
			}
			if !tt.expectNil && got["field"] != tt.expectField {
				t.Errorf("GetContext()[\"field\"] = %v, want %v", got["field"], tt.expectField)
			}
		})
	}
}

func TestBackendError(t *testing.T) {
	cause := errors.New("connection refused")
	err := BackendError("claude", cause)

	if err.Code != ErrCodeBackendExecution {
		t.Errorf("BackendError() Code = %v, want %v", err.Code, ErrCodeBackendExecution)
	}

	if err.Cause != cause {
		t.Errorf("BackendError() Cause = %v, want %v", err.Cause, cause)
	}

	if err.Context["backend"] != "claude" {
		t.Errorf("BackendError() context[backend] = %v, want 'claude'", err.Context["backend"])
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("email", "invalid email format")

	if err.Code != ErrCodeValidation {
		t.Errorf("ValidationError() Code = %v, want %v", err.Code, ErrCodeValidation)
	}

	if err.Message != "invalid email format" {
		t.Errorf("ValidationError() Message = %v, want 'invalid email format'", err.Message)
	}

	if err.Context["field"] != "email" {
		t.Errorf("ValidationError() context[field] = %v, want 'email'", err.Context["field"])
	}
}

func TestConfigError(t *testing.T) {
	cause := errors.New("yaml parse error")
	err := ConfigError("failed to parse config", cause)

	if err.Code != ErrCodeConfigInvalid {
		t.Errorf("ConfigError() Code = %v, want %v", err.Code, ErrCodeConfigInvalid)
	}

	if err.Cause != cause {
		t.Errorf("ConfigError() Cause = %v, want %v", err.Cause, cause)
	}
}

func TestErrorsIs(t *testing.T) {
	// Test that errors.Is works correctly with wrapped AppErrors
	wrapped := Wrap(ErrCodeBackendNotFound, "wrapped error", nil)

	if !errors.Is(wrapped, ErrBackendNotFound) {
		t.Error("errors.Is() should match sentinel error by code")
	}

	if errors.Is(wrapped, ErrSessionNotFound) {
		t.Error("errors.Is() should not match different sentinel error")
	}
}

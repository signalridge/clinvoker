package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/signalridge/clinvoker/internal/errors"
)

func decodeAPIError(t *testing.T, body []byte) APIError {
	t.Helper()

	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		t.Fatalf("unmarshal api error failed: %v", err)
	}
	return apiErr
}

func TestNewAPIError(t *testing.T) {
	appErr := apperrors.New(apperrors.ErrCodeValidation, "invalid field").
		WithContext("field", "name")

	apiErr := NewAPIError(appErr)
	if apiErr.Error != string(apperrors.ErrCodeValidation) {
		t.Errorf("Error = %q, want %q", apiErr.Error, apperrors.ErrCodeValidation)
	}
	if apiErr.Code != string(apperrors.ErrCodeValidation) {
		t.Errorf("Code = %q, want %q", apiErr.Code, apperrors.ErrCodeValidation)
	}
	if apiErr.Message != "invalid field" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "invalid field")
	}
	if got := apiErr.Details["field"]; got != "name" {
		t.Errorf("Details[field] = %v, want %q", got, "name")
	}
}

func TestHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name string
		code apperrors.ErrorCode
		want int
	}{
		{name: "backend unavailable", code: apperrors.ErrCodeBackendUnavailable, want: http.StatusServiceUnavailable},
		{name: "backend not found", code: apperrors.ErrCodeBackendNotFound, want: http.StatusServiceUnavailable},
		{name: "backend timeout", code: apperrors.ErrCodeBackendTimeout, want: http.StatusGatewayTimeout},
		{name: "invalid request", code: apperrors.ErrCodeInvalidRequest, want: http.StatusBadRequest},
		{name: "missing required", code: apperrors.ErrCodeMissingRequired, want: http.StatusBadRequest},
		{name: "validation", code: apperrors.ErrCodeValidation, want: http.StatusBadRequest},
		{name: "session not found", code: apperrors.ErrCodeSessionNotFound, want: http.StatusNotFound},
		{name: "config not found", code: apperrors.ErrCodeConfigNotFound, want: http.StatusNotFound},
		{name: "session conflict", code: apperrors.ErrCodeSessionConflict, want: http.StatusConflict},
		{name: "permission", code: apperrors.ErrCodePermission, want: http.StatusForbidden},
		{name: "session expired", code: apperrors.ErrCodeSessionExpired, want: http.StatusGone},
		{name: "config invalid", code: apperrors.ErrCodeConfigInvalid, want: http.StatusUnprocessableEntity},
		{name: "config parse", code: apperrors.ErrCodeConfigParse, want: http.StatusUnprocessableEntity},
		{name: "io error", code: apperrors.ErrCodeIOError, want: http.StatusInternalServerError},
		{name: "backend execution", code: apperrors.ErrCodeBackendExecution, want: http.StatusInternalServerError},
		{name: "internal", code: apperrors.ErrCodeInternal, want: http.StatusInternalServerError},
		{name: "unknown", code: apperrors.ErrCodeUnknown, want: http.StatusInternalServerError},
		{name: "default unknown code", code: apperrors.ErrorCode("new_code"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HTTPStatusCode(tt.code)
			if got != tt.want {
				t.Errorf("HTTPStatusCode(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestWriteError_WithAppError(t *testing.T) {
	rr := httptest.NewRecorder()
	appErr := apperrors.New(apperrors.ErrCodeInvalidRequest, "bad input").
		WithContext("field", "model")

	WriteError(rr, appErr)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	apiErr := decodeAPIError(t, rr.Body.Bytes())
	if apiErr.Code != string(apperrors.ErrCodeInvalidRequest) {
		t.Errorf("Code = %q, want %q", apiErr.Code, apperrors.ErrCodeInvalidRequest)
	}
	if apiErr.Message != "bad input" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "bad input")
	}
	if got := apiErr.Details["field"]; got != "model" {
		t.Errorf("Details[field] = %v, want %q", got, "model")
	}
}

func TestWriteError_WithGenericError(t *testing.T) {
	rr := httptest.NewRecorder()

	WriteError(rr, errors.New("boom"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	apiErr := decodeAPIError(t, rr.Body.Bytes())
	if apiErr.Code != string(apperrors.ErrCodeUnknown) {
		t.Errorf("Code = %q, want %q", apiErr.Code, apperrors.ErrCodeUnknown)
	}
	if apiErr.Message != "boom" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "boom")
	}
}

func TestErrorHandler_PassThrough(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	})

	handler := ErrorHandler(next)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}
	if body := rr.Body.String(); body != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}

func TestErrorHandler_PanicRecovery(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "panic error", value: errors.New("panic error")},
		{name: "panic string", value: "panic string"},
		{name: "panic unknown", value: 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				panic(tt.value)
			})
			handler := ErrorHandler(next)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/panic", http.NoBody)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
			}

			apiErr := decodeAPIError(t, rr.Body.Bytes())
			if apiErr.Code != string(apperrors.ErrCodeInternal) {
				t.Errorf("Code = %q, want %q", apiErr.Code, apperrors.ErrCodeInternal)
			}
			if apiErr.Message != "internal server error" {
				t.Errorf("Message = %q, want %q", apiErr.Message, "internal server error")
			}
		})
	}
}

func TestResponseWriter_StatusCapture(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rr,
		statusCode:     http.StatusOK,
	}

	rw.WriteHeader(http.StatusCreated)

	if got := rw.StatusCode(); got != http.StatusCreated {
		t.Errorf("StatusCode() = %d, want %d", got, http.StatusCreated)
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("wrapped writer status = %d, want %d", rr.Code, http.StatusCreated)
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		err := BadRequest("bad payload")
		if err.Code != apperrors.ErrCodeInvalidRequest {
			t.Errorf("Code = %q, want %q", err.Code, apperrors.ErrCodeInvalidRequest)
		}
		if err.Message != "bad payload" {
			t.Errorf("Message = %q, want %q", err.Message, "bad payload")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		err := NotFound("session", "abc")
		if err.Code != apperrors.ErrCodeSessionNotFound {
			t.Errorf("Code = %q, want %q", err.Code, apperrors.ErrCodeSessionNotFound)
		}
		if err.Message != "session not found" {
			t.Errorf("Message = %q, want %q", err.Message, "session not found")
		}
		if got := err.Context["id"]; got != "abc" {
			t.Errorf("Context[id] = %v, want %q", got, "abc")
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		cause := errors.New("database down")
		err := InternalError("failed", cause)
		if err.Code != apperrors.ErrCodeInternal {
			t.Errorf("Code = %q, want %q", err.Code, apperrors.ErrCodeInternal)
		}
		if err.Message != "failed" {
			t.Errorf("Message = %q, want %q", err.Message, "failed")
		}
		if !errors.Is(err, cause) {
			t.Errorf("expected wrapped cause %v", cause)
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := ValidationError("prompt", "required")
		if err.Code != apperrors.ErrCodeValidation {
			t.Errorf("Code = %q, want %q", err.Code, apperrors.ErrCodeValidation)
		}
		if err.Message != "required" {
			t.Errorf("Message = %q, want %q", err.Message, "required")
		}
		if got := err.Context["field"]; got != "prompt" {
			t.Errorf("Context[field] = %v, want %q", got, "prompt")
		}
	})
}

package mcp

import (
	"errors"
	"testing"
)

func TestMapExecutorError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantCode     int
		wantContains string
	}{
		{
			name:         "nil error",
			err:          nil,
			wantCode:     CodeInternalError,
			wantContains: "unknown error",
		},
		{
			name:         "backend not available",
			err:          errors.New("backend 'gemini' is not available"),
			wantCode:     CodeBackendUnavailable,
			wantContains: "not available",
		},
		{
			name:         "backend disabled",
			err:          errors.New("backend 'claude' is disabled in config"),
			wantCode:     CodeBackendUnavailable,
			wantContains: "disabled",
		},
		{
			name:         "unknown backend",
			err:          errors.New("unknown backend \"foo\" (available: [claude])"),
			wantCode:     CodeInvalidParams,
			wantContains: "unknown backend",
		},
		{
			name:         "not found",
			err:          errors.New("session '123' not found"),
			wantCode:     CodeToolNotFound,
			wantContains: "not found",
		},
		{
			name:         "timed out",
			err:          errors.New("operation timed out"),
			wantCode:     CodeTimeout,
			wantContains: "timed out",
		},
		{
			name:         "deadline exceeded",
			err:          errors.New("context deadline exceeded"),
			wantCode:     CodeTimeout,
			wantContains: "deadline exceeded",
		},
		{
			name:         "invalid parameter",
			err:          errors.New("invalid backend name"),
			wantCode:     CodeInvalidParams,
			wantContains: "invalid",
		},
		{
			name:         "required parameter",
			err:          errors.New("required parameter 'name' missing"),
			wantCode:     CodeInvalidParams,
			wantContains: "required",
		},
		{
			name:         "generic error",
			err:          errors.New("something went wrong"),
			wantCode:     CodeToolExecutionError,
			wantContains: "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := MapExecutorError(tt.err)

			if code != tt.wantCode {
				t.Errorf("code = %d, want %d", code, tt.wantCode)
			}
			if !contains(msg, tt.wantContains) {
				t.Errorf("message = %q, should contain %q", msg, tt.wantContains)
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	// Verify custom error codes are in the valid application range (-32000 to -32099)
	codes := []struct {
		name string
		code int
	}{
		{"CodeToolNotFound", CodeToolNotFound},
		{"CodeToolExecutionError", CodeToolExecutionError},
		{"CodeBackendUnavailable", CodeBackendUnavailable},
		{"CodeTimeout", CodeTimeout},
	}

	for _, tc := range codes {
		t.Run(tc.name, func(t *testing.T) {
			if tc.code > -32000 || tc.code < -32099 {
				t.Errorf("%s = %d, should be in range [-32099, -32000]", tc.name, tc.code)
			}
		})
	}
}

func TestErrorCodes_Uniqueness(t *testing.T) {
	codes := map[int]string{
		CodeToolNotFound:       "CodeToolNotFound",
		CodeToolExecutionError: "CodeToolExecutionError",
		CodeBackendUnavailable: "CodeBackendUnavailable",
		CodeTimeout:            "CodeTimeout",
	}

	seen := make(map[int]string)
	for code, name := range codes {
		if prevName, exists := seen[code]; exists {
			t.Errorf("duplicate error code %d: %s and %s", code, prevName, name)
		}
		seen[code] = name
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

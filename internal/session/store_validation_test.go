package session

import (
	"testing"
)

func TestValidateSessionID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid 16-char hex ID",
			id:      "a1b2c3d4e5f67890",
			wantErr: false,
		},
		{
			name:    "valid short prefix",
			id:      "a1b2c3",
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
		},
		{
			name:    "path traversal with ..",
			id:      "../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path traversal with /",
			id:      "a1b2c3/d4e5f6",
			wantErr: true,
		},
		{
			name:    "path traversal with backslash",
			id:      "a1b2c3\\d4e5f6",
			wantErr: true,
		},
		{
			name:    "invalid 16-char - uppercase",
			id:      "A1B2C3D4E5F67890",
			wantErr: true,
		},
		{
			name:    "invalid 16-char - non-hex",
			id:      "g1b2c3d4e5f67890",
			wantErr: true,
		},
		{
			name:    "valid shorter prefix - allowed",
			id:      "a1b2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSessionID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSessionID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSessionPrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{
			name:    "valid full hex ID",
			prefix:  "a1b2c3d4e5f67890",
			wantErr: false,
		},
		{
			name:    "valid short prefix",
			prefix:  "a1b2c3",
			wantErr: false,
		},
		{
			name:    "valid single char",
			prefix:  "a",
			wantErr: false,
		},
		{
			name:    "empty prefix",
			prefix:  "",
			wantErr: true,
		},
		{
			name:    "path traversal with ..",
			prefix:  "../etc",
			wantErr: true,
		},
		{
			name:    "path traversal with /",
			prefix:  "a1b2/c3",
			wantErr: true,
		},
		{
			name:    "path traversal with backslash",
			prefix:  "a1b2\\c3",
			wantErr: true,
		},
		{
			name:    "invalid - uppercase letters",
			prefix:  "A1B2C3",
			wantErr: true,
		},
		{
			name:    "invalid - non-hex characters",
			prefix:  "g1b2c3",
			wantErr: true,
		},
		{
			name:    "invalid - special characters",
			prefix:  "a1b2-c3",
			wantErr: true,
		},
		{
			name:    "invalid - spaces",
			prefix:  "a1b2 c3",
			wantErr: true,
		},
		{
			name:    "valid all digits",
			prefix:  "123456",
			wantErr: false,
		},
		{
			name:    "valid all letters",
			prefix:  "abcdef",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSessionPrefix(tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSessionPrefix(%q) error = %v, wantErr %v", tt.prefix, err, tt.wantErr)
			}
		})
	}
}

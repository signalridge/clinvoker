package backend

import (
	"testing"
)

func TestValidateExtraFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		wantErr bool
	}{
		{
			name:    "empty flags",
			flags:   []string{},
			wantErr: false,
		},
		{
			name:    "valid flags with values",
			flags:   []string{"--model=gpt-4", "-v"},
			wantErr: false,
		},
		{
			name:    "valid flag with value",
			flags:   []string{"--output-format=json"},
			wantErr: false,
		},
		{
			name:    "dangerous flag --dangerously-skip-permissions",
			flags:   []string{"--dangerously-skip-permissions"},
			wantErr: true,
		},
		{
			name:    "dangerous flag --no-verify",
			flags:   []string{"--no-verify"},
			wantErr: true,
		},
		{
			name:    "dangerous flag --force",
			flags:   []string{"--force"},
			wantErr: true,
		},
		{
			name:    "dangerous flag -f",
			flags:   []string{"-f"},
			wantErr: true,
		},
		{
			name:    "invalid flag format - no dash",
			flags:   []string{"model"},
			wantErr: true,
		},
		{
			name:    "dangerous flag with value",
			flags:   []string{"--force=true"},
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid",
			flags:   []string{"--model=gpt-4", "--force"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExtraFlags(tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExtraFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

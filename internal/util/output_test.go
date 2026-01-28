package util

import "testing"

func TestSelectOutput(t *testing.T) {
	tests := []struct {
		name     string
		stdout   string
		stderr   string
		exitCode int
		want     string
	}{
		{
			name:     "success with stdout",
			stdout:   "output",
			stderr:   "",
			exitCode: 0,
			want:     "output",
		},
		{
			name:     "success prefers stdout over stderr",
			stdout:   "output",
			stderr:   "warning",
			exitCode: 0,
			want:     "output",
		},
		{
			name:     "error with stderr returns stderr",
			stdout:   "output",
			stderr:   "error message",
			exitCode: 1,
			want:     "error message",
		},
		{
			name:     "error without stderr falls back to stdout",
			stdout:   "output",
			stderr:   "",
			exitCode: 1,
			want:     "output",
		},
		{
			name:     "empty stdout returns stderr",
			stdout:   "",
			stderr:   "stderr content",
			exitCode: 0,
			want:     "stderr content",
		},
		{
			name:     "both empty returns empty",
			stdout:   "",
			stderr:   "",
			exitCode: 0,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectOutput(tt.stdout, tt.stderr, tt.exitCode)
			if got != tt.want {
				t.Errorf("SelectOutput(%q, %q, %d) = %q, want %q",
					tt.stdout, tt.stderr, tt.exitCode, got, tt.want)
			}
		})
	}
}

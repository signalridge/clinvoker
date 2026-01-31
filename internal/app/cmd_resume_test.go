package app

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "just now - 0 seconds",
			duration: 0,
			want:     "just now",
		},
		{
			name:     "just now - 30 seconds",
			duration: 30 * time.Second,
			want:     "just now",
		},
		{
			name:     "just now - 59 seconds",
			duration: 59 * time.Second,
			want:     "just now",
		},
		{
			name:     "1 minute ago",
			duration: 1 * time.Minute,
			want:     "1m ago",
		},
		{
			name:     "5 minutes ago",
			duration: 5 * time.Minute,
			want:     "5m ago",
		},
		{
			name:     "30 minutes ago",
			duration: 30 * time.Minute,
			want:     "30m ago",
		},
		{
			name:     "59 minutes ago",
			duration: 59 * time.Minute,
			want:     "59m ago",
		},
		{
			name:     "1 hour ago",
			duration: 1 * time.Hour,
			want:     "1h ago",
		},
		{
			name:     "2 hours ago",
			duration: 2 * time.Hour,
			want:     "2h ago",
		},
		{
			name:     "12 hours ago",
			duration: 12 * time.Hour,
			want:     "12h ago",
		},
		{
			name:     "23 hours ago",
			duration: 23 * time.Hour,
			want:     "23h ago",
		},
		{
			name:     "1 day ago",
			duration: 24 * time.Hour,
			want:     "1d ago",
		},
		{
			name:     "3 days ago",
			duration: 3 * 24 * time.Hour,
			want:     "3d ago",
		},
		{
			name:     "6 days ago",
			duration: 6 * 24 * time.Hour,
			want:     "6d ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a time that is tt.duration ago from now
			testTime := time.Now().Add(-tt.duration)
			got := formatTimeAgo(testTime)
			if got != tt.want {
				t.Errorf("formatTimeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTimeAgo_DateFormat(t *testing.T) {
	// Test dates older than 7 days - should return YYYY-MM-DD format
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "8 days ago shows date",
			time: time.Now().Add(-8 * 24 * time.Hour),
			want: time.Now().Add(-8 * 24 * time.Hour).Format("2006-01-02"),
		},
		{
			name: "30 days ago shows date",
			time: time.Now().Add(-30 * 24 * time.Hour),
			want: time.Now().Add(-30 * 24 * time.Hour).Format("2006-01-02"),
		},
		{
			name: "1 year ago shows date",
			time: time.Now().Add(-365 * 24 * time.Hour),
			want: time.Now().Add(-365 * 24 * time.Hour).Format("2006-01-02"),
		},
		{
			name: "specific date format",
			time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			want: "2024-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimeAgo(tt.time)
			if got != tt.want {
				t.Errorf("formatTimeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTimeAgo_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		wantType string // "just now", "minutes", "hours", "days", "date"
	}{
		{
			name:     "exactly 1 minute boundary",
			duration: 1 * time.Minute,
			wantType: "minutes",
		},
		{
			name:     "just under 1 hour",
			duration: 59*time.Minute + 59*time.Second,
			wantType: "minutes",
		},
		{
			name:     "exactly 1 hour boundary",
			duration: 1 * time.Hour,
			wantType: "hours",
		},
		{
			name:     "just under 24 hours",
			duration: 23*time.Hour + 59*time.Minute,
			wantType: "hours",
		},
		{
			name:     "exactly 24 hours boundary",
			duration: 24 * time.Hour,
			wantType: "days",
		},
		{
			name:     "just under 7 days",
			duration: 6*24*time.Hour + 23*time.Hour,
			wantType: "days",
		},
		{
			name:     "exactly 7 days boundary",
			duration: 7 * 24 * time.Hour,
			wantType: "date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Now().Add(-tt.duration)
			got := formatTimeAgo(testTime)

			var isCorrectType bool
			switch tt.wantType {
			case "just now":
				isCorrectType = got == "just now"
			case "minutes":
				isCorrectType = got != "" && got[len(got)-1] == 'o' && got[len(got)-5:] == "m ago"
			case "hours":
				isCorrectType = got != "" && got[len(got)-1] == 'o' && got[len(got)-5:] == "h ago"
			case "days":
				isCorrectType = got != "" && got[len(got)-1] == 'o' && got[len(got)-5:] == "d ago"
			case "date":
				// Date format: YYYY-MM-DD (10 characters)
				isCorrectType = len(got) == 10 && got[4] == '-' && got[7] == '-'
			}

			if !isCorrectType {
				t.Errorf("formatTimeAgo() = %q, expected type %q", got, tt.wantType)
			}
		})
	}
}

func TestResumeCmd_Structure(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{
			name:  "Use field is set correctly",
			field: "Use",
			want:  "resume [session-id] [prompt]",
		},
		{
			name:  "Short field is set correctly",
			field: "Short",
			want:  "Resume a previous session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.field {
			case "Use":
				got = resumeCmd.Use
			case "Short":
				got = resumeCmd.Short
			}

			if got != tt.want {
				t.Errorf("resumeCmd.%s = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestResumeCmd_IsInitialized(t *testing.T) {
	if resumeCmd == nil {
		t.Fatal("resumeCmd should not be nil")
	}

	if resumeCmd.RunE == nil {
		t.Error("resumeCmd.RunE should not be nil")
	}
}

func TestResumeCmd_HasLongDescription(t *testing.T) {
	if resumeCmd.Long == "" {
		t.Error("resumeCmd.Long should not be empty")
	}
}

func TestResumeCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args succeeds",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg succeeds (session-id)",
			args:    []string{"abc123"},
			wantErr: false,
		},
		{
			name:    "two args succeeds (session-id and prompt)",
			args:    []string{"abc123", "continue working"},
			wantErr: false,
		},
		{
			name:    "three args fails",
			args:    []string{"abc123", "prompt", "extra"},
			wantErr: true,
		},
		{
			name:    "four args fails",
			args:    []string{"one", "two", "three", "four"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cobra.MaximumNArgs(2)(resumeCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args validation with %v: error = %v, wantErr = %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestResumeCmd_Flags(t *testing.T) {
	t.Run("--last flag exists", func(t *testing.T) {
		flag := resumeCmd.Flags().Lookup("last")
		if flag == nil {
			t.Fatal("--last flag should exist")
		}
		if flag.DefValue != "false" {
			t.Errorf("--last default = %q, want %q", flag.DefValue, "false")
		}
		if flag.Usage != "resume the most recent session" {
			t.Errorf("--last usage = %q, want %q", flag.Usage, "resume the most recent session")
		}
	})

	t.Run("--backend flag exists", func(t *testing.T) {
		flag := resumeCmd.Flags().Lookup("backend")
		if flag == nil {
			t.Fatal("--backend flag should exist")
		}
		if flag.DefValue != "" {
			t.Errorf("--backend default = %q, want %q", flag.DefValue, "")
		}
		if flag.Shorthand != "b" {
			t.Errorf("--backend shorthand = %q, want %q", flag.Shorthand, "b")
		}
		if flag.Usage != "filter sessions by backend" {
			t.Errorf("--backend usage = %q, want %q", flag.Usage, "filter sessions by backend")
		}
	})

	t.Run("--here flag exists", func(t *testing.T) {
		flag := resumeCmd.Flags().Lookup("here")
		if flag == nil {
			t.Fatal("--here flag should exist")
		}
		if flag.DefValue != "false" {
			t.Errorf("--here default = %q, want %q", flag.DefValue, "false")
		}
		if flag.Usage != "filter sessions by current working directory" {
			t.Errorf("--here usage = %q, want %q", flag.Usage, "filter sessions by current working directory")
		}
	})

	t.Run("--interactive flag exists", func(t *testing.T) {
		flag := resumeCmd.Flags().Lookup("interactive")
		if flag == nil {
			t.Fatal("--interactive flag should exist")
		}
		if flag.DefValue != "false" {
			t.Errorf("--interactive default = %q, want %q", flag.DefValue, "false")
		}
		if flag.Shorthand != "i" {
			t.Errorf("--interactive shorthand = %q, want %q", flag.Shorthand, "i")
		}
		if flag.Usage != "show interactive session picker" {
			t.Errorf("--interactive usage = %q, want %q", flag.Usage, "show interactive session picker")
		}
	})
}

func TestResumeCmd_FlagTypes(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "--last is bool",
			flagName: "last",
			wantType: "bool",
		},
		{
			name:     "--backend is string",
			flagName: "backend",
			wantType: "string",
		},
		{
			name:     "--here is bool",
			flagName: "here",
			wantType: "bool",
		},
		{
			name:     "--interactive is bool",
			flagName: "interactive",
			wantType: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := resumeCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("--%s flag should exist", tt.flagName)
			}
			if flag.Value.Type() != tt.wantType {
				t.Errorf("--%s type = %q, want %q", tt.flagName, flag.Value.Type(), tt.wantType)
			}
		})
	}
}

func TestResumeCmd_FlagShorthands(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{
			name:      "--backend has -b shorthand",
			flagName:  "backend",
			shorthand: "b",
		},
		{
			name:      "--interactive has -i shorthand",
			flagName:  "interactive",
			shorthand: "i",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := resumeCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("--%s flag should exist", tt.flagName)
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("--%s shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestResumeCmd_NoShorthandForSomeFlags(t *testing.T) {
	// Verify that --last and --here do not have shorthands
	tests := []struct {
		name     string
		flagName string
	}{
		{
			name:     "--last has no shorthand",
			flagName: "last",
		},
		{
			name:     "--here has no shorthand",
			flagName: "here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := resumeCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("--%s flag should exist", tt.flagName)
			}
			if flag.Shorthand != "" {
				t.Errorf("--%s should have no shorthand, got %q", tt.flagName, flag.Shorthand)
			}
		})
	}
}

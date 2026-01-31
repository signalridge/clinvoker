package app

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSessionsCmd_Structure(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{
			name:  "Use field is set correctly",
			field: "Use",
			want:  "sessions",
		},
		{
			name:  "Short field is set correctly",
			field: "Short",
			want:  "Manage sessions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.field {
			case "Use":
				got = sessionsCmd.Use
			case "Short":
				got = sessionsCmd.Short
			}

			if got != tt.want {
				t.Errorf("sessionsCmd.%s = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestSessionsListCmd_Structure(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{
			name:  "Use field is set correctly",
			field: "Use",
			want:  "list",
		},
		{
			name:  "Short field is set correctly",
			field: "Short",
			want:  "List all sessions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.field {
			case "Use":
				got = sessionsListCmd.Use
			case "Short":
				got = sessionsListCmd.Short
			}

			if got != tt.want {
				t.Errorf("sessionsListCmd.%s = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestSessionsListCmd_Flags(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{
			name:      "backend flag exists with shorthand",
			flagName:  "backend",
			shorthand: "b",
			defValue:  "",
		},
		{
			name:      "status flag exists without shorthand",
			flagName:  "status",
			shorthand: "",
			defValue:  "",
		},
		{
			name:      "limit flag exists with shorthand",
			flagName:  "limit",
			shorthand: "n",
			defValue:  "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := sessionsListCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default value = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestSessionsListCmd_HasRunE(t *testing.T) {
	if sessionsListCmd.RunE == nil {
		t.Error("sessionsListCmd.RunE should not be nil")
	}
}

func TestSessionsShowCmd_Structure(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{
			name:  "Use field is set correctly",
			field: "Use",
			want:  "show <session-id>",
		},
		{
			name:  "Short field is set correctly",
			field: "Short",
			want:  "Show session details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.field {
			case "Use":
				got = sessionsShowCmd.Use
			case "Short":
				got = sessionsShowCmd.Short
			}

			if got != tt.want {
				t.Errorf("sessionsShowCmd.%s = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestSessionsShowCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "exactly one arg succeeds",
			args:    []string{"session-id"},
			wantErr: false,
		},
		{
			name:    "no args fails",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "two args fails",
			args:    []string{"session-id", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cobra.ExactArgs(1)(sessionsShowCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args validation with %v: error = %v, wantErr = %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestSessionsShowCmd_HasRunE(t *testing.T) {
	if sessionsShowCmd.RunE == nil {
		t.Error("sessionsShowCmd.RunE should not be nil")
	}
}

func TestSessionsDeleteCmd_Structure(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{
			name:  "Use field is set correctly",
			field: "Use",
			want:  "delete <session-id>",
		},
		{
			name:  "Short field is set correctly",
			field: "Short",
			want:  "Delete a session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.field {
			case "Use":
				got = sessionsDeleteCmd.Use
			case "Short":
				got = sessionsDeleteCmd.Short
			}

			if got != tt.want {
				t.Errorf("sessionsDeleteCmd.%s = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestSessionsDeleteCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "exactly one arg succeeds",
			args:    []string{"session-id"},
			wantErr: false,
		},
		{
			name:    "no args fails",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "two args fails",
			args:    []string{"session-id", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cobra.ExactArgs(1)(sessionsDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args validation with %v: error = %v, wantErr = %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestSessionsDeleteCmd_HasRunE(t *testing.T) {
	if sessionsDeleteCmd.RunE == nil {
		t.Error("sessionsDeleteCmd.RunE should not be nil")
	}
}

func TestSessionsCleanCmd_Structure(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{
			name:  "Use field is set correctly",
			field: "Use",
			want:  "clean",
		},
		{
			name:  "Short field is set correctly",
			field: "Short",
			want:  "Clean up old sessions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.field {
			case "Use":
				got = sessionsCleanCmd.Use
			case "Short":
				got = sessionsCleanCmd.Short
			}

			if got != tt.want {
				t.Errorf("sessionsCleanCmd.%s = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestSessionsCleanCmd_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{
			name:     "older-than flag exists",
			flagName: "older-than",
			defValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := sessionsCleanCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default value = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestSessionsCleanCmd_HasRunE(t *testing.T) {
	if sessionsCleanCmd.RunE == nil {
		t.Error("sessionsCleanCmd.RunE should not be nil")
	}
}

func TestSessionsCmd_HasSubcommands(t *testing.T) {
	tests := []struct {
		name        string
		subcommand  *cobra.Command
		wantPresent bool
	}{
		{
			name:        "sessionsListCmd is added",
			subcommand:  sessionsListCmd,
			wantPresent: true,
		},
		{
			name:        "sessionsShowCmd is added",
			subcommand:  sessionsShowCmd,
			wantPresent: true,
		},
		{
			name:        "sessionsDeleteCmd is added",
			subcommand:  sessionsDeleteCmd,
			wantPresent: true,
		},
		{
			name:        "sessionsCleanCmd is added",
			subcommand:  sessionsCleanCmd,
			wantPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := sessionsCmd.Commands()
			found := false
			for _, cmd := range commands {
				if cmd == tt.subcommand {
					found = true
					break
				}
			}
			if found != tt.wantPresent {
				t.Errorf("subcommand present = %v, want %v", found, tt.wantPresent)
			}
		})
	}
}

func TestSessionsCmd_SubcommandCount(t *testing.T) {
	expectedCount := 4 // list, show, delete, clean
	commands := sessionsCmd.Commands()
	if len(commands) != expectedCount {
		t.Errorf("sessionsCmd has %d subcommands, want %d", len(commands), expectedCount)
	}
}

func TestSessionsCmd_IsInitialized(t *testing.T) {
	if sessionsCmd == nil {
		t.Fatal("sessionsCmd should not be nil")
	}
}

func TestSessionsListCmd_IsInitialized(t *testing.T) {
	if sessionsListCmd == nil {
		t.Fatal("sessionsListCmd should not be nil")
	}
}

func TestSessionsShowCmd_IsInitialized(t *testing.T) {
	if sessionsShowCmd == nil {
		t.Fatal("sessionsShowCmd should not be nil")
	}
}

func TestSessionsDeleteCmd_IsInitialized(t *testing.T) {
	if sessionsDeleteCmd == nil {
		t.Fatal("sessionsDeleteCmd should not be nil")
	}
}

func TestSessionsCleanCmd_IsInitialized(t *testing.T) {
	if sessionsCleanCmd == nil {
		t.Fatal("sessionsCleanCmd should not be nil")
	}
}

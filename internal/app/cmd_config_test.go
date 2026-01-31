package app

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestConfigCmd_Structure(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Use field",
			want: "config",
		},
		{
			name: "Short field",
			want: "Manage configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.name {
			case "Use field":
				got = configCmd.Use
			case "Short field":
				got = configCmd.Short
			}
			if got != tt.want {
				t.Errorf("configCmd.%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestConfigShowCmd_Structure(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Use field",
			want: "show",
		},
		{
			name: "Short field",
			want: "Show current configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.name {
			case "Use field":
				got = configShowCmd.Use
			case "Short field":
				got = configShowCmd.Short
			}
			if got != tt.want {
				t.Errorf("configShowCmd.%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestConfigSetCmd_Structure(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Use field",
			want: "set <key> <value>",
		},
		{
			name: "Short field",
			want: "Set a configuration value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.name {
			case "Use field":
				got = configSetCmd.Use
			case "Short field":
				got = configSetCmd.Short
			}
			if got != tt.want {
				t.Errorf("configSetCmd.%s = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestConfigSetCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "exactly two args succeeds",
			args:    []string{"key", "value"},
			wantErr: false,
		},
		{
			name:    "no args fails",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg fails",
			args:    []string{"key"},
			wantErr: true,
		},
		{
			name:    "three args fails",
			args:    []string{"key", "value", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cobra.ExactArgs(2)(configSetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args validation with %v: error = %v, wantErr = %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestConfigCmd_HasSubcommands(t *testing.T) {
	tests := []struct {
		name        string
		subcommand  *cobra.Command
		wantPresent bool
	}{
		{
			name:        "configShowCmd is added",
			subcommand:  configShowCmd,
			wantPresent: true,
		},
		{
			name:        "configSetCmd is added",
			subcommand:  configSetCmd,
			wantPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := configCmd.Commands()
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

func TestConfigShowCmd_HasRunE(t *testing.T) {
	if configShowCmd.RunE == nil {
		t.Error("configShowCmd.RunE should not be nil")
	}
}

func TestConfigSetCmd_HasRunE(t *testing.T) {
	if configSetCmd.RunE == nil {
		t.Error("configSetCmd.RunE should not be nil")
	}
}

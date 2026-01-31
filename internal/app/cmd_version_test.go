package app

import "testing"

func TestVersionCmd_Structure(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{
			name:  "Use field is set correctly",
			field: "Use",
			want:  "version",
		},
		{
			name:  "Short field is set correctly",
			field: "Short",
			want:  "Print version information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.field {
			case "Use":
				got = versionCmd.Use
			case "Short":
				got = versionCmd.Short
			}

			if got != tt.want {
				t.Errorf("versionCmd.%s = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestVersionCmd_IsInitialized(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}

	if versionCmd.Run == nil {
		t.Error("versionCmd.Run should not be nil")
	}
}

package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateWorkDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "workdir-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file for testing non-directory case
	tmpFile := filepath.Join(tmpDir, "testfile")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Create a subdirectory for testing path resolution
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	tests := []struct {
		name    string
		workDir string
		wantErr bool
	}{
		{
			name:    "empty workDir is allowed",
			workDir: "",
			wantErr: false,
		},
		{
			name:    "valid existing directory",
			workDir: tmpDir,
			wantErr: false,
		},
		{
			name:    "path with .. resolving to valid directory is allowed",
			workDir: filepath.Join(subDir, ".."), // resolves to tmpDir
			wantErr: false,
		},
		{
			name:    "path with .. resolving to non-existent is rejected",
			workDir: filepath.Join(subDir, "nonexistent_path_xyz"), // doesn't exist
			wantErr: true,
		},
		{
			name:    "relative path not allowed",
			workDir: "relative/path",
			wantErr: true,
		},
		{
			name:    "non-existent directory",
			workDir: "/nonexistent/directory/path",
			wantErr: true,
		},
		{
			name:    "file instead of directory",
			workDir: tmpFile,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorkDir(tt.workDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWorkDir(%q) error = %v, wantErr %v", tt.workDir, err, tt.wantErr)
			}
		})
	}
}

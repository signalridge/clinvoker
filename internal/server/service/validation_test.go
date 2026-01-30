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

func TestValidateWorkDirWithConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "workdir-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks for comparison
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	tests := []struct {
		name           string
		workDir        string
		allowedPrefix  []string
		blockedPrefix  []string
		wantErr        bool
		wantErrContain string
	}{
		{
			name:          "empty workDir is always allowed",
			workDir:       "",
			allowedPrefix: []string{"/allowed"},
			blockedPrefix: []string{"/blocked"},
			wantErr:       false,
		},
		{
			name:          "valid dir with no restrictions",
			workDir:       tmpDir,
			allowedPrefix: nil,
			blockedPrefix: nil,
			wantErr:       false,
		},
		{
			name:          "valid dir matching allowed prefix",
			workDir:       tmpDir,
			allowedPrefix: []string{filepath.Dir(tmpDir)},
			blockedPrefix: nil,
			wantErr:       false,
		},
		{
			name:           "valid dir not matching allowed prefix",
			workDir:        tmpDir,
			allowedPrefix:  []string{"/nonexistent/allowed"},
			blockedPrefix:  nil,
			wantErr:        true,
			wantErrContain: "not in allowed paths",
		},
		{
			name:           "valid dir matching blocked prefix",
			workDir:        tmpDir,
			allowedPrefix:  nil,
			blockedPrefix:  []string{filepath.Dir(tmpDir)},
			wantErr:        true,
			wantErrContain: "blocked path",
		},
		{
			name:           "custom blocked path blocks valid directory",
			workDir:        tmpDir,
			allowedPrefix:  nil,
			blockedPrefix:  []string{tmpDir}, // explicitly block the test dir
			wantErr:        true,
			wantErrContain: "blocked path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorkDirWithConfig(tt.workDir, tt.allowedPrefix, tt.blockedPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWorkDirWithConfig(%q, %v, %v) error = %v, wantErr %v",
					tt.workDir, tt.allowedPrefix, tt.blockedPrefix, err, tt.wantErr)
			}
			if tt.wantErrContain != "" && err != nil {
				if !contains(err.Error(), tt.wantErrContain) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErrContain)
				}
			}
		})
	}
}

func TestDefaultBlockedWorkDirPrefixes(t *testing.T) {
	// Verify that all default blocked prefixes are present
	expectedBlocked := []string{
		"/etc", "/var/run", "/root", "/sys", "/proc",
		"/dev", "/usr/bin", "/sbin", "/boot",
	}

	for _, prefix := range expectedBlocked {
		found := false
		for _, blocked := range defaultBlockedWorkDirPrefixes {
			if blocked == prefix {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to be in defaultBlockedWorkDirPrefixes", prefix)
		}
	}
}

func TestHasPathPrefix(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		prefix   string
		expected bool
	}{
		{
			name:     "exact match",
			path:     "/etc",
			prefix:   "/etc",
			expected: true,
		},
		{
			name:     "child path matches",
			path:     "/etc/passwd",
			prefix:   "/etc",
			expected: true,
		},
		{
			name:     "nested child path matches",
			path:     "/etc/nginx/nginx.conf",
			prefix:   "/etc",
			expected: true,
		},
		{
			name:     "similar prefix but different path should NOT match (boundary bypass fix)",
			path:     "/etcfoo",
			prefix:   "/etc",
			expected: false,
		},
		{
			name:     "similar prefix with suffix should NOT match",
			path:     "/etc_backup",
			prefix:   "/etc",
			expected: false,
		},
		{
			name:     "unrelated path should NOT match",
			path:     "/home/user",
			prefix:   "/etc",
			expected: false,
		},
		{
			name:     "prefix with trailing slash - exact match",
			path:     "/etc",
			prefix:   "/etc/",
			expected: false, // /etc does not start with /etc/
		},
		{
			name:     "prefix with trailing slash - child matches",
			path:     "/etc/passwd",
			prefix:   "/etc/",
			expected: true,
		},
		{
			name:     "root path",
			path:     "/",
			prefix:   "/",
			expected: true,
		},
		{
			name:     "everything under root matches",
			path:     "/anything",
			prefix:   "/",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPathPrefix(tt.path, tt.prefix)
			if result != tt.expected {
				t.Errorf("hasPathPrefix(%q, %q) = %v, want %v", tt.path, tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestPathPrefixBoundaryBypass(t *testing.T) {
	// This test specifically verifies that the boundary bypass bug is fixed.
	// Previously, using strings.HasPrefix("/etcfoo", "/etc") would return true,
	// allowing a path like "/etcfoo" to be considered as matching the blocked "/etc" prefix.
	// The fix ensures proper path boundary checking.

	// Create temp directories to test with
	tmpDir, err := os.MkdirTemp("", "boundary-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory that has a similar name to a blocked prefix
	// e.g., if /blocked is blocked, /blockedfoo should NOT be blocked
	blockedPrefix := filepath.Join(tmpDir, "blocked")
	similarPath := filepath.Join(tmpDir, "blockedfoo")
	childPath := filepath.Join(tmpDir, "blocked", "child")

	if err := os.MkdirAll(blockedPrefix, 0755); err != nil {
		t.Fatalf("failed to create blocked dir: %v", err)
	}
	if err := os.MkdirAll(similarPath, 0755); err != nil {
		t.Fatalf("failed to create similar dir: %v", err)
	}
	if err := os.MkdirAll(childPath, 0755); err != nil {
		t.Fatalf("failed to create child dir: %v", err)
	}

	// Resolve symlinks
	blockedPrefix, _ = filepath.EvalSymlinks(blockedPrefix)
	similarPath, _ = filepath.EvalSymlinks(similarPath)
	childPath, _ = filepath.EvalSymlinks(childPath)

	// Test with explicit blocked prefix
	blockList := []string{blockedPrefix}

	// The exact blocked path should be blocked
	err = validateWorkDirWithConfig(blockedPrefix, nil, blockList)
	if err == nil {
		t.Error("exact blocked path should be rejected")
	}

	// Child of blocked path should be blocked
	err = validateWorkDirWithConfig(childPath, nil, blockList)
	if err == nil {
		t.Error("child of blocked path should be rejected")
	}

	// Similar path (blockedfoo vs blocked) should NOT be blocked
	err = validateWorkDirWithConfig(similarPath, nil, blockList)
	if err != nil {
		t.Errorf("similar path %q should NOT be blocked by %q: %v", similarPath, blockedPrefix, err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

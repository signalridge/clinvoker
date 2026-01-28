// Package util provides shared utility functions across the application.
package util

// SelectOutput determines which output stream to use based on exit code.
// If exit code is non-zero and stderr has content, prefer stderr (likely error message).
// Otherwise use stdout, falling back to stderr if stdout is empty.
func SelectOutput(stdout, stderr string, exitCode int) string {
	if exitCode != 0 && stderr != "" {
		return stderr
	}
	if stdout == "" {
		return stderr
	}
	return stdout
}

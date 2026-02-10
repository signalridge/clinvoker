package mcp

import (
	"testing"

	"github.com/signalridge/clinvoker/internal/config"
)

func setTempHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	config.Reset()
}

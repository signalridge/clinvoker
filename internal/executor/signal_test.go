package executor

import (
	"os"
	"testing"
)

func TestNewSignalHandler(t *testing.T) {
	h := NewSignalHandler(nil, nil)

	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.sigChan == nil {
		t.Error("expected non-nil sigChan")
	}
	if h.done == nil {
		t.Error("expected non-nil done")
	}
}

func TestSignalHandler_SetProcess(t *testing.T) {
	h := NewSignalHandler(nil, nil)

	// Create a dummy process (we won't actually use it)
	proc := &os.Process{Pid: 12345}
	h.SetProcess(proc)

	h.mu.Lock()
	if h.process != proc {
		t.Error("process not set correctly")
	}
	h.mu.Unlock()
}

func TestSignalHandler_StartStop(t *testing.T) {
	h := NewSignalHandler(nil, nil)

	// Start should work
	h.Start()
	if !h.started {
		t.Error("handler should be started")
	}

	// Start again should be no-op
	h.Start()

	// Stop should work
	h.Stop()

	// Stop again should be no-op
	h.Stop()
}

func TestSignalHandler_StartIdempotent(t *testing.T) {
	h := NewSignalHandler(nil, nil)

	h.Start()
	h.Start()
	h.Start()

	// Should not panic or have issues
	h.Stop()
}

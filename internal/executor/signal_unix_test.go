//go:build !windows

package executor

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func startSleepProcess(t *testing.T) (*exec.Cmd, chan error) {
	t.Helper()

	// Run sleep directly so forwarded signals target the long-running process
	// itself; shell wrappers can absorb/transform signals on some platforms.
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start sleep process: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	return cmd, done
}

func ensureProcessStopped(t *testing.T, cmd *exec.Cmd, done chan error) {
	t.Helper()

	if cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGKILL)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for process cleanup")
	}
}

func TestSignalHandler_ForwardSignalToProcess(t *testing.T) {
	cmd, done := startSleepProcess(t)
	defer func() {
		if done != nil {
			ensureProcessStopped(t, cmd, done)
		}
	}()

	h := NewSignalHandler(cmd.Process, nil)
	h.forwardSignal(syscall.SIGINT)

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected process to exit due to forwarded signal")
		}
		done = nil
	case <-time.After(2 * time.Second):
		t.Fatal("process did not exit after forwarded signal")
	}
}

func TestSignalHandler_HandleTermSignal_StopsOnDone(t *testing.T) {
	cmd, done := startSleepProcess(t)
	defer ensureProcessStopped(t, cmd, done)

	h := NewSignalHandler(cmd.Process, nil)
	close(h.done)
	h.handleTermSignal(syscall.SIGTERM, cmd.Process)

	time.Sleep(100 * time.Millisecond)
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatalf("process should still be running after done is closed, got %v", err)
	}
}

func TestSignalHandler_HandleResize_ReturnsOnDone(t *testing.T) {
	h := NewSignalHandler(nil, nil)
	resizeChan := make(chan os.Signal, 1)

	h.HandleResize(resizeChan)
	resizeChan <- syscall.SIGWINCH
	close(h.done)

	time.Sleep(50 * time.Millisecond)
}

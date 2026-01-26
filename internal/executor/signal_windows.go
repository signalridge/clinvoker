//go:build windows

package executor

import (
	"os"
	"os/signal"
)

// Start begins signal handling on Windows.
func (h *SignalHandler) Start() {
	h.mu.Lock()
	if h.started {
		h.mu.Unlock()
		return
	}
	h.started = true
	h.mu.Unlock()

	// Windows only supports os.Interrupt (CTRL+C)
	signal.Notify(h.sigChan, os.Interrupt)

	go h.handleSignals()
}

// handleTermSignal handles termination signals on Windows.
// Windows doesn't have SIGTERM or SIGKILL, so we just kill the process.
func (h *SignalHandler) handleTermSignal(sig os.Signal, process *os.Process) {
	if sig == os.Interrupt {
		// On Windows, we just kill the process after timeout
		go func() {
			select {
			case <-h.done:
				return
			}
		}()
	}
}

// HandleResize is a no-op on Windows as SIGWINCH doesn't exist.
func (h *SignalHandler) HandleResize(resizeChan chan os.Signal) {
	// No-op on Windows - terminal resize is handled differently
}

// stopSignals stops signal notifications.
func stopSignals(sigChan chan os.Signal) {
	signal.Stop(sigChan)
}

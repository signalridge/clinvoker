//go:build !windows

package executor

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Start begins signal handling on Unix systems.
func (h *SignalHandler) Start() {
	h.mu.Lock()
	if h.started {
		h.mu.Unlock()
		return
	}
	h.started = true
	h.mu.Unlock()

	signal.Notify(h.sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go h.handleSignals()
}

// handleTermSignal handles SIGTERM on Unix by implementing graceful shutdown.
func (h *SignalHandler) handleTermSignal(sig os.Signal, process *os.Process) {
	if sig == syscall.SIGTERM {
		go h.gracefulShutdown(process)
	}
}

func (h *SignalHandler) gracefulShutdown(process *os.Process) {
	timer := time.NewTimer(GracefulShutdownTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		// Timeout: send SIGKILL
		process.Signal(syscall.SIGKILL)
	case <-h.done:
		// Process already exited
	}
}

// HandleResize handles terminal resize signals on Unix.
func (h *SignalHandler) HandleResize(resizeChan chan os.Signal) {
	signal.Notify(resizeChan, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-resizeChan:
				if h.pty != nil {
					// PTY resize is handled by the pty package
				}
			case <-h.done:
				signal.Stop(resizeChan)
				return
			}
		}
	}()
}

// stopSignals stops signal notifications.
func stopSignals(sigChan chan os.Signal) {
	signal.Stop(sigChan)
}

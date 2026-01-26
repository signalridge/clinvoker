package executor

import (
	"os"
	"sync"
	"time"
)

const (
	// GracefulShutdownTimeout is the time to wait before sending SIGKILL.
	GracefulShutdownTimeout = 5 * time.Second
)

// SignalHandler manages signal forwarding to child processes.
type SignalHandler struct {
	process *os.Process
	pty     *os.File
	sigChan chan os.Signal
	done    chan struct{}
	mu      sync.Mutex
	started bool
}

// NewSignalHandler creates a new signal handler.
func NewSignalHandler(process *os.Process, pty *os.File) *SignalHandler {
	return &SignalHandler{
		process: process,
		pty:     pty,
		sigChan: make(chan os.Signal, 1),
		done:    make(chan struct{}),
	}
}

// SetProcess sets the process to forward signals to.
func (h *SignalHandler) SetProcess(p *os.Process) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.process = p
}

// Stop stops signal handling.
func (h *SignalHandler) Stop() {
	h.mu.Lock()
	if !h.started {
		h.mu.Unlock()
		return
	}
	h.started = false
	h.mu.Unlock()

	stopSignals(h.sigChan)
	close(h.done)
}

func (h *SignalHandler) handleSignals() {
	for {
		select {
		case sig := <-h.sigChan:
			h.forwardSignal(sig)
		case <-h.done:
			return
		}
	}
}

func (h *SignalHandler) forwardSignal(sig os.Signal) {
	h.mu.Lock()
	process := h.process
	h.mu.Unlock()

	if process == nil {
		return
	}

	// Forward the signal to the child process
	if err := process.Signal(sig); err != nil {
		return
	}

	// Platform-specific handling for SIGTERM
	h.handleTermSignal(sig, process)
}

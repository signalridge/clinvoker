// Package resilience provides fault tolerance patterns like circuit breakers.
package resilience

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	// StateClosed means the circuit is operating normally.
	StateClosed CircuitState = iota
	// StateOpen means the circuit has tripped and requests are blocked.
	StateOpen
	// StateHalfOpen means the circuit is testing if it can close.
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Common errors returned by the circuit breaker.
var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements the circuit breaker pattern.
// It tracks failures and temporarily blocks requests when too many failures occur.
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	failureThreshold int           // Number of failures to trip the circuit
	successThreshold int           // Number of successes to close the circuit in half-open state
	timeout          time.Duration // Time to wait before transitioning from open to half-open
	halfOpenMaxCalls int           // Maximum concurrent calls allowed in half-open state

	// State
	state             CircuitState
	failureCount      int
	successCount      int
	halfOpenCallCount int
	lastFailureTime   time.Time
	lastStateChange   time.Time

	// Metrics
	totalFailures  int64
	totalSuccesses int64
	totalRejected  int64

	// Callbacks (optional)
	onStateChange func(name string, from, to CircuitState)
	name          string
}

// CircuitBreakerConfig contains configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// Name is an identifier for this circuit breaker (for logging/metrics).
	Name string

	// FailureThreshold is the number of consecutive failures to trip the circuit.
	// Default: 5
	FailureThreshold int

	// SuccessThreshold is the number of successes needed to close the circuit.
	// Default: 2
	SuccessThreshold int

	// Timeout is how long to wait before allowing a test request.
	// Default: 30 seconds
	Timeout time.Duration

	// HalfOpenMaxCalls is the max concurrent calls in half-open state.
	// Default: 1
	HalfOpenMaxCalls int

	// OnStateChange is called when the circuit breaker changes state.
	OnStateChange func(name string, from, to CircuitState)
}

// DefaultConfig returns a default circuit breaker configuration.
func DefaultConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:             name,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		HalfOpenMaxCalls: 1,
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 2
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.HalfOpenMaxCalls <= 0 {
		cfg.HalfOpenMaxCalls = 1
	}

	return &CircuitBreaker{
		name:             cfg.Name,
		failureThreshold: cfg.FailureThreshold,
		successThreshold: cfg.SuccessThreshold,
		timeout:          cfg.Timeout,
		halfOpenMaxCalls: cfg.HalfOpenMaxCalls,
		state:            StateClosed,
		lastStateChange:  time.Now(),
		onStateChange:    cfg.OnStateChange,
	}
}

// Allow checks if a request should be allowed to proceed.
// Returns true if the request is allowed, false if it should be rejected.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has elapsed
		if now.Sub(cb.lastStateChange) > cb.timeout {
			cb.transitionTo(StateHalfOpen)
			cb.halfOpenCallCount = 1
			return true
		}
		cb.totalRejected++
		return false

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenCallCount < cb.halfOpenMaxCalls {
			cb.halfOpenCallCount++
			return true
		}
		cb.totalRejected++
		return false

	default:
		return false
	}
}

// RecordSuccess records a successful call.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalSuccesses++

	switch cb.state {
	case StateClosed:
		cb.failureCount = 0

	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.transitionTo(StateClosed)
		}

	case StateOpen:
		// Shouldn't happen, but reset if it does
		cb.transitionTo(StateClosed)
	}
}

// RecordFailure records a failed call.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalFailures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.transitionTo(StateOpen)
		}

	case StateHalfOpen:
		// Single failure in half-open state trips back to open
		cb.transitionTo(StateOpen)

	case StateOpen:
		// Already open, nothing to do
	}
}

// transitionTo changes the circuit breaker state.
// Must be called with lock held.
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenCallCount = 0

	if cb.onStateChange != nil && oldState != newState {
		// Call callback outside the lock to avoid deadlocks
		go cb.onStateChange(cb.name, oldState, newState)
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns statistics about the circuit breaker.
type Stats struct {
	State          CircuitState
	FailureCount   int
	TotalFailures  int64
	TotalSuccesses int64
	TotalRejected  int64
	LastFailure    time.Time
	LastChange     time.Time
}

// Stats returns current statistics.
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Stats{
		State:          cb.state,
		FailureCount:   cb.failureCount,
		TotalFailures:  cb.totalFailures,
		TotalSuccesses: cb.totalSuccesses,
		TotalRejected:  cb.totalRejected,
		LastFailure:    cb.lastFailureTime,
		LastChange:     cb.lastStateChange,
	}
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenCallCount = 0
	cb.lastStateChange = time.Now()
}

// Execute runs the given function with circuit breaker protection.
// Returns ErrCircuitOpen if the circuit is open and the request is rejected.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.Allow() {
		return ErrCircuitOpen
	}

	err := fn()
	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// CircuitBreakerRegistry manages circuit breakers for multiple backends.
type CircuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	config   CircuitBreakerConfig
}

// NewCircuitBreakerRegistry creates a new registry with default configuration.
func NewCircuitBreakerRegistry(defaultConfig CircuitBreakerConfig) *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
		config:   defaultConfig,
	}
}

// Get returns the circuit breaker for the given name, creating one if needed.
func (r *CircuitBreakerRegistry) Get(name string) *CircuitBreaker {
	r.mu.RLock()
	cb, ok := r.breakers[name]
	r.mu.RUnlock()

	if ok {
		return cb
	}

	// Create new circuit breaker
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, ok = r.breakers[name]; ok {
		return cb
	}

	cfg := r.config
	cfg.Name = name
	cb = NewCircuitBreaker(cfg)
	r.breakers[name] = cb

	return cb
}

// AllStats returns stats for all circuit breakers.
func (r *CircuitBreakerRegistry) AllStats() map[string]Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Stats, len(r.breakers))
	for name, cb := range r.breakers {
		result[name] = cb.Stats()
	}
	return result
}

// ResetAll resets all circuit breakers.
func (r *CircuitBreakerRegistry) ResetAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cb := range r.breakers {
		cb.Reset()
	}
}

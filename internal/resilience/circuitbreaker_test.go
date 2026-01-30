package resilience

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker(DefaultConfig("test"))

	// Should allow requests when closed
	if !cb.Allow() {
		t.Error("expected Allow() to return true in closed state")
	}

	// State should be closed
	if cb.State() != StateClosed {
		t.Errorf("expected state %v, got %v", StateClosed, cb.State())
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 3
	cb := NewCircuitBreaker(cfg)

	// Record failures up to threshold
	for i := 0; i < 3; i++ {
		if !cb.Allow() {
			t.Errorf("iteration %d: expected Allow() to return true", i)
		}
		cb.RecordFailure()
	}

	// Circuit should now be open
	if cb.State() != StateOpen {
		t.Errorf("expected state %v, got %v", StateOpen, cb.State())
	}

	// Should reject requests
	if cb.Allow() {
		t.Error("expected Allow() to return false in open state")
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 3
	cb := NewCircuitBreaker(cfg)

	// Record 2 failures (below threshold)
	cb.Allow()
	cb.RecordFailure()
	cb.Allow()
	cb.RecordFailure()

	// Record success
	cb.Allow()
	cb.RecordSuccess()

	// Record 2 more failures
	cb.Allow()
	cb.RecordFailure()
	cb.Allow()
	cb.RecordFailure()

	// Circuit should still be closed (success reset the counter)
	if cb.State() != StateClosed {
		t.Errorf("expected state %v, got %v", StateClosed, cb.State())
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 1
	cfg.Timeout = 10 * time.Millisecond
	cb := NewCircuitBreaker(cfg)

	// Trip the circuit
	cb.Allow()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("expected state %v, got %v", StateOpen, cb.State())
	}

	// Wait for timeout
	time.Sleep(15 * time.Millisecond)

	// Should transition to half-open on Allow()
	if !cb.Allow() {
		t.Error("expected Allow() to return true after timeout")
	}

	if cb.State() != StateHalfOpen {
		t.Errorf("expected state %v, got %v", StateHalfOpen, cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToClosedOnSuccess(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 1
	cfg.SuccessThreshold = 2
	cfg.Timeout = 10 * time.Millisecond
	cb := NewCircuitBreaker(cfg)

	// Trip and wait
	cb.Allow()
	cb.RecordFailure()
	time.Sleep(15 * time.Millisecond)

	// First success in half-open
	cb.Allow()
	cb.RecordSuccess()
	if cb.State() != StateHalfOpen {
		t.Errorf("expected state %v after first success, got %v", StateHalfOpen, cb.State())
	}

	// Second success closes the circuit
	cb.Allow()
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Errorf("expected state %v after second success, got %v", StateClosed, cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 1
	cfg.Timeout = 10 * time.Millisecond
	cb := NewCircuitBreaker(cfg)

	// Trip and wait
	cb.Allow()
	cb.RecordFailure()
	time.Sleep(15 * time.Millisecond)

	// Get to half-open
	cb.Allow()

	// Fail in half-open
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("expected state %v after failure in half-open, got %v", StateOpen, cb.State())
	}
}

func TestCircuitBreaker_Execute(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 2
	cb := NewCircuitBreaker(cfg)

	// Successful execution
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Failed execution
	testErr := errors.New("test error")
	err = cb.Execute(func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("expected %v, got %v", testErr, err)
	}

	// Another failure trips the circuit
	err = cb.Execute(func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("expected %v, got %v", testErr, err)
	}

	// Circuit is now open
	err = cb.Execute(func() error {
		return nil
	})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 3
	cb := NewCircuitBreaker(cfg)

	// Record some activity
	cb.Allow()
	cb.RecordSuccess()
	cb.Allow()
	cb.RecordFailure()
	cb.Allow()
	cb.RecordSuccess()

	stats := cb.Stats()

	if stats.State != StateClosed {
		t.Errorf("expected state %v, got %v", StateClosed, stats.State)
	}
	if stats.TotalSuccesses != 2 {
		t.Errorf("expected 2 successes, got %d", stats.TotalSuccesses)
	}
	if stats.TotalFailures != 1 {
		t.Errorf("expected 1 failure, got %d", stats.TotalFailures)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 1
	cb := NewCircuitBreaker(cfg)

	// Trip the circuit
	cb.Allow()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("expected state %v, got %v", StateOpen, cb.State())
	}

	// Reset
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("expected state %v after reset, got %v", StateClosed, cb.State())
	}

	// Should allow requests
	if !cb.Allow() {
		t.Error("expected Allow() to return true after reset")
	}
}

func TestCircuitBreaker_Concurrent(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 100 // High threshold to avoid tripping
	cb := NewCircuitBreaker(cfg)

	var wg sync.WaitGroup
	concurrency := 50
	iterations := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if cb.Allow() {
					if j%2 == 0 {
						cb.RecordSuccess()
					} else {
						cb.RecordFailure()
					}
				}
			}
		}(i)
	}

	wg.Wait()

	stats := cb.Stats()
	total := stats.TotalSuccesses + stats.TotalFailures
	expected := int64(concurrency * iterations)

	if total != expected {
		t.Errorf("expected %d total calls, got %d", expected, total)
	}
}

func TestCircuitBreaker_HalfOpenMaxCalls(t *testing.T) {
	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 1
	cfg.Timeout = 10 * time.Millisecond
	cfg.HalfOpenMaxCalls = 2
	cb := NewCircuitBreaker(cfg)

	// Trip and wait
	cb.Allow()
	cb.RecordFailure()
	time.Sleep(15 * time.Millisecond)

	// First call should be allowed (transitions to half-open)
	if !cb.Allow() {
		t.Error("first call should be allowed")
	}

	// Second call should be allowed (within limit)
	if !cb.Allow() {
		t.Error("second call should be allowed")
	}

	// Third call should be rejected
	if cb.Allow() {
		t.Error("third call should be rejected")
	}
}

func TestCircuitBreakerRegistry(t *testing.T) {
	registry := NewCircuitBreakerRegistry(DefaultConfig(""))

	// Get creates new circuit breaker
	cb1 := registry.Get("backend1")
	if cb1 == nil {
		t.Fatal("expected non-nil circuit breaker")
	}

	// Get returns same circuit breaker
	cb2 := registry.Get("backend1")
	if cb1 != cb2 {
		t.Error("expected same circuit breaker instance")
	}

	// Different name creates different circuit breaker
	cb3 := registry.Get("backend2")
	if cb1 == cb3 {
		t.Error("expected different circuit breaker for different name")
	}

	// AllStats returns stats for all
	cb1.Allow()
	cb1.RecordFailure()

	stats := registry.AllStats()
	if len(stats) != 2 {
		t.Errorf("expected 2 stats, got %d", len(stats))
	}
	if stats["backend1"].TotalFailures != 1 {
		t.Errorf("expected 1 failure for backend1, got %d", stats["backend1"].TotalFailures)
	}
}

func TestCircuitBreakerRegistry_Concurrent(t *testing.T) {
	registry := NewCircuitBreakerRegistry(DefaultConfig(""))

	var wg sync.WaitGroup
	concurrency := 20

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cb := registry.Get("shared")
			cb.Allow()
			cb.RecordSuccess()
		}(i)
	}

	wg.Wait()

	stats := registry.AllStats()
	if stats["shared"].TotalSuccesses != int64(concurrency) {
		t.Errorf("expected %d successes, got %d", concurrency, stats["shared"].TotalSuccesses)
	}
}

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state CircuitState
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("CircuitState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	var changes []struct {
		name     string
		from, to CircuitState
	}
	var mu sync.Mutex

	cfg := DefaultConfig("test")
	cfg.FailureThreshold = 1
	cfg.Timeout = 10 * time.Millisecond
	cfg.OnStateChange = func(name string, from, to CircuitState) {
		mu.Lock()
		changes = append(changes, struct {
			name     string
			from, to CircuitState
		}{name, from, to})
		mu.Unlock()
	}
	cb := NewCircuitBreaker(cfg)

	// Trip the circuit
	cb.Allow()
	cb.RecordFailure()

	// Wait for callback
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	if len(changes) != 1 {
		t.Errorf("expected 1 state change, got %d", len(changes))
	}
	if len(changes) > 0 {
		if changes[0].from != StateClosed || changes[0].to != StateOpen {
			t.Errorf("expected closed->open, got %v->%v", changes[0].from, changes[0].to)
		}
	}
	mu.Unlock()
}

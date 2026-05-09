package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

const (
	StateClosed   = "closed"
	StateOpen     = "open"
	StateHalfOpen = "half-open"
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu          sync.Mutex
	maxFailures uint
	openTimeout time.Duration
	failures    uint
	state       string
	openedAt    time.Time
}

// New creates a CircuitBreaker with the given failure threshold and open timeout in seconds.
func New(maxFailures uint, openTimeoutSecs int) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		openTimeout: time.Duration(openTimeoutSecs) * time.Second,
		state:       StateClosed,
	}
}

// Allow checks whether a request is permitted.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case StateOpen:
		if time.Since(cb.openedAt) >= cb.openTimeout {
			cb.state = StateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

// RecordSuccess records a successful call, closing the circuit if half-open.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure records a failed call and may open the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		cb.openedAt = time.Now()
	}
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

// State returns the current state string.
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

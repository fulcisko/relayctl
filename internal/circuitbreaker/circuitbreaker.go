package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

// Breaker is a simple circuit breaker for a single backend.
type Breaker struct {
	mu           sync.Mutex
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	state        State
	openedAt     time.Time
}

// New creates a new Breaker with the given failure threshold and reset timeout.
func New(maxFailures int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Allow returns nil if the request is allowed, or ErrCircuitOpen if the
// circuit is open. It transitions to HalfOpen after the reset timeout.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateOpen:
		if time.Since(b.openedAt) >= b.resetTimeout {
			b.state = StateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

// RecordSuccess resets the breaker on a successful call.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure increments the failure counter and opens the circuit if the
// threshold is reached.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.maxFailures {
		b.state = StateOpen
		b.openedAt = time.Now()
	}
}

// State returns the current state of the breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

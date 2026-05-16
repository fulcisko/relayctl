package upstream

import (
	"fmt"
	"sync"
)

// CircuitPolicy defines per-backend circuit breaker thresholds.
type CircuitPolicy struct {
	FailureThreshold int     `json:"failure_threshold"`
	SuccessThreshold int     `json:"success_threshold"`
	TimeoutSeconds   float64 `json:"timeout_seconds"`
}

// CircuitPolicyRegistry stores per-backend circuit policies.
type CircuitPolicyRegistry struct {
	mu       sync.RWMutex
	policies map[string]CircuitPolicy
}

// NewCircuitPolicyRegistry returns an empty registry.
func NewCircuitPolicyRegistry() *CircuitPolicyRegistry {
	return &CircuitPolicyRegistry{
		policies: make(map[string]CircuitPolicy),
	}
}

// Set stores a policy for the given backend. Returns an error if backend is empty
// or thresholds are invalid.
func (r *CircuitPolicyRegistry) Set(backend string, p CircuitPolicy) error {
	if backend == "" {
		return fmt.Errorf("circuit_policy: backend must not be empty")
	}
	if p.FailureThreshold < 1 {
		return fmt.Errorf("circuit_policy: failure_threshold must be >= 1")
	}
	if p.SuccessThreshold < 1 {
		return fmt.Errorf("circuit_policy: success_threshold must be >= 1")
	}
	if p.TimeoutSeconds <= 0 {
		return fmt.Errorf("circuit_policy: timeout_seconds must be > 0")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.policies[backend] = p
	return nil
}

// Get returns the policy for a backend and whether it was found.
func (r *CircuitPolicyRegistry) Get(backend string) (CircuitPolicy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.policies[backend]
	return p, ok
}

// Delete removes the policy for a backend.
func (r *CircuitPolicyRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.policies, backend)
}

// Snapshot returns a copy of all current policies.
func (r *CircuitPolicyRegistry) Snapshot() map[string]CircuitPolicy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]CircuitPolicy, len(r.policies))
	for k, v := range r.policies {
		out[k] = v
	}
	return out
}

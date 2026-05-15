package upstream

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

// ErrNoTimeoutConfig is returned when no timeout is configured for a backend.
var ErrNoTimeoutConfig = errors.New("no timeout config for backend")

// TimeoutConfig holds per-backend timeout settings.
type TimeoutConfig struct {
	Dial           time.Duration `json:"dial_ms"`
	ResponseHeader time.Duration `json:"response_header_ms"`
	Idle           time.Duration `json:"idle_ms"`
}

// TimeoutRegistry stores per-backend timeout configurations and builds
// http.Transport instances with those timeouts applied.
type TimeoutRegistry struct {
	mu      sync.RWMutex
	configs map[string]TimeoutConfig
}

// NewTimeoutRegistry returns an empty TimeoutRegistry.
func NewTimeoutRegistry() *TimeoutRegistry {
	return &TimeoutRegistry{
		configs: make(map[string]TimeoutConfig),
	}
}

// Set stores a TimeoutConfig for the given backend URL.
func (r *TimeoutRegistry) Set(backend string, cfg TimeoutConfig) error {
	if backend == "" {
		return errors.New("backend must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[backend] = cfg
	return nil
}

// Get returns the TimeoutConfig for the given backend, or an error if absent.
func (r *TimeoutRegistry) Get(backend string) (TimeoutConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg, ok := r.configs[backend]
	if !ok {
		return TimeoutConfig{}, ErrNoTimeoutConfig
	}
	return cfg, nil
}

// Delete removes the timeout config for the given backend.
func (r *TimeoutRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.configs, backend)
}

// Transport returns an *http.Transport configured with the timeouts for the
// given backend. Falls back to http.DefaultTransport if not found.
func (r *TimeoutRegistry) Transport(backend string) http.RoundTripper {
	cfg, err := r.Get(backend)
	if err != nil {
		return http.DefaultTransport
	}
	return &http.Transport{
		DialContext:           (&net.Dialer{Timeout: cfg.Dial}).DialContext,
		ResponseHeaderTimeout: cfg.ResponseHeader,
		IdleConnTimeout:       cfg.Idle,
	}
}

// Snapshot returns a copy of all registered timeout configs.
func (r *TimeoutRegistry) Snapshot() map[string]TimeoutConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]TimeoutConfig, len(r.configs))
	for k, v := range r.configs {
		out[k] = v
	}
	return out
}

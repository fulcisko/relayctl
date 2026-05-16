package upstream

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
)

// TracingConfig holds per-backend tracing settings.
type TracingConfig struct {
	InjectRequestID bool   `json:"inject_request_id"`
	HeaderName      string `json:"header_name"`
}

// TracingRegistry stores tracing configuration per backend URL.
type TracingRegistry struct {
	mu      sync.RWMutex
	entries map[string]TracingConfig
}

// NewTracingRegistry returns an empty TracingRegistry.
func NewTracingRegistry() *TracingRegistry {
	return &TracingRegistry{entries: make(map[string]TracingConfig)}
}

// Set registers a TracingConfig for the given backend.
func (r *TracingRegistry) Set(backend string, cfg TracingConfig) error {
	if backend == "" {
		return errEmptyBackend
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-Request-ID"
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = cfg
	return nil
}

// Get returns the TracingConfig for the given backend, if present.
func (r *TracingRegistry) Get(backend string) (TracingConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg, ok := r.entries[backend]
	return cfg, ok
}

// Delete removes the tracing config for the given backend.
func (r *TracingRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// Snapshot returns a copy of all tracing entries.
func (r *TracingRegistry) Snapshot() map[string]TracingConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]TracingConfig, len(r.entries))
	for k, v := range r.entries {
		out[k] = v
	}
	return out
}

// Apply injects trace headers into req for the given backend.
// If a request-id header already exists it is preserved.
func (r *TracingRegistry) Apply(backend string, req *http.Request) {
	cfg, ok := r.Get(backend)
	if !ok || !cfg.InjectRequestID {
		return
	}
	if req.Header.Get(cfg.HeaderName) == "" {
		req.Header.Set(cfg.HeaderName, newRequestID())
	}
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

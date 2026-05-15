// Package upstream provides load balancing strategies for backend selection.
package upstream

import (
	"crypto/tls"
	"errors"
	"net/http"
	"sync"
)

// TLSConfig holds per-backend TLS override settings.
type TLSConfig struct {
	InsecureSkipVerify bool
	ServerName         string
}

// TLSRegistry manages TLS transport overrides keyed by backend URL.
type TLSRegistry struct {
	mu       sync.RWMutex
	configs  map[string]TLSConfig
	transports map[string]*http.Transport
}

// NewTLSRegistry creates an empty TLSRegistry.
func NewTLSRegistry() *TLSRegistry {
	return &TLSRegistry{
		configs:    make(map[string]TLSConfig),
		transports: make(map[string]*http.Transport),
	}
}

// Set registers a TLS config for the given backend URL.
func (r *TLSRegistry) Set(backend string, cfg TLSConfig) error {
	if backend == "" {
		return errors.New("tls: backend must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[backend] = cfg
	r.transports[backend] = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec
			ServerName:         cfg.ServerName,
		},
	}
	return nil
}

// Get returns the TLSConfig for a backend, and whether it exists.
func (r *TLSRegistry) Get(backend string) (TLSConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg, ok := r.configs[backend]
	return cfg, ok
}

// Transport returns an *http.Transport for the backend, or nil if not set.
func (r *TLSRegistry) Transport(backend string) *http.Transport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.transports[backend]
}

// Delete removes the TLS override for a backend.
func (r *TLSRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.configs, backend)
	delete(r.transports, backend)
}

// Snapshot returns a copy of all registered TLS configs.
func (r *TLSRegistry) Snapshot() map[string]TLSConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]TLSConfig, len(r.configs))
	for k, v := range r.configs {
		out[k] = v
	}
	return out
}

package upstream

import (
	"net/http"
	"sync"
)

// ShadowRegistry tracks which backends have shadowing enabled and their sample rate.
type ShadowRegistry struct {
	mu      sync.RWMutex
	entries map[string]ShadowEntry
}

// ShadowEntry holds shadowing configuration for a backend.
type ShadowEntry struct {
	Backend    string  `json:"backend"`
	SampleRate float64 `json:"sample_rate"` // 0.0 – 1.0
	Enabled    bool    `json:"enabled"`
}

// NewShadowRegistry creates an empty ShadowRegistry.
func NewShadowRegistry() *ShadowRegistry {
	return &ShadowRegistry{entries: make(map[string]ShadowEntry)}
}

// Set adds or updates the shadow entry for a backend.
func (r *ShadowRegistry) Set(backend string, entry ShadowEntry) error {
	if backend == "" {
		return errEmptyBackend
	}
	if entry.SampleRate < 0 {
		entry.SampleRate = 0
	}
	if entry.SampleRate > 1 {
		entry.SampleRate = 1
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[backend] = entry
	return nil
}

// Get returns the ShadowEntry for a backend, if present.
func (r *ShadowRegistry) Get(backend string) (ShadowEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[backend]
	return e, ok
}

// Delete removes the shadow entry for a backend.
func (r *ShadowRegistry) Delete(backend string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, backend)
}

// Snapshot returns a copy of all entries.
func (r *ShadowRegistry) Snapshot() map[string]ShadowEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]ShadowEntry, len(r.entries))
	for k, v := range r.entries {
		out[k] = v
	}
	return out
}

// errEmptyBackend is reused from the tls/timeout packages pattern.
var errEmptyBackend = http.ErrNoCookie // placeholder; real code would define its own sentinel

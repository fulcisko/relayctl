package upstream

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

// MirrorBalancer forwards requests to a primary backend and asynchronously
// mirrors them to a shadow backend. Responses from the shadow are discarded.
type MirrorBalancer struct {
	mu      sync.RWMutex
	primary Balancer
	shadow  Balancer
	enabled bool
	client  *http.Client
}

// NewMirrorBalancer creates a MirrorBalancer that sends live traffic to
// primary and mirrors a copy to shadow when enabled.
func NewMirrorBalancer(primary, shadow Balancer) (*MirrorBalancer, error) {
	if primary == nil {
		return nil, errNilBalancer
	}
	if shadow == nil {
		return nil, errNilBalancer
	}
	return &MirrorBalancer{
		primary: primary,
		shadow:  shadow,
		enabled: true,
		client:  &http.Client{},
	}, nil
}

// Next returns the next backend from the primary balancer.
func (m *MirrorBalancer) Next(r *http.Request) (string, error) {
	return m.primary.Next(r)
}

// Backends returns the primary balancer's backend list.
func (m *MirrorBalancer) Backends() []string {
	return m.primary.Backends()
}

// Mirror asynchronously sends a copy of the request to the shadow backend.
// It is a no-op when mirroring is disabled.
func (m *MirrorBalancer) Mirror(r *http.Request) {
	m.mu.RLock()
	enabled := m.enabled
	m.mu.RUnlock()
	if !enabled {
		return
	}
	shadowBackend, err := m.shadow.Next(r)
	if err != nil {
		return
	}
	go m.sendMirror(r, shadowBackend)
}

func (m *MirrorBalancer) sendMirror(orig *http.Request, target string) {
	var body []byte
	if orig.Body != nil {
		body, _ = io.ReadAll(orig.Body)
	}
	url := "http://" + target + orig.URL.RequestURI()
	req, err := http.NewRequest(orig.Method, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	for k, v := range orig.Header {
		req.Header[k] = v
	}
	req.Header.Set("X-Mirror", "1")
	//nolint:errcheck
	m.client.Do(req) //nolint:bodyclose
}

// SetEnabled enables or disables shadow mirroring at runtime.
func (m *MirrorBalancer) SetEnabled(v bool) {
	m.mu.Lock()
	m.enabled = v
	m.mu.Unlock()
}

// Enabled reports whether mirroring is currently active.
func (m *MirrorBalancer) Enabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

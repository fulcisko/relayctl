package metrics_test

import (
	"testing"

	"github.com/user/relayctl/internal/metrics"
)

func TestNew(t *testing.T) {
	c := metrics.New()
	if c == nil {
		t.Fatal("expected non-nil Collector")
	}
	s := c.Snapshot()
	if s.TotalRequests != 0 {
		t.Errorf("expected 0 total requests, got %d", s.TotalRequests)
	}
	if s.ActiveRequests != 0 {
		t.Errorf("expected 0 active requests, got %d", s.ActiveRequests)
	}
}

func TestIncRequest(t *testing.T) {
	c := metrics.New()
	c.IncRequest("/api")
	c.IncRequest("/api")
	c.IncRequest("/health")

	s := c.Snapshot()
	if s.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", s.TotalRequests)
	}
	if s.ActiveRequests != 3 {
		t.Errorf("expected 3 active requests, got %d", s.ActiveRequests)
	}
	if s.RouteHits["/api"] != 2 {
		t.Errorf("expected 2 hits for /api, got %d", s.RouteHits["/api"])
	}
	if s.RouteHits["/health"] != 1 {
		t.Errorf("expected 1 hit for /health, got %d", s.RouteHits["/health"])
	}
}

func TestDecActive(t *testing.T) {
	c := metrics.New()
	c.IncRequest("/api")
	c.IncRequest("/api")
	c.DecActive()

	s := c.Snapshot()
	if s.TotalRequests != 2 {
		t.Errorf("expected 2 total requests, got %d", s.TotalRequests)
	}
	if s.ActiveRequests != 1 {
		t.Errorf("expected 1 active request, got %d", s.ActiveRequests)
	}
}

func TestReset(t *testing.T) {
	c := metrics.New()
	c.IncRequest("/api")
	c.Reset()

	s := c.Snapshot()
	if s.TotalRequests != 0 {
		t.Errorf("expected 0 after reset, got %d", s.TotalRequests)
	}
	if len(s.RouteHits) != 0 {
		t.Errorf("expected empty route hits after reset, got %v", s.RouteHits)
	}
}

func TestSnapshot_Uptime(t *testing.T) {
	c := metrics.New()
	s := c.Snapshot()
	if s.UptimeSeconds < 0 {
		t.Errorf("expected non-negative uptime, got %f", s.UptimeSeconds)
	}
}

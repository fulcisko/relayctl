package admin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/relayctl/internal/admin"
	"github.com/user/relayctl/internal/metrics"
)

func TestMetricsHandler_OK(t *testing.T) {
	c := metrics.New()
	c.IncRequest("/api")
	c.IncRequest("/api")
	c.IncRequest("/health")
	c.DecActive()

	h := admin.NewMetricsHandler(c)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var snap metrics.Snapshot
	if err := json.NewDecoder(rr.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if snap.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", snap.TotalRequests)
	}
	if snap.ActiveRequests != 2 {
		t.Errorf("expected 2 active requests, got %d", snap.ActiveRequests)
	}
	if snap.RouteHits["/api"] != 2 {
		t.Errorf("expected 2 hits for /api, got %d", snap.RouteHits["/api"])
	}
}

func TestMetricsHandler_MethodNotAllowed(t *testing.T) {
	c := metrics.New()
	h := admin.NewMetricsHandler(c)

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/metrics", nil)
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rr.Code)
		}
	}
}

func TestMetricsHandler_EmptyCollector(t *testing.T) {
	c := metrics.New()
	h := admin.NewMetricsHandler(c)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var snap metrics.Snapshot
	if err := json.NewDecoder(rr.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if snap.TotalRequests != 0 {
		t.Errorf("expected 0 requests, got %d", snap.TotalRequests)
	}
}

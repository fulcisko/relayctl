package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/relayctl/internal/config"
)

// mockRouteCollector implements RouteCollector for testing.
type mockRouteCollector struct {
	cfg *config.Config
}

func (m *mockRouteCollector) CurrentConfig() *config.Config { return m.cfg }

func sampleConfig() *config.Config {
	return &config.Config{
		Addr: ":8080",
		Rules: []config.Rule{
			{Path: "/api", Backend: "http://localhost:9001"},
			{Path: "/web", Backend: "http://localhost:9002"},
		},
	}
}

func TestRoutesHandler_OK(t *testing.T) {
	h := NewRoutesHandler(&mockRouteCollector{cfg: sampleConfig()})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/routes", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp routesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if resp.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", resp.Addr)
	}
	if len(resp.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(resp.Routes))
	}
	if resp.Routes[0].Path != "/api" || resp.Routes[0].Backend != "http://localhost:9001" {
		t.Errorf("unexpected first route: %+v", resp.Routes[0])
	}
}

func TestRoutesHandler_MethodNotAllowed(t *testing.T) {
	h := NewRoutesHandler(&mockRouteCollector{cfg: sampleConfig()})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/routes", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestRoutesHandler_NilConfig(t *testing.T) {
	h := NewRoutesHandler(&mockRouteCollector{cfg: nil})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/routes", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestRoutesHandler_ContentType(t *testing.T) {
	h := NewRoutesHandler(&mockRouteCollector{cfg: sampleConfig()})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/routes", nil)
	h.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

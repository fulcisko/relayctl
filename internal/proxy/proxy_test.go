package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relayctl/internal/config"
)

func testConfig(backend string) *config.Config {
	return &config.Config{
		Addr: ":8080",
		Routes: []config.Route{
			{Prefix: "/api", Backend: backend},
		},
	}
}

func TestNew_InvalidBackend(t *testing.T) {
	cfg := &config.Config{
		Addr: ":8080",
		Routes: []config.Route{
			{Prefix: "/bad", Backend: "://not-a-url"},
		},
	}
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for invalid backend URL, got nil")
	}
}

func TestServeHTTP_NoMatchingRoute(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	rp, err := New(testConfig(backend.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()
	rp.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestServeHTTP_MatchingRoute(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	rp, err := New(testConfig(backend.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	rp.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestReload(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	rp, err := New(testConfig(backend.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	newCfg := &config.Config{
		Addr: ":8080",
		Routes: []config.Route{
			{Prefix: "/v2", Backend: backend.URL},
		},
	}
	if err := rp.Reload(newCfg); err != nil {
		t.Fatalf("Reload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/old", nil)
	rec := httptest.NewRecorder()
	rp.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Errorf("old route should be gone, expected 502, got %d", rec.Code)
	}
}

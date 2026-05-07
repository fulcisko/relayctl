package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHealthCollector struct {
	statuses map[string]bool
}

func (m *mockHealthCollector) Statuses() map[string]bool {
	return m.statuses
}

func TestHealthCheckHandler_OK(t *testing.T) {
	collector := &mockHealthCollector{
		statuses: map[string]bool{
			"http://backend1:8080": true,
			"http://backend2:8080": false,
		},
	}

	h := NewHealthCheckHandler(collector)
	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json content-type, got %s", ct)
	}

	var results []struct {
		URL     string `json:"url"`
		Healthy bool   `json:"healthy"`
	}
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestHealthCheckHandler_MethodNotAllowed(t *testing.T) {
	h := NewHealthCheckHandler(&mockHealthCollector{})
	req := httptest.NewRequest(http.MethodPost, "/admin/health", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHealthCheckHandler_EmptyCollector(t *testing.T) {
	h := NewHealthCheckHandler(&mockHealthCollector{
		statuses: map[string]bool{},
	})
	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var results []interface{}
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

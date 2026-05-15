package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relayctl/relayctl/internal/upstream"
)

func newTestGeoBalancer(t *testing.T) *upstream.GeoBalancer {
	t.Helper()
	fb, err := upstream.New([]string{"http://fallback:80"})
	if err != nil {
		t.Fatalf("upstream.New: %v", err)
	}
	resolver := func(_ string) string { return "us" }
	usB, _ := upstream.New([]string{"http://us:80"})
	geo, err := upstream.NewGeoBalancer(map[string]upstream.Balancer{"us": usB}, fb, resolver)
	if err != nil {
		t.Fatalf("NewGeoBalancer: %v", err)
	}
	return geo
}

func TestGeoHandler_Get(t *testing.T) {
	geo := newTestGeoBalancer(t)
	h := NewGeoHandler(geo)

	req := httptest.NewRequest(http.MethodGet, "/admin/geo", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var snap struct {
		Backends []string `json:"backends"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(snap.Backends) == 0 {
		t.Error("expected at least one backend")
	}
}

func TestGeoHandler_Put_Valid(t *testing.T) {
	geo := newTestGeoBalancer(t)
	h := NewGeoHandler(geo)

	body, _ := json.Marshal(map[string]interface{}{
		"region":   "eu",
		"backends": []string{"http://eu:80"},
	})
	req := httptest.NewRequest(http.MethodPut, "/admin/geo", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGeoHandler_Put_MissingRegion(t *testing.T) {
	geo := newTestGeoBalancer(t)
	h := NewGeoHandler(geo)

	body, _ := json.Marshal(map[string]interface{}{"backends": []string{"http://x:80"}})
	req := httptest.NewRequest(http.MethodPut, "/admin/geo", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGeoHandler_MethodNotAllowed(t *testing.T) {
	geo := newTestGeoBalancer(t)
	h := NewGeoHandler(geo)

	req := httptest.NewRequest(http.MethodDelete, "/admin/geo", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestGeoHandler_NilBalancer(t *testing.T) {
	h := NewGeoHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/admin/geo", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relayctl/internal/upstream"
)

type simpleTestBalancer struct {
	backends []string
}

func (s *simpleTestBalancer) Next(_ *http.Request) (string, error) {
	if len(s.backends) == 0 {
		return "", upstream.ErrNoBackends
	}
	return s.backends[0], nil
}
func (s *simpleTestBalancer) Backends() []string { return append([]string{}, s.backends...) }

func newTestMirrorBalancer(t *testing.T) *upstream.MirrorBalancer {
	t.Helper()
	mb, err := upstream.NewMirrorBalancer(
		&simpleTestBalancer{backends: []string{"primary:80"}},
		&simpleTestBalancer{backends: []string{"shadow:80"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	return mb
}

func TestMirrorHandler_Get(t *testing.T) {
	mb := newTestMirrorBalancer(t)
	h := NewMirrorHandler(mb)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/mirror", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["enabled"]; !ok {
		t.Fatal("expected 'enabled' field in response")
	}
}

func TestMirrorHandler_Put_Valid(t *testing.T) {
	mb := newTestMirrorBalancer(t)
	h := NewMirrorHandler(mb)
	body := bytes.NewBufferString(`{"enabled":false}`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/mirror", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if mb.Enabled() {
		t.Fatal("expected mirroring to be disabled")
	}
}

func TestMirrorHandler_Put_InvalidJSON(t *testing.T) {
	mb := newTestMirrorBalancer(t)
	h := NewMirrorHandler(mb)
	body := bytes.NewBufferString(`not-json`)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/admin/mirror", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestMirrorHandler_MethodNotAllowed(t *testing.T) {
	mb := newTestMirrorBalancer(t)
	h := NewMirrorHandler(mb)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/admin/mirror", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestMirrorHandler_NilBalancer(t *testing.T) {
	h := NewMirrorHandler(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/mirror", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

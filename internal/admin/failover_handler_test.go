package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeFailover struct {
	backends []string
}

func (f *fakeFailover) Backends() []string { return f.backends }
func (f *fakeFailover) Update(bs []string) error {
	if len(bs) == 0 {
		return errEmptyBackends
	}
	f.backends = bs
	return nil
}

var errEmptyBackends = fmt.Errorf("upstream: failover requires at least one backend")

func newTestFailover() *fakeFailover {
	return &fakeFailover{backends: []string{"http://primary", "http://secondary"}}
}

func TestFailoverHandler_Get(t *testing.T) {
	h := NewFailoverHandler(newTestFailover())
	req := httptest.NewRequest(http.MethodGet, "/admin/failover", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var resp struct {
		Backends []string `json:"backends"`
	}
	json.NewDecoder(rw.Body).Decode(&resp)
	if len(resp.Backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(resp.Backends))
	}
}

func TestFailoverHandler_Put_Valid(t *testing.T) {
	h := NewFailoverHandler(newTestFailover())
	body, _ := json.Marshal(map[string][]string{"backends": {"http://new"}})
	req := httptest.NewRequest(http.MethodPut, "/admin/failover", bytes.NewReader(body))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
}

func TestFailoverHandler_Put_Empty(t *testing.T) {
	h := NewFailoverHandler(newTestFailover())
	body, _ := json.Marshal(map[string][]string{"backends": {}})
	req := httptest.NewRequest(http.MethodPut, "/admin/failover", bytes.NewReader(body))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestFailoverHandler_Put_InvalidJSON(t *testing.T) {
	h := NewFailoverHandler(newTestFailover())
	req := httptest.NewRequest(http.MethodPut, "/admin/failover", bytes.NewReader([]byte("not-json")))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestFailoverHandler_MethodNotAllowed(t *testing.T) {
	h := NewFailoverHandler(newTestFailover())
	req := httptest.NewRequest(http.MethodDelete, "/admin/failover", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/relayctl/relayctl/internal/upstream"
)

func newTestCache() *upstream.ResponseCache {
	return upstream.NewResponseCache(5 * time.Second)
}

func TestCacheHandler_Get_Empty(t *testing.T) {
	h := NewCacheHandler(newTestCache())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/admin/cache", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var s struct {
		Entries int    `json:"entries"`
		TTL     string `json:"ttl"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.Entries != 0 {
		t.Fatalf("expected 0 entries, got %d", s.Entries)
	}
}

func TestCacheHandler_Get_WithEntries(t *testing.T) {
	c := newTestCache()
	c.Set("/foo", &upstream.CacheEntry{StatusCode: 200, Body: []byte("ok")})
	h := NewCacheHandler(c)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/admin/cache", nil))
	var s struct{ Entries int `json:"entries"` }
	json.NewDecoder(rr.Body).Decode(&s)
	if s.Entries != 1 {
		t.Fatalf("expected 1 entry, got %d", s.Entries)
	}
}

func TestCacheHandler_Delete_SingleKey(t *testing.T) {
	c := newTestCache()
	c.Set("/bar", &upstream.CacheEntry{StatusCode: 200})
	h := NewCacheHandler(c)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/admin/cache?key=/bar", nil))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if c.Len() != 0 {
		t.Fatal("expected cache to be empty after delete")
	}
}

func TestCacheHandler_Delete_Flush(t *testing.T) {
	c := newTestCache()
	c.Set("/a", &upstream.CacheEntry{StatusCode: 200})
	c.Set("/b", &upstream.CacheEntry{StatusCode: 200})
	h := NewCacheHandler(c)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/admin/cache?flush=true", nil))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if c.Len() != 0 {
		t.Fatalf("expected empty cache, got %d", c.Len())
	}
}

func TestCacheHandler_Delete_MissingKey(t *testing.T) {
	h := NewCacheHandler(newTestCache())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/admin/cache", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCacheHandler_MethodNotAllowed(t *testing.T) {
	h := NewCacheHandler(newTestCache())
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/admin/cache", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestCacheHandler_NilCache(t *testing.T) {
	h := NewCacheHandler(nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/admin/cache", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

package upstream

import (
	"net/http"
	"testing"
	"time"
)

func TestNewResponseCache_DefaultTTL(t *testing.T) {
	c := NewResponseCache(0)
	if c.ttl != 30*time.Second {
		t.Fatalf("expected default ttl 30s, got %v", c.ttl)
	}
}

func TestResponseCache_SetAndGet(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	e := &CacheEntry{Body: []byte("hello"), StatusCode: 200, Headers: http.Header{}}
	c.Set("/foo", e)
	got := c.Get("/foo")
	if got == nil {
		t.Fatal("expected cache hit")
	}
	if string(got.Body) != "hello" {
		t.Fatalf("unexpected body: %s", got.Body)
	}
}

func TestResponseCache_Miss(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	if c.Get("/missing") != nil {
		t.Fatal("expected cache miss")
	}
}

func TestResponseCache_Expired(t *testing.T) {
	c := NewResponseCache(1 * time.Millisecond)
	c.Set("/bar", &CacheEntry{Body: []byte("x"), StatusCode: 200})
	time.Sleep(5 * time.Millisecond)
	if c.Get("/bar") != nil {
		t.Fatal("expected expired entry to return nil")
	}
}

func TestResponseCache_Delete(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	c.Set("/del", &CacheEntry{StatusCode: 200})
	c.Delete("/del")
	if c.Get("/del") != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestResponseCache_Flush(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	c.Set("/a", &CacheEntry{StatusCode: 200})
	c.Set("/b", &CacheEntry{StatusCode: 200})
	c.Flush()
	if c.Len() != 0 {
		t.Fatalf("expected 0 entries after flush, got %d", c.Len())
	}
}

func TestResponseCache_Len(t *testing.T) {
	c := NewResponseCache(5 * time.Second)
	c.Set("/x", &CacheEntry{StatusCode: 200})
	c.Set("/y", &CacheEntry{StatusCode: 200})
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}

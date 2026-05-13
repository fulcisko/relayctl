package upstream

import (
	"net/http"
	"sync"
	"testing"
)

func TestNewLeastConnBalancer_Empty(t *testing.T) {
	_, err := NewLeastConnBalancer([]string{})
	if err == nil {
		t.Fatal("expected error for empty backends")
	}
}

func TestNewLeastConnBalancer_Valid(t *testing.T) {
	lb, err := NewLeastConnBalancer([]string{"http://a", "http://b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := lb.Backends(); len(got) != 2 {
		t.Fatalf("expected 2 backends, got %d", len(got))
	}
}

func TestLeastConn_PicksLeastLoaded(t *testing.T) {
	lb, _ := NewLeastConnBalancer([]string{"http://a", "http://b"})

	// Simulate 2 active connections on http://a
	_, doneA1 := lb.Next(nil)
	_, doneA2 := lb.Next(nil)
	_ = doneA1
	_ = doneA2

	// Next pick should prefer http://b (0 connections)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	url, done := lb.Next(req)
	defer done()
	if url != "http://b" {
		t.Errorf("expected http://b, got %s", url)
	}
}

func TestLeastConn_DoneDecrementsCount(t *testing.T) {
	lb, _ := NewLeastConnBalancer([]string{"http://a"})
	_, done := lb.Next(nil)
	if lb.ActiveCount("http://a") != 1 {
		t.Fatalf("expected count 1")
	}
	done()
	if lb.ActiveCount("http://a") != 0 {
		t.Fatalf("expected count 0 after done")
	}
}

func TestLeastConn_ActiveCount_Unknown(t *testing.T) {
	lb, _ := NewLeastConnBalancer([]string{"http://a"})
	if got := lb.ActiveCount("http://unknown"); got != -1 {
		t.Errorf("expected -1 for unknown backend, got %d", got)
	}
}

func TestLeastConn_Concurrent(t *testing.T) {
	lb, _ := NewLeastConnBalancer([]string{"http://a", "http://b", "http://c"})
	var wg sync.WaitGroup
	for i := 0; i < 60; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, done := lb.Next(nil)
			done()
		}()
	}
	wg.Wait()
	for _, url := range lb.Backends() {
		if c := lb.ActiveCount(url); c != 0 {
			t.Errorf("expected 0 active for %s after all done, got %d", url, c)
		}
	}
}

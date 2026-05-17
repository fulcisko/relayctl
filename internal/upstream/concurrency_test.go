package upstream

import (
	"sync"
	"testing"
)

func TestNewConcurrencyRegistry_Empty(t *testing.T) {
	reg := NewConcurrencyRegistry()
	_, _, ok := reg.Get("http://backend:8080")
	if ok {
		t.Fatal("expected no entry for unknown backend")
	}
}

func TestConcurrencyRegistry_Set_And_Get(t *testing.T) {
	reg := NewConcurrencyRegistry()
	if err := reg.Set("http://b:9000", 5); err != nil {
		t.Fatalf("Set: %v", err)
	}
	max, active, ok := reg.Get("http://b:9000")
	if !ok || max != 5 || active != 0 {
		t.Fatalf("unexpected: ok=%v max=%d active=%d", ok, max, active)
	}
}

func TestConcurrencyRegistry_Set_EmptyBackend(t *testing.T) {
	reg := NewConcurrencyRegistry()
	if err := reg.Set("", 10); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestConcurrencyRegistry_Set_InvalidMax(t *testing.T) {
	reg := NewConcurrencyRegistry()
	if err := reg.Set("http://b:9000", 0); err == nil {
		t.Fatal("expected error for max=0")
	}
}

func TestConcurrencyRegistry_Acquire_NoLimit(t *testing.T) {
	reg := NewConcurrencyRegistry()
	tok, err := reg.Acquire("http://unknown")
	if err != nil {
		t.Fatalf("expected no error for unconfigured backend: %v", err)
	}
	tok.Release() // should not panic
}

func TestConcurrencyRegistry_Acquire_UnderLimit(t *testing.T) {
	reg := NewConcurrencyRegistry()
	_ = reg.Set("http://b", 3)
	tok, err := reg.Acquire("http://b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, active, _ := reg.Get("http://b")
	if active != 1 {
		t.Fatalf("expected active=1, got %d", active)
	}
	tok.Release()
	_, active, _ = reg.Get("http://b")
	if active != 0 {
		t.Fatalf("expected active=0 after release, got %d", active)
	}
}

func TestConcurrencyRegistry_Acquire_ExceedsLimit(t *testing.T) {
	reg := NewConcurrencyRegistry()
	_ = reg.Set("http://b", 2)
	tok1, _ := reg.Acquire("http://b")
	tok2, _ := reg.Acquire("http://b")
	_, err := reg.Acquire("http://b")
	if err != ErrConcurrencyLimitExceeded {
		t.Fatalf("expected ErrConcurrencyLimitExceeded, got %v", err)
	}
	tok1.Release()
	tok2.Release()
}

func TestConcurrencyRegistry_Delete(t *testing.T) {
	reg := NewConcurrencyRegistry()
	_ = reg.Set("http://b", 5)
	reg.Delete("http://b")
	_, _, ok := reg.Get("http://b")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestConcurrencyRegistry_Concurrent(t *testing.T) {
	reg := NewConcurrencyRegistry()
	_ = reg.Set("http://b", 10)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tok, err := reg.Acquire("http://b")
			if err == nil {
				tok.Release()
			}
		}()
	}
	wg.Wait()
	_, active, _ := reg.Get("http://b")
	if active != 0 {
		t.Fatalf("expected active=0 after all releases, got %d", active)
	}
}

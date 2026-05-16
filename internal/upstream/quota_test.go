package upstream

import (
	"testing"
	"time"
)

func TestNewQuotaRegistry_Empty(t *testing.T) {
	r := NewQuotaRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Fatal("expected empty snapshot")
	}
}

func TestQuotaRegistry_Set_And_Get(t *testing.T) {
	r := NewQuotaRegistry()
	err := r.Set("http://backend:8080", QuotaConfig{MaxRequests: 100, Window: time.Minute})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, ok := r.Get("http://backend:8080")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if cfg.MaxRequests != 100 {
		t.Errorf("expected 100, got %d", cfg.MaxRequests)
	}
}

func TestQuotaRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewQuotaRegistry()
	if err := r.Set("", QuotaConfig{MaxRequests: 10}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestQuotaRegistry_Set_InvalidMaxRequests(t *testing.T) {
	r := NewQuotaRegistry()
	if err := r.Set("http://backend:8080", QuotaConfig{MaxRequests: 0}); err == nil {
		t.Fatal("expected error for zero max_requests")
	}
}

func TestQuotaRegistry_Set_DefaultsWindow(t *testing.T) {
	r := NewQuotaRegistry()
	_ = r.Set("http://b:9000", QuotaConfig{MaxRequests: 5})
	cfg, _ := r.Get("http://b:9000")
	if cfg.Window != time.Minute {
		t.Errorf("expected default window 1m, got %v", cfg.Window)
	}
}

func TestQuotaRegistry_Allow_UnderLimit(t *testing.T) {
	r := NewQuotaRegistry()
	_ = r.Set("http://b:8080", QuotaConfig{MaxRequests: 3, Window: time.Minute})
	for i := 0; i < 3; i++ {
		if !r.Allow("http://b:8080") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestQuotaRegistry_Allow_ExceedsLimit(t *testing.T) {
	r := NewQuotaRegistry()
	_ = r.Set("http://b:8080", QuotaConfig{MaxRequests: 2, Window: time.Minute})
	r.Allow("http://b:8080")
	r.Allow("http://b:8080")
	if r.Allow("http://b:8080") {
		t.Fatal("third request should be denied")
	}
}

func TestQuotaRegistry_Allow_NoConfig(t *testing.T) {
	r := NewQuotaRegistry()
	if !r.Allow("http://unknown:9999") {
		t.Fatal("unknown backend should be allowed by default")
	}
}

func TestQuotaRegistry_Allow_WindowReset(t *testing.T) {
	r := NewQuotaRegistry()
	_ = r.Set("http://b:8080", QuotaConfig{MaxRequests: 1, Window: 10 * time.Millisecond})
	r.Allow("http://b:8080") // consume quota
	if r.Allow("http://b:8080") {
		t.Fatal("should be denied before window reset")
	}
	time.Sleep(20 * time.Millisecond)
	if !r.Allow("http://b:8080") {
		t.Fatal("should be allowed after window reset")
	}
}

func TestQuotaRegistry_Delete(t *testing.T) {
	r := NewQuotaRegistry()
	_ = r.Set("http://b:8080", QuotaConfig{MaxRequests: 1, Window: time.Minute})
	r.Delete("http://b:8080")
	_, ok := r.Get("http://b:8080")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

package upstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHashBalancer_Empty(t *testing.T) {
	_, err := NewHashBalancer(nil, IPKeyFn)
	if err == nil {
		t.Fatal("expected error for empty backends")
	}
}

func TestNewHashBalancer_NilKeyFn(t *testing.T) {
	_, err := NewHashBalancer([]string{"http://localhost:8001"}, nil)
	if err == nil {
		t.Fatal("expected error for nil keyFn")
	}
}

func TestHashBalancer_Next_Deterministic(t *testing.T) {
	backends := []string{"http://a:8001", "http://b:8002", "http://c:8003"}
	hb, err := NewHashBalancer(backends, IPKeyFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.10:54321"

	first, err := hb.Next(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 10; i++ {
		got, _ := hb.Next(req)
		if got != first {
			t.Errorf("iteration %d: got %q, want %q", i, got, first)
		}
	}
}

func TestHashBalancer_Next_DifferentKeys(t *testing.T) {
	backends := []string{"http://a:8001", "http://b:8002", "http://c:8003"}
	hb, _ := NewHashBalancer(backends, IPKeyFn)

	seen := map[string]bool{}
	for i := 0; i < 30; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0." + string(rune('0'+i%10)) + ":1234"
		b, _ := hb.Next(req)
		seen[b] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected multiple backends to be selected, got %v", seen)
	}
}

func TestHashBalancer_Backends_ReturnsCopy(t *testing.T) {
	backends := []string{"http://a:8001", "http://b:8002"}
	hb, _ := NewHashBalancer(backends, IPKeyFn)

	got := hb.Backends()
	got[0] = "http://mutated"
	if hb.Backends()[0] == "http://mutated" {
		t.Error("Backends() should return a copy, not the internal slice")
	}
}

func TestHashBalancer_Update_Valid(t *testing.T) {
	hb, _ := NewHashBalancer([]string{"http://old:9000"}, IPKeyFn)
	newBackends := []string{"http://new1:9001", "http://new2:9002"}
	if err := hb.Update(newBackends); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := hb.Backends(); len(got) != 2 {
		t.Errorf("expected 2 backends, got %d", len(got))
	}
}

func TestHashBalancer_Update_Empty(t *testing.T) {
	hb, _ := NewHashBalancer([]string{"http://a:8001"}, IPKeyFn)
	if err := hb.Update(nil); err == nil {
		t.Error("expected error when updating with empty backends")
	}
}

func TestHeaderKeyFn_FallsBackToRemoteAddr(t *testing.T) {
	keyFn := HeaderKeyFn("X-User-ID")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"

	key := keyFn(req)
	if key != "1.2.3.4:5678" {
		t.Errorf("expected remote addr fallback, got %q", key)
	}
}

func TestHeaderKeyFn_UsesHeader(t *testing.T) {
	keyFn := HeaderKeyFn("X-User-ID")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "user-42")

	key := keyFn(req)
	if key != "user-42" {
		t.Errorf("expected header value, got %q", key)
	}
}

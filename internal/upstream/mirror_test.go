package upstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newSimpleMirrorBalancer(backends ...string) *simpleBalancer {
	return &simpleBalancer{backends: backends, idx: 0}
}

func TestNewMirrorBalancer_NilPrimary(t *testing.T) {
	_, err := NewMirrorBalancer(nil, newSimpleMirrorBalancer("b:80"))
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNewMirrorBalancer_NilShadow(t *testing.T) {
	_, err := NewMirrorBalancer(newSimpleMirrorBalancer("a:80"), nil)
	if err == nil {
		t.Fatal("expected error for nil shadow")
	}
}

func TestMirrorBalancer_Next_UsesPrimary(t *testing.T) {
	m, err := NewMirrorBalancer(
		newSimpleMirrorBalancer("primary:80"),
		newSimpleMirrorBalancer("shadow:80"),
	)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got, err := m.Next(req)
	if err != nil {
		t.Fatal(err)
	}
	if got != "primary:80" {
		t.Fatalf("expected primary:80, got %s", got)
	}
}

func TestMirrorBalancer_Mirror_SendsToShadow(t *testing.T) {
	received := make(chan struct{}, 1)
	shadowSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Mirror") == "1" {
			received <- struct{}{}
		}
	}))
	defer shadowSrv.Close()

	shadowHost := shadowSrv.Listener.Addr().String()
	m, _ := NewMirrorBalancer(
		newSimpleMirrorBalancer("primary:80"),
		newSimpleMirrorBalancer(shadowHost),
	)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	m.Mirror(req)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("shadow did not receive mirrored request")
	}
}

func TestMirrorBalancer_SetEnabled_DisablesMirroring(t *testing.T) {
	hit := make(chan struct{}, 1)
	shadowSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit <- struct{}{}
	}))
	defer shadowSrv.Close()

	m, _ := NewMirrorBalancer(
		newSimpleMirrorBalancer("primary:80"),
		newSimpleMirrorBalancer(shadowSrv.Listener.Addr().String()),
	)
	m.SetEnabled(false)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	m.Mirror(req)

	select {
	case <-hit:
		t.Fatal("shadow should not receive request when disabled")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestMirrorBalancer_Backends_ReturnsPrimary(t *testing.T) {
	m, _ := NewMirrorBalancer(
		newSimpleMirrorBalancer("a:80", "b:80"),
		newSimpleMirrorBalancer("s:80"),
	)
	bs := m.Backends()
	if len(bs) != 2 {
		t.Fatalf("expected 2 backends, got %d", len(bs))
	}
}

package upstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func staticResolver(region string) RegionResolver {
	return func(_ string) string { return region }
}

func newSimpleGeoBalancer(t *testing.T, backends []string) Balancer {
	t.Helper()
	b, err := New(backends)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return b
}

func TestNewGeoBalancer_NilFallback(t *testing.T) {
	_, err := NewGeoBalancer(nil, nil, staticResolver("us"))
	if err == nil {
		t.Fatal("expected error for nil fallback")
	}
}

func TestNewGeoBalancer_NilResolver(t *testing.T) {
	fb := newSimpleGeoBalancer(t, []string{"http://fb:80"})
	_, err := NewGeoBalancer(nil, fb, nil)
	if err == nil {
		t.Fatal("expected error for nil resolver")
	}
}

func TestNewGeoBalancer_NilRegionBalancer(t *testing.T) {
	fb := newSimpleGeoBalancer(t, []string{"http://fb:80"})
	_, err := NewGeoBalancer(map[string]Balancer{"us": nil}, fb, staticResolver("us"))
	if err == nil {
		t.Fatal("expected error for nil region balancer")
	}
}

func TestGeoBalancer_Next_MatchesRegion(t *testing.T) {
	usBackend := "http://us-backend:80"
	euBackend := "http://eu-backend:80"

	usB := newSimpleGeoBalancer(t, []string{usBackend})
	euB := newSimpleGeoBalancer(t, []string{euBackend})
	fb := newSimpleGeoBalancer(t, []string{"http://default:80"})

	geo, err := NewGeoBalancer(map[string]Balancer{"us": usB, "eu": euB}, fb, staticResolver("us"))
	if err != nil {
		t.Fatalf("NewGeoBalancer: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got, err := geo.Next(req)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if got != usBackend {
		t.Errorf("got %q, want %q", got, usBackend)
	}
}

func TestGeoBalancer_Next_FallsBackWhenNoRegion(t *testing.T) {
	defaultBackend := "http://default:80"
	fb := newSimpleGeoBalancer(t, []string{defaultBackend})

	geo, err := NewGeoBalancer(map[string]Balancer{}, fb, staticResolver("unknown"))
	if err != nil {
		t.Fatalf("NewGeoBalancer: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got, err := geo.Next(req)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if got != defaultBackend {
		t.Errorf("got %q, want %q", got, defaultBackend)
	}
}

func TestGeoBalancer_Backends_Union(t *testing.T) {
	usB := newSimpleGeoBalancer(t, []string{"http://us:80"})
	fb := newSimpleGeoBalancer(t, []string{"http://fb:80"})

	geo, _ := NewGeoBalancer(map[string]Balancer{"us": usB}, fb, staticResolver("us"))
	backends := geo.Backends()
	if len(backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(backends))
	}
}

func TestGeoBalancer_SetRegion(t *testing.T) {
	fb := newSimpleGeoBalancer(t, []string{"http://fb:80"})
	geo, _ := NewGeoBalancer(nil, fb, staticResolver("ap"))

	apB := newSimpleGeoBalancer(t, []string{"http://ap:80"})
	if err := geo.SetRegion("ap", apB); err != nil {
		t.Fatalf("SetRegion: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got, _ := geo.Next(req)
	if got != "http://ap:80" {
		t.Errorf("got %q, want http://ap:80", got)
	}
}

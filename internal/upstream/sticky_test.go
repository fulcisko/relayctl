package upstream

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newStickyBalancer(t *testing.T) *StickyBalancer {
	t.Helper()
	b, err := New([]string{"http://backend1", "http://backend2", "http://backend3"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return NewStickyBalancer(b, "relay_session")
}

func TestStickyBalancer_NoCookie_FallsThrough(t *testing.T) {
	sb := newStickyBalancer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	backend, ok := sb.Next(req)
	if !ok {
		t.Fatal("expected a backend to be selected")
	}
	if backend == "" {
		t.Fatal("expected non-empty backend")
	}
}

func TestStickyBalancer_SameCookie_SameBackend(t *testing.T) {
	sb := newStickyBalancer(t)

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.AddCookie(&http.Cookie{Name: "relay_session", Value: "abc123"})

	first, ok := sb.Next(req1)
	if !ok {
		t.Fatal("expected backend on first call")
	}

	// Second request with same cookie should return the same backend.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "relay_session", Value: "abc123"})

	second, ok := sb.Next(req2)
	if !ok {
		t.Fatal("expected backend on second call")
	}
	if first != second {
		t.Errorf("sticky session mismatch: got %q, want %q", second, first)
	}
}

func TestStickyBalancer_DifferentCookies_MayDiffer(t *testing.T) {
	sb := newStickyBalancer(t)

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.AddCookie(&http.Cookie{Name: "relay_session", Value: "user-A"})
	sb.Next(req1) //nolint:errcheck

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "relay_session", Value: "user-B"})
	sb.Next(req2) //nolint:errcheck

	if sb.SessionCount() != 2 {
		t.Errorf("expected 2 sessions, got %d", sb.SessionCount())
	}
}

func TestStickyBalancer_Forget(t *testing.T) {
	sb := newStickyBalancer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "relay_session", Value: "to-forget"})
	sb.Next(req)

	if sb.SessionCount() != 1 {
		t.Fatalf("expected 1 session before forget, got %d", sb.SessionCount())
	}

	sb.Forget("to-forget")

	if sb.SessionCount() != 0 {
		t.Errorf("expected 0 sessions after forget, got %d", sb.SessionCount())
	}
}

func TestStickyBalancer_SessionCount_Empty(t *testing.T) {
	sb := newStickyBalancer(t)
	if sb.SessionCount() != 0 {
		t.Errorf("expected 0 sessions on new balancer, got %d", sb.SessionCount())
	}
}

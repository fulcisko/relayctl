package upstream

import (
	"testing"
)

func TestNewShadowRegistry_Empty(t *testing.T) {
	r := NewShadowRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.Snapshot()) != 0 {
		t.Fatal("expected empty snapshot")
	}
}

func TestShadowRegistry_Set_And_Get(t *testing.T) {
	r := NewShadowRegistry()
	entry := ShadowEntry{Backend: "http://shadow:9090", SampleRate: 0.5, Enabled: true}
	if err := r.Set("http://primary:8080", entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("http://primary:8080")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.SampleRate != 0.5 {
		t.Errorf("expected 0.5, got %v", got.SampleRate)
	}
}

func TestShadowRegistry_Set_EmptyBackend(t *testing.T) {
	r := NewShadowRegistry()
	if err := r.Set("", ShadowEntry{}); err == nil {
		t.Fatal("expected error for empty backend")
	}
}

func TestShadowRegistry_ClampsRate(t *testing.T) {
	r := NewShadowRegistry()
	_ = r.Set("http://a:1", ShadowEntry{SampleRate: 5.0})
	got, _ := r.Get("http://a:1")
	if got.SampleRate != 1.0 {
		t.Errorf("expected clamped to 1.0, got %v", got.SampleRate)
	}
	_ = r.Set("http://b:2", ShadowEntry{SampleRate: -1.0})
	got2, _ := r.Get("http://b:2")
	if got2.SampleRate != 0.0 {
		t.Errorf("expected clamped to 0.0, got %v", got2.SampleRate)
	}
}

func TestShadowRegistry_Delete(t *testing.T) {
	r := NewShadowRegistry()
	_ = r.Set("http://x:1", ShadowEntry{SampleRate: 0.1, Enabled: true})
	r.Delete("http://x:1")
	if _, ok := r.Get("http://x:1"); ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestShadowRegistry_Snapshot(t *testing.T) {
	r := NewShadowRegistry()
	_ = r.Set("http://a:1", ShadowEntry{SampleRate: 0.2, Enabled: true})
	_ = r.Set("http://b:2", ShadowEntry{SampleRate: 0.8, Enabled: false})
	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
}

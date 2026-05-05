package watcher_test

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/user/relayctl/internal/watcher"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestNew_InvalidPath(t *testing.T) {
	_, err := watcher.New("/nonexistent/path/config.yaml", func() {})
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestNew_ValidPath(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\n")
	w, err := watcher.New(path, func() {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Stop()
}

func TestWatcher_TriggerOnWrite(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\n")

	var mu sync.Mutex
	called := 0
	w, err := watcher.New(path, func() {
		mu.Lock()
		called++
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Start()
	defer w.Stop()

	// Trigger a write event
	if err := os.WriteFile(path, []byte("addr: :9090\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Wait for debounce + callback
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if called == 0 {
		t.Error("expected onChange to be called at least once")
	}
}

func TestWatcher_Stop(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\n")
	w, err := watcher.New(path, func() {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Start()
	// Should not panic or block
	w.Stop()
}

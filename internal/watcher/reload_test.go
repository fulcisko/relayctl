package watcher_test

import (
	"errors"
	"os"
	"testing"

	"github.com/user/relayctl/internal/config"
	"github.com/user/relayctl/internal/watcher"
)

func validConfigFile(t *testing.T) string {
	t.Helper()
	content := `addr: ":8080"
routes:
  - path: /api
    backend: http://localhost:3000
`
	f, err := os.CreateTemp(t.TempDir(), "relay-*.yaml")
	if err != nil {
		t.Fatalf("createTemp: %v", err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestReloader_Reload_Success(t *testing.T) {
	path := validConfigFile(t)
	var received *config.Config
	r := watcher.NewReloader(path, func(cfg *config.Config) error {
		received = cfg
		return nil
	})
	if err := r.Reload(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received == nil {
		t.Fatal("expected config to be passed to callback")
	}
	if received.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", received.Addr)
	}
}

func TestReloader_Reload_BadFile(t *testing.T) {
	r := watcher.NewReloader("/no/such/file.yaml", func(cfg *config.Config) error {
		return nil
	})
	if err := r.Reload(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReloader_Reload_CallbackError(t *testing.T) {
	path := validConfigFile(t)
	sentinel := errors.New("callback failed")
	r := watcher.NewReloader(path, func(cfg *config.Config) error {
		return sentinel
	})
	err := r.Reload()
	if err == nil {
		t.Fatal("expected error from callback")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
}

func TestReloader_MustReload_NoError(t *testing.T) {
	path := validConfigFile(t)
	r := watcher.NewReloader(path, func(cfg *config.Config) error {
		return nil
	})
	// Should not panic
	r.MustReload()
}

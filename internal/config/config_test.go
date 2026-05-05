package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/relayctl/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "relay.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return p
}

func TestLoad_Valid(t *testing.T) {
	yaml := `
server:
  addr: ":8080"
routes:
  - name: api
    match: /api/
    backend: http://localhost:9000
    strip_prefix: true
`
	cfg, err := config.Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %q", cfg.Server.Addr)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
	if cfg.Routes[0].Backend != "http://localhost:9000" {
		t.Errorf("unexpected backend %q", cfg.Routes[0].Backend)
	}
}

func TestLoad_MissingAddr(t *testing.T) {
	yaml := `
server:
  addr: ""
routes:
  - name: web
    match: /
    backend: http://localhost:3000
`
	_, err := config.Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoad_MissingBackend(t *testing.T) {
	yaml := `
server:
  addr: ":8080"
routes:
  - name: broken
    match: /broken
    backend: ""
`
	_, err := config.Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("expected validation error for missing backend")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/relay.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	invalid := `
server:
  addr: ":8080"
routes:
  - name: bad
    match: /
  backend: not-indented-correctly
    extra: [unclosed
`
	_, err := config.Load(writeTemp(t, invalid))
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

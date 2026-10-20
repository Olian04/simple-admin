package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Olian04/simple-admin/internal/config"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

const manualBackend = `
backends:
  - base_url: 'http://localhost:8090'
    auth:
      type: 'apikey'
      in: 'header'
      name: 'X-API-Key'
    resources:
      - name: 'posts'
        schema: './schemas/post.json'
        paths:
          collection: '/api/collections/posts/records'
          item: '/api/collections/posts/records/{id}'
        field_hints:
          rating: 'relation'
          score: { widget: slider, min: 0, max: 5, step: 1 }
          kind: { widget: select, options: [draft, { value: live, label: Live, tone: ok }] }
`

func TestManualBackendDefaultsAndParse(t *testing.T) {
	cfg, err := config.Load(writeConfig(t, manualBackend))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.Backends) != 1 {
		t.Fatalf("want 1 backend, got %d", len(cfg.Backends))
	}
	b := cfg.Backends[0]
	if b.ID == "" {
		t.Fatal("backend id should be derived when omitted")
	}
	if b.Label == "" {
		t.Fatal("backend label should be defaulted")
	}
	r := b.Resources[0]
	if r.Label != "Posts" {
		t.Fatalf("resource label defaulted wrong: %q", r.Label)
	}
	if r.IDField != "id" {
		t.Fatalf("id_field default wrong: %q", r.IDField)
	}
	if r.Schema.Path != "./schemas/post.json" {
		t.Fatalf("schema path wrong: %q", r.Schema.Path)
	}
	if h := r.FieldHints["rating"]; h.Widget != "relation" {
		t.Fatalf("shorthand field hint wrong: %#v", h)
	}
	if h := r.FieldHints["score"]; h.Widget != "slider" || h.Max != 5 || h.Step != 1 {
		t.Fatalf("slider hint wrong: %#v", h)
	}
	if h := r.FieldHints["kind"]; len(h.Options) != 2 || h.Options[0].Value != "draft" || h.Options[1].Label != "Live" || h.Options[1].Tone != "ok" {
		t.Fatalf("select hint wrong: %#v", h)
	}
}

func TestSchemaURLForm(t *testing.T) {
	cfg, err := config.Load(writeConfig(t, `
backends:
  - id: 'x'
    base_url: 'http://h'
    resources:
      - name: 'a'
        schema: { url: 'http://h/a.json' }
        paths: { collection: '/a', item: '/a/{id}' }
`))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := cfg.Backends[0].Resources[0].Schema.URL; got != "http://h/a.json" {
		t.Fatalf("schema url wrong: %q", got)
	}
}

func TestDiscoveryEnvExpansionAndDefaults(t *testing.T) {
	t.Setenv("DISCO_KEY", "s3cr3t")
	cfg, err := config.Load(writeConfig(t, `
backends:
  - id: 'gw'
    base_url: 'http://h'
    discovery:
      url: 'http://h/openapi.json'
      auth:
        type: 'apikey'
        in: 'header'
        name: 'X-API-Key'
        value: '${DISCO_KEY}'
`))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	d := cfg.Backends[0].Discovery
	if d.PollInterval != "5m" {
		t.Fatalf("poll_interval default wrong: %q", d.PollInterval)
	}
	if d.Auth.Value != "s3cr3t" {
		t.Fatalf("env not expanded: %q", d.Auth.Value)
	}
}

func TestRejectBothModes(t *testing.T) {
	_, err := config.Load(writeConfig(t, `
backends:
  - id: 'x'
    base_url: 'http://h'
    resources:
      - name: 'a'
        schema: './a.json'
        paths: { collection: '/a', item: '/a/{id}' }
    discovery:
      url: 'http://h/openapi.json'
`))
	if err == nil {
		t.Fatal("expected error when both resources and discovery set")
	}
}

func TestRejectNeitherMode(t *testing.T) {
	_, err := config.Load(writeConfig(t, `
backends:
  - id: 'x'
    base_url: 'http://h'
`))
	if err == nil {
		t.Fatal("expected error when neither resources nor discovery set")
	}
}

func TestRejectAuthInDiscoveryMode(t *testing.T) {
	_, err := config.Load(writeConfig(t, `
backends:
  - id: 'x'
    base_url: 'http://h'
    auth:
      type: 'apikey'
      in: 'header'
      name: 'X-API-Key'
    discovery:
      url: 'http://h/openapi.json'
`))
	if err == nil {
		t.Fatal("expected error: backend-level auth forbidden in discovery mode")
	}
}

func TestRejectBackendAuthValue(t *testing.T) {
	_, err := config.Load(writeConfig(t, `
backends:
  - id: 'x'
    base_url: 'http://h'
    auth:
      type: 'apikey'
      in: 'header'
      name: 'X-API-Key'
      value: 'leaked'
    resources:
      - name: 'a'
        schema: './a.json'
        paths: { collection: '/a', item: '/a/{id}' }
`))
	if err == nil {
		t.Fatal("expected error: backend auth.value not allowed (browser-side secret)")
	}
}

func TestRejectUnsupportedAuthType(t *testing.T) {
	_, err := config.Load(writeConfig(t, `
backends:
  - id: 'x'
    base_url: 'http://h'
    auth:
      type: 'bearer'
    resources:
      - name: 'a'
        schema: './a.json'
        paths: { collection: '/a', item: '/a/{id}' }
`))
	if err == nil {
		t.Fatal("expected error: bearer not yet supported")
	}
}

func TestExampleConfigLoads(t *testing.T) {
	// The shipped example must always parse + validate.
	path := filepath.Join("..", "..", "..", "configs", "config.example.yaml")
	if _, err := config.Load(path); err != nil {
		t.Fatalf("configs/config.example.yaml failed to load: %v", err)
	}
}

func TestRejectItemPathWithoutID(t *testing.T) {
	_, err := config.Load(writeConfig(t, `
backends:
  - id: 'x'
    base_url: 'http://h'
    resources:
      - name: 'a'
        schema: './a.json'
        paths: { collection: '/a', item: '/a' }
`))
	if err == nil {
		t.Fatal("expected error: paths.item must contain {id}")
	}
}

func TestInvalidFieldHintIsRejected(t *testing.T) {
	_, err := config.Load(writeConfig(t, `
backends:
  - base_url: 'http://h'
    resources:
      - name: 'posts'
        schema: './s.json'
        paths: { collection: '/posts', item: '/posts/{id}' }
        field_hints:
          body: 'code'
`))
	if err == nil {
		t.Fatal("unknown widget should fail validation")
	}
}

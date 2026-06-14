package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var slugRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// BackendSection is one upstream backend the admin manages. Exactly one of
// Resources (manual mode) or Discovery (discovery mode) must be set.
type BackendSection struct {
	ID        string            `yaml:"id,omitempty"`
	Label     string            `yaml:"label,omitempty"`
	BaseURL   string            `yaml:"base_url"`
	Auth      *AuthScheme       `yaml:"auth,omitempty"`      // manual mode only (scheme, no secret)
	Resources []ResourceSection `yaml:"resources,omitempty"` // manual mode
	Discovery *DiscoverySection `yaml:"discovery,omitempty"` // discovery mode
}

// ResourceSection is a manually-declared resource (manual mode only).
type ResourceSection struct {
	Name           string                   `yaml:"name"`
	Label          string                   `yaml:"label,omitempty"`
	Schema         SchemaSource             `yaml:"schema"`
	Paths          ResourcePaths            `yaml:"paths"`
	IDField        string                   `yaml:"id_field,omitempty"`
	FieldTemplates map[string]FieldTemplate `yaml:"field_templates,omitempty"`
}

// ResourcePaths are the upstream CRUD paths (resolved against base_url).
type ResourcePaths struct {
	Collection string `yaml:"collection"`
	Item       string `yaml:"item"`
}

// DiscoverySection points at an OpenAPI spec from which resources + auth scheme
// are derived. Its Auth carries the server-side credential for the fetch.
type DiscoverySection struct {
	URL          string      `yaml:"url"`
	PollInterval string      `yaml:"poll_interval,omitempty"`
	Auth         *AuthScheme `yaml:"auth,omitempty"`
}

// AuthScheme is auth metadata. Value is only valid on discovery.auth
// (the server-side discovery-request credential); backend-level auth must not
// carry a secret (those live in the browser).
type AuthScheme struct {
	Type  string `yaml:"type,omitempty"`  // none|apikey|bearer|basic|oauth
	In    string `yaml:"in,omitempty"`    // header|query
	Name  string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"` // discovery.auth only; env-expanded at load
}

// SchemaSource is a JSON Schema location: a disk path (scalar string) or a
// {url: ...} mapping fetched from the backend/owner.
type SchemaSource struct {
	Path string
	URL  string
}

// UnmarshalYAML accepts either a scalar path or a `{ url: ... }` mapping.
func (s *SchemaSource) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		s.Path = node.Value
		return nil
	case yaml.MappingNode:
		var aux struct {
			URL string `yaml:"url"`
		}
		if err := node.Decode(&aux); err != nil {
			return err
		}
		s.URL = aux.URL
		return nil
	default:
		return fmt.Errorf("schema must be a file path or { url: ... }")
	}
}

// FieldTemplate is a per-field widget override: a shorthand built-in id
// (scalar), a `{ template: id, args: {...} }` mapping, or a
// `{ template_file: path }` mapping for a user-supplied template.
type FieldTemplate struct {
	Template     string                 `yaml:"template,omitempty"`
	TemplateFile string                 `yaml:"template_file,omitempty"`
	Args         map[string]TemplateArg `yaml:"args,omitempty"`
}

func (t *FieldTemplate) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		t.Template = node.Value
		return nil
	case yaml.MappingNode:
		var aux struct {
			Template     string                 `yaml:"template"`
			TemplateFile string                 `yaml:"template_file"`
			Args         map[string]TemplateArg `yaml:"args"`
		}
		if err := node.Decode(&aux); err != nil {
			return err
		}
		t.Template = aux.Template
		t.TemplateFile = aux.TemplateFile
		t.Args = aux.Args
		return nil
	default:
		return fmt.Errorf("field template must be an id or a mapping")
	}
}

// TemplateArg is a static value or a `{ jsonpath: '$.x' }` selector resolved
// against the record at render time.
type TemplateArg struct {
	Static   any
	JSONPath string
}

func (a *TemplateArg) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.MappingNode {
		var aux struct {
			JSONPath string `yaml:"jsonpath"`
		}
		if err := node.Decode(&aux); err == nil && aux.JSONPath != "" {
			a.JSONPath = aux.JSONPath
			return nil
		}
	}
	return node.Decode(&a.Static)
}

func (b BackendSection) isDiscovery() bool { return b.Discovery != nil }

func (b BackendSection) WithDefaults() BackendSection {
	if b.ID == "" {
		b.ID = deriveBackendID(b.BaseURL)
	}
	if b.Label == "" {
		b.Label = titleize(b.ID)
	}
	if b.Auth != nil {
		b.Auth = b.Auth.WithDefaults()
	}
	for i := range b.Resources {
		b.Resources[i] = b.Resources[i].WithDefaults()
	}
	if b.Discovery != nil {
		d := *b.Discovery
		if d.PollInterval == "" {
			d.PollInterval = DefaultPollInterval
		}
		if d.Auth != nil {
			d.Auth = d.Auth.WithDefaults()
		}
		b.Discovery = &d
	}
	return b
}

func (b BackendSection) Validate() error {
	if !slugRE.MatchString(b.ID) {
		return fmt.Errorf("backend id %q must be a lowercase slug (a-z0-9-)", b.ID)
	}
	if _, err := url.ParseRequestURI(b.BaseURL); err != nil {
		return fmt.Errorf("backend %q: invalid base_url %q: %w", b.ID, b.BaseURL, err)
	}

	hasResources := len(b.Resources) > 0
	hasDiscovery := b.isDiscovery()
	if hasResources == hasDiscovery {
		return fmt.Errorf("backend %q: set exactly one of `resources` (manual) or `discovery`", b.ID)
	}

	if hasDiscovery {
		if b.Auth != nil {
			return fmt.Errorf("backend %q: top-level `auth` is forbidden in discovery mode (scheme is derived from the spec)", b.ID)
		}
		return b.Discovery.Validate(b.ID)
	}

	// manual mode
	if b.Auth != nil {
		if err := b.Auth.Validate(false); err != nil {
			return fmt.Errorf("backend %q auth: %w", b.ID, err)
		}
	}
	seen := map[string]bool{}
	for _, r := range b.Resources {
		if err := r.Validate(b.ID); err != nil {
			return err
		}
		if seen[r.Name] {
			return fmt.Errorf("backend %q: duplicate resource name %q", b.ID, r.Name)
		}
		seen[r.Name] = true
	}
	return nil
}

func (r ResourceSection) WithDefaults() ResourceSection {
	if r.Label == "" {
		r.Label = titleize(r.Name)
	}
	if r.IDField == "" {
		r.IDField = DefaultIDField
	}
	return r
}

func (r ResourceSection) Validate(backendID string) error {
	if !slugRE.MatchString(r.Name) {
		return fmt.Errorf("backend %q resource name %q must be a lowercase slug (a-z0-9-)", backendID, r.Name)
	}
	if r.Schema.Path == "" && r.Schema.URL == "" {
		return fmt.Errorf("backend %q resource %q: `schema` (path or url) is required in manual mode", backendID, r.Name)
	}
	if strings.TrimSpace(r.Paths.Collection) == "" {
		return fmt.Errorf("backend %q resource %q: paths.collection is required", backendID, r.Name)
	}
	if !strings.Contains(r.Paths.Item, "{id}") {
		return fmt.Errorf("backend %q resource %q: paths.item must contain {id}", backendID, r.Name)
	}
	for field, ft := range r.FieldTemplates {
		if ft.Template == "" && ft.TemplateFile == "" {
			return fmt.Errorf("backend %q resource %q field %q: template or template_file required", backendID, r.Name, field)
		}
	}
	return nil
}

func (d DiscoverySection) Validate(backendID string) error {
	if _, err := url.ParseRequestURI(d.URL); err != nil {
		return fmt.Errorf("backend %q discovery.url %q invalid: %w", backendID, d.URL, err)
	}
	if _, err := time.ParseDuration(d.PollInterval); err != nil {
		return fmt.Errorf("backend %q discovery.poll_interval %q invalid: %w", backendID, d.PollInterval, err)
	}
	if d.Auth != nil {
		if err := d.Auth.Validate(true); err != nil {
			return fmt.Errorf("backend %q discovery.auth: %w", backendID, err)
		}
	}
	return nil
}

func (a *AuthScheme) WithDefaults() *AuthScheme {
	if a.Type == "" {
		a.Type = DefaultAuthType
	}
	return a
}

// Validate checks the auth scheme. allowValue permits the Value secret
// (true only for discovery.auth).
func (a AuthScheme) Validate(allowValue bool) error {
	switch a.Type {
	case "none":
	case "apikey":
		if a.In != "header" && a.In != "query" {
			return fmt.Errorf("apikey auth requires in: header|query")
		}
		if strings.TrimSpace(a.Name) == "" {
			return fmt.Errorf("apikey auth requires name")
		}
	case "bearer", "basic", "oauth":
		return fmt.Errorf("auth type %q is not yet supported", a.Type)
	default:
		return fmt.Errorf("unknown auth type %q (want none|apikey|bearer|basic|oauth)", a.Type)
	}
	if !allowValue && a.Value != "" {
		return fmt.Errorf("auth.value is not allowed here (credentials are stored in the browser); only discovery.auth may set value")
	}
	return nil
}

func deriveBackendID(baseURL string) string {
	sum := sha256.Sum256([]byte(baseURL))
	return "b-" + hex.EncodeToString(sum[:])[:8]
}

func titleize(s string) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "-", " "), "_", " ")
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

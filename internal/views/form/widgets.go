package form

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*.html.tmpl
var builtinFS embed.FS

// FieldContext is the data passed to a custom field-widget template. Built-in
// templates and user-supplied templates both receive this.
type FieldContext struct {
	Name       string
	Label      string
	InputID    string
	Value      string
	Required   bool
	Args       map[string]any
	OptionsURL string // relation widgets: admin endpoint returning [{value,label}]
}

// Arg returns a string arg with a fallback default.
func (c FieldContext) Arg(key, def string) string {
	if v, ok := c.Args[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return def
}

// WidgetRegistry holds built-in and user-supplied field-widget templates.
type WidgetRegistry struct {
	templates map[string]*template.Template
}

// NewWidgetRegistry parses the embedded built-in widgets plus any user-supplied
// template files. Each user file is keyed by the exact path string from config
// (resolved against baseDir when relative), so a field's Custom.TemplateFile is
// its widget id. Fails fast on parse errors.
func NewWidgetRegistry(baseDir string, userFiles []string) (*WidgetRegistry, error) {
	reg := &WidgetRegistry{templates: map[string]*template.Template{}}

	entries, err := builtinFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("read built-in templates: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		id := strings.TrimSuffix(name, ".html.tmpl")
		raw, err := builtinFS.ReadFile("templates/" + name)
		if err != nil {
			return nil, err
		}
		t, err := template.New(id).Parse(string(raw))
		if err != nil {
			return nil, fmt.Errorf("parse built-in widget %q: %w", id, err)
		}
		reg.templates[id] = t
	}

	for _, rel := range userFiles {
		path := rel
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		// #nosec G304 -- template path is operator config, not request-derived.
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read template_file %q: %w", rel, err)
		}
		t, err := template.New(rel).Parse(string(raw))
		if err != nil {
			return nil, fmt.Errorf("parse template_file %q: %w", rel, err)
		}
		reg.templates[rel] = t
	}
	return reg, nil
}

// Has reports whether a widget id is registered.
func (r *WidgetRegistry) Has(id string) bool {
	_, ok := r.templates[id]
	return ok
}

// Render executes a widget template and returns its HTML. Record data passed in
// the context is contextually auto-escaped by html/template; only the operator-
// authored template markup is trusted.
func (r *WidgetRegistry) Render(id string, ctx FieldContext) (template.HTML, error) {
	t, ok := r.templates[id]
	if !ok {
		return "", fmt.Errorf("unknown field widget %q", id)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("render widget %q: %w", id, err)
	}
	return template.HTML(buf.String()), nil //nolint:gosec // operator template output
}

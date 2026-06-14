package form

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/a-h/templ"
)

// FieldContext is the data passed to a field widget (built-in Templ component or
// user-supplied html/template).
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

// builtinWidgets maps a widget id to its Templ component. Only datetime +
// relation are real; the rest fall back to a plain input/textarea stub.
var builtinWidgets = map[string]func(FieldContext) templ.Component{
	"datetime":    widgetDatetime,
	"relation":    widgetRelation,
	"code":        widgetCode,
	"markdown":    widgetMarkdown,
	"json":        widgetJSON,
	"tags":        widgetTags,
	"color":       widgetColor,
	"geo":         widgetStubTextarea,
	"richtext":    widgetStubTextarea,
	"file":        widgetStubInput,
	"image":       widgetStubInput,
	"multiselect": widgetStubInput,
	"duration":    widgetStubInput,
	"rating":      widgetStubInput,
	"slug":        widgetStubInput,
	"currency":    widgetStubInput,
	"percent":     widgetStubInput,
	"phone":       widgetStubInput,
	"email":       widgetStubInput,
	"url":         widgetStubInput,
	"password":    widgetStubInput,
	"country":     widgetStubInput,
	"timezone":    widgetStubInput,
}

// WidgetRegistry resolves a widget id to a Templ component. Built-ins are
// compiled Templ; user-supplied template_files are runtime html/template,
// wrapped as templ.Raw (record data is contextually escaped by html/template).
type WidgetRegistry struct {
	user map[string]*template.Template
}

// NewWidgetRegistry parses any user-supplied template files (keyed by the exact
// config path, resolved against baseDir when relative). Built-ins need no setup.
func NewWidgetRegistry(baseDir string, userFiles []string) (*WidgetRegistry, error) {
	reg := &WidgetRegistry{user: map[string]*template.Template{}}
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
		reg.user[rel] = t
	}
	return reg, nil
}

// Has reports whether a widget id is registered (built-in or user-supplied).
func (r *WidgetRegistry) Has(id string) bool {
	if _, ok := builtinWidgets[id]; ok {
		return true
	}
	_, ok := r.user[id]
	return ok
}

// Component returns the rendered widget for id, or ok=false to fall back to the
// default input.
func (r *WidgetRegistry) Component(id string, ctx FieldContext) (templ.Component, bool) {
	if fn, ok := builtinWidgets[id]; ok {
		return fn(ctx), true
	}
	if t, ok := r.user[id]; ok {
		var buf bytes.Buffer
		if err := t.Execute(&buf, ctx); err != nil {
			return nil, false
		}
		return templ.Raw(buf.String()), true
	}
	return nil, false
}

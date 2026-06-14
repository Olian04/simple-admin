package form

import (
	"strings"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// FieldContext is the single data value passed to EVERY field widget — both
// built-in Templ components and user-supplied html/template files get exactly
// this (plus their configured Args). It carries the field's own schema metadata
// (Field) and a queryable view of its sibling fields, so a widget can work on
// its own data and look up related fields (e.g. a boolean flag or an id).
type FieldContext struct {
	Name       string // field name (== Field.Name)
	Label      string
	InputID    string
	Value      string
	Required   bool
	Args       map[string]any
	OptionsURL string // relation widgets: admin endpoint returning [{value,label}]

	Field    resource.Field // this field's full schema metadata (kind, format, enum, …)
	Siblings []Sibling      // the other editable fields on the same record + values
}

// Sibling is another field on the same record, with its current value.
type Sibling struct {
	Field resource.Field
	Value string
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

// Sibling returns the named sibling, if present.
func (c FieldContext) Sibling(name string) (Sibling, bool) {
	for _, s := range c.Siblings {
		if s.Field.Name == name {
			return s, true
		}
	}
	return Sibling{}, false
}

// HasSibling reports whether a sibling with the given name exists (template-friendly).
func (c FieldContext) HasSibling(name string) bool {
	_, ok := c.Sibling(name)
	return ok
}

// SiblingValue returns the named sibling's value, or "" if absent (template-friendly).
func (c FieldContext) SiblingValue(name string) string {
	if s, ok := c.Sibling(name); ok {
		return s.Value
	}
	return ""
}

// SiblingsOfKind returns siblings whose field kind matches (e.g. "boolean", "enum").
func (c FieldContext) SiblingsOfKind(kind string) []Sibling {
	var out []Sibling
	for _, s := range c.Siblings {
		if string(s.Field.Kind) == kind {
			out = append(out, s)
		}
	}
	return out
}

// BoolSiblings returns the boolean siblings (e.g. feature flags).
func (c FieldContext) BoolSiblings() []Sibling {
	return c.SiblingsOfKind(string(resource.KindBoolean))
}

// IDSiblings returns siblings that look like identifiers / references
// (name ends in "id"), e.g. foreign keys.
func (c FieldContext) IDSiblings() []Sibling {
	var out []Sibling
	for _, s := range c.Siblings {
		if strings.HasSuffix(strings.ToLower(s.Field.Name), "id") {
			out = append(out, s)
		}
	}
	return out
}

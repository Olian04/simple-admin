package ui

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Olian04/webui/pkg/webui"

	"github.com/Olian04/simple-admin/internal/backend"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

type rec = resource.Record

// accOpts say where an accessor is used. A column is read-only and never
// filtered (the upstream contract has no filtering); a form field is editable
// unless the schema marks it read-only.
type accOpts struct {
	editable bool
	column   bool
	first    bool // first table column: a RowClick owns it, so it cannot be a URL
}

func tone(t string) webui.Tone {
	switch t {
	case "ok":
		return webui.ToneOK
	case "warn":
		return webui.ToneWarning
	case "bad":
		return webui.ToneCritical
	default:
		return webui.ToneNeutral
	}
}

func one(v any) []string {
	if s := str(v); s != "" {
		return []string{s}
	}
	return nil
}

func first(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// accessor maps one schema field to the webui accessor that shows (and, in a
// form, edits) it. label is the field's unique label within its table or form.
func (b *builder) accessor(rt *resource.Runtime, f resource.Field, label string, o accOpts) webui.Accessor[rec] {
	editable := o.editable && !f.ReadOnly
	name := f.Name
	noFilter := o.column

	var store func(*rec, string)
	if editable {
		store = func(m *rec, v string) {
			if v == "" {
				set(m, name, nil)
				return
			}
			set(m, name, v)
		}
	}

	if f.Hint != nil {
		switch f.Hint.Widget {
		case resource.WidgetSlider:
			return b.slider(f, label, editable, noFilter)
		case resource.WidgetSelect:
			return b.hintSelect(f, label, editable, noFilter)
		case resource.WidgetRelation:
			if acc, ok := b.relation(rt, f, label, editable, noFilter); ok {
				return acc
			}
		}
	} else if f.Kind == resource.KindString && f.Format == "" && f.Name != rt.Descriptor.IDField {
		if acc, ok := b.relation(rt, f, label, editable, noFilter); ok {
			return acc
		}
	}

	switch f.Kind {
	case resource.KindBoolean:
		bl := webui.Bool[rec]{
			Label: label, Help: f.Description, NoFilter: noFilter,
			Load: func(m rec) *bool {
				if v, ok := m[name].(bool); ok {
					return &v
				}
				return nil
			},
		}
		if editable {
			bl.Required = f.Required
			bl.Store = func(m *rec, v *bool) {
				if v == nil {
					set(m, name, nil)
					return
				}
				set(m, name, *v)
			}
		}
		return bl

	case resource.KindEnum:
		opts := make([]webui.Option, 0, len(f.Enum))
		for _, e := range f.Enum {
			if e != "" {
				opts = append(opts, webui.Option{Value: e})
			}
		}
		sel := webui.Select[rec]{
			Label: label, Help: f.Description, Options: opts, NoFilter: noFilter,
			Load: func(m rec) []string { return one(m[name]) },
		}
		if editable {
			sel.Required = f.Required
			sel.Store = func(m *rec, v []string) {
				if s := first(v); s != "" {
					set(m, name, s)
				} else {
					set(m, name, nil)
				}
			}
		}
		return sel

	case resource.KindInteger:
		in := webui.OptionalInt[rec]{
			Label: label, Help: f.Description, NoFilter: noFilter,
			Load: func(m rec) *int {
				n, ok := num(m[name])
				if !ok {
					return nil
				}
				v := int(n)
				return &v
			},
		}
		if editable {
			in.Rules = webui.NumberRules[int]{Required: f.Required, Min: intPtr(f.Min), Max: intPtr(f.Max)}
			in.Store = func(m *rec, n *int) {
				if n == nil {
					set(m, name, nil)
					return
				}
				set(m, name, float64(*n))
			}
		}
		return in

	case resource.KindNumber:
		fl := webui.OptionalFloat[rec]{
			Label: label, Help: f.Description, NoFilter: noFilter,
			Load: func(m rec) *float64 {
				n, ok := num(m[name])
				if !ok {
					return nil
				}
				return &n
			},
		}
		if editable {
			fl.Rules = webui.NumberRules[float64]{Required: f.Required, Min: f.Min, Max: f.Max}
			fl.Store = func(m *rec, n *float64) {
				if n == nil {
					set(m, name, nil)
					return
				}
				set(m, name, *n)
			}
		}
		return fl

	case resource.KindArray:
		scalar := f.Items != nil && f.Items.Kind != resource.KindArray && f.Items.Kind != resource.KindObject
		s := webui.String[rec]{Label: label, Help: f.Description, NoFilter: noFilter, NoSort: true}
		if scalar {
			s.Load = func(m rec) string { return listText(m[name]) }
			s.Placeholder = "comma, separated, values"
			if editable {
				item := f.Items.Kind
				s.Store = func(m *rec, v string) { set(m, name, parseList(v, item)) }
			}
		} else {
			s.Load = func(m rec) string { return str(m[name]) }
			s.Multiline = true
			s.Placeholder = "JSON"
			if editable {
				s.Store = func(m *rec, v string) { set(m, name, parseJSON(v)) }
			}
		}
		return s

	case resource.KindObject:
		s := webui.String[rec]{
			Label: label, Help: f.Description, NoFilter: noFilter, NoSort: true, Placeholder: "JSON", Multiline: true,
			Load: func(m rec) string { return str(m[name]) },
		}
		if editable {
			s.Store = func(m *rec, v string) { set(m, name, parseJSON(v)) }
		}
		return s
	}

	// strings
	switch f.Format {
	case "date-time", "date":
		dt := webui.Datetime[rec]{
			Label: label, Help: f.Description, NoFilter: noFilter, DateOnly: f.Format == "date",
			Load: func(m rec) string { return str(m[name]) },
		}
		if editable {
			dt.Store = func(m *rec, v string) {
				if v == "" {
					set(m, name, nil)
					return
				}
				set(m, name, v)
			}
		}
		return dt
	case "email":
		em := webui.Email[rec]{
			Label: label, Help: f.Description, NoFilter: noFilter,
			Load: func(m rec) string { return str(m[name]) },
		}
		if editable {
			em.Rules = webui.StringRules{Required: f.Required}
			em.Store = store
		}
		return em
	case "color":
		// A color picker cannot be empty, so an optional color stays text.
		if !editable || f.Required {
			cp := webui.ColorPicker[rec]{
				Label: label, Help: f.Description, NoFilter: noFilter,
				Load: func(m rec) string { return str(m[name]) },
			}
			if editable {
				cp.Store = store
			}
			return cp
		}
	case "uri":
		if !editable && !o.first {
			return webui.URL[rec]{
				Label: label, NoFilter: noFilter,
				Load: func(m rec) string { return str(m[name]) },
			}
		}
		if o.first {
			break // a RowClick owns the first column, which then cannot be a link
		}
		ws := webui.Website[rec]{
			Label: label, Help: f.Description, NoFilter: noFilter,
			Load: func(m rec) string { return str(m[name]) },
		}
		if editable {
			ws.Rules = webui.StringRules{Required: f.Required}
			ws.Store = store
		}
		return ws
	}

	s := webui.String[rec]{
		Label: label, Help: f.Description, NoFilter: noFilter,
		Load:  func(m rec) string { return str(m[name]) },
		Store: store,
	}
	if longText(f) {
		s.Multiline = true
	}
	if editable {
		s.Rules = webui.StringRules{Required: f.Required}
		if f.MaxLength != nil {
			s.Rules.MaxLen = *f.MaxLength
		}
		// A schema pattern is arbitrary ECMA-262; webui takes only the portable
		// ones (a text area takes none). The JSON Schema validator still checks
		// the rest on the server.
		if f.Pattern != "" && !s.Multiline && webui.ValidatePattern(f.Pattern) == nil {
			s.Rules.Pattern = &webui.PatternRule{Expr: f.Pattern, Message: "must match " + f.Pattern}
		}
	}
	return s
}

// proseNames are field names that hold paragraphs, which a schema with no
// length bound cannot say by itself.
var proseNames = map[string]bool{
	"body": true, "description": true, "content": true, "text": true,
	"notes": true, "note": true, "summary": true, "bio": true, "comment": true, "message": true,
}

// longText reports whether a plain string field is prose rather than a short
// value: a bound beyond a single line's worth, or a name that holds paragraphs.
func longText(f resource.Field) bool {
	if f.Format != "" || f.Kind != resource.KindString {
		return false
	}
	if f.MaxLength != nil {
		return *f.MaxLength > 200
	}
	return proseNames[strings.ToLower(f.Name)]
}

func intPtr(f *float64) *int {
	if f == nil {
		return nil
	}
	n := int(*f)
	return &n
}

func (b *builder) slider(f resource.Field, label string, editable, noFilter bool) webui.Accessor[rec] {
	name := f.Name
	sl := webui.OptionalSlider[rec]{
		Label: label, Help: f.Description, Min: f.Hint.Min, Max: f.Hint.Max, StepSize: f.Hint.Step, NoFilter: noFilter,
		Load: func(m rec) *float64 {
			n, ok := num(m[name])
			if !ok {
				return nil
			}
			return &n
		},
	}
	if editable {
		sl.Store = func(m *rec, n *float64) {
			if n == nil {
				set(m, name, nil)
				return
			}
			set(m, name, *n)
		}
	}
	return sl
}

func (b *builder) hintSelect(f resource.Field, label string, editable, noFilter bool) webui.Accessor[rec] {
	name := f.Name
	opts := make([]webui.Option, 0, len(f.Hint.Options))
	for _, o := range f.Hint.Options {
		opts = append(opts, webui.Option{Value: o.Value, Label: o.Label, Tone: tone(o.Tone)})
	}
	sel := webui.Select[rec]{
		Label: label, Help: f.Description, Options: opts, NoFilter: noFilter,
		Load: func(m rec) []string { return one(m[name]) },
	}
	if editable {
		sel.Required = f.Required
		sel.Store = func(m *rec, v []string) {
			if s := first(v); s != "" {
				set(m, name, s)
			} else {
				set(m, name, nil)
			}
		}
	}
	return sel
}

// relation makes a Select of the records of another resource on the same
// backend, chosen by a hint or by the field's name ("author", "author_id" →
// "authors"). ok is false when there is no such resource.
func (b *builder) relation(rt *resource.Runtime, f resource.Field, label string, editable, noFilter bool) (webui.Accessor[rec], bool) {
	target, labelField, ok := b.relationTarget(rt, f)
	if !ok {
		return nil, false
	}
	name := f.Name
	sel := webui.Select[rec]{
		Label: label, Help: f.Description, NoFilter: noFilter,
		Load: func(m rec) []string { return one(m[name]) },
		LoadOptions: func(ctx context.Context) ([]webui.Option, error) {
			return b.relationOptions(ctx, target, labelField)
		},
	}
	if editable {
		sel.Required = f.Required
		sel.Store = func(m *rec, v []string) {
			if s := first(v); s != "" {
				set(m, name, s)
			} else {
				set(m, name, nil)
			}
		}
	}
	return sel, true
}

func (b *builder) relationTarget(rt *resource.Runtime, f resource.Field) (*resource.Runtime, string, bool) {
	backendID := rt.Descriptor.Backend.ID
	var target *resource.Runtime
	labelField := ""
	if f.Hint != nil && f.Hint.Widget == resource.WidgetRelation {
		labelField = f.Hint.LabelField
		if f.Hint.Resource != "" {
			t, ok := b.reg.Lookup(backendID, f.Hint.Resource)
			if !ok {
				b.log.Warn("relation hint names an unknown resource", "backend", backendID, "field", f.Name, "resource", f.Hint.Resource)
				return nil, "", false
			}
			target = t
		}
	}
	if target == nil {
		target = matchResource(f.Name, b.reg.ResourcesFor(backendID))
	}
	if target == nil || !target.Descriptor.Capabilities.List {
		return nil, "", false
	}
	if labelField == "" {
		labelField = firstStringField(target.Descriptor)
	}
	return target, labelField, true
}

// matchResource finds the resource a field name likely refers to:
// "author", "author_id" and "authorId" all name "authors".
func matchResource(fieldName string, candidates []*resource.Runtime) *resource.Runtime {
	name := strings.ToLower(fieldName)
	base := strings.TrimSuffix(strings.TrimSuffix(name, "id"), "_")
	for _, c := range candidates {
		rl := strings.ToLower(c.Descriptor.Name)
		singular := strings.TrimSuffix(rl, "s")
		if name == rl || name == singular || base == rl || base == singular {
			return c
		}
	}
	return nil
}

// firstStringField is the field of a resource that labels its records in a
// relation picker: the first writable plain string.
func firstStringField(d resource.ResourceDescriptor) string {
	for _, f := range d.Schema.Fields {
		if f.Name == d.IDField || f.ReadOnly {
			continue
		}
		if f.Kind == resource.KindString && f.Format == "" {
			return f.Name
		}
	}
	return ""
}

const (
	relationLimit    = 200
	relationCacheTTL = 10 * time.Second
)

// relationOptions lists the first records of the target as options, cached for
// a few seconds because every page that draws, checks or exports the select
// asks for them.
func (b *builder) relationOptions(ctx context.Context, target *resource.Runtime, labelField string) ([]webui.Option, error) {
	d := target.Descriptor
	key := d.Key() + "\x00" + labelField + "\x00" + b.sessions.Credential(ctx, d.Backend.ID)
	if opts, ok := b.cache.get(key); ok {
		return opts, nil
	}
	res, err := target.Service.List(b.scope(ctx, d.Backend.ID), resource.ListParams{PerPage: relationLimit})
	if errors.Is(err, backend.ErrUnauthorized) {
		// No options yet, and not an error: the page's own load then tells the
		// visitor, with a link, that a credential is needed. A failed options
		// read would hide that behind "Could not load".
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	opts := make([]webui.Option, 0, len(res.Items))
	for _, r := range res.Items {
		id := str(r[d.IDField])
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		label := id
		if labelField != "" {
			if v := str(r[labelField]); v != "" {
				label = v
			}
		}
		opts = append(opts, webui.Option{Value: id, Label: label})
	}
	b.cache.put(key, opts, relationCacheTTL)
	return opts, nil
}

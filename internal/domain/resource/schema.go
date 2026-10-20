package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// FieldsFromJSONSchema derives an ordered []Field from a JSON Schema object
// (raw bytes). Property declaration order is preserved via a token-stream pass,
// since Go maps are unordered. This is the single owner of schema→field
// structure; sources populate Schema.Fields with it and ui/render consumes it.
func FieldsFromJSONSchema(raw []byte) ([]Field, error) {
	var doc struct {
		Required   []string        `json:"required"`
		Properties json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse json schema: %w", err)
	}
	if len(doc.Properties) == 0 {
		return nil, nil
	}

	required := map[string]bool{}
	for _, r := range doc.Required {
		required[r] = true
	}

	order, props, err := orderedProperties(doc.Properties)
	if err != nil {
		return nil, err
	}

	fields := make([]Field, 0, len(order))
	for _, name := range order {
		f, err := fieldFromProp(name, props[name], required[name])
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, nil
}

// orderedProperties returns property names in declaration order plus a map of
// their raw schemas.
func orderedProperties(propsRaw json.RawMessage) ([]string, map[string]json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(propsRaw))
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("read properties: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, nil, fmt.Errorf("properties must be an object")
	}

	var order []string
	props := map[string]json.RawMessage{}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, nil, fmt.Errorf("property key not a string")
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, nil, err
		}
		order = append(order, key)
		props[key] = raw
	}
	return order, props, nil
}

type propSchema struct {
	Type        any             `json:"type"`
	Format      string          `json:"format"`
	Enum        []any           `json:"enum"`
	Minimum     *float64        `json:"minimum"`
	Maximum     *float64        `json:"maximum"`
	MaxLength   *int            `json:"maxLength"`
	Pattern     string          `json:"pattern"`
	ReadOnly    bool            `json:"readOnly"`
	Description string          `json:"description"`
	Items       json.RawMessage `json:"items"`
}

func fieldFromProp(name string, raw json.RawMessage, required bool) (Field, error) {
	var ps propSchema
	if err := json.Unmarshal(raw, &ps); err != nil {
		return Field{}, fmt.Errorf("parse property %q: %w", name, err)
	}

	f := Field{
		Name:        name,
		Label:       titleizeField(name),
		Format:      ps.Format,
		Required:    required,
		ReadOnly:    ps.ReadOnly,
		Min:         ps.Minimum,
		Max:         ps.Maximum,
		MaxLength:   ps.MaxLength,
		Pattern:     ps.Pattern,
		Description: ps.Description,
	}

	typ := firstType(ps.Type)
	switch {
	case len(ps.Enum) > 0:
		f.Kind = KindEnum
		f.Enum = stringifyEnum(ps.Enum)
	case typ == "boolean":
		f.Kind = KindBoolean
	case typ == "integer":
		f.Kind = KindInteger
	case typ == "number":
		f.Kind = KindNumber
	case typ == "array":
		f.Kind = KindArray
		if len(ps.Items) > 0 {
			item, err := fieldFromProp(name+"[]", ps.Items, false)
			if err != nil {
				return Field{}, err
			}
			f.Items = &item
		}
	case typ == "object":
		f.Kind = KindObject
	default:
		f.Kind = KindString
	}
	return f, nil
}

// firstType returns the schema "type", taking the first non-null entry when it
// is a list (e.g. ["string","null"]).
func firstType(t any) string {
	switch v := t.(type) {
	case string:
		return v
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "null" {
				return s
			}
		}
	}
	return ""
}

// titleizeField turns a field name into a human label: "publishedAt" ->
// "Published At", "id_field" -> "Id Field".
func titleizeField(name string) string {
	var b strings.Builder
	prevLower := false
	for i, r := range name {
		switch {
		case r == '_' || r == '-':
			b.WriteByte(' ')
			prevLower = false
			continue
		case unicode.IsUpper(r) && prevLower:
			b.WriteByte(' ')
		}
		if i == 0 {
			r = unicode.ToUpper(r)
		}
		b.WriteRune(r)
		prevLower = unicode.IsLower(r)
	}
	// Collapse to a clean spaced title.
	return strings.Join(strings.Fields(b.String()), " ")
}

func stringifyEnum(values []any) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		switch s := v.(type) {
		case string:
			out = append(out, s)
		case float64:
			out = append(out, strconv.FormatFloat(s, 'f', -1, 64))
		case bool:
			out = append(out, strconv.FormatBool(s))
		default:
			out = append(out, fmt.Sprintf("%v", v))
		}
	}
	return out
}

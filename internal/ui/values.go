package ui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// str renders a JSON-decoded value as text. Booleans read "true"/"false"; nil
// and values with no text form read "".
func str(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case json.Number:
		return val.String()
	case map[string]any, []any:
		raw, err := json.Marshal(val)
		if err != nil {
			return ""
		}
		return string(raw)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// num reads a JSON number (or numeric text) from a record value.
func num(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// set writes v under name, or removes the key when v is nil, so an emptied
// optional input is left out of the request as the old form did.
func set(m *resource.Record, name string, v any) {
	if *m == nil {
		*m = resource.Record{}
	}
	if v == nil {
		delete(*m, name)
		return
	}
	(*m)[name] = v
}

// listText renders an array of scalars as comma separated text.
func listText(v any) string {
	items, ok := v.([]any)
	if !ok {
		return str(v)
	}
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, str(it))
	}
	return strings.Join(parts, ", ")
}

// parseList splits comma separated text into scalars of the item kind.
func parseList(text string, item resource.FieldKind) any {
	var out []any
	for _, part := range strings.Split(text, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		switch item {
		case resource.KindNumber, resource.KindInteger:
			if f, err := strconv.ParseFloat(part, 64); err == nil {
				out = append(out, f)
				continue
			}
		case resource.KindBoolean:
			if b, err := strconv.ParseBool(part); err == nil {
				out = append(out, b)
				continue
			}
		}
		out = append(out, part)
	}
	if out == nil {
		return nil
	}
	return out
}

// parseJSON reads text as JSON, falling back to the text itself so the schema
// validator can report it against the field.
func parseJSON(text string) any {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var v any
	if json.Unmarshal([]byte(text), &v) == nil {
		return v
	}
	return text
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

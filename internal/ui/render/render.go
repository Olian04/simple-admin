// Package render turns a resource's schema + data into UI models (table
// columns, form fields) and display/input helpers. It is pure Go and
// transport-agnostic; Templ components consume what it produces.
package render

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// Column is one table column in the list view.
type Column struct {
	Name  string
	Label string
}

// Columns returns the list-view columns: top-level scalar/enum fields.
// Arrays and objects are omitted (rendered as a badge/summary elsewhere).
func Columns(fields []resource.Field) []Column {
	var cols []Column
	for _, f := range fields {
		if f.Kind == resource.KindArray || f.Kind == resource.KindObject {
			continue
		}
		cols = append(cols, Column{Name: f.Name, Label: f.Label})
	}
	return cols
}

// FormField is an editable field seeded with its current value + any error.
type FormField struct {
	resource.Field
	Value any
	Error string
}

// FormFields returns the editable fields for a create/edit form. Read-only
// fields (id, timestamps) are excluded; values and validation errors are seeded.
func FormFields(fields []resource.Field, data resource.Record, errs map[string]string) []FormField {
	out := make([]FormField, 0, len(fields))
	for _, f := range fields {
		if f.ReadOnly {
			continue
		}
		out = append(out, FormField{Field: f, Value: data[f.Name], Error: errs[f.Name]})
	}
	return out
}

// Cell renders a record value for a table cell.
func Cell(rec resource.Record, col Column) string {
	return DisplayValue(rec[col.Name])
}

// DisplayValue formats a JSON-decoded value for display.
func DisplayValue(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case bool:
		if val {
			return "✓"
		}
		return "✗"
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case json.Number:
		return val.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ValueString renders a field value for a form input's value attribute.
func ValueString(v any) string {
	if v == nil {
		return ""
	}
	return DisplayValue(v)
}

// InputType maps a field to an HTML <input type> for the built-in text-ish
// inputs. Enum/boolean/array render via dedicated controls, not this.
func InputType(f resource.Field) string {
	if f.Kind == resource.KindNumber || f.Kind == resource.KindInteger {
		return "number"
	}
	switch f.Format {
	case "email":
		return "email"
	case "uri":
		return "url"
	case "password":
		return "password"
	case "date":
		return "date"
	case "date-time":
		return "datetime-local"
	case "color":
		return "color"
	default:
		return "text"
	}
}

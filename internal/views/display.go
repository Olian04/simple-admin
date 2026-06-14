package views

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

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

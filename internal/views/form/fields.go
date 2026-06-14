// Package form is the form-rendering package: the create/edit form Templ
// component plus its view models, field helpers, and the field-widget registry
// (built-in widgets + user-supplied template files).
package form

import (
	"html/template"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

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

// FormFieldView is one input in a create/edit form.
type FormFieldView struct {
	Name       string
	Label      string
	Kind       string // string|number|integer|boolean|enum|array|object
	InputType  string // HTML input type for text-ish kinds
	Value      string
	Checked    bool // for boolean
	Enum       []string
	Required   bool
	Error      string
	CustomHTML template.HTML // pre-rendered custom widget; when set, overrides the kind switch
}

// FormView is the data for a create/edit form.
type FormView struct {
	Backend   resource.Backend
	Name      string
	Label     string
	Mode      string // create|edit
	ID        string
	Action    string // admin route the form submits to
	UsePut    bool   // edit submits via hx-put; create via hx-post
	CancelURL string
	FormError string // non-field-level error
	Fields    []FormFieldView
}

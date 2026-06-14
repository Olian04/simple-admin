package form

import (
	"strings"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// AutoSelectWidget picks a built-in widget for a field by convention when no
// explicit field template is configured. It returns the widget id and, for the
// relation widget, the target resource name. ok is false when no convention
// matches (the field renders with its default built-in input).
func AutoSelectWidget(f resource.Field, knownResources []string) (id, target string, ok bool) {
	// Date/time fields get the datetime widget.
	switch f.Format {
	case "date", "date-time", "time":
		return "datetime", "", true
	}
	// Fields that reference another resource on the same backend get a relation
	// picker: "author" -> "authors", "authorId" -> "authors", etc.
	if t, matched := relationTarget(f.Name, knownResources); matched {
		return "relation", t, true
	}
	return "", "", false
}

// relationTarget matches a field name to a known resource it likely references.
func relationTarget(fieldName string, knownResources []string) (string, bool) {
	name := strings.ToLower(fieldName)
	base := strings.TrimSuffix(strings.TrimSuffix(name, "id"), "_")
	for _, r := range knownResources {
		rl := strings.ToLower(r)
		singular := strings.TrimSuffix(rl, "s")
		if name == rl || name == singular || base == rl || base == singular {
			return r, true
		}
	}
	return "", false
}

// DatetimeInputType maps the datetime widget's mode/format to an HTML input type.
func DatetimeInputType(mode string) string {
	switch mode {
	case "date":
		return "date"
	case "time":
		return "time"
	default:
		return "datetime-local"
	}
}

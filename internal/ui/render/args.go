package render

import (
	"github.com/ohler55/ojg/jp"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// ResolveArgs resolves a field template's configured args against a record:
// static values pass through; {jsonpath: "$.x"} selectors are evaluated against
// the record (first match wins).
func ResolveArgs(spec map[string]resource.TemplateArg, record resource.Record) map[string]any {
	if len(spec) == 0 {
		return nil
	}
	out := make(map[string]any, len(spec))
	for key, arg := range spec {
		if arg.JSONPath != "" {
			if expr, err := jp.ParseString(arg.JSONPath); err == nil {
				if vals := expr.Get(map[string]any(record)); len(vals) > 0 {
					out[key] = vals[0]
				}
			}
			continue
		}
		out[key] = arg.Static
	}
	return out
}

// Package openapi generates an OpenAPI 3 spec from PocketBase collection
// schemas, mapping each base collection to its real CRUD REST endpoints so the
// simple-admin UI can drive it as a compliant backend.
package openapi

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// Generate builds an OpenAPI 3.0 document (as a JSON-serialisable map) from the
// app's base collections. When advertiseAPIKey is true, an apiKey security
// scheme (header X-API-Key) is declared and applied to every operation.
func Generate(app core.App, advertiseAPIKey bool) (map[string]any, error) {
	collections, err := app.FindAllCollections("base")
	if err != nil {
		return nil, err
	}

	schemas := map[string]any{}
	paths := map[string]any{}

	var security []any
	securitySchemes := map[string]any{}
	if advertiseAPIKey {
		securitySchemes["ApiKeyAuth"] = map[string]any{
			"type": "apiKey",
			"in":   "header",
			"name": "X-API-Key",
		}
		security = []any{map[string]any{"ApiKeyAuth": []any{}}}
	}

	for _, c := range collections {
		if c.System {
			continue
		}

		schemaName := schemaNameFor(c.Name)
		schemas[schemaName] = collectionSchema(c)

		ref := map[string]any{"$ref": "#/components/schemas/" + schemaName}
		collectionPath := "/api/collections/" + c.Name + "/records"
		itemPath := collectionPath + "/{id}"

		paths[collectionPath] = map[string]any{
			"get": op("List "+c.Name, listResponse(ref), nil, security),
			"post": op("Create "+c.Name, ref, requestBody(ref), security),
		}
		paths[itemPath] = map[string]any{
			"parameters": []any{idParam()},
			"get":        op("Get "+c.Name, ref, nil, security),
			"patch":      op("Update "+c.Name, ref, requestBody(ref), security),
			"delete":     deleteOp("Delete "+c.Name, security),
		}
	}

	components := map[string]any{"schemas": schemas}
	if advertiseAPIKey {
		components["securitySchemes"] = securitySchemes
	}

	doc := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "simple-admin example backend",
			"version":     "0.1.0",
			"description": "Reference compliant backend (PocketBase) for simple-admin.",
		},
		"servers":    []any{map[string]any{"url": "/"}},
		"paths":      paths,
		"components": components,
	}
	if advertiseAPIKey {
		doc["security"] = security
	}

	return doc, nil
}

func collectionSchema(c *core.Collection) map[string]any {
	properties := map[string]any{}
	var required []any

	for _, f := range c.Fields {
		if f.GetHidden() {
			continue
		}
		prop, readOnly := fieldSchema(f)
		if readOnly {
			prop["readOnly"] = true
		}
		properties[f.GetName()] = prop

		if isRequired(f) && !readOnly {
			required = append(required, f.GetName())
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// fieldSchema maps a PocketBase field to a JSON Schema fragment and reports
// whether it is read-only (system/autodate fields the client cannot set).
func fieldSchema(f core.Field) (map[string]any, bool) {
	switch field := f.(type) {
	case *core.TextField:
		s := map[string]any{"type": "string"}
		if field.Max > 0 {
			s["maxLength"] = field.Max
		}
		return s, field.System
	case *core.EditorField:
		return map[string]any{"type": "string", "x-field": "richtext"}, field.System
	case *core.EmailField:
		return map[string]any{"type": "string", "format": "email"}, field.System
	case *core.URLField:
		return map[string]any{"type": "string", "format": "uri"}, field.System
	case *core.NumberField:
		s := map[string]any{"type": "number"}
		if field.Min != nil {
			s["minimum"] = *field.Min
		}
		if field.Max != nil {
			s["maximum"] = *field.Max
		}
		return s, field.System
	case *core.BoolField:
		return map[string]any{"type": "boolean"}, field.System
	case *core.DateField:
		return map[string]any{"type": "string", "format": "date-time"}, field.System
	case *core.AutodateField:
		return map[string]any{"type": "string", "format": "date-time"}, true
	case *core.SelectField:
		enum := make([]any, len(field.Values))
		for i, v := range field.Values {
			enum[i] = v
		}
		if field.MaxSelect > 1 {
			return map[string]any{"type": "array", "items": map[string]any{"type": "string", "enum": enum}}, field.System
		}
		return map[string]any{"type": "string", "enum": enum}, field.System
	case *core.RelationField:
		if field.MaxSelect > 1 {
			return map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, field.System
		}
		return map[string]any{"type": "string"}, field.System
	case *core.JSONField:
		return map[string]any{}, field.System
	default:
		// Unknown/system field types (incl. the system "id"): render as string.
		return map[string]any{"type": "string"}, f.GetSystem()
	}
}

func isRequired(f core.Field) bool {
	switch field := f.(type) {
	case *core.TextField:
		return field.Required
	case *core.EditorField:
		return field.Required
	case *core.EmailField:
		return field.Required
	case *core.URLField:
		return field.Required
	case *core.NumberField:
		return field.Required
	case *core.BoolField:
		return field.Required
	case *core.DateField:
		return field.Required
	case *core.SelectField:
		return field.Required
	case *core.RelationField:
		return field.Required
	case *core.JSONField:
		return field.Required
	default:
		return false
	}
}

func op(summary string, responseSchema map[string]any, body map[string]any, security []any) map[string]any {
	o := map[string]any{
		"summary": summary,
		"responses": map[string]any{
			"200": map[string]any{
				"description": "OK",
				"content": map[string]any{
					"application/json": map[string]any{"schema": responseSchema},
				},
			},
		},
	}
	if body != nil {
		o["requestBody"] = body
	}
	if security != nil {
		o["security"] = security
	}
	return o
}

func deleteOp(summary string, security []any) map[string]any {
	o := map[string]any{
		"summary": summary,
		"responses": map[string]any{
			"204": map[string]any{"description": "No Content"},
		},
	}
	if security != nil {
		o["security"] = security
	}
	return o
}

func requestBody(ref map[string]any) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{"schema": ref},
		},
	}
}

// listResponse describes PocketBase's paginated list envelope.
func listResponse(itemRef map[string]any) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"page":       map[string]any{"type": "integer"},
			"perPage":    map[string]any{"type": "integer"},
			"totalItems": map[string]any{"type": "integer"},
			"totalPages": map[string]any{"type": "integer"},
			"items":      map[string]any{"type": "array", "items": itemRef},
		},
	}
}

func idParam() map[string]any {
	return map[string]any{
		"name":     "id",
		"in":       "path",
		"required": true,
		"schema":   map[string]any{"type": "string"},
	}
}

func schemaNameFor(collection string) string {
	if collection == "" {
		return "Resource"
	}
	return strings.ToUpper(collection[:1]) + collection[1:]
}

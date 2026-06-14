package openapi_test

import (
	"testing"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/source/openapi"
)

const fixtureSpec = `{
  "openapi": "3.0.3",
  "paths": {
    "/api/collections/posts/records": {
      "get": {
        "responses": {"200": {"content": {"application/json": {"schema": {
          "type": "object",
          "properties": {"items": {"type": "array", "items": {"$ref": "#/components/schemas/Post"}}}
        }}}}}
      },
      "post": {
        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Post"}}}},
        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Post"}}}}}
      }
    },
    "/api/collections/posts/records/{id}": {
      "get": {"responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Post"}}}}}},
      "patch": {"responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Post"}}}}}},
      "delete": {"responses": {"204": {"description": "No Content"}}}
    }
  },
  "components": {
    "schemas": {
      "Post": {
        "type": "object",
        "required": ["title"],
        "properties": {
          "id": {"type": "string", "readOnly": true},
          "title": {"type": "string"},
          "status": {"type": "string", "enum": ["draft", "published"]}
        }
      }
    },
    "securitySchemes": {
      "ApiKeyAuth": {"type": "apiKey", "in": "header", "name": "X-API-Key"}
    }
  }
}`

func TestParseDerivesDescriptor(t *testing.T) {
	backend := resource.Backend{ID: "gw", Label: "GW", BaseURL: "http://h"}
	descs, err := openapi.Parse([]byte(fixtureSpec), backend)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(descs) != 1 {
		t.Fatalf("want 1 descriptor, got %d", len(descs))
	}
	d := descs[0]
	if d.Name != "post" || d.Label != "Post" {
		t.Fatalf("name/label wrong: %q/%q", d.Name, d.Label)
	}
	caps := d.Capabilities
	if !caps.List || !caps.Create || !caps.Read || !caps.Delete {
		t.Fatalf("capabilities missing: %+v", caps)
	}
	if !caps.Update || caps.UpdateVerb != "PATCH" {
		t.Fatalf("update verb wrong: %+v", caps)
	}
	if d.AuthScheme.Type != "apikey" || d.AuthScheme.In != "header" || d.AuthScheme.Name != "X-API-Key" {
		t.Fatalf("auth wrong: %+v", d.AuthScheme)
	}
	if d.CRUDPaths.Collection != "/api/collections/posts/records" || d.CRUDPaths.Item != "/api/collections/posts/records/{id}" {
		t.Fatalf("paths wrong: %+v", d.CRUDPaths)
	}
	if len(d.Schema.Fields) != 3 {
		t.Fatalf("want 3 fields, got %d", len(d.Schema.Fields))
	}
}

package manual_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/source/manual"
)

const schemaJSON = `{
  "type": "object",
  "required": ["title"],
  "properties": {
    "id": {"type":"string","readOnly":true},
    "title": {"type":"string","maxLength":200},
    "status": {"type":"string","enum":["draft","published"]},
    "body": {"type":"string"},
    "rating": {"type":"number"}
  }
}`

func TestManualSourceBuildsDescriptor(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "post.json")
	if err := os.WriteFile(schemaPath, []byte(schemaJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	backends := []config.BackendSection{{
		ID:      "example",
		Label:   "Example",
		BaseURL: "http://h",
		Auth:    &config.AuthScheme{Type: "apikey", In: "header", Name: "X-API-Key"},
		Resources: []config.ResourceSection{{
			Name:       "posts",
			Label:      "Posts",
			Schema:     config.SchemaSource{Path: schemaPath},
			Paths:      config.ResourcePaths{Collection: "/posts", Item: "/posts/{id}"},
			IDField:    "id",
			UpdateVerb: "PATCH",
			FieldTemplates: map[string]config.FieldTemplate{
				"body": {Template: "code"},
			},
		}},
	}}

	descs, err := manual.New(backends, dir).Descriptors(context.Background())
	if err != nil {
		t.Fatalf("descriptors: %v", err)
	}
	if len(descs) != 1 {
		t.Fatalf("want 1 descriptor, got %d", len(descs))
	}
	d := descs[0]
	if d.Name != "posts" || d.Backend.ID != "example" {
		t.Fatalf("descriptor identity wrong: %+v", d)
	}
	if !d.Capabilities.Create || d.Capabilities.UpdateVerb != "PATCH" {
		t.Fatalf("capabilities wrong: %+v", d.Capabilities)
	}
	if d.AuthScheme.Type != "apikey" || d.AuthScheme.Name != "X-API-Key" {
		t.Fatalf("auth wrong: %+v", d.AuthScheme)
	}

	// Field order preserved; title required; body carries custom template.
	if len(d.Schema.Fields) != 5 {
		t.Fatalf("want 5 fields, got %d", len(d.Schema.Fields))
	}
	if d.Schema.Fields[0].Name != "id" || !d.Schema.Fields[0].ReadOnly {
		t.Fatalf("id field wrong: %+v", d.Schema.Fields[0])
	}
	byName := map[string]int{}
	for i, f := range d.Schema.Fields {
		byName[f.Name] = i
	}
	title := d.Schema.Fields[byName["title"]]
	if !title.Required || title.MaxLength == nil || *title.MaxLength != 200 {
		t.Fatalf("title field wrong: %+v", title)
	}
	status := d.Schema.Fields[byName["status"]]
	if status.Kind != "enum" || len(status.Enum) != 2 {
		t.Fatalf("status enum wrong: %+v", status)
	}
	body := d.Schema.Fields[byName["body"]]
	if body.Custom == nil || body.Custom.TemplateID != "code" {
		t.Fatalf("body custom template wrong: %+v", body.Custom)
	}
}

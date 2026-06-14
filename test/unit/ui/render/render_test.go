package render_test

import (
	"testing"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/ui/render"
)

func fields() []resource.Field {
	maxLen := 200
	return []resource.Field{
		{Name: "id", Kind: resource.KindString, ReadOnly: true},
		{Name: "title", Kind: resource.KindString, Required: true, MaxLength: &maxLen},
		{Name: "status", Kind: resource.KindEnum, Enum: []string{"draft", "published"}},
		{Name: "rating", Kind: resource.KindNumber},
		{Name: "featured", Kind: resource.KindBoolean},
		{Name: "tags", Kind: resource.KindArray},
		{Name: "meta", Kind: resource.KindObject},
		{Name: "email", Kind: resource.KindString, Format: "email"},
	}
}

func TestColumnsExcludeArrayObject(t *testing.T) {
	cols := render.Columns(fields())
	names := map[string]bool{}
	for _, c := range cols {
		names[c.Name] = true
	}
	if names["tags"] || names["meta"] {
		t.Fatalf("array/object should not be columns: %+v", cols)
	}
	if !names["id"] || !names["title"] {
		t.Fatalf("scalar columns missing: %+v", cols)
	}
}

func TestFormFieldsExcludeReadOnlyAndSeed(t *testing.T) {
	data := resource.Record{"title": "Hello", "rating": float64(5)}
	errs := map[string]string{"title": "too short"}
	ff := render.FormFields(fields(), data, errs)
	for _, f := range ff {
		if f.Name == "id" {
			t.Fatal("read-only id must be excluded from form")
		}
		if f.Name == "title" {
			if f.Value != "Hello" || f.Error != "too short" {
				t.Fatalf("title not seeded: %+v", f)
			}
		}
	}
}

func TestDisplayValue(t *testing.T) {
	cases := map[any]string{
		nil:          "",
		true:         "✓",
		false:        "✗",
		"hi":         "hi",
		float64(5):   "5",
		float64(2.5): "2.5",
	}
	for in, want := range cases {
		if got := render.DisplayValue(in); got != want {
			t.Fatalf("DisplayValue(%v)=%q want %q", in, got, want)
		}
	}
}

func TestInputType(t *testing.T) {
	cases := []struct {
		f    resource.Field
		want string
	}{
		{resource.Field{Kind: resource.KindNumber}, "number"},
		{resource.Field{Kind: resource.KindInteger}, "number"},
		{resource.Field{Kind: resource.KindString, Format: "email"}, "email"},
		{resource.Field{Kind: resource.KindString, Format: "date-time"}, "datetime-local"},
		{resource.Field{Kind: resource.KindString}, "text"},
	}
	for _, c := range cases {
		if got := render.InputType(c.f); got != c.want {
			t.Fatalf("InputType(%+v)=%q want %q", c.f, got, c.want)
		}
	}
}

package render_test

import (
	"testing"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/ui/render"
)

func TestAutoSelectDatetime(t *testing.T) {
	id, _, ok := render.AutoSelectWidget(resource.Field{Name: "publishedAt", Format: "date-time"}, nil)
	if !ok || id != "datetime" {
		t.Fatalf("want datetime, got %q ok=%v", id, ok)
	}
}

func TestAutoSelectRelation(t *testing.T) {
	id, target, ok := render.AutoSelectWidget(
		resource.Field{Name: "author", Kind: resource.KindString},
		[]string{"posts", "authors"},
	)
	if !ok || id != "relation" || target != "authors" {
		t.Fatalf("want relation->authors, got id=%q target=%q ok=%v", id, target, ok)
	}
}

func TestAutoSelectNoneForPlainField(t *testing.T) {
	if _, _, ok := render.AutoSelectWidget(resource.Field{Name: "title", Kind: resource.KindString}, []string{"posts"}); ok {
		t.Fatal("plain field should not auto-select a widget")
	}
}

func TestResolveArgsStaticAndJSONPath(t *testing.T) {
	spec := map[string]resource.TemplateArg{
		"language": {Static: "json"},
		"max":      {JSONPath: "$.maxRating"},
	}
	rec := resource.Record{"maxRating": float64(10)}
	args := render.ResolveArgs(spec, rec)
	if args["language"] != "json" {
		t.Fatalf("static arg wrong: %#v", args["language"])
	}
	if args["max"] != float64(10) {
		t.Fatalf("jsonpath arg wrong: %#v", args["max"])
	}
}

func TestRegistryBuiltins(t *testing.T) {
	reg, err := render.NewRegistry("", nil)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	for _, id := range []string{"datetime", "relation", "code", "geo", "rating"} {
		if !reg.Has(id) {
			t.Fatalf("expected built-in widget %q", id)
		}
	}
	html, err := reg.Render("datetime", render.FieldContext{Name: "x", Value: "2020-01-01T00:00", Args: map[string]any{"type": "datetime-local"}})
	if err != nil {
		t.Fatalf("render datetime: %v", err)
	}
	if html == "" {
		t.Fatal("expected datetime html")
	}
}

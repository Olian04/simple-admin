package form_test

import (
	"testing"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/views/form"
)

func TestFormFieldsExcludeReadOnlyAndSeed(t *testing.T) {
	fields := []resource.Field{
		{Name: "id", Kind: resource.KindString, ReadOnly: true},
		{Name: "title", Kind: resource.KindString, Required: true},
	}
	data := resource.Record{"title": "Hello"}
	errs := map[string]string{"title": "too short"}
	ff := form.FormFields(fields, data, errs)
	for _, f := range ff {
		if f.Name == "id" {
			t.Fatal("read-only id must be excluded from form")
		}
		if f.Name == "title" && (f.Value != "Hello" || f.Error != "too short") {
			t.Fatalf("title not seeded: %+v", f)
		}
	}
}

func TestInputType(t *testing.T) {
	cases := []struct {
		f    resource.Field
		want string
	}{
		{resource.Field{Kind: resource.KindNumber}, "number"},
		{resource.Field{Kind: resource.KindString, Format: "email"}, "email"},
		{resource.Field{Kind: resource.KindString, Format: "date-time"}, "datetime-local"},
		{resource.Field{Kind: resource.KindString}, "text"},
	}
	for _, c := range cases {
		if got := form.InputType(c.f); got != c.want {
			t.Fatalf("InputType(%+v)=%q want %q", c.f, got, c.want)
		}
	}
}

func TestAutoSelectDatetime(t *testing.T) {
	id, _, ok := form.AutoSelectWidget(resource.Field{Name: "publishedAt", Format: "date-time"}, nil)
	if !ok || id != "datetime" {
		t.Fatalf("want datetime, got %q ok=%v", id, ok)
	}
}

func TestAutoSelectRelation(t *testing.T) {
	id, target, ok := form.AutoSelectWidget(
		resource.Field{Name: "author", Kind: resource.KindString},
		[]string{"posts", "authors"},
	)
	if !ok || id != "relation" || target != "authors" {
		t.Fatalf("want relation->authors, got id=%q target=%q ok=%v", id, target, ok)
	}
}

func TestAutoSelectNoneForPlainField(t *testing.T) {
	if _, _, ok := form.AutoSelectWidget(resource.Field{Name: "title", Kind: resource.KindString}, []string{"posts"}); ok {
		t.Fatal("plain field should not auto-select a widget")
	}
}

func TestResolveArgsStaticAndJSONPath(t *testing.T) {
	spec := map[string]resource.TemplateArg{
		"language": {Static: "json"},
		"max":      {JSONPath: "$.maxRating"},
	}
	rec := resource.Record{"maxRating": float64(10)}
	args := form.ResolveArgs(spec, rec)
	if args["language"] != "json" {
		t.Fatalf("static arg wrong: %#v", args["language"])
	}
	if args["max"] != float64(10) {
		t.Fatalf("jsonpath arg wrong: %#v", args["max"])
	}
}

func TestWidgetRegistryBuiltins(t *testing.T) {
	reg, err := form.NewWidgetRegistry("", nil)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	for _, id := range []string{"datetime", "relation", "code", "rating"} {
		if !reg.Has(id) {
			t.Fatalf("expected built-in widget %q", id)
		}
	}
}

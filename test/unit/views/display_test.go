package views_test

import (
	"testing"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/views"
)

func TestColumnsExcludeArrayObject(t *testing.T) {
	fields := []resource.Field{
		{Name: "id", Kind: resource.KindString},
		{Name: "title", Kind: resource.KindString},
		{Name: "tags", Kind: resource.KindArray},
		{Name: "meta", Kind: resource.KindObject},
	}
	cols := views.Columns(fields)
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
		if got := views.DisplayValue(in); got != want {
			t.Fatalf("DisplayValue(%v)=%q want %q", in, got, want)
		}
	}
}

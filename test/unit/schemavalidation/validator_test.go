package schemavalidation_test

import (
	"errors"
	"testing"

	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/schemavalidation"
)

const schema = `{
  "type": "object",
  "required": ["title"],
  "properties": {
    "title": {"type": "string", "maxLength": 5},
    "rating": {"type": "number"},
    "status": {"type": "string", "enum": ["draft", "published"]}
  }
}`

func newValidator(t *testing.T) *schemavalidation.Validator {
	t.Helper()
	v, err := schemavalidation.New([]byte(schema))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return v
}

func TestValidatePasses(t *testing.T) {
	v := newValidator(t)
	if err := v.Validate(resource.Record{"title": "ok", "status": "draft", "rating": float64(3)}); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidateMissingRequired(t *testing.T) {
	v := newValidator(t)
	err := v.Validate(resource.Record{"status": "draft"})
	var verr *resource.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("want *ValidationError, got %v", err)
	}
	if _, ok := verr.Fields["title"]; !ok {
		t.Fatalf("expected title required error, got %+v", verr.Fields)
	}
}

func TestValidateFieldConstraintsMapToField(t *testing.T) {
	v := newValidator(t)
	err := v.Validate(resource.Record{"title": "toolong", "status": "bogus"})
	var verr *resource.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("want *ValidationError, got %v", err)
	}
	if _, ok := verr.Fields["title"]; !ok {
		t.Fatalf("expected title (maxLength) error, got %+v", verr.Fields)
	}
	if _, ok := verr.Fields["status"]; !ok {
		t.Fatalf("expected status (enum) error, got %+v", verr.Fields)
	}
}

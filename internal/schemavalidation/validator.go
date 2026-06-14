// Package schemavalidation wraps santhosh-tekuri/jsonschema to validate record
// bodies against a resource's JSON Schema, returning field-keyed errors.
package schemavalidation

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/Olian04/simple-admin/internal/domain/resource"
)

var quotedRE = regexp.MustCompile(`'([^']+)'`)

// Validator validates records against a compiled JSON Schema.
type Validator struct {
	schema *jsonschema.Schema
}

// New compiles a JSON Schema (raw bytes) into a Validator.
func New(raw []byte) (*Validator, error) {
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", doc); err != nil {
		return nil, fmt.Errorf("add schema: %w", err)
	}
	sch, err := c.Compile("schema.json")
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return &Validator{schema: sch}, nil
}

// Validate implements resource.Validator. On failure it returns a
// *resource.ValidationError keyed by field name.
func (v *Validator) Validate(rec resource.Record) error {
	err := v.schema.Validate(map[string]any(rec))
	if err == nil {
		return nil
	}
	var ve *jsonschema.ValidationError
	if !errors.As(err, &ve) {
		return err
	}

	fields := map[string]string{}
	collect(ve.BasicOutput(), fields)
	if len(fields) == 0 {
		fields["_"] = ve.Error()
	}
	return &resource.ValidationError{Fields: fields}
}

func collect(u *jsonschema.OutputUnit, fields map[string]string) {
	if u == nil {
		return
	}
	if u.Error != nil && !u.Valid {
		msg := u.Error.String()
		if name := fieldName(u.InstanceLocation); name != "" {
			if _, ok := fields[name]; !ok {
				fields[name] = msg
			}
		} else if strings.Contains(msg, "missing propert") {
			// Root-level "missing properties 'title'": map each to a required error.
			for _, m := range quotedRE.FindAllStringSubmatch(msg, -1) {
				if _, ok := fields[m[1]]; !ok {
					fields[m[1]] = "is required"
				}
			}
		}
	}
	for i := range u.Errors {
		collect(&u.Errors[i], fields)
	}
}

// fieldName turns a JSON-pointer instance location ("/title", "/author/0")
// into the top-level field name.
func fieldName(loc string) string {
	s := strings.TrimPrefix(loc, "/")
	if s == "" {
		return ""
	}
	if idx := strings.IndexByte(s, '/'); idx >= 0 {
		s = s[:idx]
	}
	return s
}

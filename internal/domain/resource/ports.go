package resource

import (
	"context"
	"sort"
	"strings"
)

// Record is a single resource instance (a decoded JSON object).
type Record map[string]any

// ListParams controls a list query.
type ListParams struct {
	Page    int
	PerPage int
}

// ListResult is a page of records.
type ListResult struct {
	Items      []Record
	TotalItems int
	Page       int
	PerPage    int
}

// Store performs CRUD against an upstream backend for a single resource.
// The credential for auth passthrough travels on the context.
type Store interface {
	List(ctx context.Context, params ListParams) (ListResult, error)
	Get(ctx context.Context, id string) (Record, error)
	Create(ctx context.Context, body Record) (Record, error)
	Update(ctx context.Context, id string, body Record) (Record, error)
	Delete(ctx context.Context, id string) error
}

// Validator validates a record body against a resource schema.
type Validator interface {
	// Validate returns nil on success or a *ValidationError on failure.
	Validate(body Record) error
}

// DescriptorSource produces resource descriptors (manual config or discovery).
type DescriptorSource interface {
	Descriptors(ctx context.Context) ([]ResourceDescriptor, error)
}

// ValidationError carries per-field validation messages keyed by field path,
// so handlers can render inline form errors.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Fields) == 0 {
		return "validation failed"
	}
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+": "+e.Fields[k])
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

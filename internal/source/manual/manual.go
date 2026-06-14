// Package manual builds resource descriptors from manually-configured
// backends: each resource declares its own JSON Schema (disk path or url) and
// CRUD paths. No OpenAPI is involved.
package manual

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// Source is a resource.DescriptorSource over the manual backends in config.
type Source struct {
	backends []config.BackendSection
	baseDir  string // resolves relative schema file paths (the config file's dir)
	client   *http.Client
}

// New returns a manual source. baseDir resolves relative schema paths.
func New(backends []config.BackendSection, baseDir string) *Source {
	return &Source{
		backends: backends,
		baseDir:  baseDir,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Descriptors implements resource.DescriptorSource.
func (s *Source) Descriptors(ctx context.Context) ([]resource.ResourceDescriptor, error) {
	var out []resource.ResourceDescriptor
	for _, b := range s.backends {
		if b.Discovery != nil {
			continue // discovery backends are handled by the openapi source
		}
		backend := resource.Backend{ID: b.ID, Label: b.Label, BaseURL: b.BaseURL}
		auth := authScheme(b.Auth)

		for _, r := range b.Resources {
			raw, err := s.loadSchema(ctx, r.Schema)
			if err != nil {
				return nil, fmt.Errorf("backend %q resource %q schema: %w", b.ID, r.Name, err)
			}
			fields, err := resource.FieldsFromJSONSchema(raw)
			if err != nil {
				return nil, fmt.Errorf("backend %q resource %q: %w", b.ID, r.Name, err)
			}
			applyFieldTemplates(fields, r.FieldTemplates)

			out = append(out, resource.ResourceDescriptor{
				Backend: backend,
				Name:    r.Name,
				Label:   r.Label,
				Schema:  resource.Schema{Raw: raw, Fields: fields},
				Capabilities: resource.Capabilities{
					List:       true,
					Read:       true,
					Create:     true,
					Update:     true,
					Delete:     true,
					UpdateVerb: r.UpdateVerb,
				},
				AuthScheme: auth,
				IDField:    r.IDField,
				CRUDPaths:  resource.CRUDPaths{Collection: r.Paths.Collection, Item: r.Paths.Item},
			})
		}
	}
	return out, nil
}

func (s *Source) loadSchema(ctx context.Context, src config.SchemaSource) ([]byte, error) {
	if src.URL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch schema %s: status %d", src.URL, resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}

	path := src.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.baseDir, path)
	}
	// #nosec G304 -- schema path is operator config, not request-derived.
	return os.ReadFile(path)
}

func authScheme(a *config.AuthScheme) resource.AuthScheme {
	if a == nil {
		return resource.AuthScheme{Type: "none"}
	}
	return resource.AuthScheme{Type: a.Type, In: a.In, Name: a.Name}
}

func applyFieldTemplates(fields []resource.Field, templates map[string]config.FieldTemplate) {
	if len(templates) == 0 {
		return
	}
	for i := range fields {
		ft, ok := templates[fields[i].Name]
		if !ok {
			continue
		}
		fields[i].Custom = &resource.FieldTemplate{
			TemplateID:   ft.Template,
			TemplateFile: ft.TemplateFile,
			Args:         convertArgs(ft.Args),
		}
	}
}

func convertArgs(args map[string]config.TemplateArg) map[string]resource.TemplateArg {
	if len(args) == 0 {
		return nil
	}
	out := make(map[string]resource.TemplateArg, len(args))
	for k, v := range args {
		out[k] = resource.TemplateArg{Static: v.Static, JSONPath: v.JSONPath}
	}
	return out
}

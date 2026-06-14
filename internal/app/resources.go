package app

import (
	"context"
	"fmt"

	"github.com/Olian04/simple-admin/internal/backend"
	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/schemavalidation"
	"github.com/Olian04/simple-admin/internal/source/manual"
	"github.com/Olian04/simple-admin/internal/source/openapi"
	"github.com/Olian04/simple-admin/internal/views/form"
)

// buildRegistry runs the descriptor sources (manual + discovery) and returns a
// populated registry. configDir resolves relative manual schema paths. It fails
// fast if any configured source cannot load.
//
// Each descriptor becomes a live Runtime: a backend.Client (Store), a compiled
// JSON Schema validator, and the descriptor's Service.
func buildRegistry(ctx context.Context, cfg config.Config, configDir string) (*resource.Registry, error) {
	sources := []resource.DescriptorSource{
		manual.New(cfg.Backends, configDir),
		openapi.New(cfg.Backends),
	}

	var all []resource.ResourceDescriptor
	for _, src := range sources {
		descs, err := src.Descriptors(ctx)
		if err != nil {
			return nil, fmt.Errorf("descriptor source: %w", err)
		}
		all = append(all, descs...)
	}

	runtimes := make([]*resource.Runtime, 0, len(all))
	for i := range all {
		d := all[i]
		validator, err := schemavalidation.New(d.Schema.Raw)
		if err != nil {
			return nil, fmt.Errorf("backend %q resource %q schema: %w", d.Backend.ID, d.Name, err)
		}
		svc := &resource.Service{Descriptor: d, Store: backend.New(d), Validator: validator}
		runtimes = append(runtimes, &resource.Runtime{Descriptor: d, Service: svc})
	}

	reg := resource.NewRegistry()
	reg.Replace(runtimes)
	return reg, nil
}

// buildWidgets builds the field-widget registry: embedded built-ins plus any
// user-supplied template_file widgets referenced in config (resolved against
// configDir). Fails fast on parse errors.
func buildWidgets(cfg config.Config, configDir string) (*form.WidgetRegistry, error) {
	var userFiles []string
	seen := map[string]bool{}
	for _, b := range cfg.Backends {
		for _, r := range b.Resources {
			for _, ft := range r.FieldTemplates {
				if ft.TemplateFile != "" && !seen[ft.TemplateFile] {
					seen[ft.TemplateFile] = true
					userFiles = append(userFiles, ft.TemplateFile)
				}
			}
		}
	}
	return form.NewWidgetRegistry(configDir, userFiles)
}

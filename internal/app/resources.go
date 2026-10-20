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
)

// buildRegistry runs the descriptor sources (manual + discovery) and returns a
// populated registry. configDir resolves relative manual schema paths. It fails
// fast if any configured source cannot load.
//
// Each descriptor becomes a live Runtime: a backend.Client (Store), a compiled
// JSON Schema validator, and the descriptor's Service.
func buildRegistry(ctx context.Context, cfg config.Config, configDir string) (*resource.Registry, error) {
	runtimes, err := assembleRuntimes(ctx, cfg, configDir)
	if err != nil {
		return nil, err
	}
	reg := resource.NewRegistry()
	reg.Replace(runtimes)
	return reg, nil
}

// assembleRuntimes runs the descriptor sources and turns each descriptor into a
// live Runtime (backend.Client Store + compiled validator + Service). It is
// reused by the startup build and the discovery poller.
func assembleRuntimes(ctx context.Context, cfg config.Config, configDir string) ([]*resource.Runtime, error) {
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
	return runtimes, nil
}

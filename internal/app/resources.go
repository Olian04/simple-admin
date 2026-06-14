package app

import (
	"context"
	"fmt"

	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/domain/resource"
	"github.com/Olian04/simple-admin/internal/source/manual"
	"github.com/Olian04/simple-admin/internal/source/openapi"
)

// buildRegistry runs the descriptor sources (manual + discovery) and returns a
// populated registry. configDir resolves relative manual schema paths. It fails
// fast if any configured source cannot load.
//
// Per-resource Store/Validator wiring (turning descriptors into live Runtimes
// with CRUD) is added in a later step; for now Runtimes carry only descriptors
// so the sidebar/navigation render.
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
		runtimes = append(runtimes, &resource.Runtime{Descriptor: all[i]})
	}

	reg := resource.NewRegistry()
	reg.Replace(runtimes)
	return reg, nil
}

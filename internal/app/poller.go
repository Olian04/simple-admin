package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/domain/resource"
)

// startDiscoveryPoller periodically re-runs the descriptor sources and swaps the
// registry and calls onChange, so discovery-backed resources stay in sync without
// a restart (the UI is rebuilt from the registry). It is
// a no-op when no backend uses discovery. Stops on ctx cancellation.
func startDiscoveryPoller(ctx context.Context, cfg config.Config, configDir string, reg *resource.Registry, onChange func()) {
	interval := discoveryInterval(cfg)
	if interval <= 0 {
		return // no discovery backends → nothing to poll
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runtimes, err := assembleRuntimes(ctx, cfg, configDir)
				if err != nil {
					slog.Warn("discovery poll failed; keeping current resources", "error", err)
					continue
				}
				reg.Replace(runtimes)
				if onChange != nil {
					onChange()
				}
				slog.Info("discovery refreshed", "resources", len(runtimes))
			}
		}
	}()
}

// discoveryInterval returns the smallest poll_interval among discovery backends,
// or 0 if none use discovery.
func discoveryInterval(cfg config.Config) time.Duration {
	var smallest time.Duration
	for _, b := range cfg.Backends {
		if b.Discovery == nil {
			continue
		}
		d, err := time.ParseDuration(b.Discovery.PollInterval)
		if err != nil || d <= 0 {
			continue
		}
		if smallest == 0 || d < smallest {
			smallest = d
		}
	}
	return smallest
}

package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/observability/metrics"
	httptransport "github.com/Olian04/simple-admin/internal/transport/http"
	"github.com/Olian04/simple-admin/internal/transport/metricshttp"
)

func Run(ctx context.Context, cfg config.Config, configDir string) error {
	cfg = cfg.WithDefaults()
	registry, err := metrics.NewRegistry(
		metrics.WithNamespace(cfg.Metrics.MetricPrefix),
		metrics.WithConstLabels(map[string]string(cfg.Labels)),
	)
	if err != nil {
		return fmt.Errorf("metrics registry: %w", err)
	}

	resources, err := buildRegistry(ctx, cfg, configDir)
	if err != nil {
		return fmt.Errorf("build resources: %w", err)
	}

	widgets, err := buildWidgets(cfg, configDir)
	if err != nil {
		return fmt.Errorf("build widgets: %w", err)
	}

	handler := httptransport.Router(registry, resources, widgets, cfg.UI)

	n := 1
	metricsOn := cfg.MetricsEnabled()
	if metricsOn {
		n++
	}
	errCh := make(chan error, n)

	go func() { errCh <- serveHTTP(ctx, cfg.HTTP.ListenAddr, handler) }()

	if metricsOn {
		go func() { errCh <- metricshttp.Serve(ctx, cfg.Metrics.ListenAddr, registry.Handler()) }()
		slog.Info("simple-admin started", "http_addr", cfg.HTTP.ListenAddr, "metrics_addr", cfg.Metrics.ListenAddr)
	} else {
		slog.Info("simple-admin started", "http_addr", cfg.HTTP.ListenAddr, "metrics_enabled", false)
	}

	select {
	case <-ctx.Done():
		// Signal-driven (SIGINT/SIGTERM) cancellation is a graceful shutdown,
		// not an error — servers drain via their own ctx.Done() handlers.
		slog.Info("shutting down", "reason", "signal")
		return nil
	case err := <-errCh:
		return err
	}
}

func serveHTTP(ctx context.Context, addr string, handler http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("http shutdown", "error", err)
		}
	}()
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

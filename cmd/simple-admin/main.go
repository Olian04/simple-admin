package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/Olian04/simple-admin/cmd/simple-admin/version"
	"github.com/Olian04/simple-admin/internal/app"
	"github.com/Olian04/simple-admin/internal/config"
	"github.com/Olian04/simple-admin/internal/observability/logging"
)

func main() {
	vi := version.Info()

	cli.VersionPrinter = printVersion(vi)

	root := &cli.Command{
		Name:    "simple-admin",
		Usage:   "Portable schema-driven admin UI",
		Version: vi.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "path to YAML config file",
				Sources: cli.EnvVars("APP_CONFIG_FILE"),
			},
		},
		Action: runCLI,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := root.Run(ctx, os.Args); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("exited with error", "error_message", err)
		os.Exit(1)
	}
}

func runCLI(ctx context.Context, c *cli.Command) error {
	configPath := c.String("config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	configDir := "."
	if configPath != "" {
		configDir = filepath.Dir(configPath)
	}

	logLabels := maps.Clone(map[string]string(cfg.Labels))
	if logLabels == nil {
		logLabels = make(map[string]string, 1)
	}
	logLabels["software_version"] = version.Info().Version

	cleanup, err := logging.Setup(
		logging.WithLevel(cfg.Logging.Level),
		logging.WithFormat(cfg.Logging.Format),
		logging.WithStream(cfg.Logging.Stream),
		logging.WithLabels(logLabels),
	)
	if err != nil {
		slog.Error("setup logging", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	return app.Run(ctx, cfg, configDir)
}

func printVersion(vi version.VersionInfo) func(cmd *cli.Command) {
	return func(cmd *cli.Command) {
		_, err := fmt.Fprintf(cmd.Root().Writer, "%s version %s\nrevision %s\nbuild_time %s\n",
			cmd.Name, vi.Version, vi.Revision, vi.BuildTime)
		if err != nil {
			slog.Error("write version", "error_message", err.Error())
		}
	}
}

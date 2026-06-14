package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
)

type Option func(*setupOptions)

type setupOptions struct {
	level  string
	format string
	stream string
	labels map[string]string
}

func WithLevel(s string) Option {
	return func(o *setupOptions) { o.level = s }
}

func WithFormat(s string) Option {
	return func(o *setupOptions) { o.format = s }
}

func WithStream(s string) Option {
	return func(o *setupOptions) { o.stream = s }
}

func WithLabels(m map[string]string) Option {
	return func(o *setupOptions) { o.labels = m }
}

func Setup(opts ...Option) (func(), error) {
	var o setupOptions
	for _, opt := range opts {
		opt(&o)
	}

	if o.level == "" {
		return nil, fmt.Errorf("logging.Setup: WithLevel is required")
	}
	if o.format == "" {
		return nil, fmt.Errorf("logging.Setup: WithFormat is required")
	}
	if o.stream == "" {
		return nil, fmt.Errorf("logging.Setup: WithStream is required")
	}

	lvl, err := parseLevel(o.level)
	if err != nil {
		return nil, err
	}
	out, err := parseStream(o.stream)
	if err != nil {
		return nil, err
	}
	handler, err := newHandler(o.format, out, lvl)
	if err != nil {
		return nil, err
	}

	logger := slog.New(handler)

	keys := make([]string, 0, len(o.labels))
	for k := range o.labels {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	args := make([]any, 0, 2*len(keys))
	for _, k := range keys {
		args = append(args, k, o.labels[k])
	}
	if len(args) > 0 {
		logger = logger.With(args...)
	}

	old := slog.Default()
	slog.SetDefault(logger)
	return func() { slog.SetDefault(old); _, _ = io.Discard.Write(nil) }, nil
}

func parseLevel(s string) (slog.Level, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid logging.level %q: want debug|info|warn|error", s)
	}
}

func parseStream(s string) (*os.File, error) {
	switch s {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("invalid logging.stream %q: want stdout|stderr", s)
	}
}

func newHandler(format string, out io.Writer, lvl slog.Level) (slog.Handler, error) {
	opts := &slog.HandlerOptions{Level: lvl}
	switch format {
	case "json":
		return slog.NewJSONHandler(out, opts), nil
	case "text":
		return slog.NewTextHandler(out, opts), nil
	default:
		return nil, fmt.Errorf("invalid logging.format %q: want json|text", format)
	}
}

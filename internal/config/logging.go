package config

import (
	"fmt"
	"strings"
)

// LoggingSection maps YAML block `logging`.
type LoggingSection struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
	Stream string `yaml:"stream,omitempty"`
}

func (l LoggingSection) WithDefaults() LoggingSection {
	l.Level = strings.TrimSpace(l.Level)
	l.Format = strings.TrimSpace(l.Format)
	l.Stream = strings.TrimSpace(l.Stream)
	if l.Level == "" {
		l.Level = DefaultLoggingLevel
	}
	if l.Format == "" {
		l.Format = DefaultLoggingFormat
	}
	if l.Stream == "" {
		l.Stream = DefaultLoggingStream
	}
	return l
}

func (l LoggingSection) Validate() error {
	switch l.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("logging.level %q invalid: want debug|info|warn|error", l.Level)
	}
	switch l.Format {
	case "json", "text":
	default:
		return fmt.Errorf("logging.format %q invalid: want json|text", l.Format)
	}
	switch l.Stream {
	case "stdout", "stderr":
	default:
		return fmt.Errorf("logging.stream %q invalid: want stdout|stderr", l.Stream)
	}
	return nil
}

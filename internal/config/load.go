package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (Config, error) {
	if path == "" {
		c := Config{}.WithDefaults()
		if err := c.Validate(); err != nil {
			return Config{}, err
		}
		return c, nil
	}
	// #nosec G304 -- config path is explicit operator input, not derived from request data.
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		c := Config{}.WithDefaults()
		if err := c.Validate(); err != nil {
			return Config{}, err
		}
		return c, nil
	}

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}
	cfg = cfg.WithDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

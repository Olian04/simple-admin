// Package config holds root YAML loading and subsection types (`labels`, `http`, `metrics`, `logging`, `backends`).
package config

import "fmt"

// Config maps the root YAML file — see configs/config.example.yaml.
type Config struct {
	Labels   Labels           `yaml:"labels,omitempty"`
	HTTP     HTTPSection      `yaml:"http,omitempty"`
	Metrics  MetricsSection   `yaml:"metrics,omitempty"`
	Logging  LoggingSection   `yaml:"logging,omitempty"`
	Backends []BackendSection `yaml:"backends,omitempty"`
}

func (c Config) WithDefaults() Config {
	c.Labels = c.Labels.WithDefaults()
	c.HTTP = c.HTTP.WithDefaults()
	c.Metrics = c.Metrics.WithDefaults()
	c.Logging = c.Logging.WithDefaults()
	backends := make([]BackendSection, len(c.Backends))
	for i := range c.Backends {
		backends[i] = c.Backends[i].WithDefaults()
	}
	c.Backends = backends
	return c
}

func (c Config) Validate() error {
	if err := c.Labels.Validate(); err != nil {
		return err
	}
	if err := c.HTTP.Validate(); err != nil {
		return err
	}
	if err := c.Metrics.Validate(); err != nil {
		return err
	}
	if err := c.Logging.Validate(); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, b := range c.Backends {
		if err := b.Validate(); err != nil {
			return err
		}
		if seen[b.ID] {
			return fmt.Errorf("duplicate backend id %q", b.ID)
		}
		seen[b.ID] = true
	}
	return nil
}

func (c Config) MetricsEnabled() bool {
	if c.Metrics.Enabled == nil {
		return true
	}
	return *c.Metrics.Enabled
}

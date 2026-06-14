// Package config holds root YAML loading and subsection types (`labels`, `http`, `metrics`, `logging`).
package config

// Config maps the root YAML file — see configs/config.example.yaml.
type Config struct {
	Labels  Labels          `yaml:"labels,omitempty"`
	HTTP    HTTPSection     `yaml:"http,omitempty"`
	Metrics MetricsSection  `yaml:"metrics,omitempty"`
	Logging LoggingSection  `yaml:"logging,omitempty"`
}

func (c Config) WithDefaults() Config {
	c.Labels = c.Labels.WithDefaults()
	c.HTTP = c.HTTP.WithDefaults()
	c.Metrics = c.Metrics.WithDefaults()
	c.Logging = c.Logging.WithDefaults()
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
	return c.Logging.Validate()
}

func (c Config) MetricsEnabled() bool {
	if c.Metrics.Enabled == nil {
		return true
	}
	return *c.Metrics.Enabled
}

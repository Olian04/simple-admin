package config

import (
	"fmt"
	"strings"

	"github.com/prometheus/common/model"
)

// MetricsSection maps YAML block `metrics`.
type MetricsSection struct {
	Enabled      *bool  `yaml:"enabled,omitempty"`
	ListenAddr   string `yaml:"listen_addr,omitempty"`
	MetricPrefix string `yaml:"metric_prefix,omitempty"`
}

func (m MetricsSection) WithDefaults() MetricsSection {
	if m.ListenAddr == "" {
		m.ListenAddr = DefaultMetricsListenAddr
	}
	if m.Enabled == nil {
		t := DefaultMetricsEnabled
		m.Enabled = &t
	}
	if m.MetricPrefix == "" {
		m.MetricPrefix = DefaultMetricPrefix
	}
	return m
}

func (m MetricsSection) Validate() error {
	if strings.TrimSpace(m.ListenAddr) == "" {
		return fmt.Errorf("metrics.listen_addr must be non-empty")
	}
	if m.Enabled == nil {
		return fmt.Errorf("metrics.enabled must be set after WithDefaults: bug")
	}
	if strings.TrimSpace(m.MetricPrefix) == "" {
		return fmt.Errorf("metrics.metric_prefix must be non-empty")
	}
	composed := m.MetricPrefix + "_http_requests_total"
	if !model.LegacyValidation.IsValidMetricName(composed) {
		return fmt.Errorf("metrics.metric_prefix %q: composed metric %q must satisfy legacy Prometheus metric name rules", m.MetricPrefix, composed)
	}
	return nil
}

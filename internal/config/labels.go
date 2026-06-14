package config

import (
	"fmt"

	"github.com/prometheus/common/model"
)

// Labels attach to slog and Prometheus const labels on metrics.
type Labels map[string]string

func (l Labels) WithDefaults() Labels {
	if l == nil {
		return Labels{}
	}
	out := make(Labels, len(l))
	for k, v := range l {
		out[k] = v
	}
	return out
}

func (l Labels) Validate() error {
	for k := range l {
		if !model.LegacyValidation.IsValidLabelName(k) {
			return fmt.Errorf("invalid labels key %q: must satisfy legacy Prometheus label name rules", k)
		}
	}
	return nil
}

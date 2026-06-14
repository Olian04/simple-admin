package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option func(*registryOptions)

type registryOptions struct {
	namespace   string
	constLabels map[string]string
}

func WithNamespace(s string) Option {
	return func(o *registryOptions) { o.namespace = s }
}

// WithConstLabels sets static metric labels (Prometheus naming rules apply).
func WithConstLabels(m map[string]string) Option {
	return func(o *registryOptions) { o.constLabels = m }
}

type Registry struct {
	reg      *prometheus.Registry
	requests prometheus.Counter
}

func NewRegistry(opts ...Option) (*Registry, error) {
	var o registryOptions
	for _, opt := range opts {
		opt(&o)
	}

	if o.namespace == "" {
		return nil, fmt.Errorf("metrics.NewRegistry: WithNamespace is required")
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
	)

	req := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   o.namespace,
		Name:        "http_requests_total",
		Help:        "Total HTTP requests handled.",
		ConstLabels: prometheus.Labels(o.constLabels),
	})
	reg.MustRegister(req)
	return &Registry{reg: reg, requests: req}, nil
}

func (r *Registry) IncRequests() {
	r.requests.Inc()
}

func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{Registry: r.reg})
}

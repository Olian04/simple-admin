package config

const (
	DefaultHTTPListenAddr    = ":8080"
	DefaultMetricsListenAddr = ":9090"
	DefaultMetricPrefix      = "simple_admin"
	DefaultLoggingLevel      = "info"
	DefaultLoggingFormat     = "json"
	DefaultLoggingStream     = "stdout"

	DefaultAuthType     = "none"
	DefaultIDField      = "id"
	DefaultPollInterval = "5m"
	DefaultUpdateVerb   = "PUT"
)

const DefaultMetricsEnabled = true

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

	DefaultMonacoCDN  = "https://cdn.jsdelivr.net/npm/monaco-editor@0.52.2/min/vs"
	DefaultEasyMDECDN = "https://cdn.jsdelivr.net/npm/easymde@2.18.0/dist"
	DefaultPageSize   = 25
)

const DefaultMetricsEnabled = true

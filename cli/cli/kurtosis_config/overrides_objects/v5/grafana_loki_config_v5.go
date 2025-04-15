package v5

type GrafanaLokiConfig struct {
	// ShouldStartBeforeEngine starts Grafana and Loki before the engine if true.
	// Equivalent to running `grafloki start` before `engine start`.
	// Useful for making them the default logging setup in Kurtosis.
	ShouldStartBeforeEngine bool   `yaml:"should-start-before-engine,omitempty"`
	GrafanaImage            string `yaml:"grafana-image,omitempty"`
	LokiImage               string `yaml:"loki-image,omitempty"`
}

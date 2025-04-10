package v5

type GraflokiConfig struct {
	ShouldStartBeforeEngine bool   `yaml:"should-start-before-engine,omitempty"`
	GrafanaImage            string `yaml:"grafana-image,omitempty"`
	LokiImage               string `yaml:"loki-image,omitempty"`
}

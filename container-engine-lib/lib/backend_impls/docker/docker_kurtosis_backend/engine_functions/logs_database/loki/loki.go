package loki

import (
	"github.com/kurtosis-tech/stacktrace"
	"gopkg.in/yaml.v3"
)


type Loki struct {
	config *LokiConfig
}

func NewLoki(httpListenPort uint16) *Loki {
	lokiConfig := newDefaultLokiConfigForKurtosisCentralizedLogs(httpListenPort)
	return &Loki{lokiConfig}
}

func (loki *Loki) GetConfigContent() (string, error) {
	lokiConfigYAMLContent, err := yaml.Marshal(loki.config)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred marshalling Loki config '%+v'", loki.config)
	}
	lokiConfigYAMLContentStr := string(lokiConfigYAMLContent)
	return lokiConfigYAMLContentStr, nil
}

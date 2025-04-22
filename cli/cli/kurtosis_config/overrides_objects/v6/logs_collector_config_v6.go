package v6

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"

type LogsCollectorConfigV6 struct {
	Parsers []logs_collector.Parser `yaml:"parsers,omitempty"`
	Filters []logs_collector.Filter `yaml:"filters,omitempty"`
}

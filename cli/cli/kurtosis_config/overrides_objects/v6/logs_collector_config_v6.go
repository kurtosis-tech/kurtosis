package v6

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"

type LogsCollectorConfigV6 struct {
	Filters []logs_collector.Filter `yaml:"filters,omitempty"`
}

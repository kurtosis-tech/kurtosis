package v7

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
                           DO NOT CHANGE THIS FILE!
  If you change this file, it will break config for users who have instantiated an
           overrides file with this version of config overrides!
    Instead, to make changes, you will need to add a new version of the config
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
*/

// LogsCollectorConfigV7 is the configuration for the logs collector.
// Kurtosis leverages a logs collector and logs aggregator to collect, aggregate, and logs from services in enclaves.
// The logs collector picks up logs from services in enclaves and sends them to the logs aggregator.
type LogsCollectorConfigV7 struct {
	Parsers []logs_collector.Parser `yaml:"parsers,omitempty"`
	Filters []logs_collector.Filter `yaml:"filters,omitempty"`
}

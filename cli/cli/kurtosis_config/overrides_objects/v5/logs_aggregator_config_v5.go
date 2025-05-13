package v5

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
                           DO NOT CHANGE THIS FILE!
  If you change this file, it will break config for users who have instantiated an
           overrides file with this version of config overrides!
    Instead, to make changes, you will need to add a new version of the config
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
*/

// LogsAggregatorConfigV5 is the configuration for the logs aggregator.
// Kurtosis leverages a logs collector and logs aggregator to collect, aggregate, and logs from services in enclaves.
// The logs aggregator aggregates logs forwarded to it by the logs collector and sends them to the configured sinks for storage and downstream processing.
type LogsAggregatorConfigV5 struct {
	Sinks map[string]map[string]interface{} `yaml:"sinks,omitempty"`
}

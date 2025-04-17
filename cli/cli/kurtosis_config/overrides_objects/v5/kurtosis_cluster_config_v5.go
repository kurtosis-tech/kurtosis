package v5

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
                           DO NOT CHANGE THIS FILE!
  If you change this file, it will break config for users who have instantiated an
           overrides file with this version of config overrides!
    Instead, to make changes, you will need to add a new version of the config
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
*/

type KurtosisClusterConfigV5 struct {
	Type *string `yaml:"type,omitempty"`
	// If we ever get another type of cluster that has configuration, this will need to be polymorphically deserialized
	Config            *KubernetesClusterConfigV5 `yaml:"config,omitempty"`
	LogsAggregator    *LogsAggregatorConfigV5    `yaml:"logs-aggregator,omitempty"`
	GrafanaLokiConfig *GrafanaLokiConfig         `yaml:"grafana-loki,omitempty"`

	// ShouldEnableDefaultLogsSink controls use of PersistentVolumeLogsDB (default: true) as the storage location for logs.
	// Useful for saving storage when using custom or Grafana Loki-based logging.
	ShouldEnableDefaultLogsSink *bool `yaml:"should-enable-default-logs-sink,omitempty"`
}

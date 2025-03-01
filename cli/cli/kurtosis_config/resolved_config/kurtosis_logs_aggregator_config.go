package resolved_config

type KurtosisLogsAggregatorConfig struct {
	Image string
	Sinks map[string]map[string]interface{}
}

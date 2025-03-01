package v3

type KurtosisLogsAggregatorConfigV3 struct {
	Sinks map[string]map[string]interface{} `yaml:"sinks,omitempty"`
}

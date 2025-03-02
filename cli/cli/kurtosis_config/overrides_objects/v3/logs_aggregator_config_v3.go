package v3

type LogsAggregatorConfigV3 struct {
	Sinks map[string]map[string]interface{} `yaml:"sinks,omitempty"`
}

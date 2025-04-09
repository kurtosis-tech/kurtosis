package v5

type LogsAggregatorConfigV5 struct {
	Sinks map[string]map[string]interface{} `yaml:"sinks,omitempty"`
}

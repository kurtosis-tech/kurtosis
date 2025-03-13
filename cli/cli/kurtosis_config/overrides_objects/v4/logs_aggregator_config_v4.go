package v4

type LogsAggregatorConfigV4 struct {
	Sinks map[string]map[string]interface{} `yaml:"sinks,omitempty"`
}

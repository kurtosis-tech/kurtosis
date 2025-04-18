package v6

type LogsAggregatorConfigV6 struct {
	Sinks map[string]map[string]interface{} `yaml:"sinks,omitempty"`
}

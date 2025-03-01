package v3

type KurtosisLogsAggregatorConfigV3 struct {
	Image *string                           `yaml:"image,omitempty"`
	Sinks map[string]map[string]interface{} `yaml:"sinks,omitempty"`
}

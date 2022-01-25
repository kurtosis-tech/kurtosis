package kurtosis_config

type KurtosisConfig struct {
	//We set public fields because YAML marshalling needs it on this way
	ShouldSendMetrics bool `yaml:"should-send-metrics"`
}

func NewKurtosisConfig(doesUserAcceptSendingMetrics bool) *KurtosisConfig {
	return &KurtosisConfig{ShouldSendMetrics: doesUserAcceptSendingMetrics}
}

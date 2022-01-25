package kurtosis_config
// NOTE: All new YAML property names here should be kebab-case because
//a) it's easier to read b) it's easier to write
//c) it's consistent with previous properties and changing the format of
//an already-written config file is very difficult
type KurtosisConfig struct {
	//We set public fields because YAML marshalling needs it on this way
	ShouldSendMetrics bool `yaml:"should-send-metrics"`
}

func NewKurtosisConfig(doesUserAcceptSendingMetrics bool) *KurtosisConfig {
	return &KurtosisConfig{ShouldSendMetrics: doesUserAcceptSendingMetrics}
}

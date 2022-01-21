package kurtosis_config

type KurtosisConfig struct {
	shouldSendMetrics bool
}

func NewKurtosisConfig(userAcceptSendingMetrics bool) *KurtosisConfig {
	return &KurtosisConfig{shouldSendMetrics: userAcceptSendingMetrics}
}

func (config *KurtosisConfig) IsUserAcceptSendingMetrics() bool {
	return config.shouldSendMetrics
}

func (config *KurtosisConfig) SetUserAcceptSendingMetrics(userAcceptSendingMetrics bool) {
	config.shouldSendMetrics = userAcceptSendingMetrics
}

package config

type KurtosisConfig struct {
	userAcceptSendingMetrics bool
}

func NewKurtosisConfig(userAcceptSendingMetrics bool) *KurtosisConfig {
	return &KurtosisConfig{userAcceptSendingMetrics: userAcceptSendingMetrics}
}

func (config *KurtosisConfig) IsUserAcceptSendingMetrics() bool {
	return config.userAcceptSendingMetrics
}

func (config *KurtosisConfig) SetUserAcceptSendingMetrics(userAcceptSendingMetrics bool) {
	config.userAcceptSendingMetrics = userAcceptSendingMetrics
}

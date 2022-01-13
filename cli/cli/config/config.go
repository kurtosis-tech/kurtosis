package config

type Config interface {
	HasMetricsConsentPromptBeenDisplayed() bool
	MetricsConsentPromptHasBeenDisplayed()
	HasUserAcceptedSendingMetrics() bool
	UserAcceptSendingMetrics()
	UserDoNotAcceptSendingMetrics()
	Save() error
	String() string
}

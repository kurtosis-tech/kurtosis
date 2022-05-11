package kurtosis_config

/*
	KurtosisConfig should be the interface other modules use to access
	the latest configuration values available in Kurtosis CLI configuration.

	This prevents code using configuration from needing to completely change
	everytime configuration versions change.
 */

type KurtosisConfig struct {
	versionSpecificConfig *KurtosisConfigV1
}

func NewDefaultKurtosisConfig(doesUserAcceptSendingMetrics *bool) *KurtosisConfig {
	return &KurtosisConfig{
		versionSpecificConfig: NewDefaultKurtosisConfigV1(doesUserAcceptSendingMetrics),
	}
}

func NewKurtosisConfigFromConfigV0(v0 *KurtosisConfigV0) *KurtosisConfig {
	return &KurtosisConfig{
		versionSpecificConfig: NewDefaultKurtosisConfigV1(v0.ShouldSendMetrics),
	}
}

func NewKurtosisConfigFromConfigV1(v1 *KurtosisConfigV1) *KurtosisConfig {
	return &KurtosisConfig{
		versionSpecificConfig: v1,
	}
}

func (kurtosisConfig *KurtosisConfig) GetConfigVersion() int {
	return *kurtosisConfig.versionSpecificConfig.ConfigVersion
}

func (kurtosisConfig *KurtosisConfig) GetShouldSendMetrics() bool {
	return *kurtosisConfig.versionSpecificConfig.ShouldSendMetrics
}

func (kurtosisConfig *KurtosisConfig) GetKurtosisClusters() map[string]*KurtosisClusterV1 {
	return *kurtosisConfig.versionSpecificConfig.KurtosisClusters
}

func (kurtosisConfig *KurtosisConfig) GetVersionSpecificConfig() *KurtosisConfigV1 {
	return kurtosisConfig.versionSpecificConfig
}
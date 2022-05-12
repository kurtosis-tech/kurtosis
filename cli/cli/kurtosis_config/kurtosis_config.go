package kurtosis_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v1"
)

/*
	KurtosisConfig should be the interface other modules use to access
	the latest configuration values available in Kurtosis CLI configuration.

	This prevents code using configuration from needing to completely change
	everytime configuration versions change.
 */

type KurtosisConfig struct {
	versionSpecificConfig *v1.KurtosisConfigV1
}

type KurtosisClusterConfig struct {
	versionSpecificClusterConfig *v1.KurtosisClusterV1
}

func InitializeKurtosisConfigFromUserInput(didUserAcceptSendingMetrics bool) *KurtosisConfig {
	versionSpecificConfig := v1.NewDefaultKurtosisConfigV1()
	overrides := &v1.KurtosisConfigV1{ShouldSendMetrics: &didUserAcceptSendingMetrics}
	versionSpecificConfig.OverlayOverrides(overrides)
	return &KurtosisConfig{
		versionSpecificConfig: versionSpecificConfig,
	}
}

func NewKurtosisConfig(versionSpecificConfig *v1.KurtosisConfigV1) *KurtosisConfig {
	return &KurtosisConfig{
		versionSpecificConfig: versionSpecificConfig,
	}
}

func (kurtosisConfig *KurtosisConfig) Validate() error {
	return kurtosisConfig.versionSpecificConfig.Validate()
}

func (kurtosisConfig *KurtosisConfig) GetConfigVersion() int {
	return *kurtosisConfig.versionSpecificConfig.ConfigVersion
}

func (kurtosisConfig *KurtosisConfig) GetShouldSendMetrics() bool {
	return *kurtosisConfig.versionSpecificConfig.ShouldSendMetrics
}

func (kurtosisConfig *KurtosisConfig) GetKurtosisClusters() map[string]*KurtosisClusterConfig {
	clusterConfigMap := map[string]*KurtosisClusterConfig{}
	for clusterId, clusterConfigV1 := range *kurtosisConfig.versionSpecificConfig.KurtosisClusters {
		clusterConfigMap[clusterId] = &KurtosisClusterConfig{versionSpecificClusterConfig: clusterConfigV1}
	}
	return clusterConfigMap
}

func (kurtosisConfig *KurtosisConfig) GetVersionSpecificConfig() *v1.KurtosisConfigV1 {
	return kurtosisConfig.versionSpecificConfig
}
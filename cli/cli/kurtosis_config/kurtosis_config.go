package kurtosis_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v1"
	"github.com/kurtosis-tech/stacktrace"
)

/*
	KurtosisConfig should be the interface other modules use to access
	the latest configuration values available in Kurtosis CLI configuration.

	This prevents code using configuration from needing to completely change
	everytime configuration versions change.
 */

type KurtosisConfig struct {
	renderedConfig *v1.KurtosisConfigV1
	overrides      *v1.KurtosisConfigV1
}

type KurtosisClusterConfig struct {
	versionSpecificClusterConfig *v1.KurtosisClusterV1
}

func NewDefaultKurtosisConfig() *KurtosisConfig {
	defaultConfig := v1.NewDefaultKurtosisConfigV1()
	return &KurtosisConfig{
		renderedConfig: defaultConfig,
	}
}

func InitializeKurtosisConfigFromUserInput(didUserAcceptSendingMetrics bool) (*KurtosisConfig, error) {
	kurtosisConfig := NewDefaultKurtosisConfig()
	overrides := &v1.KurtosisConfigV1{ShouldSendMetrics: &didUserAcceptSendingMetrics}
	kurtosisConfigWithInput, err := kurtosisConfig.WithOverrides(overrides)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to render Kurtosis config from user input %t", didUserAcceptSendingMetrics)
	}
	return kurtosisConfigWithInput, nil
}

func (kurtosisConfig *KurtosisConfig) WithOverrides(overrides *v1.KurtosisConfigV1) (*KurtosisConfig, error) {
	renderedKurtosisConfig, err := kurtosisConfig.renderedConfig.OverlayOverrides(overrides)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to overlay configuration overrides on default Kurtosis configuration.")
	}
	return &KurtosisConfig{
		renderedConfig: renderedKurtosisConfig,
		overrides:      overrides,
	}, nil
}

func (kurtosisConfig *KurtosisConfig) Validate() error {
	return kurtosisConfig.renderedConfig.Validate()
}

func (kurtosisConfig *KurtosisConfig) GetOverrides() *v1.KurtosisConfigV1 {
	return kurtosisConfig.overrides
}

func (kurtosisConfig *KurtosisConfig) GetConfigVersion() config_version.ConfigVersion {
	return *kurtosisConfig.renderedConfig.ConfigVersion
}

func (kurtosisConfig *KurtosisConfig) GetShouldSendMetrics() bool {
	return *kurtosisConfig.renderedConfig.ShouldSendMetrics
}

func (kurtosisConfig *KurtosisConfig) GetKurtosisClusters() map[string]*KurtosisClusterConfig {
	clusterConfigMap := map[string]*KurtosisClusterConfig{}
	for clusterId, clusterConfigV1 := range *kurtosisConfig.renderedConfig.KurtosisClusters {
		clusterConfigMap[clusterId] = &KurtosisClusterConfig{versionSpecificClusterConfig: clusterConfigV1}
	}
	return clusterConfigMap
}
package kurtosis_config

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type KurtosisConfigProvider struct {
	configStore *kurtosisConfigStore
}

func NewKurtosisConfigProvider(configStore *kurtosisConfigStore) *KurtosisConfigProvider {
	return &KurtosisConfigProvider{configStore: configStore}
}

func NewDefaultKurtosisConfigProvider() *KurtosisConfigProvider {
	configStore := GetKurtosisConfigStore()
	configProvider := NewKurtosisConfigProvider(configStore)
	return configProvider
}

func (configProvider *KurtosisConfigProvider) GetOrInitializeConfig() (*KurtosisConfig, error){

	var (
		kurtosisConfig *KurtosisConfig
		err error
	)

	hasConfig := configProvider.configStore.HasConfig()
	if hasConfig {
		kurtosisConfig, err = configProvider.configStore.GetConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting config")
		}
	} else {
		kurtosisConfig, err = InitInteractiveConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred executing init interactive config")
		}
		if err = configProvider.configStore.SetConfig(kurtosisConfig); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
		}
	}
	logrus.Debugf("Loaded Kurtosis Config  %+v", kurtosisConfig)
	return kurtosisConfig, nil
}

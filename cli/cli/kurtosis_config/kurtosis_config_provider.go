package kurtosis_config

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type KurtosisConfigProvider struct {
	configStore *KurtosisConfigStore
}

func NewKurtosisConfigProvider(configStore *KurtosisConfigStore) *KurtosisConfigProvider {
	return &KurtosisConfigProvider{configStore: configStore}
}

func NewDefaultKurtosisConfigProvider() *KurtosisConfigProvider {
	configStore := newKurtosisConfigStore()
	configProvider := NewKurtosisConfigProvider(configStore)
	return configProvider
}

func (configProvider *KurtosisConfigProvider) IsConfigAlreadyCreated() bool {
	return configProvider.configStore.HasConfig()
}

func (configProvider *KurtosisConfigProvider) GetOrInitializeConfig() (*KurtosisConfig, error) {

	var (
		kurtosisConfig *KurtosisConfig
		err            error
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

		//Saving config
		if err = configProvider.configStore.SetConfig(kurtosisConfig); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
		}
	}
	logrus.Debugf("Loaded Kurtosis Config  %+v", kurtosisConfig)
	return kurtosisConfig, nil
}

func (configProvider *KurtosisConfigProvider) SetConfig(kurtosisConfig *KurtosisConfig) error {
	if err := configProvider.configStore.SetConfig(kurtosisConfig); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
	}
	return nil
}

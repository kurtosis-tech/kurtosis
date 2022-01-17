package config

import "github.com/kurtosis-tech/stacktrace"

type ConfigProvider struct {
	configStore *ConfigStore
	configInitializer *ConfigInitializer
}

func NewConfigProvider(configStore *ConfigStore, configInitializer *ConfigInitializer) *ConfigProvider {
	return &ConfigProvider{configStore: configStore, configInitializer: configInitializer}
}

func (configProvider *ConfigProvider) GetOrInitializeConfig() (*KurtosisConfig, error){

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
		kurtosisConfig, err = configProvider.configInitializer.InitInteractiveConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred executing init interactive config")
		}
		if err = configProvider.configStore.SetConfig(kurtosisConfig); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
		}
	}

	return kurtosisConfig, nil
}

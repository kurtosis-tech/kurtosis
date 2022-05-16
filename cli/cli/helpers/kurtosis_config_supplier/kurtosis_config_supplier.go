package kurtosis_config_supplier

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/stacktrace"
)

func GetKurtosisConfig() (*resolved_config.KurtosisConfig, error) {
	configStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(configStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		 return nil, stacktrace.Propagate(err, "An error occurred getting or initializing the Kurtosis config")
	}
	return kurtosisConfig, nil
}

package kurtosis_config

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/email_collector"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type KurtosisConfigProvider struct {
	configStore *kurtosisConfigStore
}

func NewKurtosisConfigProvider(configStore *kurtosisConfigStore) *KurtosisConfigProvider {
	return &KurtosisConfigProvider{configStore: configStore}
}

func (configProvider *KurtosisConfigProvider) GetOrInitializeConfig() (*resolved_config.KurtosisConfig, error) {

	var (
		kurtosisConfig *resolved_config.KurtosisConfig
		err            error
	)

	doesKurtosisConfigAlreadyExists, err := configProvider.configStore.HasConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking if Kurtosis config already exists")
	}
	if doesKurtosisConfigAlreadyExists {
		kurtosisConfig, err = configProvider.configStore.GetConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting config")
		}
	} else {
		kurtosisConfig, err = initConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred executing init interactive config")
		}

		// only ask for input on interactive terminal
		if interactive_terminal_decider.IsInteractiveTerminal() {
			email_collector.AskUserForEmailAndLogIt()
		}

		if err = configProvider.configStore.SetConfig(kurtosisConfig); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
		}
	}
	logrus.Debugf("Loaded Kurtosis Config  %+v", kurtosisConfig)
	return kurtosisConfig, nil
}

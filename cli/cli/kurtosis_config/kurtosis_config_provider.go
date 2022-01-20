package kurtosis_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/machine_id_provider"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_tracker"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/metrics-library/golang/lib/client/snow_plow_client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type KurtosisConfigProvider struct {
	configStore *KurtosisConfigStore
	configInitializer *KurtosisConfigInitializer
}

func NewKurtosisConfigProvider(configStore *KurtosisConfigStore, configInitializer *KurtosisConfigInitializer) *KurtosisConfigProvider {
	return &KurtosisConfigProvider{configStore: configStore, configInitializer: configInitializer}
}

func NewDefaultKurtosisConfigProvider() *KurtosisConfigProvider {
	configStore := newKurtosisConfigStore()
	promptDisplayer := prompt_displayer.NewPromptDisplayer()
	configInitializer := newKurtosisConfigInitializer(promptDisplayer)

	configProvider := NewKurtosisConfigProvider(configStore, configInitializer)
	return configProvider
}

func (configProvider *KurtosisConfigProvider) IsConfigAlreadyCreated() bool {
	return configProvider.configStore.HasConfig()
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
		kurtosisConfig, err = configProvider.configInitializer.InitInteractiveConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred executing init interactive config")
		}
		if err = configProvider.configStore.SetConfig(kurtosisConfig); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
		}

		userId, err := machine_id_provider.GetProtectedMachineID()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting protected machine ID")
		}

		metricsClient, err := snow_plow_client.NewSnowPlowClient(source.KurtosisCLISource, userId)
		if err != nil {
			//If tracking fails, we don't throw and error, because we don't want to interrupt user's execution
			logrus.Debugf("An error occurred creating SnowPlow metrics client\n%v", err)
		} else {
			metricsTracker := metrics_tracker.NewMetricsTracker(metricsClient)

			if err = metricsTracker.TrackUserAcceptSendingMetrics(kurtosisConfig.IsUserAcceptSendingMetrics()); err != nil {
				//If tracking fails, we don't throw and error, because we don't want to interrupt user's execution
				logrus.Debugf("An error occurred knowing if user accept sending metrics\n%v", err)
			}

			if !kurtosisConfig.IsUserAcceptSendingMetrics() {
				metricsTracker.DisableTracking()
			}
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

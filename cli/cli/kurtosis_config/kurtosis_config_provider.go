package kurtosis_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_tracker"
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

		//Tracking user metrics consent
		metricsClient, err := metrics_tracker.CreateMetricsClient()
		if err != nil {
			//We don't throw and error if this fails, because we don't want to interrupt user's execution
			logrus.Debugf("An error occurred creating SnowPlow metrics client\n%v", err)
		} else {
			metricsTracker := metrics_tracker.NewMetricsTracker(metricsClient)
			if err = metricsTracker.TrackUserAcceptSendingMetrics(kurtosisConfig.IsUserAcceptSendingMetrics()); err != nil {
				//We don't throw and error if this fails, because we don't want to interrupt user's execution
				logrus.Debugf("An error occurred knowing if user accept sending metrics\n%v", err)
			}
			if !kurtosisConfig.IsUserAcceptSendingMetrics() {
				//If user reject sending metrics the feature will be disabled
				metricsTracker.DisableTracking()
			}
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

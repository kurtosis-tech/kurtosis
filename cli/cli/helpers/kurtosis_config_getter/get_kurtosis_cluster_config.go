package kurtosis_config_getter

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultClusterName = resolved_config.DefaultDockerClusterName
)

func getKurtosisClusterName() (string, error) {
	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()

	isClusterSet, err := clusterSettingStore.HasClusterSetting()
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to check if cluster setting has been set.")
	}
	var clusterName string
	if !isClusterSet {
		// If the user has not yet set a cluster, use default
		clusterName = defaultClusterName
	} else {
		clusterName, err = clusterSettingStore.GetClusterSetting()
		if err != nil {
			return "", stacktrace.Propagate(err, "Expected to be able to get cluster s.")
		}
	}
	return clusterName, nil
}

func getKurtosisConfig() (*resolved_config.KurtosisConfig, error) {
	configStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(configStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting or initializing the Kurtosis config")
	}
	return kurtosisConfig, nil
}

func GetKurtosisClusterConfig() (*resolved_config.KurtosisClusterConfig, error) {
	clusterName, err := getKurtosisClusterName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get Kurtosis Cluster Name from Kurtosis settings, instead a non-nil error was returned")
	}

	kurtosisConfig, err := getKurtosisConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting Kurtosis configuration")
	}

	clusterConfig, found := kurtosisConfig.GetKurtosisClusters()[clusterName]
	if !found {
		return nil, stacktrace.NewError("Expected to find Kurtosis configuration for cluster '%v', instead found nothing", clusterName)
	}

	return clusterConfig, nil
}

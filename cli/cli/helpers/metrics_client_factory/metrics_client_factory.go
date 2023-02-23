package metrics_client_factory

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shouldFlushMetricsClientQueueOnEachEvent           = false
	defaultSendingUserMetricsIfConfigurationIsNotFound = true
)

func GetMetricsClient() (metrics_client.MetricsClient, func() error, error) {

	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()
	isClusterSet, err := clusterSettingStore.HasClusterSetting()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to check if cluster setting has been set.")
	}
	clusterType := resolved_config.DefaultDockerClusterName
	if isClusterSet {
		clusterType, err = clusterSettingStore.GetClusterSetting()
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Cluster is set but config couldn't be fetched")
		}
	}

	kurtosisConfigStore := kurtosis_config.GetKurtosisConfigStore()
	hasConfig, err := kurtosisConfigStore.HasConfig()
	if err != nil {
		return nil, nil, stacktrace.NewError("An error occurred while determining whether configuration already exists")
	}
	var sendUserMetrics bool
	if hasConfig {
		kurtosisConfig, err := kurtosisConfigStore.GetConfig()
		if err != nil {
			return nil, nil, stacktrace.NewError("An error occurred while fetching stored configuration")
		}
		sendUserMetrics = kurtosisConfig.GetShouldSendMetrics()
	} else {
		sendUserMetrics = defaultSendingUserMetricsIfConfigurationIsNotFound
	}

	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while getting the users metrics id")
	}

	metricsClient, metricsClientCloseFunc, err := metrics_client.CreateMetricsClient(
		source.KurtosisCLISource,
		kurtosis_version.KurtosisVersion,
		metricsUserId,
		clusterType,
		sendUserMetrics,
		shouldFlushMetricsClientQueueOnEachEvent,
		newDoNothingMetricsClientCallback(),
	)

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while creating the metrics client")
	}

	return metricsClient, metricsClientCloseFunc, nil
}

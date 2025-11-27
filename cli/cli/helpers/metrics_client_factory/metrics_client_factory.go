package metrics_client_factory

import (
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/defaults"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/metrics_cloud_user_instance_id_helper"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/metrics_user_id_store"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/dzobbe/PoTE-kurtosis/kurtosis_version"
	"github.com/dzobbe/PoTE-kurtosis/metrics-library/golang/lib/analytics_logger"
	"github.com/dzobbe/PoTE-kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/dzobbe/PoTE-kurtosis/metrics-library/golang/lib/source"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	shouldFlushMetricsClientQueueOnEachEvent = false
)

func GetMetricsClient() (metrics_client.MetricsClient, func() error, error) {
	metricsUserId, clusterType, err := getMetricsUserIdAndClusterType()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "an error occurred while getting metrics user id and cluster type")
	}

	kurtosisConfigStore := kurtosis_config.GetKurtosisConfigStore()
	hasConfig, err := kurtosisConfigStore.HasConfig()
	if err != nil {
		return nil, nil, stacktrace.NewError("An error occurred while determining whether configuration already exists")
	}

	maybeCloudUserId, maybeCloudInstanceId := metrics_cloud_user_instance_id_helper.GetMaybeCloudUserAndInstanceID()

	var sendUserMetrics bool
	if hasConfig {
		kurtosisConfig, err := kurtosisConfigStore.GetConfig()
		if err != nil {
			return nil, nil, stacktrace.NewError("An error occurred while fetching stored configuration")
		}
		sendUserMetrics = kurtosisConfig.GetShouldSendMetrics()
	} else {
		sendUserMetrics = defaults.SendMetricsByDefault
	}

	logger := logrus.StandardLogger()
	metricsClient, metricsClientCloseFunc, err := metrics_client.CreateMetricsClient(
		metrics_client.NewMetricsClientCreatorOption(
			source.KurtosisCLISource,
			kurtosis_version.KurtosisVersion,
			metricsUserId,
			clusterType,
			sendUserMetrics,
			shouldFlushMetricsClientQueueOnEachEvent,
			metrics_client.DoNothingMetricsClientCallback{},
			analytics_logger.ConvertLogrusLoggerToAnalyticsLogger(logger),
			metrics_client.IsCI(), maybeCloudUserId, maybeCloudInstanceId),
	)

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while creating the metrics client")
	}

	return metricsClient, metricsClientCloseFunc, nil
}

// GetSegmentClient use this method only if you are sure that you want to send metrics otherwise use GetMetricsClient
func GetSegmentClient() (metrics_client.MetricsClient, func() error, error) {
	metricsUserId, clusterType, err := getMetricsUserIdAndClusterType()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "an error occurred while getting metrics user id and cluster type")
	}
	// this is force set to true in order to get the segment client
	sendUserMetrics := true

	maybeCloudUserId, maybeCloudInstanceId := metrics_cloud_user_instance_id_helper.GetMaybeCloudUserAndInstanceID()

	logger := logrus.StandardLogger()
	metricsClient, metricsClientCloseFunc, err := metrics_client.CreateMetricsClient(
		metrics_client.NewMetricsClientCreatorOption(source.KurtosisCLISource,
			kurtosis_version.KurtosisVersion,
			metricsUserId,
			clusterType,
			sendUserMetrics,
			shouldFlushMetricsClientQueueOnEachEvent,
			metrics_client.DoNothingMetricsClientCallback{},
			analytics_logger.ConvertLogrusLoggerToAnalyticsLogger(logger),
			metrics_client.IsCI(), maybeCloudUserId, maybeCloudInstanceId),
	)

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while creating the metrics client")
	}

	return metricsClient, metricsClientCloseFunc, nil
}

func getMetricsUserIdAndClusterType() (string, string, error) {
	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()
	isClusterSet, err := clusterSettingStore.HasClusterSetting()
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to check if cluster setting has been set.")
	}
	clusterType := resolved_config.DefaultDockerClusterName
	if isClusterSet {
		clusterType, err = clusterSettingStore.GetClusterSetting()
		if err != nil {
			return "", "", stacktrace.Propagate(err, "Cluster is set but config couldn't be fetched")
		}
	}

	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred while getting the users metrics id")
	}

	return metricsUserId, clusterType, nil
}

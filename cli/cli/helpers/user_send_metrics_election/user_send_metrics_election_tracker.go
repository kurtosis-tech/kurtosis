package user_send_metrics_election

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_cloud_user_instance_id_helper"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_send_metrics_election/user_metrics_election_event_backlog"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/analytics_logger"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/source"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	shouldFlushMetricsClientQueueOnEachEvent                 = true
	didUserAcceptSendingMetricsValueForMetricsClientCreation = true
)

func SendAnyBackloggedUserMetricsElectionEvent() error {

	userMetricsElectionEventBacklog := user_metrics_election_event_backlog.GetUserMetricsElectionEventBacklog()
	shouldSendMetrics, hasBackloggedEvent, err := userMetricsElectionEventBacklog.Get()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if a user-consent-to-send-metrics-election backlog exists")
	}

	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()

	isClusterSet, err := clusterSettingStore.HasClusterSetting()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to check if cluster setting has been set.")
	}

	clusterType := resolved_config.DefaultDockerClusterName
	if isClusterSet {
		clusterType, err = clusterSettingStore.GetClusterSetting()
		if err != nil {
			return stacktrace.Propagate(err, "Cluster is set but config couldn't be fetched")
		}
	}

	maybeCloudUserID, maybeCloudInstanceID := metrics_cloud_user_instance_id_helper.GetMaybeCloudUserAndInstanceID()

	if hasBackloggedEvent {
		metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()

		metricsUserId, err := metricsUserIdStore.GetUserID()
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting metrics user ID")
		}

		metricsClientCallback := &metricsElectionEventBacklogClearingCallback{}

		logger := logrus.StandardLogger()
		// This is a special metrics client that, will record their decision about whether to send metrics or not
		metricsClient, metricsClientCloseFunc, err := metrics_client.CreateMetricsClient(
			metrics_client.NewMetricsClientCreatorOption(
				source.KurtosisCLISource,
				kurtosis_version.KurtosisVersion,
				metricsUserId,
				clusterType,
				didUserAcceptSendingMetricsValueForMetricsClientCreation,
				shouldFlushMetricsClientQueueOnEachEvent,
				metricsClientCallback,
				analytics_logger.ConvertLogrusLoggerToAnalyticsLogger(logger),
				metrics_client.IsCI(),
				maybeCloudUserID,
				maybeCloudInstanceID,
			),
		)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating the metrics client for recording send-metrics election")
		}
		defer func() {
			if err := metricsClientCloseFunc(); err != nil {
				logrus.Debugf("We tried to close the metrics client, but doing so threw an error:\n%v", err)
			}
		}()

		if err := metricsClient.TrackShouldSendMetricsUserElection(shouldSendMetrics); err != nil {
			return stacktrace.Propagate(err, "An error occurred tracking should-send-metrics user election event")
		}
	}

	return nil
}

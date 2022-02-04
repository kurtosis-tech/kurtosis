package user_consent_to_send_metrics_election

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	shouldFlushMetricsClientQueueOnEachEvent                                = true
	didUserAcceptSendingMetricsValueIfUserConsentToSendMetricsElectionExist = true
)

func TrackUserConsentToSendingMetricsElection() error {

	userConsentToSendMetricsElectionStore := GetUserConsentToSendMetricsElectionStore()
	doesUserConsentToSendMetricsElectionExist, err := userConsentToSendMetricsElectionStore.Exist()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if user consent to send metrics election exist")
	}

	if doesUserConsentToSendMetricsElectionExist {
		metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()

		metricsUserId, err := metricsUserIdStore.GetUserID()
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting metrics user ID")
		}

		mtrClientCallback := &metricsClientCallbackObject{}

		// This is a special metrics client that, if the user allows, will record their decision about whether to send metrics or not
		metricsClient, metricsClientCloseFunc, err := metrics_client.CreateMetricsClient(
			source.KurtosisCLISource,
			kurtosis_cli_version.KurtosisCLIVersion,
			metricsUserId,
			didUserAcceptSendingMetricsValueIfUserConsentToSendMetricsElectionExist,
			shouldFlushMetricsClientQueueOnEachEvent,
			mtrClientCallback,
		)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating the metrics client")
		}
		defer func() {
			if err := metricsClientCloseFunc(); err != nil {
				logrus.Debugf("We tried to close the metrics client, but doing so threw an error:\n%v", err)
			}
		}()

		if err := metricsClient.TrackShouldSendMetricsUserElection(didUserAcceptSendingMetricsValueIfUserConsentToSendMetricsElectionExist); err != nil {
			return stacktrace.Propagate(err, "An error occurred tracking should send metrics user election event")
		}
	}

	return nil
}

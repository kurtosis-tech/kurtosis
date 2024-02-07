package email_collector

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_cloud_user_instance_id_helper"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/analytics_logger"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/source"
	"github.com/sirupsen/logrus"
)

const (
	defaultEmailValue     = ""
	emailValueInputPrompt = "(Optional) Share your email address for occasional updates & outreach for product feedback from Kurtosis"
	sendUserMetrics       = true
	flushQueueOnEachEvent = false
)

func AskUserForEmailAndLogIt() {
	userEmail, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(emailValueInputPrompt, defaultEmailValue)
	if err != nil {
		logrus.Debugf("The user tried to input his email address but it failed")
	}

	if userEmail != defaultEmailValue {
		logUserEmailAddressAsMetric(userEmail)
	}

}

// TODO this recreates a metrics client instead of using the factory as there are circular dependencies - clean this up
func logUserEmailAddressAsMetric(userEmail string) {
	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		logrus.Debugf("an error occurred while getting users metrics id:\n%v", err)
		return
	}
	logger := logrus.StandardLogger()

	maybeCloudUserID, maybeCloudInstanceID := metrics_cloud_user_instance_id_helper.GetMaybeCloudUserAndInstanceID()

	metricsClient, metricsClientCloseFunc, err := metrics_client.CreateMetricsClient(
		metrics_client.NewMetricsClientCreatorOption(
			source.KurtosisCLISource,
			kurtosis_version.GetVersion(),
			metricsUserId,
			// TODO this isn't relevant for the metric also this only runs at first install;
			// The user hasn't ever used Kurtosis yet so it has to be the default cluster
			resolved_config.DefaultDockerClusterName,
			sendUserMetrics,
			flushQueueOnEachEvent,
			metrics_client.DoNothingMetricsClientCallback{},
			analytics_logger.ConvertLogrusLoggerToAnalyticsLogger(logger),
			metrics_client.IsCI(),
			maybeCloudUserID,
			maybeCloudInstanceID),
	)
	if err != nil {
		logrus.Debugf("tried creating a metrics client but failed with error:\n%v", err)
		return
	}
	defer func() {
		err = metricsClientCloseFunc()
		if err != nil {
			logrus.Debugf("an error occurred while closing the metrics client:\n%v", err)
		}
	}()
	if err = metricsClient.TrackUserSharedEmailAddress(userEmail); err != nil {
		logrus.Debugf("tried sending user email address as metric but failed:\n%v", err)
		return
	}
}

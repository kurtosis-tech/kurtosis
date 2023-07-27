package email_collector

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/do_nothing_metrics_client_callback"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/logrus_logger_converter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	"github.com/sirupsen/logrus"
)

const (
	defaultEmailValue     = ""
	emailValueInputPrompt = "If you wish to share your email address Kurtosis will occasionally share updates with you"
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

func logUserEmailAddressAsMetric(userEmail string) {
	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	logger := logrus.StandardLogger()

	metricsClient, metricsClientCloseFunc, err := metrics_client.CreateMetricsClient(
		source.KurtosisCLISource,
		kurtosis_version.KurtosisVersion,
		metricsUserId,
		resolved_config.DefaultDockerClusterName,
		sendUserMetrics,
		flushQueueOnEachEvent,
		do_nothing_metrics_client_callback.NewDoNothingMetricsClientCallback(),
		logrus_logger_converter.ConvertLogrusLoggerToAnalyticsLogger(logger),
	)
	if err != nil {
		logrus.Debugf("tried creating a metrics client but failed with error:\n%v", err)
		return
	}
	defer metricsClientCloseFunc()
	if err = metricsClient.TrackUserSharedEmailAddress(userEmail); err != nil {
		logrus.Debugf("tried sending user email address as metric but failed:\n%v", err)
		return
	}
}

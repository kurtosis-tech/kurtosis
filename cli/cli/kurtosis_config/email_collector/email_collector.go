package email_collector

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_client_factory"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/prompt_displayer"
	"github.com/sirupsen/logrus"
)

const (
	defaultEmailValue     = ""
	emailValueInputPrompt = "If you wish to share your email address Kurtosis will occasionally share updates with you"
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
	metricsClient, metricsClientCloser, err := metrics_client_factory.GetMetricsClient()
	if err != nil {
		logrus.Debugf("tried creating a metrics client log user email address but failed:\n%v", err)
		return
	}
	defer metricsClientCloser()
	if err = metricsClient.TrackUserSharedEmailAddress(userEmail); err != nil {
		logrus.Debugf("tried sending user email address as metric but failed:\n%v", err)
	}
}

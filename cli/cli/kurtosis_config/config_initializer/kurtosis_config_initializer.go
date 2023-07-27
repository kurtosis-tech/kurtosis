package config_initializer

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_client_factory"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_send_metrics_election"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_send_metrics_election/user_metrics_election_event_backlog"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultEmailValue     = ""
	emailValueInputPrompt = "If you wish to share your email address Kurtosis will occasionally share updates with you"
)

func InitConfig() (*resolved_config.KurtosisConfig, error) {
	printMetricsPreface()

	userEmail, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(emailValueInputPrompt, defaultEmailValue)
	if err != nil {
		logrus.Debugf("The user tried to input his email address but it failed")
	}

	if userEmail != defaultEmailValue {
		logUserEmailAddressAsMetric(userEmail)
	}

	userMetricsElectionEventBacklog := user_metrics_election_event_backlog.GetUserMetricsElectionEventBacklog()
	if err := userMetricsElectionEventBacklog.Set(defaults.SendMetricsByDefault); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Debugf("An error occurred creating user-consent-to-send-metrics election file\n%v", err)
	}
	//Here we are trying to send this metric for first time, but if it fails we'll continue to retry every time the CLI runs
	if err := user_send_metrics_election.SendAnyBackloggedUserMetricsElectionEvent(); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Debugf("An error occurred tracking user-consent-to-send-metrics election\n%v", err)
	}

	kurtosisConfig, err := resolved_config.NewKurtosisConfigFromRequiredFields(defaults.SendMetricsByDefault)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to initialize Kurtosis configuration from user input %t.", defaults.SendMetricsByDefault)
	}
	return kurtosisConfig, nil
}

func printMetricsPreface() {
	fmt.Println("")
	fmt.Println("The Kurtosis CLI collects user metrics by default. These metrics are anonymized, private & obfuscated. These metrics help us better understand what features are used, what features to invest in and what features might be buggy.")
	fmt.Println("In case you wish to not send metrics, you can do so by running - kurtosis analytics disable")
	fmt.Printf("Read more at %v\n", user_support_constants.MetricsPhilosophyDocs)
	fmt.Println("")
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

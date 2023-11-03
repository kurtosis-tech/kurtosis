package kurtosis_config

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_send_metrics_election"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_send_metrics_election/user_metrics_election_event_backlog"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func initConfig() (*resolved_config.KurtosisConfig, error) {
	printMetricsPreface()

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

package kurtosis_config

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_optin"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	metricsConsentPromptLabel       = "Is it okay to send anonymized metrics purely to improve the product?"
	secondMetricsConsentPromptLabel = "That's okay; we understand. Would it be alright if we send a one-time event recording your opt-out so we can see how much users dislike the metrics? Regardless of your choice, no other events will be tracked per your election."

	shouldSendMetricsDefaultValue = true
	shouldSendMetricsOptOutEventDefaultValue = true

	shouldFlushMetricsClientQueueOnEachEvent = true
)

func initInteractiveConfig() (*KurtosisConfig, error) {

	fmt.Println(metrics_optin.WhyKurtosisCollectMetricsDescriptionNote)

	didUserAcceptSendingMetrics, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(metricsConsentPromptLabel, shouldSendMetricsDefaultValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
	}
	didUserConsentToSendMetricsElectionEvent := didUserAcceptSendingMetrics

	if !didUserAcceptSendingMetrics {
		didUserConsentToSendMetricsElectionEvent, err = prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(secondMetricsConsentPromptLabel, shouldSendMetricsOptOutEventDefaultValue)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
		}
	}

	if didUserConsentToSendMetricsElectionEvent {
		metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()

		metricsUserId, err := metricsUserIdStore.GetUserID()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting metrics user ID")
		}

		// This is a special metrics client that, if the user allows, will record their decision about whether to send metrics or not
		metricsClient, err := metrics_client.CreateMetricsClient(source.KurtosisCLISource, kurtosis_cli_version.KurtosisCLIVersion, metricsUserId, didUserConsentToSendMetricsElectionEvent, shouldFlushMetricsClientQueueOnEachEvent)
		if err != nil {
			return nil,  stacktrace.Propagate(err, "An error occurred creating the metrics client")
		}
		defer func() {
			if err := metricsClient.Close(); err != nil {
				logrus.Warnf("We tried to close the metrics client, but doing so threw an error:\n%v", err)
			}
		}()

		if err := metricsClient.TrackShouldSendMetricsUserElection(didUserAcceptSendingMetrics); err != nil {
			//We don't want to interrupt users flow if something fails when tracking metrics
			logrus.Errorf("An error occurred tracking should send metrics user election event\n%v",err)
		}
	}

	kurtosisConfig := NewKurtosisConfig(didUserAcceptSendingMetrics)
	return kurtosisConfig, nil
}

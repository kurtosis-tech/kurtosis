package kurtosis_config

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_optin"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/user_consent_to_send_metrics_election"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	metricsConsentPromptLabel       = "Is it okay to send anonymized metrics purely to improve the product?"
	secondMetricsConsentPromptLabel = "That's okay; we understand. Would it be alright if we send a one-time event recording your opt-out so we can see how much users dislike the metrics? Regardless of your choice, no other events will be tracked per your election."

	shouldSendMetricsDefaultValue = true
	shouldSendMetricsOptOutEventDefaultValue = true
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
		userConsentToSendMetricsElectionStore := user_consent_to_send_metrics_election.GetUserConsentToSendMetricsElectionStore()
		if err := userConsentToSendMetricsElectionStore.Create(); err != nil {
			//We don't want to interrupt users flow if something fails when tracking metrics
			logrus.Debugf("An error occurred creating user consent to send metrics election file\n%v",err)
		}
	}

	kurtosisConfig := NewKurtosisConfig(didUserAcceptSendingMetrics)
	return kurtosisConfig, nil
}

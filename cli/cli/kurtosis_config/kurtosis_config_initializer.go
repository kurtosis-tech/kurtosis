package kurtosis_config

import (
	"fmt"
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
	secondMetricsConsentPromptLabel = "Is it alright if we record your opt-out as a one-time event so we can see how much users dislike metrics?"

	shouldSendMetricsDefaultValue = true
	shouldSendMetricsOptOutEventDefaultValue = true

	shouldFlushMetricsClientQueueOnEachEvent = true
)

func initInteractiveConfig() (*KurtosisConfig, error) {
	printMetricsPreface()

	didUserAcceptSendingMetrics, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(metricsConsentPromptLabel, shouldSendMetricsDefaultValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
	}
	didUserConsentToSendMetricsElectionEvent := didUserAcceptSendingMetrics

	if !didUserAcceptSendingMetrics {
		fmt.Println("That's okay; we understand. No product analytic metrics will be collected from this point forward.")
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

func printMetricsPreface() {
	fmt.Println("")
	fmt.Println("============================================================================================")
	fmt.Println("                               Metrics Election Preface")
	fmt.Println("============================================================================================")
	fmt.Println("The Kurtosis CLI has the potential collect anonymized product analytics metrics, and I want")
	fmt.Println("to explain why this big message is popping up in your face right now.")
	fmt.Println("")
	fmt.Println("User metrics are useful for detecting product bugs and seeing feature usage. However, they're")
	fmt.Println("also heavily abused to invade our privacy. I hate it as much as I'm guessing you do.")
	fmt.Println("")
	fmt.Println("Kurtosis is a small startup so product analytics metrics are vital for us to determine")
	fmt.Println("where to put resources, but it was important to me that we do this ethically.")
	fmt.Println("")
	fmt.Println("Therefore, we've made our metrics:")
	fmt.Println(" - private: we will *never* give or sell your data to third parties")
	fmt.Println(" - opt-in: we require you to make a choice about collection rather than assuming you want in")
	fmt.Println(" - anonymized: your user ID is a hash; we don't know who you are")
	fmt.Println(" - obfuscated: potentially-sensitive parameters (like module exec params) are hashed as well")
	fmt.Println("")
	fmt.Println("If that sounds fair to you, we'd really appreciate you helping us get the data to make our product")
	fmt.Println("better. In exchange, you have my personal word to honor the trust you place in us by fulfilling")
	fmt.Println("our metrics promises above.")
	fmt.Println("")
	fmt.Println("Sincerely,")
	fmt.Println("Kevin Today, Kurtosis CTO")
	fmt.Println("")
}
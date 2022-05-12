package kurtosis_config

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/user_send_metrics_election"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/user_send_metrics_election/user_metrics_election_event_backlog"
	"github.com/kurtosis-tech/kurtosis-cli/cli/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	metricsConsentPromptLabel       = "Is it okay to send anonymized metrics purely to improve the product?"
	secondMetricsConsentPromptLabel = "Is it alright if we record your opt-out as a one-time event so we can see how much users dislike metrics?"

	shouldSendMetricsDefaultValue = true
	shouldSendMetricsOptOutEventDefaultValue = true
)

func initInteractiveConfig() (*KurtosisConfig, error) {
	// Check if we're actually running in interactive mode (i.e. STDOUT is a terminal)
	if !interactive_terminal_decider.IsInteractiveTerminal() {
		return nil, stacktrace.NewError(
			"The Kurtosis config isn't initialized so we'd initialize it interactively here except STDOUT isn't " +
				"a terminal (indicating that this is probably running in CI) which means that you'll need to manually " +
				"initialize the config using the instructions here: %v",
			user_support_constants.CLISetupDocsUrl,
		)
	}

	printMetricsPreface()

	didUserAcceptSendingMetrics, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(metricsConsentPromptLabel, shouldSendMetricsDefaultValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred displaying user-metrics-consent prompt")
	}
	didUserConsentToSendMetricsElectionEvent := didUserAcceptSendingMetrics

	if !didUserAcceptSendingMetrics {
		fmt.Println("That's okay; we understand. No product analytic metrics will be collected from this point forward.")
		didUserConsentToSendMetricsElectionEvent, err = prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(secondMetricsConsentPromptLabel, shouldSendMetricsOptOutEventDefaultValue)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred displaying user-metrics-consent prompt")
		}
	}

	if didUserConsentToSendMetricsElectionEvent {
		userMetricsElectionEventBacklog := user_metrics_election_event_backlog.GetUserMetricsElectionEventBacklog()
		if err := userMetricsElectionEventBacklog.Set(didUserAcceptSendingMetrics); err != nil {
			//We don't want to interrupt users flow if something fails when tracking metrics
			logrus.Debugf("An error occurred creating user-consent-to-send-metrics election file\n%v",err)
		}
		//Here we are trying to send this metric for first time, but if it fails we'll continue to retry every time the CLI runs
		if err := user_send_metrics_election.SendAnyBackloggedUserMetricsElectionEvent(); err != nil {
			//We don't want to interrupt users flow if something fails when tracking metrics
			logrus.Debugf("An error occurred tracking user-consent-to-send-metrics election\n%v",err)
		}
	}

	kurtosisConfig, err := InitializeKurtosisConfigFromUserInput(didUserAcceptSendingMetrics)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to initialize Kurtosis configuration from user input %t.", didUserAcceptSendingMetrics)
	}
	return kurtosisConfig, nil
}

func printMetricsPreface() {
	fmt.Println("")
	fmt.Println("============================================================================================")
	fmt.Println("                               Metrics Election Preface")
	fmt.Println("============================================================================================")
	fmt.Println("The Kurtosis CLI has the potential to collect anonymized product analytics metrics, and I want")
	fmt.Println("to explain why this big message is popping up in your face right now.")
	fmt.Println("")
	fmt.Println("User metrics are useful for detecting product bugs and seeing feature usage. However, they're")
	fmt.Println("also heavily abused to invade our privacy. I hate it as much as I'm guessing you do.")
	fmt.Println("")
	fmt.Println("Kurtosis is a small startup so product analytics metrics are vital for us to determine")
	fmt.Println("where to put resources, but it was important to me that we do this ethically.")
	fmt.Println("")
	fmt.Println("Therefore, our metrics are")
	fmt.Println(" - private: we will *never* give or sell your data to third parties")
	fmt.Println(" - opt-in: we require you to make a choice about collection rather than assuming you want in")
	fmt.Println(" - anonymized: your user ID is a hash; we don't know who you are")
	fmt.Println(" - obfuscated: potentially-sensitive parameters (like module exec params) are hashed as well")
	fmt.Println("")
	fmt.Println("If that sounds fair to you, we'd really appreciate you helping us get the data to make the product")
	fmt.Println("better. In exchange, you have my personal word to honor the trust you place in us by fulfilling")
	fmt.Println("our metrics promises above.")
	fmt.Println("")
	fmt.Println("Sincerely,")
	fmt.Println("Kevin Today, Kurtosis CTO")
	fmt.Println("")
}
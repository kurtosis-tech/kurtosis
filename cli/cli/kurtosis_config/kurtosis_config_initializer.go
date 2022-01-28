package kurtosis_config

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_optin"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	metricsPromptLabel = "Is it okay to send anonymized metrics purely to improve the product?"
)

func initInteractiveConfig() (*KurtosisConfig, error) {

	fmt.Println(metrics_optin.WhyKurtosisCollectMetricsDescriptionNote)

	didUserAcceptSendingMetrics, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(metricsPromptLabel, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
	}

	kurtosisConfig := NewKurtosisConfig(didUserAcceptSendingMetrics)
	return kurtosisConfig, nil
}

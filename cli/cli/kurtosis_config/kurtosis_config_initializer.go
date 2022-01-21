package kurtosis_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	metricsPromptLabel = "Do you accept collecting and sending metrics to improve the product?"
)

func InitInteractiveConfig() (*KurtosisConfig, error) {

	userInputResult, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(metricsPromptLabel, prompt_displayer.YesInput)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
	}

	kurtosisConfig := NewKurtosisConfig(userInputResult)
	return kurtosisConfig, nil
}

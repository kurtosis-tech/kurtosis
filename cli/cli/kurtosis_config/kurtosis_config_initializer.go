package kurtosis_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/stacktrace"
)

type KurtosisConfigInitializer struct {
	promptDisplayer *prompt_displayer.PromptDisplayer
}

func newKurtosisConfigInitializer(promptDisplayer *prompt_displayer.PromptDisplayer) *KurtosisConfigInitializer {
	return &KurtosisConfigInitializer{promptDisplayer: promptDisplayer}
}

func (configInitializer *KurtosisConfigInitializer) InitInteractiveConfig() (*KurtosisConfig, error) {

	userInputResult, err := configInitializer.promptDisplayer.DisplayUserMetricsConsentPromptAndGetUserInputResult()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
	}

	kurtosisConfig := NewKurtosisConfig(userInputResult)
	return kurtosisConfig, nil
}

package config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/stacktrace"
)

type ConfigInitializer struct {
	promptDisplayer *prompt_displayer.PromptDisplayer
}

func NewConfigInitializer(promptDisplayer *prompt_displayer.PromptDisplayer) *ConfigInitializer {
	return &ConfigInitializer{promptDisplayer: promptDisplayer}
}

func (configInitializer *ConfigInitializer) InitInteractiveConfig() (*KurtosisConfig, error) {

	userInputResult, err := configInitializer.promptDisplayer.DisplayUserMetricsConsentPromptAndGetUserInputResult()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
	}

	kurtosisConfig := NewKurtosisConfig(userInputResult)
	return kurtosisConfig, nil
}

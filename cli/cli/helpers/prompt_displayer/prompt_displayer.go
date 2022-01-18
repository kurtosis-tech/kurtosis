package prompt_displayer

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/user_input_validations"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

const (
	metricsPromptLabel             = "Do you accept collecting and sending metrics to improve the product?"
	defaultMetricsPromptInputValue = user_input_validations.YesInput

	overrideConfigPromptLabel             = "The Kurtosis Config is already created, Do you want to override it?"
	defaultOverrideConfigPromptInputValue = user_input_validations.NotInput
)

type PromptDisplayer struct {
}

func NewPromptDisplayer() *PromptDisplayer {
	return &PromptDisplayer{}
}

func (promptDisplayer *PromptDisplayer) DisplayOverrideKurtosisConfigConfirmationPromptAndGetUserInputResult() (bool, error) {

	prompt := promptui.Prompt{
		Label:    overrideConfigPromptLabel,
		Default:  string(defaultOverrideConfigPromptInputValue),
		Validate: user_input_validations.ValidateConfirmationInput,
	}

	userOverrideKurtosisConfigInput, err := prompt.Run()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred running Kurtosis config override prompt")
	}
	logrus.Debugf("Kurtosis config confirmation prompt user input: '%v'", userOverrideKurtosisConfigInput)

	userConfirmOverrideKurtosisConfig := user_input_validations.IsConfirmationInput(userOverrideKurtosisConfigInput)

	return userConfirmOverrideKurtosisConfig, nil
}

func (promptDisplayer *PromptDisplayer) DisplayUserMetricsConsentPromptAndGetUserInputResult() (bool, error) {

	prompt := promptui.Prompt{
		Label:    metricsPromptLabel,
		Default:  string(defaultMetricsPromptInputValue),
		Validate: user_input_validations.ValidateMetricsConsentInput,
	}

	userAcceptSendingMetricsInput, err := prompt.Run()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred running metrics consent prompt")
	}
	logrus.Debugf("User metrics consent prompt user input: '%v'", userAcceptSendingMetricsInput)

	userAcceptSendingMetrics := user_input_validations.IsAcceptSendingMetricsInput(userAcceptSendingMetricsInput)

	return userAcceptSendingMetrics, nil
}

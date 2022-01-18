package prompt_displayer

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/user_input_validations"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

const (
	metricsPromptLabel             = "Do you accept collecting and sending metrics to improve the product?"
	defaultMetricsPromptInputValue = "yes"
)

type PromptDisplayer struct {
}

func NewPromptDisplayer() *PromptDisplayer {
	return &PromptDisplayer{}
}

func (promptDisplayer *PromptDisplayer) DisplayUserMetricsConsentPromptAndGetUserInputResult() (bool, error) {

	prompt := promptui.Prompt{
		Label:    metricsPromptLabel,
		Default:  defaultMetricsPromptInputValue,
		Validate: user_input_validations.ValidateMetricsConsentInput,
	}

	userAcceptSendingMetricsInput, err := prompt.Run()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred running metrics consent prompt")
	}
	logrus.Debugf("User choose %q\n", userAcceptSendingMetricsInput)

	userAcceptSendingMetrics := user_input_validations.IsAcceptedSendingMetricsValidInput(userAcceptSendingMetricsInput)

	return userAcceptSendingMetrics, nil
}

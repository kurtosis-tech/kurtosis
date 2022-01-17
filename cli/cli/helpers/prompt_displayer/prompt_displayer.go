package prompt_displayer

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	metricsPromptLabel             = "Do you accept collecting and sending metrics to improve the product?"
	defaultMetricsPromptInputValue = "yes"
)

var userAcceptSendingMetricsValidInputs = []string{"y", "yes"}
var userDoNotAcceptSendingMetricsValidInputs = []string{"n", "no"}
var allAcceptSendingMetricsValidInputs = append(userAcceptSendingMetricsValidInputs, userDoNotAcceptSendingMetricsValidInputs...)

type PromptDisplayer struct {
}

func NewPromptDisplayer() *PromptDisplayer {
	return &PromptDisplayer{}
}

func (promptDisplayer *PromptDisplayer) DisplayUserMetricsConsentPromptAndGetUserInputResult() (bool, error) {

	prompt := promptui.Prompt{
		Label:   metricsPromptLabel,
		Default: defaultMetricsPromptInputValue,
		Validate: validateMetricsConsentPromptInput,
	}

	userAcceptSendingMetricsInput, err := prompt.Run()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred running metrics consent prompt")
	}
	logrus.Debugf("User choose %q\n", userAcceptSendingMetricsInput)

	if contains(userAcceptSendingMetricsValidInputs, userAcceptSendingMetricsInput) {
		return true, nil
	}

	return false, nil
}

func validateMetricsConsentPromptInput(input string) error {
	input = strings.ToLower(input)
	isValid := contains(allAcceptSendingMetricsValidInputs, input)
	if !isValid {
		return stacktrace.NewError("Yo have entered an invalid input '%v'. Valid inputs: '%+v'", input, allAcceptSendingMetricsValidInputs)
	}
	return nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.ToLower(v) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

package prompt_displayer

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

const (
	metricsPromptLabel             = "Do you accept collecting and sending metrics to improve the product? yes/no"
	defaultMetricsPromptInputValue = "yes"
)

type PromptDisplayer struct {
	cliConfig config.Config
}

func NewPromptDisplayer(cliConfig config.Config) *PromptDisplayer {
	return &PromptDisplayer{cliConfig: cliConfig}
}

func (promptDisplayer *PromptDisplayer) DisplayUserMetricsConsentPrompt() (string, error) {

	prompt := promptui.Prompt{
		Label:   metricsPromptLabel,
		Default: defaultMetricsPromptInputValue,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred running metrics consent prompt")
	}
	logrus.Debugf("User choose %q\n", result)

	promptDisplayer.cliConfig.MetricsConsentPromptHasBeenDisplayed()

	return result, nil
}

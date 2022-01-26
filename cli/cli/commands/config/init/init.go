package init

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_optin"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	acceptSendingMetricsArg = "accept-sending-metrics"

	overrideConfigPromptLabel = "The Kurtosis Config is already created, Do you want to override it?"

	//Valid accept sending metrics inputs
	acceptSendingMetricsInput  = "send-metrics"
	rejectSendingMetricsInput  = "dont-send-metrics"
)

var allAcceptSendingMetricsValidInputs = []string{acceptSendingMetricsInput, rejectSendingMetricsInput}

var positionalArgs = []string{
	acceptSendingMetricsArg,
}

var InitCmd = &cobra.Command{
	Use:                   command_str_consts.InitCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Initialize the Kurtosis CLI configuration",
	// TODO Make this dynamic to display exactly what metrics we collect from the users
	Long: "Initializes the configuration file that the CLI uses with the given values.\n" +
		"\n" +
		metrics_optin.WhyKurtosisCollectMetricsDescriptionNote,
	RunE:                  run,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	acceptSendingMetricsStr := parsedPositionalArgs[acceptSendingMetricsArg]

	didUserAcceptSendingMetrics, err := validateMetricsConsentInputAndGetBooleanResult(acceptSendingMetricsStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred validating metrics consent input")
	}

	kurtosisConfigStore := kurtosis_config.GetKurtosisConfigStore()
	doesKurtosisConfigAlreadyExists, err := kurtosisConfigStore.HasConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if Kurtosis config already exists")
	}
	if doesKurtosisConfigAlreadyExists {
		shouldOverrideKurtosisConfig, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(overrideConfigPromptLabel, false)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred displaying confirmation prompt")
		}
		if !shouldOverrideKurtosisConfig {
			logrus.Infof("Skipping overriding Kurtosis config")
			return nil
		}
	}

	kurtosisConfig := kurtosis_config.NewKurtosisConfig(didUserAcceptSendingMetrics)

	if err := kurtosisConfigStore.SetConfig(kurtosisConfig); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
	}

	return nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func validateMetricsConsentInputAndGetBooleanResult(input string) (bool, error) {

	didUserAcceptSendingMetrics := false

	isValid := contains(allAcceptSendingMetricsValidInputs, input)
	if !isValid {
		 return false, stacktrace.NewError(
			"You have entered an invalid argument '%v'. "+
				"'%v' to accept sending metrics or "+
				"'%v' to skip sending metrics",
			input,
			acceptSendingMetricsInput,
			rejectSendingMetricsInput)
	}

	if input == acceptSendingMetricsInput {
		didUserAcceptSendingMetrics = true
	}

	return didUserAcceptSendingMetrics, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.ToLower(v) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

package init

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/annotations"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_tracker"
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

var annotationsMap = map[string]string{
	annotations.SkipConfigInitializationOnGlobalSetupKey: annotations.SkipConfigInitializationOnGlobalSetupValue,
}

var InitCmd = &cobra.Command{
	Use:                   command_str_consts.InitCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Initialize the Kurtosis CLI configuration",
	RunE:                  run,
	Annotations:           annotationsMap,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	acceptSendingMetricsStr := parsedPositionalArgs[acceptSendingMetricsArg]

	userAcceptSendingMetrics, err := validateMetricsConsentInputAndGetBooleanResult(acceptSendingMetricsStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred validating metrics consent input")
	}

	acceptSendingMetricsConfigValueChange := true
	configProvider := kurtosis_config.NewDefaultKurtosisConfigProvider()
	if configProvider.IsConfigAlreadyCreated() {
		userOverrideKurtosisConfig, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(overrideConfigPromptLabel, prompt_displayer.NoInput)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred overwriting Kurtosis config")
		}
		if !userOverrideKurtosisConfig {
			logrus.Infof("Skipping overriding Kurtosis config")
			return nil
		}

		currentConfig, err := configProvider.GetOrInitializeConfig()
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting or initializing config")
		}

		if userAcceptSendingMetrics == currentConfig.IsUserAcceptSendingMetrics() {
			acceptSendingMetricsConfigValueChange = false
		}
	}

	kurtosisConfig := kurtosis_config.NewKurtosisConfig(userAcceptSendingMetrics)

	//We want to track everytime that users change the metrics consent decision
	if acceptSendingMetricsConfigValueChange {
		//Tracking user metrics consent
		metricsClient, err := metrics_tracker.CreateMetricsClient()
		if err != nil {
			//We don't throw and error if this fails, because we don't want to interrupt user's execution
			logrus.Debugf("An error occurred creating SnowPlow metrics client\n%v", err)
		} else {
			metricsTracker := metrics_tracker.NewMetricsTracker(metricsClient)
			if err = metricsTracker.TrackUserAcceptSendingMetrics(kurtosisConfig.IsUserAcceptSendingMetrics()); err != nil {
				//We don't throw and error if this fails, because we don't want to interrupt user's execution
				logrus.Debugf("An error occurred knowing if user accept sending metrics\n%v", err)
			}
			if !kurtosisConfig.IsUserAcceptSendingMetrics() {
				//If user reject sending metrics the feature will be disabled
				metricsTracker.DisableTracking()
			}
		}
	}

	//Saving config
	if err := configProvider.SetConfig(kurtosisConfig); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
	}

	return nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func validateMetricsConsentInputAndGetBooleanResult(input string) (bool, error) {

	userAcceptSendingMetrics := false

	isValid := contains(allAcceptSendingMetricsValidInputs, input)
	if !isValid {
		 return userAcceptSendingMetrics, stacktrace.NewError(
			"Yo have entered an invalid 'accept sending metrics argument'. "+
				"You have to set'%v' if you accept sending metrics or"+
				"you have to set '%v' if you reject sending metrics",
			input,
			acceptSendingMetricsInput,
			rejectSendingMetricsInput)
	}

	if input == acceptSendingMetricsInput {
		userAcceptSendingMetrics = true
	}

	return userAcceptSendingMetrics, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.ToLower(v) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

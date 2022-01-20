package init

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_tracker"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/user_input_validations"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	kurtosisLogLevelArg     = "kurtosis-log-level"
	acceptSendingMetricsArg = "accept-sending-metrics"
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	acceptSendingMetricsArg,
}

var InitCmd = &cobra.Command{
	Use:                   command_str_consts.InitCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Initialize the Kurtosis CLI configuration",
	PersistentPreRunE:     persistentPreRun,
	RunE:                  run,
}

var kurtosisLogLevelStr string

func init() {
	InitCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
}

// It is empty to overwrite the inherited from root cmd
func persistentPreRun(cmd *cobra.Command, args []string) error {
	return nil
}

func run(cmd *cobra.Command, args []string) error {

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	acceptSendingMetricsStr := parsedPositionalArgs[acceptSendingMetricsArg]

	if err := user_input_validations.ValidateMetricsConsentInput(acceptSendingMetricsStr); err != nil {
		return stacktrace.Propagate(err, "An error occurred validating metrics consent input")
	}

	userAcceptSendingMetrics := user_input_validations.IsAcceptSendingMetricsInput(acceptSendingMetricsStr)

	acceptSendingMetricsConfigValueChange := true
	configProvider := kurtosis_config.NewDefaultKurtosisConfigProvider()
	if configProvider.IsConfigAlreadyCreated() {
		promptDisplayer := prompt_displayer.NewPromptDisplayer()
		userOverrideKurtosisConfig, err := promptDisplayer.DisplayOverrideKurtosisConfigConfirmationPromptAndGetUserInputResult()
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

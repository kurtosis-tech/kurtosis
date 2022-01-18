package init

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
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

// Defined and empty persistentPreRun func to overwrite the inherited from root cmd
func persistentPreRun(cmd *cobra.Command, args []string) error {
	return nil
}

func run(cmd *cobra.Command, args []string) error {

	//TODO checks if config already exist an throw a prompt to override configuration

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	acceptSendingMetricsStr := parsedPositionalArgs[acceptSendingMetricsArg]

	if err := user_input_validations.ValidateMetricsConsentInput(acceptSendingMetricsStr); err != nil {
		return stacktrace.Propagate(err, "An error occurred validating metrics consent input")
	}

	userAcceptSendingMetrics := user_input_validations.IsAcceptedSendingMetricsValidInput(acceptSendingMetricsStr)

	kurtosisConfig := kurtosis_config.NewKurtosisConfig(userAcceptSendingMetrics)

	configProvider := kurtosis_config.NewDefaultKurtosisConfigProvider()
	if err := configProvider.SetConfig(kurtosisConfig); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
	}
	return nil
}

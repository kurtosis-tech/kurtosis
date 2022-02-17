package init

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/set_selection_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	acceptSendingMetricsArgKey = "accept-sending-metrics"

	overrideConfigPromptLabel = "The Kurtosis Config is already created; do you want to override it?"

	//Valid accept sending metrics inputs
	acceptSendingMetricsInput  = "send-metrics"
	rejectSendingMetricsInput  = "dont-send-metrics"
)

var validAcceptSendingMetricsArgValues = map[string]bool{
	acceptSendingMetricsInput: true,
	rejectSendingMetricsInput: true,
}

var InitCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.InitCmdStr,
	ShortDescription:         "Initialize the Kurtosis CLI configuration",
	LongDescription:          "Initializes the configuration file that the CLI uses with the given values",
	Args:                     []*args.ArgConfig{
		set_selection_arg.NewSetSelectionArg(
			acceptSendingMetricsArgKey,
			validAcceptSendingMetricsArgValues,
		),
	},
	RunFunc:                  run,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	didUserAcceptSendingMetricsStr, err := args.GetNonGreedyArg(acceptSendingMetricsArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but none was found; this is a bug in Kurtosis!", acceptSendingMetricsArgKey)
	}

	// We get validation for free by virtue of the KurtosisCommand framework
	var didUserAcceptSendingMetrics bool
	if didUserAcceptSendingMetricsStr == acceptSendingMetricsInput {
		didUserAcceptSendingMetrics = true
	} else if didUserAcceptSendingMetricsStr == rejectSendingMetricsInput {
		didUserAcceptSendingMetrics = false
	} else {
		// If this happens, there's something wrong with the validation being done via KurtosisCommand
		return stacktrace.NewError(
			"Encountered an unrecognized 'should send metrics?' input string '%v', which should never happen; this is a bug in Kurtosis!",
			didUserAcceptSendingMetricsStr,
		)
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

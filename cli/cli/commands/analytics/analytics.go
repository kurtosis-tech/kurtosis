package analytics

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/set_selection_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	//Valid accept sending metrics inputs
	enableSendingMetrics   = "enable"
	disableSendingMetrics  = "disable"
	enableDisableDelimiter = "|"

	enableDisableStatus = enableSendingMetrics + enableDisableDelimiter + disableSendingMetrics
)

var validMetricsSendingToggleValue = map[string]bool{
	enableSendingMetrics:  true,
	disableSendingMetrics: true,
}

var AnalyticsCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.Analytics,
	ShortDescription: "Control Kurtosis's anonymous aggregate user behavior analytics",
	LongDescription:  "Control Kurtosis's anonymous aggregate user behavior analytics. Read more at\n" + user_support_constants.MetricsPhilosophyDocs,
	Args: []*args.ArgConfig{
		// the enableDisableStatus would appear as the name of the argument
		set_selection_arg.NewSetSelectionArg(
			enableDisableStatus,
			validMetricsSendingToggleValue,
		),
	},
	Flags:                    nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	didUserAcceptSendingMetricsStr, err := args.GetNonGreedyArg(enableDisableStatus)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but none was found; this is a bug in Kurtosis!", enableDisableStatus)
	}

	// We get validation for free by virtue of the KurtosisCommand framework
	var didUserAcceptSendingMetrics bool
	if didUserAcceptSendingMetricsStr == enableSendingMetrics {
		didUserAcceptSendingMetrics = true
	} else if didUserAcceptSendingMetricsStr == disableSendingMetrics {
		didUserAcceptSendingMetrics = false
	} else {
		// If this happens, there's something wrong with the validation being done via KurtosisCommand
		return stacktrace.NewError(
			"Encountered an unrecognized '%v' input string '%v', which should never happen; this is a bug in Kurtosis!",
			enableDisableStatus,
			didUserAcceptSendingMetricsStr,
		)
	}

	kurtosisConfigStore := kurtosis_config.GetKurtosisConfigStore()
	var kurtosisConfig *resolved_config.KurtosisConfig

	hasConfig, err := kurtosisConfigStore.HasConfig()
	if err != nil {
		return stacktrace.NewError("An error occurred while determining whether configuration already exists")
	}

	if hasConfig {
		kurtosisConfig, err = kurtosisConfigStore.GetConfig()
		if err != nil {
			return stacktrace.NewError("An error occurred while fetching stored configuration")
		}
		kurtosisConfig = resolved_config.NewKurtosisConfigWithMetricsSetFromExistingConfig(kurtosisConfig, didUserAcceptSendingMetrics)
	} else {
		kurtosisConfig, err = resolved_config.NewKurtosisConfigFromRequiredFields(didUserAcceptSendingMetrics)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to initialize Kurtosis configuration from user input %t", didUserAcceptSendingMetrics)
		}
	}

	if err := kurtosisConfigStore.SetConfig(kurtosisConfig); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting analytics configuration")
	}

	logrus.Infof("Analytics tracking is now %v", didUserAcceptSendingMetricsStr)

	return nil
}

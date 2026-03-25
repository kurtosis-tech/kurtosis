package start

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/grafloki"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultEngineVersion                   = ""
	restartEngineOnSameVersionIfAnyRunning = true
	dontRestartAPIContainers               = false
)

var LokiStartCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.LokiStartCmdStr,
	ShortDescription:         "Starts a Loki instance.",
	LongDescription:          "Starts a Loki instance and configures Kurtosis engine to send logs to it.",
	RunFunc:                  run,
	Flags:                    nil,
	Args:                     nil,
	PostValidationAndRunFunc: nil,
	PreValidationAndRunFunc:  nil,
}

func run(
	ctx context.Context,
	_ *flags.ParsedFlags,
	_ *args.ParsedArgs,
) error {
	clusterConfig, err := kurtosis_config_getter.GetKurtosisClusterConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Kurtosis cluster config.")
	}

	lokiSink, err := grafloki.StartLoki(ctx, clusterConfig.GetClusterType(), clusterConfig.GetGraflokiConfig())
	if err != nil {
		return err
	}

	logrus.Infof("Configuring engine to send logs to Loki...")
	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	engineManager.SetSkipConfiguredGrafloki(true)
	_, engineClientCloseFunc, err := engineManager.RestartEngineIdempotently(
		ctx,
		defaults.DefaultEngineLogLevel,
		defaultEngineVersion,
		restartEngineOnSameVersionIfAnyRunning,
		defaults.DefaultEngineEnclavePoolSize,
		defaults.DefaultEnableDebugMode,
		defaults.DefaultGitHubAuthTokenOverride,
		dontRestartAPIContainers,
		defaults.DefaultDomain,
		defaults.DefaultLogRetentionPeriod,
		lokiSink,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred restarting engine to be configured to send logs to Loki.")
	}
	defer func() {
		if closeErr := engineClientCloseFunc(); closeErr != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", closeErr)
		}
	}()

	return nil
}

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
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultEngineVersion                   = ""
	restartEngineOnSameVersionIfAnyRunning = true
)

var GraflokiStartCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.GraflokiStartCmdStr,
	ShortDescription:         "Starts a Grafana/Loki instance.",
	LongDescription:          "Starts a Grafana/Loki instance and configures Kurtosis engine to send logs to it.",
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
	clusterConfig.GetClusterType()

	// NOTE(tedi  04/03/25): If you're wondering why the grafana / loki instance is being started by the CLI (and not in container-engine-lib via KurtosisBackend as with LogsCollector and LogsAggregator), here's why:
	// 1. now that Kurtosis is purely OSS, it's important to reduce maintenance surface / complexity inside Kurtosis core (Engine, APIContainer, KurtosisBackend, Starlark Engine)
	// 	  wherever possible infra not essential to Kurtosis core should be built outside of it or at the edges (e.g. client)
	// 2. the export logs feature was built in service of leveraging existing logging solutions/not rebuilding logging in Kurtosis
	// 3. having grafloki started/managed by the CLI lets us build on the export logs feature
	// In other words
	// putting it in the CLI is saying - “You could set up Grafana and Loki yourself, and then restart the engine to point to it, Kurtosis CLI will do that for you to save you a step”
	// putting it in Kurtosis core is saying - “Grafana and Loki are core a necessary part of the Kurtosis platform and supports the Kurtosis abstraction/value prop" - which is not the case
	// https://drawpaintacademy.com/the-bull/
	lokiSink, grafanaUrl, err := grafloki.StartGrafloki(ctx, clusterConfig.GetClusterType(), clusterConfig.GetGraflokiConfig())
	if err != nil {
		return err // already wrapped
	}
	logrus.Infof("Grafana running at %v", grafanaUrl)

	logrus.Infof("Configuring engine to send logs to Loki...")
	err = restartEngineWithLogsSink(ctx, lokiSink)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred restarting engine to be configured to send logs to Loki.")
	}

	return nil
}

func restartEngineWithLogsSink(ctx context.Context, sink logs_aggregator.Sinks) error {
	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	dontRestartAPIContainers := false
	_, engineClientCloseFunc, restartEngineErr := engineManager.RestartEngineIdempotently(ctx,
		defaults.DefaultEngineLogLevel,
		defaultEngineVersion,
		restartEngineOnSameVersionIfAnyRunning,
		defaults.DefaultEngineEnclavePoolSize,
		defaults.DefaultEnableDebugMode,
		defaults.DefaultGitHubAuthTokenOverride,
		dontRestartAPIContainers,
		defaults.DefaultDomain,
		defaults.DefaultLogRetentionPeriod,
		sink)
	if restartEngineErr != nil {
		return stacktrace.Propagate(restartEngineErr, "An error occurred restarting the Kurtosis engine")
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()
	return nil
}

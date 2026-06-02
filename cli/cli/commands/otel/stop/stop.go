package stop

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/otel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultEngineVersion                   = ""
	restartEngineOnSameVersionIfAnyRunning = true
	dontRestartAPIContainers               = false
)

var OtelStopCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.OtelStopCmdStr,
	ShortDescription:         "Stops Docker-only OpenTelemetry side containers.",
	LongDescription:          "Restarts a running Kurtosis engine without the OpenTelemetry Loki sink, then stops Docker-only OpenTelemetry collector and ClickHouse side containers if they exist.",
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

	if clusterConfig.GetClusterType() != resolved_config.KurtosisClusterType_Docker {
		return otel.StopOtel(ctx, clusterConfig.GetClusterType())
	}

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}

	engineStatus, _, _, err := engineManager.GetEngineStatus(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Kurtosis engine status.")
	}

	if engineStatus != engine_manager.EngineStatus_Running {
		logrus.Infof("Engine status is '%v'; skipping engine restart.", engineStatus)
	} else {
		logrus.Infof("Configuring engine to stop sending logs to otel collector...")
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
			defaults.DefaultLogsSinks,
		)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred restarting engine to remove otel collector log sink.")
		}
		defer func() {
			if closeErr := engineClientCloseFunc(); closeErr != nil {
				logrus.Warnf("Error closing the engine client:\n'%v'", closeErr)
			}
		}()
	}

	if err = otel.StopOtel(ctx, clusterConfig.GetClusterType()); err != nil {
		return err
	}

	return nil
}

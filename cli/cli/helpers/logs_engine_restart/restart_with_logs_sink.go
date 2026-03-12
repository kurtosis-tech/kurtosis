package logs_engine_restart

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultEngineVersion                   = ""
	restartEngineOnSameVersionIfAnyRunning = true
)

func RestartEngineWithLogsSink(ctx context.Context, sink logs_aggregator.Sinks, skipConfiguredGrafloki bool) error {
	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	dontRestartAPIContainers := false

	var engineClientCloseFunc func() error
	if skipConfiguredGrafloki {
		_, engineClientCloseFunc, err = engineManager.RestartEngineIdempotentlyWithoutConfiguredGrafloki(
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
			sink,
		)
	} else {
		_, engineClientCloseFunc, err = engineManager.RestartEngineIdempotently(
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
			sink,
		)
	}
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred restarting the Kurtosis engine")
	}

	defer func() {
		if closeErr := engineClientCloseFunc(); closeErr != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", closeErr)
		}
	}()

	return nil
}

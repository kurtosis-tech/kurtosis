package start

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
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultEngineVersion                   = ""
	restartEngineOnSameVersionIfAnyRunning = true
	dontRestartAPIContainers               = false
)

var OtelStartCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.OtelStartCmdStr,
	ShortDescription:         "Starts Docker-only OpenTelemetry side containers.",
	LongDescription:          "Starts Docker-only OpenTelemetry collector and ClickHouse side containers and configures Kurtosis engine to send logs to the collector.",
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

	endpoints, err := otel.StartOtel(ctx, clusterConfig.GetClusterType())
	if err != nil {
		return err
	}

	logrus.Infof("otel ClickHouse running at %v (native: %v)", endpoints.ClickHouseHTTPURL, endpoints.ClickHouseNativeAddress)
	logrus.Infof("otel collector running at %v (http: %v)", endpoints.CollectorOTLPGRPCURL, endpoints.CollectorOTLPHTTPURL)
	logrus.Infof("Configuring engine to send logs to otel collector...")
	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	engineManager.SetSkipConfiguredGrafloki(true)
	// `otel start` already started the side containers above; suppress the config-driven auto-start
	// path so RestartEngineIdempotently below doesn't try to start them a second time.
	engineManager.SetSkipConfiguredOtel(true)
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
		otel.NewLokiSink(endpoints.CollectorLokiURL),
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred restarting engine to be configured to send logs to otel collector.")
	}
	defer func() {
		if closeErr := engineClientCloseFunc(); closeErr != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", closeErr)
		}
	}()
	return nil
}

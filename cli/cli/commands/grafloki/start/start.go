package start

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/grafloki"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
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

	// NOTE(tedi  04/03/25): If you're wondering why the grafana / loki instance is being started by the CLI (and not in container-engine-lib via KurtosisBackend as with LogsCollector and LogsAggregator), here's why:
	// 1. now that Kurtosis is purely OSS, it's important to reduce maintenance surface / complexity inside Kurtosis core. wherever possible infra not essential to Kurtosis core should be built outside of it or at the edges (e.g. client)
	// 2. the export logs feature was built in service of leveraging existing logging solutions/not rebuilding logging in Kurtosis
	// 3. having grafloki started/managed by the CLI lets us build on the export logs feature
	var lokiHost string
	var grafanaUrl string
	switch clusterConfig.GetClusterType() {
	case resolved_config.KurtosisClusterType_Docker:
		lokiHost, grafanaUrl, err = grafloki.StartGrafLokiInDocker(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Docker.")
		}
	case resolved_config.KurtosisClusterType_Kubernetes:
		lokiHost, grafanaUrl, err = grafloki.StartGrafLokiInKubernetes(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Kubernetes.")
		}
	default:
		return stacktrace.NewError("Unsupported cluster type: %v", clusterConfig.GetClusterType().String())
	}

	lokiSink := map[string]map[string]interface{}{
		"loki": {
			"type":     "loki",
			"endpoint": lokiHost,
			"encoding": map[string]string{
				"codec": "json",
			},
			"labels": map[string]string{
				"job": "kurtosis",
			},
		},
	}

	logrus.Infof("Configuring engine to send logs to Loki...")
	err = RestartEngineWithLogsSink(ctx, lokiSink)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred restarting engine to be configured to send logs to Loki.")
	}

	out.PrintOutLn(fmt.Sprintf("Grafana running at %v", grafanaUrl))
	return nil
}

func RestartEngineWithLogsSink(ctx context.Context, sink logs_aggregator.Sinks) error {
	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	var engineClientCloseFunc func() error
	var restartEngineErr error
	dontRestartAPIContainers := false
	_, engineClientCloseFunc, restartEngineErr = engineManager.RestartEngineIdempotently(ctx,
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

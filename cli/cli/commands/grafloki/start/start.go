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
	CommandStr:       command_str_consts.GraflokiStartCmdStr,
	ShortDescription: "Starts a grafana/loki instance.",
	LongDescription:  "Starts a grafana/loki instance that the kurtosis engine will be configured to send logs to.",
	RunFunc:          run,
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

	var lokiHost string
	if clusterConfig.GetClusterType() == resolved_config.KurtosisClusterType_Docker {
		lokiHost, err = grafloki.StartGrafLokiInDocker(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Docker.")
		}
	} else if clusterConfig.GetClusterType() == resolved_config.KurtosisClusterType_Kubernetes {
		lokiHost, err = grafloki.StartGrafLokiInKubernetes(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Kubernetes.")
		}
	} else {
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

	out.PrintOutLn(fmt.Sprintf("Grafana running at http://localhost:%v", ""))
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
	_, engineClientCloseFunc, restartEngineErr = engineManager.RestartEngineIdempotently(ctx, defaults.DefaultEngineLogLevel, defaultEngineVersion, restartEngineOnSameVersionIfAnyRunning, defaults.DefaultEngineEnclavePoolSize, defaults.DefaultEnableDebugMode, defaults.DefaultGitHubAuthTokenOverride, dontRestartAPIContainers, defaults.DefaultDomain, defaults.DefaultLogRetentionPeriod, sink)
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

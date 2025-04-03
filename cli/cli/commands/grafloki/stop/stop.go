package stop

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/grafloki"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
)

var GraflokiStopCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.GraflokiStopCmdStr,
	ShortDescription: "Stops a grafana/loki instance.",
	LongDescription:  "Stop a grafana/loki instance if one already exists.",
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

	if clusterConfig.GetClusterType() == resolved_config.KurtosisClusterType_Docker {
		err := grafloki.StopGrafLokiInDocker(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping Grafana and Loki containers in Docker.")
		}
	} else if clusterConfig.GetClusterType() == resolved_config.KurtosisClusterType_Kubernetes {
		err := grafloki.StopGrafLokiInKubernetes(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping Grafana and Loki containers in Kubernetes.")
		}
	} else {
		return stacktrace.NewError("Unsupported cluster type: %v", clusterConfig.GetClusterType().String())
	}

	out.PrintOutLn("Successfully stopped Grafana and Loki containers.")
	return nil
}

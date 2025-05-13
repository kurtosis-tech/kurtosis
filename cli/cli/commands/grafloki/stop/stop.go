package stop

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/grafloki"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/stacktrace"
)

var GraflokiStopCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.GraflokiStopCmdStr,
	ShortDescription:         "Stops a grafana/loki instance.",
	LongDescription:          "Stop a grafana/loki instance if one already exists.",
	RunFunc:                  run,
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	PostValidationAndRunFunc: nil,
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

	if err = grafloki.StopGrafloki(ctx, clusterConfig.GetClusterType()); err != nil {
		return err // already wrapped
	}

	return nil
}

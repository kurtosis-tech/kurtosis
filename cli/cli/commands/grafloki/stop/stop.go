package stop

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
)

var GraflokiStopCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.GraflokiStopCmdStr,
	ShortDescription: "Stops a grafana/loki instance.",
	LongDescription:  "Stop a grafana/loki instance if one already exists.",
	RunFunc:          run,
}

func run(
	ctx context.Context,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	return nil
}

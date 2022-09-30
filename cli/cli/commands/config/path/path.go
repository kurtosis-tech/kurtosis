package path

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

var PathCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.PathCmdStr,
	ShortDescription:         "Shows Kurtosis config file path",
	LongDescription:          "Displays the path to the Kurtosis config file",
	RunFunc:                  run,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	configFilepath, err := host_machine_directories.GetKurtosisConfigYAMLFilepath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis config filepath")
	}
	fmt.Fprintln(logrus.StandardLogger().Out, configFilepath)
	return nil
}
package stop

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

var PortalStopCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.PortalStopCmdStr,
	ShortDescription:         "Stops Kurtosis Portal",
	LongDescription:          "Stops Kurtosis Portal if it is currently running",
	Flags:                    []*flags.FlagConfig{},
	Args:                     []*args.ArgConfig{},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	logrus.Infof("Stopping Kurtosis Portal")
	portalManager := portal_manager.NewPortalManager()
	if err := portalManager.StopExisting(ctx); err != nil {
		return stacktrace.Propagate(err, "Error stopping portal")
	}
	logrus.Infof("Kurtosis portal stopped")
	return nil
}

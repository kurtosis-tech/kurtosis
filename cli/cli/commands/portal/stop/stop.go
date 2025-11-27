package stop

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

var PortalStopCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.PortalStopCmdStr,
	ShortDescription:         "Stops Kurtosis Portal",
	LongDescription:          "Stops Kurtosis Portal daemon if one is currently running locally",
	Flags:                    []*flags.FlagConfig{},
	Args:                     []*args.ArgConfig{},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	logrus.Infof("Stopping the Kurtosis Portal")
	portalManager := portal_manager.NewPortalManager()
	if err := portalManager.StopExisting(ctx); err != nil {
		return stacktrace.Propagate(err, "Error stopping Kurtosis Portal")
	}
	logrus.Infof("Kurtosis Portal stopped")
	return nil
}

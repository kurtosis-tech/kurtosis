package start

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

var PortalStartCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.PortalStartCmdStr,
	ShortDescription:         "Starts Kurtosis Portal",
	LongDescription:          "Starts Kurtosis Portal in the background. The Portal can then be stopped or restarted using the corresponding commands",
	Flags:                    []*flags.FlagConfig{},
	Args:                     []*args.ArgConfig{},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	portalManager := portal_manager.NewPortalManager()
	currentPortalPid, _, isPortalReachable, err := portalManager.CurrentStatus(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to determine current state of Kurtosis Portal process")
	}

	// If there is a healthy running Portal, we do nothing since we don't want to break the current port forward.
	// We could save the port forward state and restart the Portal but it is not yet implemented.
	if isPortalReachable {
		logrus.Infof("Portal is currently running on PID '%d' and healthy.", currentPortalPid)
		return nil
	}

	logrus.Infof("Starting Kurtosis Portal")
	err = portalManager.StartRequiredVersion(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Error starting portal")
	}
	return nil
}

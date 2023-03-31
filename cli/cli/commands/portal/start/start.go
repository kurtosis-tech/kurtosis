package start

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

func run(ctx context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	logrus.Infof("Starting Kurtosis Portal")
	portalManager := portal_manager.NewPortalManager()

	// Checking if new version is available and potentially downloading it
	if _, err := portal_manager.DownloadLatestKurtosisPortalBinary(ctx); err != nil {
		return stacktrace.Propagate(err, "An unexpected error occurred trying to download the latest version of Kurtosis Portal")
	}

	currentPortalPid, process, isPortalReachable, err := portalManager.CurrentStatus(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to determine current state of Kurtosis Portal process")
	}
	if isPortalReachable {
		logrus.Infof("Portal is currently running on PID '%d' and healthy.", currentPortalPid)
		return nil
	}
	if process != nil {
		logrus.Warnf("A non-healthy Portal process is currently running on PID '%d'. Stop it first before starting a new one", currentPortalPid)
		return nil
	}

	startedPortalPid, err := portalManager.StartNew(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Error starting portal")
	}
	logrus.Infof("Kurtosis Portal started successfully on PID %d", startedPortalPid)
	return nil
}

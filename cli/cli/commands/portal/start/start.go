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

func run(ctx context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	logrus.Infof("Starting Kurtosis Portal")
	portalManager := portal_manager.NewPortalManager()
	err := portalManager.DownloadAndStart(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Error starting portal")
	}
	return nil
}

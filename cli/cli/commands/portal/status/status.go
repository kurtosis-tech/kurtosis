package status

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
)

var PortalStatusCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.PortalStatusCmdStr,
	ShortDescription:         "Displays the status of the Kurtosis Portal",
	LongDescription:          "Determines and displays the status of the Kurtosis Portal process running as a daemon",
	Flags:                    []*flags.FlagConfig{},
	Args:                     []*args.ArgConfig{},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	portalManager := portal_manager.NewPortalManager()

	pid, process, isPortalReachable, err := portalManager.CurrentStatus(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to determine the status of Kurtosis Portal")
	}

	if isPortalReachable {
		out.PrintOutLn(fmt.Sprintf("Kurtosis Portal is running and healthy on PID %d", pid))
		return nil
	} else if process != nil {
		out.PrintOutLn(fmt.Sprintf("Kurtosis Portal process is on PID %d but it is unreachable", pid))
		return nil
	} else if pid != 0 {
		out.PrintOutLn("Kurtosis Portal PID file exits but process is dead")
		return nil
	}
	out.PrintOutLn("Kurtosis Portal is not running")
	return nil
}

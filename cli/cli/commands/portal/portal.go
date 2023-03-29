package portal

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/portal/start"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/portal/status"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/portal/stop"
	"github.com/spf13/cobra"
)

// ContextCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var PortalCmd = &cobra.Command{
	Use:   command_str_consts.PortalCmdStr,
	Short: "Manage lifecycle of Kurtosis Portal",
	RunE:  nil,
}

func init() {
	PortalCmd.AddCommand(start.PortalStartCmd.MustGetCobraCommand())
	PortalCmd.AddCommand(stop.PortalStopCmd.MustGetCobraCommand())
	PortalCmd.AddCommand(status.PortalStatusCmd.MustGetCobraCommand())
}

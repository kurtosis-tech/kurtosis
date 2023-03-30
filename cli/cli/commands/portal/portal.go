package portal

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/portal/start"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/portal/status"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/portal/stop"
	"github.com/spf13/cobra"
)

// PortalCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var PortalCmd = &cobra.Command{
	Use:   command_str_consts.PortalCmdStr,
	Short: "Manages lifecycle of Kurtosis Portal",
	Long: "Manages the lifecycle of Kurtosis Portal. Kurtosis Portal is a lightweight daemon running locally to " +
		"enable seamless communication with Kurtosis enclaves that are running on a remote Kurtosis server. The " +
		"Portal is tightly coupled to the concept of remote Kurtosis contexts. On a local context, the portal is not " +
		"needed.",
	RunE: nil,
}

func init() {
	PortalCmd.AddCommand(start.PortalStartCmd.MustGetCobraCommand())
	PortalCmd.AddCommand(stop.PortalStopCmd.MustGetCobraCommand())
	PortalCmd.AddCommand(status.PortalStatusCmd.MustGetCobraCommand())
}

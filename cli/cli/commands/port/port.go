package port

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/port/print"
	"github.com/spf13/cobra"
)

// PortCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var PortCmd = &cobra.Command{
	Use:   command_str_consts.PortCmdStr,
	Short: "Manage ports",
	RunE:  nil,
}

func init() {
	PortCmd.AddCommand(print.PortPrintCmd.MustGetCobraCommand())
}

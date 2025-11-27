package cloud

import (
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/cloud/load"
	"github.com/spf13/cobra"
)

// CloudCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var CloudCmd = &cobra.Command{
	Use:   command_str_consts.CloudCmdStr,
	Short: "Manage Kurtosis cloud instances",
	RunE:  nil,
}

func init() {
	CloudCmd.AddCommand(load.LoadCmd.MustGetCobraCommand())
}

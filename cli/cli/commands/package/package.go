package _package

import (
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/package/init_cmd"
	"github.com/spf13/cobra"
)

// PackageCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var PackageCmd = &cobra.Command{
	Use:   command_str_consts.PackageCmdStr,
	Short: "Manage packages",
	RunE:  nil,
}

func init() {
	PackageCmd.AddCommand(init_cmd.InitCmd.MustGetCobraCommand())
}

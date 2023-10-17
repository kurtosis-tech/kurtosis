package kurtosis_context

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/ls"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/rm"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context/set"
	"github.com/spf13/cobra"
)

// ContextCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var ContextCmd = &cobra.Command{
	Use:   command_str_consts.ContextCmdStr,
	Short: "Manage Kurtosis contexts",
	RunE:  nil,
}

func init() {
	ContextCmd.AddCommand(add.ContextAddCmd.MustGetCobraCommand())
	ContextCmd.AddCommand(ls.ContextLsCmd.MustGetCobraCommand())
	ContextCmd.AddCommand(rm.ContextRmCmd.MustGetCobraCommand())
	ContextCmd.AddCommand(set.ContextSetCmd.MustGetCobraCommand())
}

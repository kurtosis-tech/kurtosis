package startosis

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/startosis/exec"
	"github.com/spf13/cobra"
)

var StartosisCmd = &cobra.Command{
	Use:   command_str_consts.StartosisCmdStr,
	Short: "Interact with Startosis scripts",
	RunE:  nil,
}

func init() {
	StartosisCmd.AddCommand(exec.StartosisExecCmd.MustGetCobraCommand())
}

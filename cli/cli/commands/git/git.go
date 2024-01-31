package git

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/git/login"
	"github.com/spf13/cobra"
)

// EngineCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "Manage git login",
	RunE:  nil,
}

func init() {
	GitCmd.AddCommand(login.LoginCmd)
}

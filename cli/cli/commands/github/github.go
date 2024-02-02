package github

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/github/login"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/github/logout"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/github/status"
	"github.com/spf13/cobra"
)

// GitHubCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var GitHubCmd = &cobra.Command{
	Use:   command_str_consts.GitHubCmdStr,
	Short: "Manage GitHub login",
	RunE:  nil,
}

func init() {
	GitHubCmd.AddCommand(login.LoginCmd.MustGetCobraCommand())
	GitHubCmd.AddCommand(logout.LogoutCmd.MustGetCobraCommand())
	GitHubCmd.AddCommand(status.StatusCmd.MustGetCobraCommand())
	GitHubCmd.AddCommand(status.StatusCmd.MustGetCobraCommand())
}

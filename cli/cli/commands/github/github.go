package github

import (
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/github/login"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/github/logout"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/github/status"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/github/token"
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
	GitHubCmd.AddCommand(token.TokenCmd.MustGetCobraCommand())
}

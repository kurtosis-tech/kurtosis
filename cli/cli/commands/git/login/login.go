package login

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/github_auth_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authorizes Kurtosis CLI on behalf of a Github user.",
	Long:  "Authorizes Kurtosis CLI to perform git operations on behalf of the user such as retrieving packages in private Github repositories.",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	gitAuthCfg, err := github_auth_config.GetGithubAuthConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving github auth configuration.")
	}
	if gitAuthCfg.GetCurrentUser() != "" {
		out.PrintOutLn(fmt.Sprintf("Logged in as github user: '%v'", gitAuthCfg.GetCurrentUser()))
		return nil
	}
	username, err := gitAuthCfg.Login()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred logging user in.")
	}
	out.PrintOutLn(fmt.Sprintf("Successfully logged in github user: '%v'", username))
	return nil
}

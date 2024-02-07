package logout

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/github_auth_store"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
)

var LogoutCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.GitHubLogoutCmdStr,
	ShortDescription:         "Logs out a GitHub user from Kurtosis CLI",
	LongDescription:          "Logs out a GitHub user from Kurtosis CLI by removing their GitHub user info and auth token from Kurtosis CLI config",
	Args:                     nil,
	Flags:                    nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	githubAuthStore, err := github_auth_store.GetGitHubAuthStore()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving GitHub auth configuration.")
	}
	username := githubAuthStore.GetUser()
	if username == "" {
		out.PrintOutLn("No GitHub user logged into Kurtosis CLI: %v")
		return nil
	}
	err = githubAuthStore.RemoveUser()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred logging out GitHub user: %v", username)
	}
	out.PrintOutLn(fmt.Sprintf("Successfully logged GitHub user '%v' out of Kurtosis CLI", username))
	return nil
}

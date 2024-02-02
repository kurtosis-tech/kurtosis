package logout

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/github_auth_config"
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
	githubAuthCfg, err := github_auth_config.GetGithubAuthConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving GitHub auth configuration.")
	}
	if !githubAuthCfg.IsLoggedIn() {
		out.PrintOutLn("No GitHub user currently logged in.")
		return nil
	}
	err = githubAuthCfg.Logout()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred logging out GitHub user: %v", githubAuthCfg.GetCurrentUser())
	}
	out.PrintOutLn("Successfully logged GitHub user out of Kurtosis CLI")
	return nil
}

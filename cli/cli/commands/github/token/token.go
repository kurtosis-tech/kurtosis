package token

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
	CommandStr:               command_str_consts.GitHubTokenCmdStr,
	ShortDescription:         "Displays GitHub auth token used if a user is logged in",
	LongDescription:          "Displays GitHub auth token used if a user is logged in",
	Args:                     nil,
	Flags:                    nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	githubAuthCfg, err := github_auth_config.GetGitHubAuthConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving GitHub auth configuration.")
	}
	if !githubAuthCfg.IsLoggedIn() {
		out.PrintOutLn("No GitHub user currently logged in.")
		return nil
	}
	githubAuthToken, err := githubAuthCfg.GetAuthToken()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving GitHub auth token for currently logged in user: %v.", githubAuthCfg.GetCurrentUser())
	}
	out.PrintOutLn(githubAuthToken)
	return nil
}

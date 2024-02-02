package login

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/github_auth_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
)

var LoginCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.GitHubLoginCmdStr,
	ShortDescription:         "Authorizes Kurtosis CLI on behalf of a Github user.",
	LongDescription:          "Authorizes Kurtosis CLI to perform git operations on behalf of a GitHub user such as retrieving packages in private repositories.",
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
	if githubAuthCfg.IsLoggedIn() {
		out.PrintOutLn(fmt.Sprintf("Logged in as GitHub user: '%v'", githubAuthCfg.GetCurrentUser()))
		return nil
	}
	err = githubAuthCfg.Login()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred logging user in.")
	}
	out.PrintOutLn(fmt.Sprintf("Successfully logged in GitHub user: '%v'", githubAuthCfg.GetCurrentUser()))
	return nil
}

package token

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/github_auth_store"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
)

var TokenCmd = &lowlevel.LowlevelKurtosisCommand{
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
	githubAuthStore, err := github_auth_store.GetGitHubAuthStore()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving GitHub auth store.")
	}
	username, err := githubAuthStore.GetUser()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user to see if user already exists.")
	}
	if username == "" {
		out.PrintOutLn("No GitHub user currently logged in.")
		return nil
	}
	authToken, err := githubAuthStore.GetAuthToken()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving GitHub auth token for user: %v.", username)
	}
	out.PrintOutLn(authToken)
	return nil
}

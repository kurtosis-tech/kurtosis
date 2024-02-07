package status

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

var StatusCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.GitHubStatusCmdStr,
	ShortDescription:         "Displays GitHub auth info",
	LongDescription:          "Displays GitHub auth info by showing a logged in user's info or whether no GitHub user is logged into Kurtosis CLI",
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
	username, err := githubAuthStore.GetUser()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user to see if user already exists.")
	}
	if username == "" {
		out.PrintOutLn("No GitHub user logged into Kurtosis CLI")
		return nil
	}
	out.PrintOutLn(fmt.Sprintf("Logged in as GitHub user: %v", username))
	return nil
}

package login

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/github_auth_store"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/oauth"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultEngineVersion                   = ""
	restartEngineOnSameVersionIfAnyRunning = true
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

func run(ctx context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	githubAuthStore, err := github_auth_store.GetGitHubAuthStore()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving GitHub auth store.")
	}
	username, err := githubAuthStore.GetUser()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user to see if user already exists.")
	}
	if username != "" {
		out.PrintOutLn(fmt.Sprintf("Logged in as GitHub user: %v", username))
		return nil
	}
	authToken, username, err := oauth.AuthFlow()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred in the Github OAuth flow.")
	}
	err = githubAuthStore.SetUser(username, authToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting GitHub user: %v", username)
	}
	out.PrintOutLn(fmt.Sprintf("Successfully logged in GitHub user: %v", username))
	logrus.Info("Restarting engine for GitHub auth to take effect...")
	err = RestartEngineAfterGitHubAuth(ctx)
	if err != nil {
		return err
	}
	logrus.Infof("Engine restarted successfully")
	return nil
}

func RestartEngineAfterGitHubAuth(ctx context.Context) error {
	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	var engineClientCloseFunc func() error
	var restartEngineErr error
	dontRestartAPIContainers := false
	_, engineClientCloseFunc, restartEngineErr = engineManager.RestartEngineIdempotently(ctx, defaults.DefaultEngineLogLevel, defaultEngineVersion, restartEngineOnSameVersionIfAnyRunning, defaults.DefaultEngineEnclavePoolSize, defaults.DefaultEnableDebugMode, defaults.DefaultGitHubAuthTokenOverride, dontRestartAPIContainers, defaults.DefaultDomain, defaults.DefaultLogRetentionPeriod)
	if restartEngineErr != nil {
		return stacktrace.Propagate(restartEngineErr, "An error occurred restarting the Kurtosis engine")
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()
	return nil
}

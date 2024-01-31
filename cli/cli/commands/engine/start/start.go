package start

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/engine/common"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"os"
	"strings"
)

const (
	engineVersionArg    = "version"
	logLevelArg         = "log-level"
	enclavePoolSizeFlag = "enclave-pool-size"
	gitAuthTokenArg     = "git-auth-token"

	defaultEngineVersion          = ""
	kurtosisTechEngineImagePrefix = "kurtosistech/engine"
	imageVersionDelimiter         = ":"
)

var engineVersion string
var logLevelStr string
var enclavePoolSize uint8
var gitAuthTokenStr string

// StartCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var StartCmd = &cobra.Command{
	Use:   command_str_consts.EngineStartCmdStr,
	Short: "Starts the Kurtosis engine",
	Long:  "Starts the Kurtosis engine, doing nothing if an engine is already running",
	RunE:  run,
}

func init() {
	StartCmd.Flags().StringVar(
		&engineVersion,
		engineVersionArg,
		defaultEngineVersion,
		"The version (Docker tag) of the Kurtosis engine that should be started (blank will start the default version)",
	)
	StartCmd.Flags().StringVar(
		&logLevelStr,
		logLevelArg,
		defaults.DefaultEngineLogLevel.String(),
		fmt.Sprintf(
			"The level that the started engine should log at (%v)",
			strings.Join(
				logrus_log_levels.GetAcceptableLogLevelStrs(),
				"|",
			),
		),
	)
	StartCmd.Flags().Uint8Var(
		&enclavePoolSize,
		enclavePoolSizeFlag,
		defaults.DefaultEngineEnclavePoolSize,
		fmt.Sprintf(
			"The enclave pool size, the default value is '%v' which means it will be disabled. CAUTION: This is only available for Kubernetes, and this command will fail if you want to use it for Docker.",
			defaults.DefaultEngineEnclavePoolSize,
		),
	)
	StartCmd.Flags().StringVar(
		&gitAuthTokenStr,
		gitAuthTokenArg,
		"",
		"A git personal access token used to provide the engine git auth access.",
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if err := common.ValidateEnclavePoolSizeFlag(enclavePoolSize); err != nil {
		return stacktrace.Propagate(err, "An error occurred validating the '%v' flag", enclavePoolSizeFlag)
	}

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing log level string '%v'", logLevelStr)
	}

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager")
	}

	// TODO: make this work if no secret is returned
	// get password
	userLogin := os.Getenv("GIT_USER")
	secret, err := keyring.Get("kurtosis-git", userLogin)
	if err != nil {
		logrus.Errorf("Unable to get token from keyring")
	}
	logrus.Debugf("Successfully retrieved git token from keyring.")
	logrus.Infof("Successfully retrieved git auth info for user: %v", userLogin)

	// setup git authentication
	gitAuth := &http.BasicAuth{
		Username: "token",
		Password: secret,
	}

	var engineClientCloseFunc func() error
	var startEngineErr error
	if engineVersion == defaultEngineVersion {
		logrus.Infof("Starting Kurtosis engine from image '%v%v%v'...", kurtosisTechEngineImagePrefix, imageVersionDelimiter, kurtosis_version.KurtosisVersion)
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, logLevel, enclavePoolSize, gitAuth)
	} else {
		logrus.Infof("Starting Kurtosis engine from image '%v%v%v'...", kurtosisTechEngineImagePrefix, imageVersionDelimiter, engineVersion)
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithCustomVersion(ctx, engineVersion, logLevel, enclavePoolSize, gitAuth)
	}
	if startEngineErr != nil {
		return stacktrace.Propagate(startEngineErr, "An error occurred starting the Kurtosis engine")
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()

	logrus.Info("Kurtosis engine started")

	return nil
}

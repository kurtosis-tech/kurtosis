package new

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-cli/cli/repl_runner"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	enclaveIDArg           = "enclave-id"
	kurtosisLogLevelArg    = "kurtosis-log-level"
	javascriptReplImageArg = "js-repl-image"
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIDArg,
}

var kurtosisLogLevelStr string
var jsReplImage string

var NewCmd = &cobra.Command{
	Use:                   "new [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Create a new Javascript REPL inside a Kurtosis Enclave",
	RunE:                  run,
}

func init() {
	NewCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
	NewCmd.Flags().StringVarP(
		&jsReplImage,
		javascriptReplImageArg,
		"r",
		defaults.DefaultJavascriptReplImage,
		"The image of the Javascript REPL to connect to the enclave with",
	)
}

func run(cmd *cobra.Command, args []string) error {
	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing Kurtosis loglevel string '%v' to a log level object", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgs(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveId := parsedPositionalArgs[enclaveIDArg]

	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, jsReplImage)

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient)

	enclaveCtx, err := enclaveManager.GetEnclaveContext(context.Background(), enclaveId, logrus.StandardLogger())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave context for enclave ID '%v'", enclaveId)
	}

	logrus.Debug("Running REPL...")
	if err := repl_runner.RunREPL(enclaveCtx, jsReplImage); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Debug("REPL exited")

	return nil
}

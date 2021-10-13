package attach

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager/enclave_context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-cli/cli/repl_container_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	kurtosisLogLevelArg      = "kurtosis-log-level"
	jsReplImageArg           = "js-repl-image"

	replContainerSuccessExitCode = 0

	// This is the directory in which the node REPL is running inside the REPL container, which is where
	//  we'll bind-mount the host machine's current directory into the container so the user can access
	//  files on their host machine
	workingDirpathInsideReplContainer = "/repl"
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	jsReplImageArg,
}

var kurtosisLogLevelStr string

var AttachCmd = &cobra.Command{
	Use:                   "attach " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Attach an interactive Javascript REPL to a Kurtosis Enclave",
	RunE:                  run,
}

func init() {
	AttachCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
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
	jsReplImage, found := parsedPositionalArgs[jsReplImageArg]
	if !found {
		return stacktrace.NewError("No '%v' positional args was found in '%+v' - this is very strange!", jsReplImageArg, parsedPositionalArgs)
	}

	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, jsReplImage)

	enclave_context.NewEnclaveContext()

	logrus.Debug("Running REPL...")
	if err := repl_container_manager.RunReplContainer(dockerManager, enclaveCtx, jsReplImage); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Debug("REPL exited")


	return nil
}

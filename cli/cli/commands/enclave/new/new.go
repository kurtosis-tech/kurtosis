package new

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	apiContainerImageArg     = "api-container-image"
	jsReplImageArg           = "js-repl-image"
	isPartitioningEnabledArg = "partition-enabled"
	kurtosisLogLevelArg      = "kurtosis-log-level"

	shouldPublishPorts = true
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	apiContainerImageArg,
	jsReplImageArg,
}

var kurtosisLogLevelStr string
var isPartitioningEnabled bool

var NewCmd = &cobra.Command{
	Use:                   "new " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Creates a new Kurtosis enclave",
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
	NewCmd.Flags().BoolVar(
		&isPartitioningEnabled,
		isPartitioningEnabledArg,
		false,
		"Enable network partition functionality",
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
	apiContainerImage, found := parsedPositionalArgs[apiContainerImageArg]
	if !found {
		return stacktrace.NewError("No '%v' positional args was found in '%+v' - this is very strange!", apiContainerImageArg, parsedPositionalArgs)
	}
	jsReplImage, found := parsedPositionalArgs[jsReplImageArg]
	if !found {
		return stacktrace.NewError("No '%v' positional args was found in '%+v' - this is very strange!", jsReplImageArg, parsedPositionalArgs)
	}

	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, apiContainerImage)
	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, jsReplImage)

	enclaveId := execution_ids.GetExecutionID()

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient, apiContainerImage)

	_, err = enclaveManager.CreateEnclave(
		context.Background(),
		logrus.StandardLogger(),
		kurtosisLogLevel,
		enclaveId,
		isPartitioningEnabled,
		shouldPublishPorts,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave")
	}
	return nil
}

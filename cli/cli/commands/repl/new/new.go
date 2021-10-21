package new

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-cli/cli/repl_runner"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/lib/kurtosis_context"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	enclaveIDArg           = "enclave-id"
	javascriptReplImageArg = "js-repl-image"
)

var positionalArgs = []string{
	enclaveIDArg,
}

var jsReplImage string

var NewCmd = &cobra.Command{
	Use:                   "new [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Create a new Javascript REPL inside the given Kurtosis enclave",
	RunE:                  run,
}

func init() {
	NewCmd.Flags().StringVarP(
		&jsReplImage,
		javascriptReplImageArg,
		"r",
		defaults.DefaultJavascriptReplImage,
		"The image of the Javascript REPL to connect to the enclave with",
	)
}

func run(cmd *cobra.Command, args []string) error {
	// TODO Set CLI loglevel from a global flag

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveId := parsedPositionalArgs[enclaveIDArg]

	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, jsReplImage)

	kurtosisContext, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis Context")
	}

	enclaves, err := kurtosisContext.GetEnclaves()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave, make sure that you already started Kurtosis Engine Sever with `kurtosis engine start` command")
	}

	enclaveCtx, found := enclaves[enclaveId]
	if !found {
		return stacktrace.Propagate(err, "An error occurred founding enclave with ID '%v' from enclaves map '%+v'", enclaveId, enclaves)
	}

	logrus.Debug("Running REPL...")
	if err := repl_runner.RunREPL(
		enclaveCtx,
		jsReplImage,
		dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Debug("REPL exited")

	return nil
}

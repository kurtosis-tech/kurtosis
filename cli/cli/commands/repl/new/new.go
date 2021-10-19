package new

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/engine_service_client"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-cli/cli/repl_runner"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
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
	ctx := context.Background()

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

	engineServiceClient, closeEngineServiceClient, err := engine_service_client.NewEngineServiceClient()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting engine service client")
	}
	defer closeEngineServiceClient()

	getEnclaveArgs := &kurtosis_engine_rpc_api_bindings.GetEnclaveArgs{
		EnclaveId: enclaveId,
	}

	enclaveCtx, err := engineServiceClient.GetEnclave(ctx, getEnclaveArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave context for enclave ID '%v'", enclaveId)
	}

	logrus.Debug("Running REPL...")
	if err := repl_runner.RunREPL(
		enclaveId,
		enclaveCtx.NetworkId,
		enclaveCtx.ApiContainerHostIp,
		enclaveCtx.ApiContainerHostPort,
		enclaveCtx.ApiContainerIpInsideNetwork,
		jsReplImage,
		dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Debug("REPL exited")

	return nil
}

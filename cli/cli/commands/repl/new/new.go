package new

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/repl_runner"
	positional_arg_parser "github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
)

const (
	enclaveIDArg           = "enclave-id"
	imageArg               = "image"
)

var positionalArgs = []string{
	enclaveIDArg,
}

var jsReplImage string

var NewCmd = &cobra.Command{
	Use:                   command_str_consts.ReplNewCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Create a new REPL inside the given Kurtosis enclave",
	RunE:                  run,
}

func init() {
	NewCmd.Flags().StringVarP(
		&jsReplImage,
		imageArg,
		"i",
		defaults.DefaultJavascriptReplImage,
		"The image of the REPL to connect to the enclave with",
	)
}

func run(cmd *cobra.Command, args []string) error {
	// TODO Set CLI loglevel from a global flag

	ctx := context.Background()

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

	engineManager := engine_manager.NewEngineManager(dockerManager)
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotently(ctx, defaults.DefaultEngineImage)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	response, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err,"An error occurred getting enclaves")
	}
	enclaveInfoMap := response.GetEnclaveInfo()
	enclaveInfo, found := enclaveInfoMap[enclaveId]
	if !found {
		return stacktrace.Propagate(err, "An error occurred finding enclave with ID '%v' on enclave info map '%+v'", enclaveId, enclaveInfoMap)
	}

	apicHostMachineIp, apicHostMachinePort, err := enclave_liveness_validator.ValidateEnclaveLiveness(enclaveInfo)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred validating that the enclave was running")
	}

	logrus.Debug("Running REPL...")
	if err := repl_runner.RunREPL(
		enclaveInfo.GetEnclaveId(),
		enclaveInfo.GetNetworkId(),
		enclaveInfo.GetApiContainerInfo().GetIpInsideEnclave(),
		enclaveInfo.GetApiContainerInfo().GetPortInsideEnclave(),
		apicHostMachineIp,
		apicHostMachinePort,
		jsReplImage,
		dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Debug("REPL exited")

	return nil
}

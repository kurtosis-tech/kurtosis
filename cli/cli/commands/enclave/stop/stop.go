package stop

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/version_checker"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	enclaveIdArg        = "enclave-id"
)

var StopCmd = &cobra.Command{
	Use:   command_str_consts.EnclaveStopCmdStr + " [flags] " + enclaveIdArg + " [" + enclaveIdArg + "...]",
	DisableFlagsInUseLine: true,
	Short: "Stops the specified enclaves",
	RunE:  run,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	version_checker.CheckLatestVersion()

	logrus.Info("Stopping enclaves...")

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)
	engineManager := engine_manager.NewEngineManager(dockerManager)
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotently(ctx, defaults.DefaultEngineImage, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	stopEnclaveErrorStrs := []string{}
	for _, enclaveId := range args {
		stopArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{EnclaveId: enclaveId}
		if _, err := engineClient.StopEnclave(ctx, stopArgs); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveId)
			stopEnclaveErrorStrs = append(stopEnclaveErrorStrs, wrappedErr.Error())
		}
	}

	if len(stopEnclaveErrorStrs) > 0 {
		joinedErrorsStr := strings.Join(
			stopEnclaveErrorStrs,
			"\n\n",
		)
		// We use this rather than stacktrace because stacktrace gets messy
		return errors.New(
			fmt.Sprintf(
				"One or more errors occurred when stopping enclaves:\n%v",
				joinedErrorsStr,
			),
		)
	}

	logrus.Info("Enclaves stopped successfully")

	return nil
}

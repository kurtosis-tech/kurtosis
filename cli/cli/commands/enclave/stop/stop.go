package stop

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	enclaveIdArg = "enclave-id"
)

var StopCmd = &cobra.Command{
	Use:                   command_str_consts.EnclaveStopCmdStr + " [flags] " + enclaveIdArg + " [" + enclaveIdArg + "...]",
	DisableFlagsInUseLine: true,
	Short:                 "Stops the specified enclaves",
	RunE:                  run,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if len(args) == 0 {
		return stacktrace.NewError("At least one enclave ID to stop must be provided")
	}

	logrus.Info("Stopping enclaves...")

	// TODO REFACTOR: we should get this backend from the config!!
	var apiContainerModeArgs *backend_creator.APIContainerModeArgs = nil  // Not an API container
	kurtosisBackend, err := backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a Kurtosis backend connected to local Docker")
	}
	engineManager := engine_manager.NewEngineManager(kurtosisBackend)
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, defaults.DefaultEngineLogLevel)
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

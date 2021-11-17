package rm

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
)

const (
	shouldForceRemoveArg = "force"
	enclaveIdArg        = "enclave-id"

	defaultShouldForceRemove = false
)

var RmCmd = &cobra.Command{
	Use:   command_str_consts.EnclaveRmCmdStr + " [flags] " + enclaveIdArg + " [" + enclaveIdArg + "...]",
	DisableFlagsInUseLine: true,
	Short: "Destroys the specified enclaves",
	Long: "Destroys the specified enclaves, removing all resources associated with them",
	RunE:  run,
}

var shouldForceRemove bool

func init() {
	RmCmd.Flags().BoolVarP(
		&shouldForceRemove,
		shouldForceRemoveArg,
		"f",
		defaultShouldForceRemove,
		"Deletes all enclaves, regardless of whether they're already stopped",
	)

}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	logrus.Info("Destroying enclaves...")

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)
	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves to check that the ones to destroy are stopped")
	}
	allEnclaveInfo := getEnclavesResp.EnclaveInfo

	enclaveDestructionErrorStrs := []string{}
	for _, enclaveId := range args {
		if err := destroyEnclave(ctx, enclaveId, allEnclaveInfo, engineClient); err != nil {
			enclaveDestructionErrorStrs = append(enclaveDestructionErrorStrs, err.Error())
		}
	}

	if len(enclaveDestructionErrorStrs) > 0 {
		errorStr := fmt.Sprintf(
			"One or more errors occurred destroying the enclaves:\n%v",
			strings.Join(enclaveDestructionErrorStrs, "\n\n"),
		)
		return errors.New(errorStr)
	}

	logrus.Info("Enclaves successfully destroyed")

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func destroyEnclave(ctx context.Context, enclaveId string, allEnclaveInfo map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo, engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient) error {
	enclaveInfo, found := allEnclaveInfo[enclaveId]
	if !found {
		return stacktrace.NewError("No enclave '%v' exists", enclaveId)
	}

	enclaveStatus := enclaveInfo.ContainersStatus
	var isEnclaveRemovableWithoutForce bool
	switch enclaveStatus {
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_EMPTY, kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED:
		isEnclaveRemovableWithoutForce = true
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING:
		isEnclaveRemovableWithoutForce = false
	default:
		return stacktrace.NewError("Unrecognized enclave status '%v'; this is a bug in Kurtosis", enclaveStatus)
	}

	if !shouldForceRemove && !isEnclaveRemovableWithoutForce {
		return stacktrace.NewError(
			"Refusing to destroy enclave '%v' because its status is '%v'; to force its removal, rerun this command with the '%v' flag",
			enclaveId,
			enclaveStatus,
			shouldForceRemoveArg,
		)
	}

	destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{EnclaveId: enclaveId}
	if _, err := engineClient.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclave '%v'", enclaveId)
	}
	return nil
}

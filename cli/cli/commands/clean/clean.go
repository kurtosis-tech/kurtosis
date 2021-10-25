package clean

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/container_status_calculator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_labels_schema"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strings"
)

const (
	shouldCleanAllArgs = "all"

	defaultShouldCleanAll = false

	shouldConsiderStoppedContainersWhenSearchingForStoppedEngines = true
)

var CleanCmd = &cobra.Command{
	Use:   command_str_consts.CleanCmdStr,
	Short: "Cleans up Kurtosis leftover artifacts",
	Long: fmt.Sprintf(
		"Removes Kurtosis stopped Kurtosis enclaves (and live ones if the '%v' flag is set), as well as stopped engine containers",
		shouldCleanAllArgs,
	),
	RunE:  run,
}

var shouldCleanAll bool

func init() {
	CleanCmd.Flags().BoolVarP(
		&shouldCleanAll,
		shouldCleanAllArgs,
		"a",
		defaultShouldCleanAll,
		"If set, removes running enclaves as well",
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)
	engineManager := engine_manager.NewEngineManager(dockerManager)
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotently(ctx, defaults.DefaultEngineImage)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	// First, clean up old engine containers




	if didCleanErrorsOccur {

	}
	return nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func cleanStoppedEngineContainers(ctx context.Context, dockerManager *docker_manager.DockerManager) (hadErrors bool) {
	engineContainers, err := dockerManager.GetContainersByLabels(ctx, engine_labels_schema.EngineContainerLabels, shouldConsiderStoppedContainersWhenSearchingForStoppedEngines)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers")
	}

	containersToDestroy := []*types.Container{}

	for _, container := range engineContainers {
		containerId := container.GetId()
		containerName := container.GetName()
		containerStatus := container.GetStatus()
		isRunning, err := container_status_calculator.IsContainerRunning(containerStatus)
		if err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred determining if engine container '%v' with status '%v' is running", containerName, containerStatus)
			removeContainerErrorStrs = append(removeContainerErrorStrs, wrappedErr.Error())
			continue
		}

		if !isRunning {
			if err := dockerManager.RemoveContainer(ctx, containerId); err != nil {
				wrappedErr := stacktrace.Propagate(err, "An error occurred removing stopped engine container '%v' with ID '%v'", containerName, containerId)
				removeContainerErrorStrs = append(removeContainerErrorStrs, wrappedErr.Error())
				continue
			}
			successfullyDestroyedEngineContainerNames = append(successfullyDestroyedEngineContainerNames, containerName)
		}
	}

	successfullyDestroyedEngineContainerNames := []string{}
	removeContainerErrorStrs := []string{}

	if len(successfullyDestroyedEngineContainerNames) > 0 {
		logrus.Infof("Removed the following old engine containers:")
		for _, destroyedContainerName := range successfullyDestroyedEngineContainerNames {
			fmt.Fprintln(logrus.StandardLogger().Out, destroyedContainerName)
		}
	}

	numErrors := len(removeContainerErrorStrs)
	if numErrors > 0 {
		logrus.Errorf(
			"One or more errors occurred destroying old engine containers:\n%v",
			strings.Join(
				removeContainerErrorStrs,
				"\n\n",
			),
		)
		return true
	}
	return false
}

func cleanEnclaves(ctx context.Context, engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient) (hadErrors bool) {
	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves to determine which need to be cleaned up")
	}

	successfullyDestroyedEnclaveIds := []string{}
	enclaveDestructionErrorStrs := []string{}
	for enclaveId, enclaveInfo := range getEnclavesResp.EnclaveInfo {
		enclaveStatus := enclaveInfo.ContainersStatus
		if shouldCleanAll || enclaveStatus == kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED {
			destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{EnclaveId: enclaveId}
			if _, err := engineClient.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
				wrappedErr := stacktrace.Propagate(err, "An error occurred destroying enclave '%v'", enclaveId)
				enclaveDestructionErrorStrs = append(enclaveDestructionErrorStrs, wrappedErr.Error())
			} else {
				successfullyDestroyedEnclaveIds = append(successfullyDestroyedEnclaveIds, enclaveId)
			}
		}
	}

	sort.Strings(successfullyDestroyedEnclaveIds)
	logrus.Info("Successfully destroyed the following enclaves:")
	for _, enclaveId := range successfullyDestroyedEnclaveIds {
		fmt.Fprintln(logrus.StandardLogger().Out, enclaveId)
	}

	didCleanErrorsOccur := false
	if len(enclaveDestructionErrorStrs) > 0 {
		logrus.Errorf(
			"One or more errors occurred destroying enclaves:\n%v",
			strings.Join(
				enclaveDestructionErrorStrs,
				"\n\n",
			),
		)
		didCleanErrorsOccur = true
	}

}
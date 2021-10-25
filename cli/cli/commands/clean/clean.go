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
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
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

	shouldCleanRunningEngineContainers = false

	// Obviously yes
	shouldFetchStoppedContainersWhenDestroyingStoppedContainers = true

	// Titles of the cleaning phases
	// Should be lowercased as they'll go into a string like "Cleaning XXXXX...."
	oldEngineCleaningPhaseTitle = "old Kurtosis engine containers"
	metadataAcquisitionTestsuitePhaseTitle = "metadata-acquiring testsuite containers"
	enclavesCleaningPhaseTitle = "enclaves"
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

	// Map of cleaning_phase_title -> (successfully_destroyed_object_id, object_destruction_errors, clean_error)
	cleaningPhaseFunctions := map[string]func() ([]string, []error, error) {
		oldEngineCleaningPhaseTitle: func() ([]string, []error, error) {
			// Don't use stacktrace b/c the only reason this function exists is to pass in the right args
			return cleanStoppedEngineContainers(ctx, dockerManager)
		},
		enclavesCleaningPhaseTitle: func() ([]string, []error, error) {
			// Don't use stacktrace b/c the only reason this function exists is to pass in the right args
			return cleanEnclaves(ctx, engineClient)
		},
		metadataAcquisitionTestsuitePhaseTitle: func() ([]string, []error, error) {
			// Don't use stacktrace b/c the only reason this function exists is to pass in the right args
			return cleanMetadataAcquisitionTestsuites(ctx, dockerManager, shouldCleanAll)
		},
	}

	phasesWithErrors := []string{}
	for phaseTitle, cleaningFunc := range cleaningPhaseFunctions {
		logrus.Infof("Cleaning %v...", phaseTitle)
		successfullyRemovedArtifactIds, removalErrors, err := cleaningFunc()
		if err != nil {
			logrus.Errorf("Errors occurred cleaning %v:\n%v", phaseTitle, err)
			phasesWithErrors = append(phasesWithErrors, phaseTitle)
			continue
		}

		if len(successfullyRemovedArtifactIds) > 0 {
			logrus.Infof("Successfully removed the following %v:", phaseTitle)
			sort.Strings(successfullyRemovedArtifactIds)
			for _, successfulArtifactId := range successfullyRemovedArtifactIds {
				fmt.Fprintln(logrus.StandardLogger().Out, successfulArtifactId)
			}
		}

		if len(removalErrors) > 0 {
			logrus.Errorf("Errors occurred removing the following %v:", phaseTitle)
			for _, err := range removalErrors {
				fmt.Fprintln(logrus.StandardLogger().Out, "")
				fmt.Fprintln(logrus.StandardLogger().Out, err.Error())
			}
			phasesWithErrors = append(phasesWithErrors, phaseTitle)
			continue
		}
		logrus.Infof("Successfully cleaned %v", phaseTitle)
	}

	if len(phasesWithErrors) > 0 {
		errorStr := "Errors occurred cleaning " + strings.Join(phasesWithErrors, ", ")
		return errors.New(errorStr)
	}
	return nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func cleanStoppedEngineContainers(ctx context.Context, dockerManager *docker_manager.DockerManager) ([]string, []error, error) {
	successfullyDestroyedContainerNames, containerDestructionErrors, err := cleanContainers(ctx, dockerManager, engine_labels_schema.EngineContainerLabels, shouldCleanRunningEngineContainers)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred cleaning stopped Kurtosis engine containers")
	}
	return successfullyDestroyedContainerNames, containerDestructionErrors, nil
}

func cleanMetadataAcquisitionTestsuites(ctx context.Context, dockerManager *docker_manager.DockerManager, shouldKillRunningContainers bool) ([]string, []error, error) {
	metadataAcquisitionTestsuiteLabels := map[string]string{
		enclave_object_labels.ContainerTypeLabel: enclave_object_labels.ContainerTypeTestsuiteContainer,
		enclave_object_labels.TestsuiteTypeLabelKey: enclave_object_labels.TestsuiteTypeLabelValue_MetadataAcquisition,
	}
	successfullyDestroyedContainerNames, containerDestructionErrors, err := cleanContainers(ctx, dockerManager, metadataAcquisitionTestsuiteLabels, shouldKillRunningContainers)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred cleaning metadata-acquisition testsuite containers")
	}
	return successfullyDestroyedContainerNames, containerDestructionErrors, nil
}

func cleanEnclaves(ctx context.Context, engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient) ([]string, []error, error) {
	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclaves to determine which need to be cleaned up")
	}

	enclaveIdsToDestroy := []string{}
	for enclaveId, enclaveInfo := range getEnclavesResp.EnclaveInfo {
		enclaveStatus := enclaveInfo.ContainersStatus
		if shouldCleanAll || enclaveStatus == kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED {
			enclaveIdsToDestroy = append(enclaveIdsToDestroy, enclaveId)
		}
	}

	successfullyDestroyedEnclaveIds := []string{}
	enclaveDestructionErrors := []error{}
	for _, enclaveId := range enclaveIdsToDestroy {
		destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{EnclaveId: enclaveId}
		if _, err := engineClient.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred removing enclave '%v'", enclaveId)
			enclaveDestructionErrors = append(enclaveDestructionErrors, wrappedErr)
			continue
		}
		successfullyDestroyedEnclaveIds = append(successfullyDestroyedEnclaveIds, enclaveId)
	}
	return successfullyDestroyedEnclaveIds, enclaveDestructionErrors, nil
}

func cleanContainers(ctx context.Context, dockerManager *docker_manager.DockerManager, searchLabels map[string]string, shouldKillRunningContainers bool) ([]string, []error, error) {
	matchingContainers, err := dockerManager.GetContainersByLabels(
		ctx,
		searchLabels,
		shouldFetchStoppedContainersWhenDestroyingStoppedContainers,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting containers matching labels '%+v'", searchLabels)
	}

	stoppedContainers := []*types.Container{}
	for _, container := range matchingContainers {
		containerName := container.GetName()
		containerStatus := container.GetStatus()
		if shouldKillRunningContainers {
			stoppedContainers = append(stoppedContainers, container)
			continue
		}

		isRunning, err := container_status_calculator.IsContainerRunning(containerStatus)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred determining if container '%v' with status '%v' is running", containerName, containerStatus)
		}
		if !isRunning {
			stoppedContainers = append(stoppedContainers, container)
		}
	}

	successfullyDestroyedContainerNames := []string{}
	removeContainerErrors := []error{}
	for _, container := range stoppedContainers {
		containerId := container.GetId()
		containerName := container.GetName()
		if err := dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred removing stopped container '%v'", containerName)
			removeContainerErrors = append(removeContainerErrors, wrappedErr)
			continue
		}
		successfullyDestroyedContainerNames = append(successfullyDestroyedContainerNames, containerName)
	}

	return successfullyDestroyedContainerNames, removeContainerErrors, nil
}
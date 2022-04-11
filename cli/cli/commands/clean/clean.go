package clean

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/container_status_calculator"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

const (
	shouldCleanRunningEnclavesFlagKey = "all"
	defaultShouldCleanRunningEnclaves = "false"

	shouldCleanRunningEngineContainers = false

	// Obviously yes
	shouldFetchStoppedContainersWhenDestroyingStoppedContainers = true

	// Titles of the cleaning phases
	// Should be lowercased as they'll go into a string like "Cleaning XXXXX...."
	oldEngineCleaningPhaseTitle = "old Kurtosis engine containers"
	enclavesCleaningPhaseTitle  = "enclaves"

	kurtosisBackendCtxKey = "kurtosis-backend"
	dockerManagerCtxKey = "docker-manager"
	engineClientCtxKey = "engine-client"
)

var CleanCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:              command_str_consts.CleanCmdStr,
	ShortDescription: "Cleans up Kurtosis leftover artifacts",
	LongDescription: fmt.Sprintf(
		"Removes Kurtosis stopped Kurtosis enclaves (and live ones if the '%v' flag is set), as well as stopped engine containers",
		shouldCleanRunningEnclavesFlagKey,
	),
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	DockerManagerContextKey: dockerManagerCtxKey,
	EngineClientContextKey:  engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:       shouldCleanRunningEnclavesFlagKey,
			Usage:     "If set, removes running enclaves as well",
			Shorthand: "a",
			Type:      flags.FlagType_Bool,
			Default:   defaultShouldCleanRunningEnclaves,
		},
	},
	RunFunc:                 run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	dockerManager *docker_manager.DockerManager,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	shouldCleanAll, err := flags.GetBool(shouldCleanRunningEnclavesFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", shouldCleanAll)
	}

	// Map of cleaning_phase_title -> (successfully_destroyed_object_id, object_destruction_errors, clean_error)
	cleaningPhaseFunctions := map[string]func() ([]string, []error, error){
		oldEngineCleaningPhaseTitle: func() ([]string, []error, error) {
			// Don't use stacktrace b/c the only reason this function exists is to pass in the right args
			return cleanStoppedEngineContainers(ctx, dockerManager)
		},
		enclavesCleaningPhaseTitle: func() ([]string, []error, error) {
			// Don't use stacktrace b/c the only reason this function exists is to pass in the right args
			return cleanEnclaves(ctx, engineClient, shouldCleanAll)
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
	engineContainerLabels := map[string]string{
		forever_constants.AppIDLabel:         forever_constants.AppIDValue,
		forever_constants.ContainerTypeLabel: forever_constants.ContainerType_EngineServer,
	}
	successfullyDestroyedContainerNames, containerDestructionErrors, err := cleanContainers(ctx, dockerManager, engineContainerLabels, shouldCleanRunningEngineContainers)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred cleaning stopped Kurtosis engine containers")
	}
	return successfullyDestroyedContainerNames, containerDestructionErrors, nil
}

func cleanEnclaves(ctx context.Context, engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient, shouldCleanAll bool) ([]string, []error, error) {
	cleanArgs := &kurtosis_engine_rpc_api_bindings.CleanArgs{ShouldCleanAll: shouldCleanAll}
	cleanResp, err := engineClient.Clean(ctx, cleanArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while calling clean")
	}

	successfullyDestroyedEnclaveIds := []string{}
	for enclaveId, _ := range cleanResp.RemovedEnclaveIds {
		successfullyDestroyedEnclaveIds = append(successfullyDestroyedEnclaveIds, enclaveId)
	}
	return successfullyDestroyedEnclaveIds, nil, nil
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

	containersToDestroy := []*types.Container{}
	for _, container := range matchingContainers {
		containerName := container.GetName()
		containerStatus := container.GetStatus()
		if shouldKillRunningContainers {
			containersToDestroy = append(containersToDestroy, container)
			continue
		}

		isRunning, err := container_status_calculator.IsContainerRunning(containerStatus)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred determining if container '%v' with status '%v' is running", containerName, containerStatus)
		}
		if !isRunning {
			containersToDestroy = append(containersToDestroy, container)
		}
	}

	successfullyDestroyedContainerNames := []string{}
	removeContainerErrors := []error{}
	for _, container := range containersToDestroy {
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

package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
)

func DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
	dockerManager *docker_manager.DockerManager,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {

	matchingEnginesByContainerId, err := getMatchingEngines(ctx, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	logsComponentsContainersByEngineContainerId, err := getLogsComponentsContainerIdsByEngineContainerIds(ctx, matchingEnginesByContainerId, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs componentes containers by engine container IDs '%+v'", matchingEnginesByContainerId)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedEnginesByContainerId := map[string]interface{}{}
	for containerId, engineObj := range matchingEnginesByContainerId {
		matchingUncastedEnginesByContainerId[containerId] = interface{}(engineObj)
	}

	var removeEngineOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		engineContainerId := dockerObjectId
		if err := dockerManager.RemoveContainer(ctx, engineContainerId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing engine container with ID '%v'", engineContainerId)
		}

		logsComponentsToRemoveContainerIDs, found := logsComponentsContainersByEngineContainerId[engineContainerId]
		if !found {
			return nil
		}

		if logsComponentsToRemoveContainerIDs.logsDatabaseContainerId != "" {
			if err := dockerManager.RemoveContainer(ctx, logsComponentsToRemoveContainerIDs.logsDatabaseContainerId); err != nil {
				return stacktrace.Propagate(err, "An error occurred removing logs database container with ID '%v'", logsComponentsToRemoveContainerIDs.logsDatabaseContainerId)
			}
		}

		if logsComponentsToRemoveContainerIDs.logsCollectorContainerId != "" {
			if err := dockerManager.RemoveContainer(ctx, logsComponentsToRemoveContainerIDs.logsCollectorContainerId); err != nil {
				return stacktrace.Propagate(err, "An error occurred removing logs collector container with ID '%v'", logsComponentsToRemoveContainerIDs.logsCollectorContainerId)
			}
		}

		return nil
	}

	successfulEngineGuidStrs, erroredEngineGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedEnginesByContainerId,
		dockerManager,
		extractEngineGuidFromUncastedEngineObj,
		removeEngineOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing engine containers matching filters '%+v'", filters)
	}

	successfulGuids := map[engine.EngineGUID]bool{}
	for guidStr := range successfulEngineGuidStrs {
		successfulGuids[engine.EngineGUID(guidStr)] = true
	}

	erroredGuids := map[engine.EngineGUID]error{}
	for guidStr, err := range erroredEngineGuidStrs {
		erroredGuids[engine.EngineGUID(guidStr)] = stacktrace.Propagate(
			err,
		"An error occurred destroying engine '%v'",
			guidStr,
		)
	}

	//Remove the logs components volumes only if are not in use
	allLogsDatabaseContainers, err := getAllLogsDatabaseContainers(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting all logs database containers")
	}
	if len(allLogsDatabaseContainers) == 0 {
		logsDatabaseVolumeName, err := getLogsDatabaseVolumeName(ctx, dockerManager)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs database volume name")
		}

		if err := dockerManager.RemoveVolume(ctx, logsDatabaseVolumeName); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred removing logs database volume '%v'", logsDatabaseVolumeName)
		}
	}

	allLogsCollectorContainers, err := shared_helpers.GetAllLogsCollectorContainers(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting all logs collector containers")
	}
	if len(allLogsCollectorContainers) == 0 {
		logsCollectorVolumeName, err := getLogsCollectorVolumeName(ctx, dockerManager)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs collector volume name")
		}

		if err := dockerManager.RemoveVolume(ctx, logsCollectorVolumeName); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred removing logs collector volume '%v'", logsCollectorVolumeName)
		}
	}

	return successfulGuids, erroredGuids, nil
}

func getLogsDatabaseVolumeName(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (string, error) {
	volumeSearchLabels :=  map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():      label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.VolumeTypeDockerLabelKey.GetString(): label_value_consts.LogsDatabaseVolumeTypeDockerLabelValue.GetString(),
	}
	foundVolumes, err := dockerManager.GetVolumesByLabels(ctx, volumeSearchLabels)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting logs database volumes matching labels '%+v'", volumeSearchLabels)
	}
	if len(foundVolumes) > 1 {
		return "", stacktrace.NewError("Found multiple logs database volumes; this should never happen")
	}
	if len(foundVolumes) == 0 {
		return "", stacktrace.NewError("No logs database volume found")
	}
	volume := foundVolumes[0]
	return volume.Name, nil
}

func getLogsCollectorVolumeName(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (string, error) {
	volumeSearchLabels :=  map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():      label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.VolumeTypeDockerLabelKey.GetString(): label_value_consts.LogsCollectorVolumeTypeDockerLabelValue.GetString(),
	}
	foundVolumes, err := dockerManager.GetVolumesByLabels(ctx, volumeSearchLabels)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting logs collector volumes matching labels '%+v'", volumeSearchLabels)
	}
	if len(foundVolumes) > 1 {
		return "", stacktrace.NewError("Found multiple logs collector volumes; this should never happen")
	}
	if len(foundVolumes) == 0 {
		return "", stacktrace.NewError("No logs collector volume found")
	}
	volume := foundVolumes[0]
	return volume.Name, nil
}




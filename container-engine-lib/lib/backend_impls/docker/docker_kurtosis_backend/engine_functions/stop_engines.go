package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
)

func StopEngines(
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
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines matching filters '%+v'", filters)
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

	var killEngineOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		engineContainerId := dockerObjectId
		if err := dockerManager.KillContainer(ctx, engineContainerId); err != nil {
			return stacktrace.Propagate(err, "An error occurred killing engine container with ID '%v'", dockerObjectId)
		}
		logsComponentsToKillContainerIDs, found := logsComponentsContainersByEngineContainerId[engineContainerId]
		if !found {
			return nil
		}

		if logsComponentsToKillContainerIDs.logsDatabaseContainerId != "" {
			if err := dockerManager.KillContainer(ctx, logsComponentsToKillContainerIDs.logsDatabaseContainerId); err != nil {
				return stacktrace.Propagate(err, "An error occurred killing logs database container with ID '%v'", logsComponentsToKillContainerIDs.logsDatabaseContainerId)
			}
		}

		if logsComponentsToKillContainerIDs.logsCollectorContainerId != "" {
			if err := dockerManager.KillContainer(ctx, logsComponentsToKillContainerIDs.logsCollectorContainerId); err != nil {
				return stacktrace.Propagate(err, "An error occurred killing logs collector container with ID '%v'", logsComponentsToKillContainerIDs.logsDatabaseContainerId)
			}
		}

		return nil
	}

	successfulEngineGuidStrs, erroredEngineGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedEnginesByContainerId,
		dockerManager,
		extractEngineGuidFromUncastedEngineObj,
		killEngineOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred killing engine containers matching filters '%+v'", filters)
	}

	successfulGuids := map[engine.EngineGUID]bool{}
	for guidStr := range successfulEngineGuidStrs {
		successfulGuids[engine.EngineGUID(guidStr)] = true
	}

	erroredGuids := map[engine.EngineGUID]error{}
	for guidStr, err := range erroredEngineGuidStrs {
		erroredGuids[engine.EngineGUID(guidStr)] = stacktrace.Propagate(
			err,
			"An error occurred stopping engine '%v'",
			guidStr,
		)
	}

	//TODO what happens with centralized logs components (logs database and logs collector) containers when the engine is stopped?

	return successfulGuids, erroredGuids, nil
}

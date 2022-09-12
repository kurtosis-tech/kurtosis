package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
	"time"
)

const (
	stopLogsComponentsContainersTimeout = 1 * time.Minute
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

	if err := removeLogsComponentesGraceFully(ctx, dockerManager); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing the logs componentes containers")
	}

	return successfulGuids, erroredGuids, nil
}

func removeLogsComponentesGraceFully(ctx context.Context, dockerManager *docker_manager.DockerManager) error {

	logsCollectorContainer, err := shared_helpers.GetLogsCollectorContainer(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs collector container")
	}

	if err := dockerManager.StopContainer(ctx, logsCollectorContainer.GetId(), stopLogsComponentsContainersTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs collector container with ID '%v'", logsCollectorContainer.GetId())
	}

	if err := dockerManager.RemoveContainer(ctx, logsCollectorContainer.GetId()); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the logs collector container with ID '%v'", logsCollectorContainer.GetId())
	}

	logsDatabaseContainer, err := getLogsDatabaseContainer(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs database container")
	}

	if err := dockerManager.StopContainer(ctx, logsDatabaseContainer.GetId(), stopLogsComponentsContainersTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs database container with ID '%v'", logsDatabaseContainer.GetId())
	}

	if err := dockerManager.RemoveContainer(ctx, logsDatabaseContainer.GetId()); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the logs database container with ID '%v'", logsDatabaseContainer.GetId())
	}

	return nil
}

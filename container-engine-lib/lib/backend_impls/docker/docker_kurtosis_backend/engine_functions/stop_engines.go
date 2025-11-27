package engine_functions

import (
	"context"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/reverse_proxy_functions"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
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

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedEnginesByContainerId := map[string]*engine.Engine{}
	for containerId, engineObj := range matchingEnginesByContainerId {
		matchingUncastedEnginesByContainerId[containerId] = engineObj
	}

	var killEngineOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		engineContainerId := dockerObjectId
		// TODO Switch to graceful stop to ensure we don't get database corruption
		if err := dockerManager.KillContainer(ctx, engineContainerId); err != nil {
			return stacktrace.Propagate(err, "An error occurred killing engine container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulEngineGuidStrs, erroredEngineGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedEnginesByContainerId,
		dockerManager,
		extractEngineGuidFromEngine,
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

	// Stop centralized logging components
	if err := logs_aggregator_functions.DestroyLogsAggregator(ctx, dockerManager); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing the logging components.")
	}

	// Stop reverse proxy
	if err := reverse_proxy_functions.DestroyReverseProxy(ctx, dockerManager); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing the reverse proxy.")
	}

	return successfulGuids, erroredGuids, nil
}

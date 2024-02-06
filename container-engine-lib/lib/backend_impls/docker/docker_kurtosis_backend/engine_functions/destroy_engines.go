package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
)

func DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
	dockerManager *docker_manager.DockerManager,
	objsAttrProvider object_attributes_provider.DockerObjectAttributesProvider,

) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {

	matchingEnginesByContainerId, err := getMatchingEngines(ctx, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
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

		return nil
	}

	successfulEngineGuidStrs, erroredEngineGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingEnginesByContainerId,
		dockerManager,
		extractEngineGuidFromEngine,
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

	// Remove GitHub Auth storage associated with successfully removed engine
	for guid, _ := range successfulGuids {
		githubAuthStorageAttrs, err := objsAttrProvider.ForGitHubAuthStorageVolume(guid)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred retrieving attributes for GitHub auth storage volume.")
		}
		githubAuthStorageVolNamStr := githubAuthStorageAttrs.GetName().GetString()
		err = dockerManager.RemoveVolume(ctx, githubAuthStorageVolNamStr)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred removing GitHub auth storage volume.")
		}
	}

	return successfulGuids, erroredGuids, nil
}

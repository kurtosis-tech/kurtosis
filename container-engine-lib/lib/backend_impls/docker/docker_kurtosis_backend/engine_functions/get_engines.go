package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
)

func GetEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
	dockerManager *docker_manager.DockerManager,
) (
	map[engine.EngineGUID]*engine.Engine,
	error,
) {
	matchingEngines, err := getMatchingEngines(ctx, filters, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	matchingEnginesByEngineGuid := map[engine.EngineGUID]*engine.Engine{}
	for _, engineObj := range matchingEngines {
		matchingEnginesByEngineGuid[engineObj.GetGUID()] = engineObj
	}

	return matchingEnginesByEngineGuid, nil
}

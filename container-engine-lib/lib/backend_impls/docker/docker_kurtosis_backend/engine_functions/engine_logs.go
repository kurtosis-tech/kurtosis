package engine_functions

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/stacktrace"
)

func GetEngineLogs(ctx context.Context, outputDirpath string, dockerManager *docker_manager.DockerManager) error {
	engineContainerSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.ContainerTypeDockerLabelKey.GetString(): label_value_consts.EngineContainerTypeDockerLabelValue.GetString(),
		// NOTE: we do NOT use the engine GUID label here, and instead do postfiltering, because Docker has no way to do disjunctive search!
	}
	allEngineContainers, err := dockerManager.GetContainersByLabels(ctx, engineContainerSearchLabels, consts.ShouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred fetching engine containers using labels: %+v", engineContainerSearchLabels)
	}

	if err = shared_helpers.DumpContainers(ctx, dockerManager, allEngineContainers, outputDirpath); err != nil {
		// the error is already wrapped properly
		return err
	}

	return nil
}

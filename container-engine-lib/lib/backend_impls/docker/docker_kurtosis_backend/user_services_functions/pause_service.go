package user_service_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func PauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	dockerManager *docker_manager.DockerManager,
) error {
	_, dockerResources, err := shared_helpers.GetSingleUserServiceObjAndResourcesNoMutex(ctx, enclaveId, serviceUuid, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get information about service '%v' from Kurtosis ", serviceUuid)
	}
	container := dockerResources.ServiceContainer
	if container == nil {
		return stacktrace.NewError("Cannot pause service '%v' as it doesn't have a container to pause", serviceUuid)
	}
	if err = dockerManager.PauseContainer(ctx, container.GetId()); err != nil {
		return stacktrace.Propagate(err, "Failed to pause container '%v' for service '%v' ", container.GetName(), serviceUuid)
	}
	return nil
}

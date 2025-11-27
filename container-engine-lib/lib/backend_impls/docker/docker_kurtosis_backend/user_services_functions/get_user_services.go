package user_service_functions

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceUUID]*service.Service,
	error,
) {
	userServices, _, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, filters, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}
	return userServices, nil
}

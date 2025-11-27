package user_service_functions

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func getSingleUserServiceObjAndResourcesNoMutex(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	userServiceUuid service.ServiceUUID,
	dockerManager *docker_manager.DockerManager,
) (
	*service.Service,
	*shared_helpers.UserServiceDockerResources,
	error,
) {
	filters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			userServiceUuid: true,
		},
		Statuses: nil,
	}
	userServices, dockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services using filters '%v'", filters)
	}
	numOfUserServices := len(userServices)
	if numOfUserServices == 0 {
		return nil, nil, stacktrace.NewError("No user service with UUID '%v' in enclave with ID '%v' was found", userServiceUuid, enclaveId)
	}
	if numOfUserServices > 1 {
		return nil, nil, stacktrace.NewError("Expected to find only one user service with UUID '%v' in enclave with ID '%v', but '%v' was found", userServiceUuid, enclaveId, numOfUserServices)
	}

	var resultService *service.Service
	for _, resultService = range userServices {
	}

	var resultDockerResources *shared_helpers.UserServiceDockerResources
	for _, resultDockerResources = range dockerResources {
	}

	return resultService, resultDockerResources, nil
}

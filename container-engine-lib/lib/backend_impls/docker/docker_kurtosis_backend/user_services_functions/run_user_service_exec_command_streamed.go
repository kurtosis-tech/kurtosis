package user_service_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func RunUserServiceExecCommandWithStreamedOutput(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	cmd []string,
	dockerManager *docker_manager.DockerManager,
) (chan string, chan *exec_result.ExecResult, error) {
	userServiceUuids := map[service.ServiceUUID]bool{}
	userServiceUuids[serviceUuid] = true
	filters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    userServiceUuids,
		Statuses: nil,
	}
	_, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	var userServiceDockerResource *shared_helpers.UserServiceDockerResources
	if dockerResource, found := allDockerResources[serviceUuid]; found {
		userServiceDockerResource = dockerResource
	} else {
		return nil, nil, stacktrace.NewError("No docker resources were found for the service with identifier: '%v'", serviceUuid)
	}

	userServiceDockerContainer := userServiceDockerResource.ServiceContainer

	execOutputLinesChan, finalExecChan, err := dockerManager.RunExecCommandWithStreamedOutput(
		ctx,
		userServiceDockerContainer.GetId(),
		cmd)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred attempting to stream exec output from docker.")
	}

	return execOutputLinesChan, finalExecChan, nil
}

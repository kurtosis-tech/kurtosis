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
	userServiceCommands map[service.ServiceUUID][]string,
	dockerManager *docker_manager.DockerManager,
) (chan string, chan *exec_result.ExecResult, error) {
	// only process 1 exec command for now
	if len(userServiceCommands) > 1 {
		return nil, nil, stacktrace.NewError("Can only stream one exec function at a time.")
	}

	userServiceDockerResources := map[service.ServiceUUID]*shared_helpers.UserServiceDockerResources{}
	userServiceUuids := map[service.ServiceUUID]bool{}
	for userServiceUuid := range userServiceCommands {
		userServiceUuids[userServiceUuid] = true
	}

	filters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    userServiceUuids,
		Statuses: nil,
	}

	_, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}
	for serviceUuid := range userServiceCommands {
		if dockerResource, found := allDockerResources[serviceUuid]; found {
			userServiceDockerResources[serviceUuid] = dockerResource
		}
		// if no container is found here, we just don't add it to the map. runExecOperationInParallel will create
		// the error to the map downstream
	}

	var userServiceDockerResource *shared_helpers.UserServiceDockerResources
	var commandArg []string
	count := 1
	for serviceUuid, cmd := range userServiceCommands {
		commandArg = cmd
		resource, found := userServiceDockerResources[serviceUuid]
		userServiceDockerResource = resource
		if found && count == 1 {
			break
		}

		// if no container is found here, we just don't add it to the map. runExecOperationInParallel will create
		// the error to the map downstream
	}

	userServiceDockerContainer := userServiceDockerResource.ServiceContainer

	execOutputLinesChan, finalExecChan := dockerManager.RunExecCommandWithStreamedOutput(
		ctx,
		userServiceDockerContainer.GetId(),
		commandArg)
	return execOutputLinesChan, finalExecChan, nil
}

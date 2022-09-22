package user_service_functions

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

// TODO Switch these to streaming so that huge command outputs don't blow up the API container memory
// NOTE: This function will block while the exec is ongoing; if we need more perf we can make it async
func RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceGUID]*exec_result.ExecResult,
	map[service.ServiceGUID]error,
	error,
) {
	userServiceGuids := map[service.ServiceGUID]bool{}
	for userServiceGuid := range userServiceCommands {
		userServiceGuids[userServiceGuid] = true
	}

	filters := &service.ServiceFilters{
		GUIDs: userServiceGuids,
	}
	_, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	// TODO Parallelize to increase perf
	succesfulUserServiceExecResults := map[service.ServiceGUID]*exec_result.ExecResult{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}
	for guid, commandArgs := range userServiceCommands {
		dockerResources, found := allDockerResources[guid]
		if !found {
			erroredUserServiceGuids[guid] = stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because no Docker resources were found for it",
				commandArgs,
				guid,
			)
			continue
		}
		container := dockerResources.ServiceContainer

		execOutputBuf := &bytes.Buffer{}
		exitCode, err := dockerManager.RunExecCommand(
			ctx,
			container.GetId(),
			commandArgs,
			execOutputBuf,
		)
		if err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred executing command '%+v' on container '%v' for user service '%v'",
				commandArgs,
				container.GetName(),
				guid,
			)
			erroredUserServiceGuids[guid] = wrappedErr
			continue
		}
		newExecResult := exec_result.NewExecResult(exitCode, execOutputBuf.String())
		succesfulUserServiceExecResults[guid] = newExecResult
	}

	return succesfulUserServiceExecResults, erroredUserServiceGuids, nil
}
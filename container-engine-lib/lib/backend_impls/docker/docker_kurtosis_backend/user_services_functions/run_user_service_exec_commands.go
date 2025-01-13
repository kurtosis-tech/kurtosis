package user_service_functions

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/stacktrace"
	"reflect"
)

// TODO Switch these to streaming so that huge command outputs don't blow up the API container memory
func RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	containerUser string,
	userServiceCommands map[service.ServiceUUID][]string,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceUUID]*exec_result.ExecResult,
	map[service.ServiceUUID]error,
	error,
) {
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

	successfulExecs, failedExecs, err := runExecOperationsInParallel(ctx, containerUser, userServiceCommands, userServiceDockerResources, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An unexpected error occurred running the exec commands in parallel")
	}
	return successfulExecs, failedExecs, nil
}

func runExecOperationsInParallel(
	ctx context.Context,
	containerUser string,
	commandArgs map[service.ServiceUUID][]string,
	userServiceDockerResources map[service.ServiceUUID]*shared_helpers.UserServiceDockerResources,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceUUID]*exec_result.ExecResult,
	map[service.ServiceUUID]error,
	error,
) {
	successfulExecs := map[service.ServiceUUID]*exec_result.ExecResult{}
	failedExecs := map[service.ServiceUUID]error{}

	execOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for serviceUuid, commandArg := range commandArgs {
		userServiceDockerResource, found := userServiceDockerResources[serviceUuid]
		if !found {
			failedExecs[serviceUuid] = stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because no Docker resources were found for it",
				commandArgs,
				serviceUuid,
			)
			continue
		}

		execOperationId := operation_parallelizer.OperationID(serviceUuid)
		execOperation := createExecOperation(ctx, serviceUuid, containerUser, userServiceDockerResource, commandArg, dockerManager)
		execOperations[execOperationId] = execOperation
	}

	successfulOperations, failedOperations := operation_parallelizer.RunOperationsInParallel(execOperations)
	for operationUuid, operationResult := range successfulOperations {
		serviceUuid := service.ServiceUUID(operationUuid)
		execResult, ok := operationResult.(*exec_result.ExecResult)
		if !ok {
			// There's no way currently for one execResult object to be of a different type, and the others being
			// proper ExecResult. So, here we fail hard and do not continue other execResult are correct.
			// This is an internal bug anyway that would point to a pretty bad issue, so failing hard is also fine in
			// this sense.
			return nil, nil, stacktrace.NewError("An error occurred processing the result of the exec command "+
				"run on service '%s'. It seems the result object is of an unexpected type ('%v'). This is a Kurtosis "+
				"internal bug.", serviceUuid, reflect.TypeOf(execResult))
		}
		successfulExecs[serviceUuid] = execResult
	}
	for operationId, err := range failedOperations {
		serviceUuid := service.ServiceUUID(operationId)
		failedExecs[serviceUuid] = err
	}
	return successfulExecs, failedExecs, nil
}

func createExecOperation(
	ctx context.Context,
	serviceUuid service.ServiceUUID,
	containerUser string,
	userServiceDockerResource *shared_helpers.UserServiceDockerResources,
	commandArg []string,
	dockerManager *docker_manager.DockerManager,
) operation_parallelizer.Operation {
	return func() (interface{}, error) {
		execOutputBuf := &bytes.Buffer{}
		userServiceDockerContainer := userServiceDockerResource.ServiceContainer

		exitCode, err := dockerManager.RunExecCommandAsUser(
			ctx,
			userServiceDockerContainer.GetId(),
			containerUser,
			commandArg,
			execOutputBuf,
		)

		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred executing command '%+v' on container '%v' for user service '%v'",
				commandArg,
				userServiceDockerContainer.GetName(),
				serviceUuid,
			)
		}
		return exec_result.NewExecResult(exitCode, execOutputBuf.String()), nil
	}
}

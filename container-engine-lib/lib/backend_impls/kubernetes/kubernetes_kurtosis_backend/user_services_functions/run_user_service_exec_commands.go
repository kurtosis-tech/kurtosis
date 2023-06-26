package user_services_functions

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/stacktrace"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

// TODO Switch these to streaming methods, so that huge command outputs don't blow up the memory of the API container
func RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	userServiceCommands map[service.ServiceUUID][]string,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	succesfulUserServiceExecResults map[service.ServiceUUID]*exec_result.ExecResult,
	erroredUserServiceGuids map[service.ServiceUUID]error,
	resultErr error,
) {
	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	requestedGuids := map[service.ServiceUUID]bool{}
	for guid := range userServiceCommands {
		requestedGuids[guid] = true
	}
	matchingServicesFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    requestedGuids,
		Statuses: nil,
	}
	matchingObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, matchingServicesFilters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching the requested UUIDs: %+v", requestedGuids)
	}

	successfulExecs, failedExecs, err := runExecOperationsInParallel(namespaceName, userServiceCommands, matchingObjectsAndResources, kubernetesManager, ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An unexpected error occurred running the exec commands in parallel")
	}
	return successfulExecs, failedExecs, nil
}

func runExecOperationsInParallel(namespaceName string, commandArgs map[service.ServiceUUID][]string, userServiceKubernetesResources map[service.ServiceUUID]*shared_helpers.UserServiceObjectsAndKubernetesResources, kubernetesManager *kubernetes_manager.KubernetesManager, ctx context.Context) (map[service.ServiceUUID]*exec_result.ExecResult, map[service.ServiceUUID]error, error) {
	successfulExecs := map[service.ServiceUUID]*exec_result.ExecResult{}
	failedExecs := map[service.ServiceUUID]error{}

	execOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for serviceUuid, commandArg := range commandArgs {
		userServiceKubernetesResource, found := userServiceKubernetesResources[serviceUuid]
		if !found {
			failedExecs[serviceUuid] = stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because no Kubernetes resources were found for it",
				commandArgs,
				serviceUuid,
			)
			continue
		}

		userServiceKubernetesService := userServiceKubernetesResource.Service
		if userServiceKubernetesService == nil {
			failedExecs[serviceUuid] = stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because the service is not started yet",
				commandArgs,
				serviceUuid,
			)
			continue
		}
		if userServiceKubernetesService.GetStatus() != container_status.ContainerStatus_Running {
			failedExecs[serviceUuid] = stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because the service status is '%v'",
				commandArgs,
				serviceUuid,
				userServiceKubernetesService.GetStatus().String(),
			)
			continue
		}

		userServiceKubernetesPod := userServiceKubernetesResource.KubernetesResources.Pod

		execOperationId := operation_parallelizer.OperationID(serviceUuid)
		execOperation := createExecOperation(namespaceName, serviceUuid, userServiceKubernetesPod, commandArg, kubernetesManager, ctx)
		execOperations[execOperationId] = execOperation
	}

	successfulOperations, failedOperations := operation_parallelizer.RunOperationsInParallel(execOperations)
	for operationUuid, operationResult := range successfulOperations {
		serviceUuid := service.ServiceUUID(operationUuid)
		execResult, ok := operationResult.(*exec_result.ExecResult)
		if !ok {
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

func createExecOperation(namespaceName string, serviceUuid service.ServiceUUID, servicePod *v1.Pod, commandArg []string, kubernetesManager *kubernetes_manager.KubernetesManager, ctx context.Context) operation_parallelizer.Operation {
	return func() (interface{}, error) {
		outputBuffer := &bytes.Buffer{}
		concurrentBuffer := concurrent_writer.NewConcurrentWriter(outputBuffer)
		exitCode, err := kubernetesManager.RunExecCommandWithContext(
			ctx,
			namespaceName,
			servicePod.Name,
			userServiceContainerName,
			commandArg,
			concurrentBuffer,
			concurrentBuffer,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to execute command '%+v' in user "+
				"service container '%v' in Kubernetes pod '%v' for Kurtosis service with guid '%v', instead a "+
				"non-nil error was returned",
				commandArg,
				userServiceContainerName,
				servicePod.Name,
				serviceUuid,
			)
		}
		return exec_result.NewExecResult(exitCode, outputBuffer.String()), nil
	}
}

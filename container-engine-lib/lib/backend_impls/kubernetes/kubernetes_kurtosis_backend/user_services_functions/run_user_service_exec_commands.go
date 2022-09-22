package user_services_functions

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/stacktrace"
)

// TODO Switch these to streaming methods, so that huge command outputs don't blow up the memory of the API container
func RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	requestedGuids := map[service.ServiceGUID]bool{}
	for guid := range userServiceCommands {
		requestedGuids[guid] = true
	}
	matchingServicesFilters := &service.ServiceFilters{
		GUIDs: requestedGuids,
	}
	matchingObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, matchingServicesFilters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching the requested GUIDs: %+v", requestedGuids)
	}

	for guid, commandArgs := range userServiceCommands {
		objectsAndResources, found := matchingObjectsAndResources[guid]
		if !found {
			return nil, nil, stacktrace.NewError(
				"Requested to execute command '%+v' on service '%v', but the service does not exist",
				commandArgs,
				guid,
			)
		}
		serviceObj := objectsAndResources.Service
		if serviceObj == nil {
			return nil, nil, stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because the service is not started yet",
				commandArgs,
				guid,
			)
		}
		if serviceObj.GetStatus() != container_status.ContainerStatus_Running {
			return nil, nil, stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because the service status is '%v'",
				commandArgs,
				guid,
				serviceObj.GetStatus().String(),
			)
		}
	}

	// TODO Parallelize for perf
	userServiceExecSuccess := map[service.ServiceGUID]*exec_result.ExecResult{}
	userServiceExecErr := map[service.ServiceGUID]error{}
	for serviceGuid, serviceCommand := range userServiceCommands {
		userServiceObjectAndResources, found := matchingObjectsAndResources[serviceGuid]
		if !found {
			// Should never happen because we validate that the object exists earlier
			return nil, nil, stacktrace.NewError("Validated that service '%v' has Kubernetes resources, but couldn't find them when we need to run the exec", serviceGuid)
		}
		// Don't need to validate that this is non-nil because we did so before we started executing
		userServicePod := userServiceObjectAndResources.KubernetesResources.Pod
		userServicePodName := userServicePod.Name

		outputBuffer := &bytes.Buffer{}
		concurrentBuffer := concurrent_writer.NewConcurrentWriter(outputBuffer)
		exitCode, err := kubernetesManager.RunExecCommand(
			namespaceName,
			userServicePodName,
			userServiceContainerName,
			serviceCommand,
			concurrentBuffer,
			concurrentBuffer,
		)
		if err != nil {
			userServiceExecErr[serviceGuid] = stacktrace.Propagate(
				err,
				"Expected to be able to execute command '%+v' in user service container '%v' in Kubernetes pod '%v' "+
					"for Kurtosis service with guid '%v', instead a non-nil error was returned",
				serviceCommand,
				userServiceContainerName,
				userServicePodName,
				serviceGuid,
			)
			continue
		}
		userServiceExecSuccess[serviceGuid] = exec_result.NewExecResult(exitCode, outputBuffer.String())
	}
	return userServiceExecSuccess, userServiceExecErr, nil
}
package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func RunUserServiceExecCommandWithStreamedOutput(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	userServiceCommands map[service.ServiceUUID][]string,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (chan string, chan *exec_result.ExecResult, error) {
	execOutputChan := make(chan string)
	finalExecResultChan := make(chan *exec_result.ExecResult)
	go func() {
		defer func() {
			close(execOutputChan)
			close(finalExecResultChan)
		}()

		// only process 1 exec command for now
		if len(userServiceCommands) > 1 {
			sendErrorAndFail(execOutputChan, stacktrace.NewError("Can only stream one exec function at a time."), "An error occurred streaming exec output")
			return
		}

		namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
		if err != nil {
			sendErrorAndFail(execOutputChan, err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
			return
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
			sendErrorAndFail(execOutputChan, err, "An error occurred getting user services matching the requested UUIDs: %v", requestedGuids)
			return
		}

		var userServiceKubernetesResource *shared_helpers.UserServiceObjectsAndKubernetesResources
		var serviceUuid service.ServiceUUID
		var commandArg []string
		count := 1
		for id, cmd := range userServiceCommands {
			serviceUuid = id
			commandArg = cmd
			resource, found := matchingObjectsAndResources[serviceUuid]
			if !found {
				sendErrorAndFail(
					execOutputChan,
					stacktrace.NewError(
						"Cannot execute command '%+v' on service '%v' because no Kubernetes resources were found for it",
						commandArg,
						serviceUuid),
					"An error was found while running exec with streamed output over kubernetes")
				return
			}
			userServiceKubernetesResource = resource
			if found && count == 1 {
				break
			} else {

			}
		}

		userServiceKubernetesService := userServiceKubernetesResource.Service
		if userServiceKubernetesService == nil {
			sendErrorAndFail(
				execOutputChan,
				err,
				"An error was found while running exec with streamed output over kubernetes for service '%s' and command '%v'.",
				commandArg,
				serviceUuid)
			return
		}
		if userServiceKubernetesService.GetStatus() != container_status.ContainerStatus_Running {
			sendErrorAndFail(execOutputChan, stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because the service status is '%v'",
				commandArg,
				serviceUuid,
				userServiceKubernetesService.GetStatus().String()),
				"An error was found while running exec with streamed output over kubernetes")
			return
		}

		userServiceKubernetesPod := userServiceKubernetesResource.KubernetesResources.Pod

		execOutputLinesChan, finalResultChan := kubernetesManager.RunExecCommandWithStreamedOutput(
			ctx,
			namespaceName,
			userServiceKubernetesPod.Name,
			userServiceContainerName,
			commandArg)
		for execOutputLine := range execOutputLinesChan {
			execOutputChan <- execOutputLine
		}
		for execResult := range finalResultChan {
			finalExecResultChan <- execResult
		}
	}()
	return execOutputChan, finalExecResultChan, nil
}

func sendErrorAndFail(destChan chan<- string, err error, msg string, msgArgs ...interface{}) {
	propagatedErr := stacktrace.Propagate(err, msg, msgArgs...)
	destChan <- propagatedErr.Error()
}

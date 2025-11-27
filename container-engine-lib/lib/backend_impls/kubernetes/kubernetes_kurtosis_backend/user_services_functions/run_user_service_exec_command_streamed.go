package user_services_functions

import (
	"context"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func RunUserServiceExecCommandWithStreamedOutput(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	cmd []string,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (chan string, chan *exec_result.ExecResult, error) {
	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	requestedGuids := map[service.ServiceUUID]bool{}
	requestedGuids[serviceUuid] = true
	matchingServicesFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    requestedGuids,
		Statuses: nil,
	}
	matchingObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, matchingServicesFilters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching the requested UUIDs: %v", requestedGuids)
	}

	var userServiceKubernetesResource *shared_helpers.UserServiceObjectsAndKubernetesResources
	if resource, found := matchingObjectsAndResources[serviceUuid]; found {
		userServiceKubernetesResource = resource
	} else {
		return nil, nil, stacktrace.NewError(
			"Cannot execute command '%+v' on service '%v' because no Kubernetes resources were found for it",
			cmd,
			serviceUuid)
	}

	userServiceKubernetesService := userServiceKubernetesResource.Service
	if userServiceKubernetesService == nil {
		return nil, nil, stacktrace.Propagate(err, "An error was found while running exec with streamed output over kubernetes for service '%s' and command '%v'.",
			cmd,
			serviceUuid)
	}
	if userServiceKubernetesService.GetContainer().GetStatus() != container.ContainerStatus_Running {
		return nil, nil, stacktrace.NewError(
			"Cannot execute command '%+v' on service '%v' because the service status is '%v'",
			cmd,
			serviceUuid,
			userServiceKubernetesService.GetContainer().GetStatus().String())
	}

	userServiceKubernetesPod := userServiceKubernetesResource.KubernetesResources.Pod

	execOutputLinesChan, finalResultChan, err := kubernetesManager.RunExecCommandWithStreamedOutput(
		ctx,
		namespaceName,
		userServiceKubernetesPod.Name,
		userServiceContainerName,
		cmd)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred attempting to stream exec output from docker.")
	}
	return execOutputLinesChan, finalResultChan, nil
}

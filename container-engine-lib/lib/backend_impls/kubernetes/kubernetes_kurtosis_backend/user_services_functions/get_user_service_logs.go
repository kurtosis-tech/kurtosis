package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
)

const (
	shouldAddTimestampsToUserServiceLogs = false
)

func GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (successfulUserServiceLogs map[service.ServiceGUID]io.ReadCloser, erroredUserServiceGuids map[service.ServiceGUID]error, resultError error) {
	serviceObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Expected to be able to get user services and Kubernetes resources, instead a non nil error was returned")
	}
	userServiceLogs := map[service.ServiceGUID]io.ReadCloser{}
	erredServiceLogs := map[service.ServiceGUID]error{}
	shouldCloseLogStreams := true
	for _, serviceObjectAndResource := range serviceObjectsAndResources {
		serviceGuid := serviceObjectAndResource.Service.GetRegistration().GetGUID()
		servicePod := serviceObjectAndResource.KubernetesResources.Pod
		if servicePod == nil {
			erredServiceLogs[serviceGuid] = stacktrace.NewError("Expected to find a pod for Kurtosis service with GUID '%v', instead no pod was found", serviceGuid)
			continue
		}
		serviceNamespaceName := serviceObjectAndResource.KubernetesResources.Service.GetNamespace()
		// Get logs
		logReadCloser, err := kubernetesManager.GetContainerLogs(ctx, serviceNamespaceName, servicePod.Name, userServiceContainerName, shouldFollowLogs, shouldAddTimestampsToUserServiceLogs)
		if err != nil {
			erredServiceLogs[serviceGuid] = stacktrace.Propagate(err, "Expected to be able to call Kubernetes to get logs for service with GUID '%v', instead a non-nil error was returned", serviceGuid)
			continue
		}
		defer func() {
			if shouldCloseLogStreams {
				logReadCloser.Close()
			}
		}()

		userServiceLogs[serviceGuid] = logReadCloser
	}

	shouldCloseLogStreams = false
	return userServiceLogs, erredServiceLogs, nil
}
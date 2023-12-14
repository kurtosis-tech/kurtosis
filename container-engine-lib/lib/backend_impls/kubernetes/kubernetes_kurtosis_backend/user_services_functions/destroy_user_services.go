package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func DestroyUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager) (resultSuccessfulGuids map[service.ServiceUUID]bool, resultErroredGuids map[service.ServiceUUID]error, resultErr error) {
	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	allObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred getting user services in enclave '%v' matching filters: %+v",
			enclaveId,
			filters,
		)
	}

	successfulGuids := map[service.ServiceUUID]bool{}
	erroredGuids := map[service.ServiceUUID]error{}
	for serviceUuid, serviceObjsAndResources := range allObjectsAndResources {
		resources := serviceObjsAndResources.KubernetesResources

		serviceToRemove := resources.Service
		if serviceToRemove != nil {
			if err := kubernetesManager.RemoveService(ctx, serviceToRemove); err != nil {
				erroredGuids[serviceUuid] = stacktrace.Propagate(
					err,
					"An error occurred removing Kubernetes service '%v' in namespace '%v'",
					serviceToRemove.Name,
					namespaceName,
				)
				continue
			}
		}
		podToRemove := resources.Pod
		if podToRemove != nil {
			if err := kubernetesManager.RemovePod(ctx, podToRemove); err != nil {
				erroredGuids[serviceUuid] = stacktrace.Propagate(
					err,
					"An error occurred removing Kubernetes pod '%v' in namespace '%v'",
					podToRemove.Name,
					namespaceName,
				)
				continue
			}
		}
		ingressToRemove := resources.Ingress
		if ingressToRemove != nil {
			if err := kubernetesManager.RemoveIngress(ctx, ingressToRemove); err != nil {
				erroredGuids[serviceUuid] = stacktrace.Propagate(
					err,
					"An error occurred removing Kubernetes ingress '%v' in namespace '%v'",
					ingressToRemove.Name,
					namespaceName,
				)
				continue
			}
		}
	}
	return successfulGuids, erroredGuids, nil
}

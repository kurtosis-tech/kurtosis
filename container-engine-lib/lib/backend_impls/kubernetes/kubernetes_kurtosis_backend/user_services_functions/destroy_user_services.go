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
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager) (resultSuccessfulGuids map[service.ServiceGUID]bool, resultErroredGuids map[service.ServiceGUID]error, resultErr error) {
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

	successfulGuids := map[service.ServiceGUID]bool{}
	erroredGuids := map[service.ServiceGUID]error{}
	for serviceGuid, serviceObjsAndResources := range allObjectsAndResources {
		resources := serviceObjsAndResources.KubernetesResources

		pod := resources.Pod
		if pod != nil {
			if err := kubernetesManager.RemovePod(ctx, pod); err != nil {
				erroredGuids[serviceGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing Kubernetes pod '%v' in namespace '%v'",
					pod.Name,
					namespaceName,
				)
				continue
			}
		}
	}
	return successfulGuids, erroredGuids, nil
}
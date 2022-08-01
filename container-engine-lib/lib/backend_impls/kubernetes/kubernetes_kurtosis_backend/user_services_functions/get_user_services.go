package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_functions"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

func GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	cliModeArgs *shared_functions.CliModeArgs,
	apiContainerModeArgs *shared_functions.ApiContainerModeArgs,
	engineServerModeArgs *shared_functions.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (successfulUserServices map[service.ServiceGUID]*service.Service, resultError error) {
	allObjectsAndResources, err := shared_functions.GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting user services in enclave '%v' matching filters: %+v",
			enclaveId,
			filters,
		)
	}
	result := map[service.ServiceGUID]*service.Service{}
	for guid, serviceObjsAndResources := range allObjectsAndResources {
		serviceObj := serviceObjsAndResources.Service
		if serviceObj == nil {
			// Indicates a registration-only service; skip
			continue
		}
		result[guid] = serviceObj
	}
	return result, nil
}
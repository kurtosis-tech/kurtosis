package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (successfulUserServices map[service.ServiceUUID]*service.Service, resultError error) {
	logrus.Debugf(
		"Getting user services in enclave '%v' matching filters: %+v",
		enclaveId,
		filters,
	)
	allObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	logrus.Debugf(
		"All found: %+v",
		allObjectsAndResources,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting user services in enclave '%v' matching filters: %+v",
			enclaveId,
			filters,
		)
	}
	result := map[service.ServiceUUID]*service.Service{}
	for guid, serviceObjsAndResources := range allObjectsAndResources {
		serviceObj := serviceObjsAndResources.Service
		if serviceObj == nil {
			logrus.Debugf(
				"Service object not found for guid, assuming reg-only service '%v'",
				guid,
			)
			// Indicates a registration-only service; skip
			continue
		}
		result[guid] = serviceObj
	}
	return result, nil
}

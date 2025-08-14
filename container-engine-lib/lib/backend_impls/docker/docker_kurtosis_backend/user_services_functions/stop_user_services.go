package user_service_functions

import (
	"context"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/service_registration"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	stopContainerTimeout = 10 * time.Second
)

func StopUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	dockerManager *docker_manager.DockerManager,
) (
	resultSuccessfulServiceUUIDs map[service.ServiceUUID]bool,
	resultErroredServiceUUIDs map[service.ServiceUUID]error,
	resultErr error,
) {
	allServiceObjs, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveUuid, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	servicesToStopByContainerId := map[string]*service.Service{}
	for uuid, serviceResources := range allDockerResources {
		serviceObj, found := allServiceObjs[uuid]
		if !found {
			// Should never happen; there should be a 1:1 mapping between service_objects:docker_resources by GUID
			return nil, nil, stacktrace.NewError("No service object found for service '%v' that had Docker resources", uuid)
		}
		servicesToStopByContainerId[serviceResources.ServiceContainer.GetId()] = serviceObj
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.StopContainer(ctx, dockerObjectId, stopContainerTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping user service container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	serviceRegistrationsByServiceUuid, err := serviceRegistrationRepository.GetAllEnclaveServiceRegistrations(enclaveUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting all enclave service registrations from the repository for enclave with UUID '%v'", enclaveUuid)
	}

	successfulUuidStrs, erroredUuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		servicesToStopByContainerId,
		dockerManager,
		extractServiceUUIDFromService,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred killing user service containers matching filters '%+v'", filters)
	}

	successfulUuids := map[service.ServiceUUID]bool{}
	for uuidStr := range successfulUuidStrs {
		serviceUuid := service.ServiceUUID(uuidStr)
		successfulUuids[serviceUuid] = true
		serviceRegistration, found := serviceRegistrationsByServiceUuid[serviceUuid]
		if !found {
			erroredUuidStrs[uuidStr] = stacktrace.NewError("Expected to find service registration by service UUID '%v' in map '%+v', but none was found; this is a bug in Kurtosis", serviceUuid, serviceRegistrationsByServiceUuid)
			delete(successfulUuidStrs, uuidStr)
			continue
		}
		serviceName := serviceRegistration.GetName()
		serviceStatus := service.ServiceStatus_Stopped
		if err := serviceRegistrationRepository.UpdateStatus(serviceName, serviceStatus); err != nil {
			erroredUuidStrs[uuidStr] = stacktrace.Propagate(err, "An error occurred while updating service status to '%s' in service registration for service '%s'", serviceStatus, serviceName)
			delete(successfulUuidStrs, uuidStr)
			continue
		}
	}

	erroredUuids := map[service.ServiceUUID]error{}
	for uuidStr, err := range erroredUuidStrs {
		erroredUuids[service.ServiceUUID(uuidStr)] = stacktrace.Propagate(
			err,
			"An error occurred stopping service '%v'",
			uuidStr,
		)
	}

	return successfulUuids, erroredUuids, nil
}

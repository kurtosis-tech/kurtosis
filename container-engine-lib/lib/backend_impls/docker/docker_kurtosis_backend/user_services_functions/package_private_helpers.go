package user_service_functions

import (
	"context"
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db/service_registration"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

func destroyUserServicesUnlocked(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	enclaveFreeIpProvidersForEnclave *free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
) (
	resultSuccessfulUuids map[service.ServiceUUID]bool,
	resultErroredUuids map[service.ServiceUUID]error,
	resultErr error,
) {
	// We filter registrations here because a registration is the canonical resource for a Kurtosis user service - no registration,
	// no Kurtosis service - and not all registrations will have Docker resources
	matchingRegistrations := map[service.ServiceUUID]*service.ServiceRegistration{}

	serviceRegistrationsForEnclave, err := serviceRegistrationRepository.GetAllEnclaveServiceRegistrations(enclaveUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting all enclave service registrations from the repository for enclave with UUID '%v'", enclaveUuid)
	}

	for uuid, registration := range serviceRegistrationsForEnclave {
		if len(filters.UUIDs) > 0 {
			if _, found := filters.UUIDs[registration.GetUUID()]; !found {
				continue
			}
		}

		if len(filters.Names) > 0 {
			if _, found := filters.Names[registration.GetName()]; !found {
				continue
			}
		}

		matchingRegistrations[uuid] = registration
	}

	// NOTE: This may end up with less results here than we have registrations, if the user registered but did not start a service,
	// though we should never end up with _more_ Docker resources
	allServiceObjs, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveUuid, filters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	if len(allServiceObjs) > len(matchingRegistrations) || len(allDockerResources) > len(matchingRegistrations) {
		return nil, nil, stacktrace.NewError("Found more Docker resources matching the following filters than service registrations; this is a bug in Kurtosis: %+v", filters)
	}

	// TODO Refactor getMatchingUserServiceObjsAndDockerResourcesNoMutex to return a single package object so we don't need to do this
	// Before deleting anything, verify we have the same keys in both maps
	if len(allServiceObjs) != len(allDockerResources) {
		return nil, nil, stacktrace.NewError(
			"Expected the number of services to remove (%v), and the number of services for which we have Docker resources for (%v), to be the same but were different!",
			len(allServiceObjs),
			len(allDockerResources),
		)
	}
	for serviceUuid := range allServiceObjs {
		if _, found := allDockerResources[serviceUuid]; !found {
			return nil, nil, stacktrace.NewError(
				"Have service object to remove '%v', which doesn't have corresponding Docker resources; this is a "+
					"bug in Kurtosis",
				serviceUuid,
			)
		}
	}

	registrationsToDeregister := map[service.ServiceUUID]*service.ServiceRegistration{}

	// Find the registrations which don't have any Docker resources and immediately add them to the list of stuff to deregister
	for uuid, registration := range matchingRegistrations {
		if _, doesRegistrationHaveResources := allDockerResources[uuid]; doesRegistrationHaveResources {
			// We'll deregister registrations-with-resources if and only if we can successfully remove their resources
			continue
		}

		// If the status filter is specified, don't deregister any registrations-without-resources
		if len(filters.Statuses) > 0 {
			continue
		}

		registrationsToDeregister[uuid] = registration
	}

	// Now try removing all the registrations-with-resources
	successfulResourceRemovalUuids, erroredResourceRemovalUuids, err := removeUserServiceDockerResources(
		ctx,
		allServiceObjs,
		allDockerResources,
		dockerManager,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred trying to remove user service Docker resources",
		)
	}

	erroredUuids := map[service.ServiceUUID]error{}
	for uuid, err := range erroredResourceRemovalUuids {
		erroredUuids[uuid] = stacktrace.Propagate(
			err,
			"An error occurred destroying Docker resources for service '%v'",
			uuid,
		)
	}

	successfulUuids := map[service.ServiceUUID]bool{}
	for uuid := range successfulResourceRemovalUuids {
		registrationsToDeregister[uuid] = matchingRegistrations[uuid]
		successfulUuids[uuid] = true
	}

	// Finalize deregistration
	for uuid, registration := range registrationsToDeregister {
		ipAddr := registration.GetPrivateIP()
		if err = enclaveFreeIpProvidersForEnclave.ReleaseIpAddr(ipAddr); err != nil {
			logrus.Errorf("Error releasing IP address '%v'", ipAddr)
		}

		serviceRegistration, found := serviceRegistrationsForEnclave[uuid]
		if !found {
			erroredUuids[uuid] = stacktrace.NewError("Failed to get service registration for service UUID '%v'. This should never happen. This is a Kurtosis bug.", uuid)
			delete(successfulUuids, uuid)
			continue
		}

		serviceName := serviceRegistration.GetName()
		if err := serviceRegistrationRepository.Delete(serviceName); err != nil {
			erroredUuids[uuid] = stacktrace.Propagate(err, "An error occurred deleting the service registration for service '%v' from the repository", serviceName)
			delete(successfulUuids, uuid)
			continue
		}
	}

	return successfulUuids, erroredUuids, nil
}

/*
Destroys the Docker containers and makes a best-effort attempt to destroy the volumes

NOTE: We make a best-effort attempt to delete files artifact expansion volumes for a service, but there's still the
possibility that some will get leaked! There's unfortunately no way around this though because:

 1. We delete the canonical resource (the user container) before the volumes out of necessity
    since Docker won't let us delete volumes unless no containers are using them
 2. We can't undo container deletion
 3. Normally whichever resource is created first (volumes) is the one we'd use as the canonical resource,
    but we can't do that since we're not guaranteed to have volumes (because a service may not have requested any
    files artifacts).

Therefore, we just make a best-effort attempt to clean up the volumes and leak the rest, though it's not THAT
big of a deal since they'll be deleted when the enclave gets deleted.
*/
func removeUserServiceDockerResources(
	ctx context.Context,
	serviceObjectsToRemove map[service.ServiceUUID]*service.Service,
	resourcesToRemove map[service.ServiceUUID]*shared_helpers.UserServiceDockerResources,
	dockerManager *docker_manager.DockerManager,
) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {

	erroredUuids := map[service.ServiceUUID]error{}

	// Before deleting anything, verify we have the same keys in both maps
	if len(serviceObjectsToRemove) != len(resourcesToRemove) {
		return nil, nil, stacktrace.NewError(
			"Expected the number of services to remove (%v), and the number of services for which we have Docker resources for (%v), to be the same but were different!",
			len(serviceObjectsToRemove),
			len(resourcesToRemove),
		)
	}
	for serviceUuid := range serviceObjectsToRemove {
		if _, found := resourcesToRemove[serviceUuid]; !found {
			return nil, nil, stacktrace.NewError(
				"Have service object to remove '%v', which doesn't have corresponding Docker resources; this is a "+
					"bug in Kurtosis",
				serviceUuid,
			)
		}
	}

	kurtosisObjectsToRemoveByContainerId := map[string]*service.Service{}
	for serviceUuid, resources := range resourcesToRemove {
		// Safe to skip the is-found check because we verified the map keys are identical earlier
		serviceObj := serviceObjectsToRemove[serviceUuid]

		containerId := resources.ServiceContainer.GetId()
		kurtosisObjectsToRemoveByContainerId[containerId] = serviceObj
	}

	// TODO Simplify this with Go generics
	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing user service container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulContainerRemoveUuidStrs, erroredContainerRemoveUuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		kurtosisObjectsToRemoveByContainerId,
		dockerManager,
		extractServiceUUIDFromService,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing user service containers in parallel")
	}

	for uuidStr, err := range erroredContainerRemoveUuidStrs {
		erroredUuids[service.ServiceUUID(uuidStr)] = stacktrace.Propagate(
			err,
			"An error occurred destroying container for service '%v'",
			uuidStr,
		)
	}

	// TODO Parallelize if we need more perf (but we shouldn't, since removing volumes way faster than containers)
	successfulVolumeRemovalUuids := map[service.ServiceUUID]bool{}
	for serviceUuidStr := range successfulContainerRemoveUuidStrs {
		serviceUuid := service.ServiceUUID(serviceUuidStr)

		// Safe to skip the is-found check because we verified that the maps have the same keys earlier
		resources := resourcesToRemove[serviceUuid]

		failedVolumeErrStrs := []string{}
		for _, volumeName := range resources.ExpanderVolumeNames {
			/*
				We try to delete as many volumes as we can here, rather than ejecting on failure, because any volumes not
				deleted here will be leaked! There's unfortunately no way around this though because:

				 1) we've already deleted the canonical resource (the user container) out of necessity
				    since Docker won't let us delete volumes unless no containers are using them
				 2) we can't undo container deletion
				 3) normally whichever resource is created first (volumes) is the one we'd use as the canonical resource,
				    but we can't do that since we're not guaranteed to have volumes

				Therefore, we just make a best-effort attempt to clean up the volumes and leak the rest :(
			*/
			if err := dockerManager.RemoveVolume(ctx, volumeName); err != nil {
				errStrBuilder := strings.Builder{}
				errStrBuilder.WriteString(fmt.Sprintf(
					">>>>>>>>>>>>>>>>>> Removal error for volume %v <<<<<<<<<<<<<<<<<<<<<<<<<<<\n",
					volumeName,
				))
				errStrBuilder.WriteString(err.Error())
				errStrBuilder.WriteString("\n")
				errStrBuilder.WriteString(fmt.Sprintf(
					">>>>>>>>>>>>>>> End removal error for volume %v <<<<<<<<<<<<<<<<<<<<<<<<<<",
					volumeName,
				))
				failedVolumeErrStrs = append(failedVolumeErrStrs, errStrBuilder.String())
			}
		}

		if len(failedVolumeErrStrs) > 0 {
			erroredUuids[serviceUuid] = stacktrace.NewError(
				"Errors occurred removing volumes for service '%v'\n"+
					"ACTION REQUIRED: You will need to manually remove these volumes, else they will stay around until the enclave is destroyed!\n"+
					"%v",
				serviceUuid,
				strings.Join(failedVolumeErrStrs, "\n\n"),
			)
			continue
		}
		successfulVolumeRemovalUuids[serviceUuid] = true
	}

	successUuids := successfulVolumeRemovalUuids
	return successUuids, erroredUuids, nil
}

func extractServiceUUIDFromService(service *service.Service) string {
	return string(service.GetRegistration().GetUUID())
}

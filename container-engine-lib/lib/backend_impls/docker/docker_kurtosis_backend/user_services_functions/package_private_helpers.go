package user_service_functions

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/free_ip_addr_tracker"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

func destroyUserServicesUnlocked(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration,
	enclaveFreeIpProviders map[enclave.EnclaveID]*free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
) (
	resultSuccessfulGuids map[service.ServiceGUID]bool,
	resultErroredGuids map[service.ServiceGUID]error,
	resultErr error,
) {

	freeIpAddrTrackerForEnclave, found := enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Cannot destroy services in enclave '%v' because no free IP address tracker is registered for it; this likely "+
				"means that the destroy user services call is being made from somewhere it shouldn't be (i.e. outside the API contianer)",
			enclaveId,
		)
	}

	registrationsForEnclave, found := serviceRegistrations[enclaveId]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v', so we cannot get service registrations matching filters: %+v",
			enclaveId,
			filters,
		)
	}

	// We filter registrations here because a registration is the canonical resource for a Kurtosis user service - no registration,
	// no Kurtosis service - and not all registrations will have Docker resources
	matchingRegistrations := map[service.ServiceGUID]*service.ServiceRegistration{}
	for guid, registration := range registrationsForEnclave {
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[registration.GetGUID()]; !found {
				continue
			}
		}

		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[registration.GetID()]; !found {
				continue
			}
		}

		matchingRegistrations[guid] = registration
	}

	// NOTE: This may end up with less results here than we have registrations, if the user registered but did not start a service,
	// though we should never end up with _more_ Docker resources
	allServiceObjs, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, filters, dockerManager)
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
	for serviceGuid := range allServiceObjs {
		if _, found := allDockerResources[serviceGuid]; !found {
			return nil, nil, stacktrace.NewError(
				"Have service object to remove '%v', which doesn't have corresponding Docker resources; this is a "+
					"bug in Kurtosis",
				serviceGuid,
			)
		}
	}

	registrationsToDeregister := map[service.ServiceGUID]*service.ServiceRegistration{}

	// Find the registrations which don't have any Docker resources and immediately add them to the list of stuff to deregister
	for guid, registration := range matchingRegistrations {
		if _, doesRegistrationHaveResources := allDockerResources[guid]; doesRegistrationHaveResources {
			// We'll deregister registrations-with-resources if and only if we can successfully remove their resources
			continue
		}

		// If the status filter is specified, don't deregister any registrations-without-resources
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			continue
		}

		registrationsToDeregister[guid] = registration
	}

	// Now try removing all the registrations-with-resources
	successfulResourceRemovalGuids, erroredResourceRemovalGuids, err := removeUserServiceDockerResources(
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

	erroredGuids := map[service.ServiceGUID]error{}
	for guid, err := range erroredResourceRemovalGuids {
		erroredGuids[guid] = stacktrace.Propagate(
			err,
			"An error occurred destroying Docker resources for service '%v'",
			guid,
		)
	}

	successfulGuids := map[service.ServiceGUID]bool{}
	for guid := range successfulResourceRemovalGuids {
		registrationsToDeregister[guid] = matchingRegistrations[guid]
		successfulGuids[guid] = true
	}

	// Finalize deregistration
	for guid, registration := range registrationsToDeregister {
		ipAddr := registration.GetPrivateIP()
		if err = freeIpAddrTrackerForEnclave.ReleaseIpAddr(ipAddr); err != nil {
			logrus.Errorf("Error releasing IP address '%v'", ipAddr)
		}
		delete(registrationsForEnclave, guid)
	}

	return successfulGuids, erroredGuids, nil
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
	serviceObjectsToRemove map[service.ServiceGUID]*service.Service,
	resourcesToRemove map[service.ServiceGUID]*shared_helpers.UserServiceDockerResources,
	dockerManager *docker_manager.DockerManager,
) (map[service.ServiceGUID]bool, map[service.ServiceGUID]error, error) {

	erroredGuids := map[service.ServiceGUID]error{}

	// Before deleting anything, verify we have the same keys in both maps
	if len(serviceObjectsToRemove) != len(resourcesToRemove) {
		return nil, nil, stacktrace.NewError(
			"Expected the number of services to remove (%v), and the number of services for which we have Docker resources for (%v), to be the same but were different!",
			len(serviceObjectsToRemove),
			len(resourcesToRemove),
		)
	}
	for serviceGuid := range serviceObjectsToRemove {
		if _, found := resourcesToRemove[serviceGuid]; !found {
			return nil, nil, stacktrace.NewError(
				"Have service object to remove '%v', which doesn't have corresponding Docker resources; this is a "+
					"bug in Kurtosis",
				serviceGuid,
			)
		}
	}

	uncastedKurtosisObjectsToRemoveByContainerId := map[string]interface{}{}
	for serviceGuid, resources := range resourcesToRemove {
		// Safe to skip the is-found check because we verified the map keys are identical earlier
		serviceObj := serviceObjectsToRemove[serviceGuid]

		containerId := resources.ServiceContainer.GetId()
		uncastedKurtosisObjectsToRemoveByContainerId[containerId] = serviceObj
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

	successfulContainerRemoveGuidStrs, erroredContainerRemoveGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		uncastedKurtosisObjectsToRemoveByContainerId,
		dockerManager,
		extractServiceGUIDFromServiceObj,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing user service containers in parallel")
	}

	for guidStr, err := range erroredContainerRemoveGuidStrs {
		erroredGuids[service.ServiceGUID(guidStr)] = stacktrace.Propagate(
			err,
			"An error occurred destroying container for service '%v'",
			guidStr,
		)
	}

	// TODO Parallelize if we need more perf (but we shouldn't, since removing volumes way faster than containers)
	successfulVolumeRemovalGuids := map[service.ServiceGUID]bool{}
	for serviceGuidStr := range successfulContainerRemoveGuidStrs {
		serviceGuid := service.ServiceGUID(serviceGuidStr)

		// Safe to skip the is-found check because we verified that the maps have the same keys earlier
		resources := resourcesToRemove[serviceGuid]

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
			erroredGuids[serviceGuid] = stacktrace.NewError(
				"Errors occurred removing volumes for service '%v'\n"+
					"ACTION REQUIRED: You will need to manually remove these volumes, else they will stay around until the enclave is destroyed!\n"+
					"%v",
				serviceGuid,
				strings.Join(failedVolumeErrStrs, "\n\n"),
			)
			continue
		}
		successfulVolumeRemovalGuids[serviceGuid] = true
	}

	successGuids := successfulVolumeRemovalGuids
	return successGuids, erroredGuids, nil
}

func extractServiceGUIDFromServiceObj(uncastedObj interface{}) (string, error) {
	castedObj, ok := uncastedObj.(*service.Service)
	if !ok {
		return "", stacktrace.NewError("An error occurred downcasting the user service object")
	}
	return string(castedObj.GetRegistration().GetGUID()), nil
}

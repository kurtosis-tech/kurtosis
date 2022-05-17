package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
)

// Any of these values being nil indicates that the resource doesn't exist
type enclaveKubernetesResources struct {
	namespace *apiv1.Namespace

	// Pods are technically not resources that define an enclave, but we need them both
	//to StopEnclave and to return an EnclaveStatus
	pods []*apiv1.Pod
}


// ====================================================================================================
//                                     		Enclave CRUD Methods
// ====================================================================================================

func (backend *KubernetesKurtosisBackend) CreateEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	isPartitioningEnabled bool,
) (
	*enclave.Enclave,
	error,
) {
	if isPartitioningEnabled {
		return nil, stacktrace.NewError("Partitioning not supported for Kubernetes-backed Kurtosis.")
	}
	teardownContext := context.Background()

	searchNamespaceLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}
	namespaceList, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, searchNamespaceLabels)
	if err != nil {
		return nil, stacktrace.NewError("Failed to list namespaces from Kubernetes, so can not verify if enclave '%v' already exists.", enclaveId)
	}
	if len(namespaceList.Items) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because an enclave with ID '%v' already exists", enclaveId, enclaveId)
	}

	// Make Enclave attributes provider
	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to generate an object attributes provider for the enclave with ID '%v'", enclaveId)
	}

	enclaveNamespaceAttrs, err := enclaveObjAttrsProvider.ForEnclaveNamespace(isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave network attributes for the enclave with ID '%v'", enclaveId)
	}

	enclaveNamespaceName := enclaveNamespaceAttrs.GetName().GetString()
	enclaveNamespaceLabels := getStringMapFromLabelMap(enclaveNamespaceAttrs.GetLabels())

	enclaveNamespace, err := backend.kubernetesManager.CreateNamespace(ctx, enclaveNamespaceName, enclaveNamespaceLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%v' for enclave '%v'", enclaveNamespaceName, enclaveId)
	}
	shouldDeleteNamespace := true
	defer func() {
		if shouldDeleteNamespace {
			if err := backend.kubernetesManager.RemoveNamespace(teardownContext, enclaveNamespaceName); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete namespace '%v' that we created but an error was thrown:\n%v", enclaveNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove namespace with name '%v'!!!!!!!", enclaveNamespaceName)
			}
		}
	}()

	enclaveDataVolumeAttrs, err := enclaveObjAttrsProvider.ForEnclaveDataVolume()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave data volume attributes for the enclave with ID '%v'", enclaveId)
	}

	persistentVolumeClaimName := enclaveDataVolumeAttrs.GetName().GetString()
	persistentVolumeClaimLabels := getStringMapFromLabelMap(enclaveDataVolumeAttrs.GetLabels())

	foundVolumes, err := backend.kubernetesManager.GetPersistentVolumeClaimsByLabels(ctx, enclaveNamespaceName, persistentVolumeClaimLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave data volumes matching labels '%+v'", persistentVolumeClaimLabels)
	}
	if len(foundVolumes.Items) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because one or more enclave data volumes for that enclave already exists", enclaveId)
	}

	_, err = backend.kubernetesManager.CreatePersistentVolumeClaim(ctx,
		enclaveNamespaceName,
		persistentVolumeClaimName,
		persistentVolumeClaimLabels,
		backend.volumeSizePerEnclaveInMegabytes,
		backend.volumeStorageClassName)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"Failed to create persistent volume claim in enclave '%v' with name '%v' and storage class name '%v'",
			enclaveNamespaceName,
			persistentVolumeClaimName,
			backend.volumeStorageClassName)
	}
	shouldDeleteVolume := true
	defer func() {
		if shouldDeleteVolume {
			if err := backend.kubernetesManager.RemovePersistentVolumeClaim(teardownContext, enclaveNamespaceName, persistentVolumeClaimName); err != nil {
				logrus.Errorf(
					"Creating the enclave didn't complete successfully, so we tried to delete enclave persistent volume claim '%v' " +
						"that we created but an error was thrown:\n%v",
					persistentVolumeClaimName,
					err,
				)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove persistent volume claim with name '%v'!!!!!!!", persistentVolumeClaimName)
			}
		}
	}()

	enclaveResources := &enclaveKubernetesResources{
		namespace: enclaveNamespace,
		pods: []*apiv1.Pod{},
	}
	enclaveObjsById, err := getEnclaveObjectsFromKubernetesResources(map[enclave.EnclaveID]*enclaveKubernetesResources{
		enclaveId: enclaveResources,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the new enclave's Kubernetes resources to enclave objects")
	}
	resultEnclave, found := enclaveObjsById[enclaveId]
	if !found {
		return nil, stacktrace.NewError("Successfully converted the new enclave's Kubernetes resources to an enclave object, but the resulting map didn't have an entry for enclave ID '%v'", enclaveId)
	}

	shouldDeleteVolume = false
	shouldDeleteNamespace = false
	return resultEnclave, nil
}

func (backend *KubernetesKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveID]*enclave.Enclave,
	error,
) {
	matchingEnclaves, _, err := backend.getMatchingEnclaveObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves matching the following filters: %+v", filters)
	}
	return matchingEnclaves, nil
}

func (backend *KubernetesKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveID]bool,
	map[enclave.EnclaveID]error,
	error,
) {

	_, matchingKubernetesResources, err := backend.getMatchingEnclaveObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclaves and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveID]error{}
	for enclaveId, resources := range matchingKubernetesResources {
		enclaveIdStr := string(enclaveId)
		if resources.namespace != nil {
			namespaceName := resources.namespace.GetName()

			enclaveWithIDMatchLabels := map[string]string{
				label_key_consts.AppIDLabelKey.GetString():     label_value_consts.AppIDLabelValue.GetString(),
				label_key_consts.EnclaveIDLabelKey.GetString(): enclaveIdStr,
			}

			// Services
			servicesByEnclaveId, err := kubernetes_resource_collectors.CollectMatchingServices(
				ctx,
				backend.kubernetesManager,
				namespaceName,
				enclaveWithIDMatchLabels,
				label_key_consts.EnclaveIDLabelKey.GetString(),
				map[string]bool{
					enclaveIdStr: true,
				},
			)
			if err != nil {
				return nil, nil, stacktrace.Propagate(err, "An error occurred getting services matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespaceName)
			}
			services, found := servicesByEnclaveId[enclaveIdStr]
			if !found {
				erroredEnclaveIds[enclaveId] = stacktrace.NewError("Expected to find enclave's services for enclave '%v' in services by enclave ID map '%+v' but was not found, this is a bug in Kurtosis", enclaveIdStr, servicesByEnclaveId)
			}

			if services != nil {
				errorsByServiceName := map[string]error{}
				for _, service := range services {
					serviceName := service.GetName()
					if err := backend.kubernetesManager.RemoveSelectorsFromService(ctx, namespaceName, serviceName); err != nil {
						errorsByServiceName[serviceName] = err
						continue
					}
				}

				if len(errorsByServiceName) > 0 {
					combinedErrorTitle := fmt.Sprintf("Namespace %v - Service", namespaceName)
					combinedError := buildCombinedError(errorsByServiceName, combinedErrorTitle)
					erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
						combinedError,
						"An error occurred removing one or more service's selectors in namespace '%v' for enclave with ID '%v'",
						namespaceName,
						enclaveId,
					)
					continue
				}
			}

			// Pods
			if resources.pods != nil {
				errorsByPodName := map[string]error{}
				for _, pod := range resources.pods {
					podName := pod.GetName()
					if err := backend.kubernetesManager.RemovePod(ctx, namespaceName, podName); err != nil {
						errorsByPodName[podName] = err
						continue
					}
				}

				if len(errorsByPodName) > 0 {
					combinedErrorTitle := fmt.Sprintf("Namespace %v - Pod", namespaceName)
					combinedError := buildCombinedError(errorsByPodName, combinedErrorTitle)
					erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
						combinedError,
						"An error occurred removing one or more pods in namespace '%v' for enclave with ID '%v'",
						namespaceName,
						enclaveId,
					)
					continue
				}
			}

			successfulEnclaveIds[enclaveId] = true
		}
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func (backend *KubernetesKurtosisBackend) DumpEnclave(ctx context.Context, enclaveId enclave.EnclaveID, outputDirpath string) error {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveID]bool, erroredEnclaveIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingEnclaveObjectsAndKubernetesResources(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveID]*enclave.Enclave,
	map[enclave.EnclaveID]*enclaveKubernetesResources,
	error,
) {
	matchingResources, err := backend.getMatchingEnclaveKubernetesResources(ctx, filters.IDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave Kubernetes resources matching IDs: %+v", filters.IDs)
	}

	enclaveObjects, err := getEnclaveObjectsFromKubernetesResources(matchingResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave objects from Kubernetes resources")
	}

	// Finally, apply the filters
	resultEnclaveObjs := map[enclave.EnclaveID]*enclave.Enclave{}
	resultKubernetesResources := map[enclave.EnclaveID]*enclaveKubernetesResources{}
	for enclaveId, enclaveObj := range enclaveObjects {
		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[enclaveObj.GetID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[enclaveObj.GetStatus()]; !found {
				continue
			}
		}

		resultEnclaveObjs[enclaveId] = enclaveObj
		if _, found := matchingResources[enclaveId]; !found {
			return nil, nil, stacktrace.NewError("Expected to find Kubernetes resources matching enclave '%v' but none was found", enclaveId)
		}
		// Okay to do because we're guaranteed a 1:1 mapping between enclave_obj:enclave_resources
		resultKubernetesResources[enclaveId] = matchingResources[enclaveId]
	}

	return resultEnclaveObjs, resultKubernetesResources, nil
}

// Get back any and all enclave's Kubernetes resources matching the given enclave IDs, where a nil or empty map == "match all enclave IDs"
func (backend *KubernetesKurtosisBackend) getMatchingEnclaveKubernetesResources(ctx context.Context, enclaveIds map[enclave.EnclaveID]bool) (
	map[enclave.EnclaveID]*enclaveKubernetesResources,
	error,
) {

	result := map[enclave.EnclaveID]*enclaveKubernetesResources{}

	enclaveMatchLabels := getEnclaveMatchLabels()

	enclaveIdsStrSet := map[string]bool{}
	for enclaveId, booleanValue := range enclaveIds {
		enclaveIdStr := string(enclaveId)
		enclaveIdsStrSet[enclaveIdStr] = booleanValue
	}

	// Namespaces
	namespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(
		ctx,
		backend.kubernetesManager,
		enclaveMatchLabels,
		label_key_consts.EnclaveIDLabelKey.GetString(),
		enclaveIdsStrSet,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespaces matching IDs '%+v'", enclaveIdsStrSet)
	}

	// Per-namespace objects
	for enclaveIdStr, namespacesForEnclaveId := range namespaces {
		if len(namespacesForEnclaveId) == 0 {
			return nil, stacktrace.NewError(
				"Ostensibly found namespaces for enclave ID '%v', but no namespace objects were returned",
				enclaveIdStr,
			)
		}
		if len(namespacesForEnclaveId) > 1 {
			return nil, stacktrace.NewError(
				"Expected at most one namespace to match enclave ID '%v', but got '%v'",
				enclaveIdStr,
				len(namespacesForEnclaveId),
			)
		}

		namespace := namespacesForEnclaveId[0]

		enclaveWithIDMatchLabels := map[string]string{
			label_key_consts.AppIDLabelKey.GetString():     label_value_consts.AppIDLabelValue.GetString(),
			label_key_consts.EnclaveIDLabelKey.GetString(): enclaveIdStr,
		}

		// Pods
		pods, err := kubernetes_resource_collectors.CollectMatchingPods(
			ctx,
			backend.kubernetesManager,
			namespace.GetName(),
			enclaveWithIDMatchLabels,
			label_key_consts.EnclaveIDLabelKey.GetString(),
			map[string]bool{
				enclaveIdStr: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting pods matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespace.GetName())
		}

		enclaveId := enclave.EnclaveID(enclaveIdStr)

		enclaveResources, found := result[enclaveId]
		if !found {
			enclaveResources = &enclaveKubernetesResources{}
		}
		enclaveResources.namespace = namespace
		enclaveResources.pods = pods[enclaveIdStr]

		result[enclaveId] = enclaveResources
	}

	return result, nil
}

func getEnclaveObjectsFromKubernetesResources(
	allResources map[enclave.EnclaveID]*enclaveKubernetesResources,
) (
	map[enclave.EnclaveID]*enclave.Enclave,
	error,
) {
	result := map[enclave.EnclaveID]*enclave.Enclave{}

	for enclaveId, resourcesForEnclaveId := range allResources {

		if resourcesForEnclaveId.namespace == nil {
			return nil, stacktrace.NewError("Cannot create an enclave object '%v' when no Kubernetes namespace exists", enclaveId)
		}

		enclaveStatus, err := getEnclaveStatusFromEnclavePods(resourcesForEnclaveId.pods)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave status from enclave pods '%+v'", resourcesForEnclaveId.pods)
		}

		enclaveObj := enclave.NewEnclave(
			enclaveId,
			enclaveStatus,
		)

		result[enclaveId] = enclaveObj
	}
	return result, nil
}

func getEnclaveStatusFromEnclavePods(enclavePods []*apiv1.Pod) (enclave.EnclaveStatus, error) {
	resultEnclaveStatus := enclave.EnclaveStatus_Empty
	if len(enclavePods) > 0 {
		resultEnclaveStatus = enclave.EnclaveStatus_Stopped
	}
	for _, enclavePod := range enclavePods {
		podPhase := enclavePod.Status.Phase

		isPodRunning, found := isPodRunningDeterminer[podPhase]
		if !found {
			// This should never happen because we enforce completeness in a unit test
			return resultEnclaveStatus, stacktrace.NewError("No is-running designation found for enclave pod phase '%v'; this is a bug in Kurtosis!", podPhase)
		}
		if isPodRunning {
			resultEnclaveStatus = enclave.EnclaveStatus_Running
			//Enclave is considered running if we found at least one pod running
			break
		}
	}

	return resultEnclaveStatus, nil
}

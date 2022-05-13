package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"os"
	"strconv"
)

const (
	// Name to give the file that we'll write for storing specs of pods, containers, etc.
	kubernetesObjectSpecFilename = "spec.json"
	containerLogsFilename        = "output.log"

	shouldFetchStoppedContainersWhenDumpingEnclave = true

	// Permissions for the files & directories we create as a result of the dump
	createdDirPerms  = 0755
	createdFilePerms = 0644

	numPodsToDumpAtOnce                      = 20

	shouldFollowPodLogsWhenDumping = false
	shouldAddTimestampsWhenDumpingPodLogs = true

	enclaveDumpJsonSerializationIndent = "  "
	enclaveDumpJsonSerializationPrefix = ""
)

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

	_, err = backend.kubernetesManager.CreateNamespace(ctx, enclaveNamespaceName, enclaveNamespaceLabels)
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

	persistentVolumeClaimName := enclaveDataVolumeAttrs.GetName()
	persistentVolumeClaimLabels := enclaveDataVolumeAttrs.GetLabels()

	enclaveVolumeLabelMap := map[string]string{}
	for kubernetesLabelKey, kubernetesLabelValue := range persistentVolumeClaimLabels {
		pvcLabelKey := kubernetesLabelKey.GetString()
		pvcLabelValue := kubernetesLabelValue.GetString()
		enclaveVolumeLabelMap[pvcLabelKey] = pvcLabelValue
	}

	foundVolumes, err := backend.kubernetesManager.GetPersistentVolumeClaimsByLabels(ctx, enclaveNamespaceName, enclaveVolumeLabelMap)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave data volumes matching labels '%+v'", enclaveVolumeLabelMap)
	}
	if len(foundVolumes.Items) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because one or more enclave data volumes for that enclave already exists", enclaveId)
	}

	_, err = backend.kubernetesManager.CreatePersistentVolumeClaim(ctx,
		enclaveNamespaceName,
		persistentVolumeClaimName.GetString(),
		enclaveVolumeLabelMap,
		strconv.Itoa(backend.volumeSizePerEnclaveInGigabytes),
		backend.volumeStorageClassName)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"Failed to create persistent volume claim in enclave '%v' with name '%v' and storage class name '%v'",
			enclaveNamespaceName,
			persistentVolumeClaimName.GetString(),
			backend.volumeStorageClassName)
	}
	shouldDeleteVolume := true
	defer func() {
		if shouldDeleteVolume {
			if err := backend.kubernetesManager.RemovePersistentVolumeClaim(teardownContext, enclaveNamespaceName,persistentVolumeClaimName.GetString()); err != nil {
				logrus.Errorf(
					"Creating the enclave didn't complete successfully, so we tried to delete enclave persistent volume claim '%v' " +
						"that we created but an error was thrown:\n%v",
					persistentVolumeClaimName.GetString(),
					err,
				)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove persistent volume claim with name '%v'!!!!!!!", persistentVolumeClaimName.GetString())
			}
		}
	}()

	enclaveObj := enclave.NewEnclave(enclaveId, enclave.EnclaveStatus_Empty)

	shouldDeleteVolume = false
	shouldDeleteNamespace = false
	return enclaveObj, nil
}

func (backend *KubernetesKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveID]*enclave.Enclave,
	error,
) {
	matchingEnclavesByNamespace, err := backend.getMatchingEnclaves(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves matching the following filters: %+v", filters)
	}

	matchingEnclavesByEnclaveId := map[enclave.EnclaveID]*enclave.Enclave{}
	for _, enclaveObj := range matchingEnclavesByNamespace {
		matchingEnclavesByEnclaveId[enclaveObj.GetID()] = enclaveObj
	}

	return matchingEnclavesByEnclaveId, nil
}

func (backend *KubernetesKurtosisBackend) StopEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveID]bool, erroredEnclaveIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DumpEnclave(ctx context.Context, enclaveId enclave.EnclaveID, outputDirpath string) error {
	searchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.IDLabelKey.GetString(): string(enclaveId),
	}
	namespacesList, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, searchLabels)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting namespaces matching labels: %+v", searchLabels)
	}
	namespaces := namespacesList.Items
	if len(namespaces) == 0 {
		return stacktrace.NewError("No enclave found with ID '%v'", enclaveId)
	}
	if len(namespaces) > 1 {
		return stacktrace.NewError("Expected one enclave matching ID '%v' but found '%v'", enclaveId, len(namespaces))
	}
	namespace := namespaces[0]

	// Create output directory
	if _, err := os.Stat(outputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", outputDirpath)
	}
	if err := os.Mkdir(outputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory at '%v'", outputDirpath)
	}

	// TODO MOVE INTO HELPER FUNCTION
	// TODO PARALLELIZE
	enclavePodsList, err := backend.kubernetesManager.GetPodsByLabels(ctx, namespace.Name, searchLabels)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting pods in enclave '%v'", enclaveId)
	}
	enclavePods := enclavePodsList.Items

	for _, pod := range enclavePods {
		for _, container := range pod.Spec.Containers {

		}

		backend.kubernetesManager.

	}
}

func (backend *KubernetesKurtosisBackend) DestroyEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveID]bool, erroredEnclaveIds map[enclave.EnclaveID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets enclaves matching the search filters, indexed by their [namespace]
func (backend *KubernetesKurtosisBackend) getMatchingEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[string]*enclave.Enclave,
	error,
) {
	matchingEnclaves := map[string]*enclave.Enclave{}

	enclaveNamespaces, err := backend.getAllEnclaveNamespaces(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting all enclave namespaces")
	}

	for _, enclaveNamespace := range enclaveNamespaces {
		enclaveNamespaceName := enclaveNamespace.GetName()
		enclaveNamespaceLabels := enclaveNamespace.GetLabels()

		enclaveIdStr, found := enclaveNamespaceLabels[label_key_consts.EnclaveIDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find a label with name '%v' in Kubernetes namespace '%v', but no such label was found", label_key_consts.EnclaveIDLabelKey.GetString(), enclaveNamespaceName)
		}
		enclaveId := enclave.EnclaveID(enclaveIdStr)
		// If the IDs filter is specified, drop enclaves not matching it
		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[enclaveId]; !found {
				continue
			}
		}

		enclavePods, err := backend.getAllEnclavePods(ctx, enclaveNamespaceName, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting all enclave pods for enclave '%v' in namespace '%v'", enclaveId, enclaveNamespaceName)
		}

		enclaveStatus, err := getEnclaveStatusFromEnclavePods(enclavePods)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave status from enclave pods '%+v'", enclavePods)
		}

		// If the Statuses filter is specified, drop enclaves not matching it
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[enclaveStatus]; !found {
				continue
			}
		}

		enclaveObj := enclave.NewEnclave(enclaveId, enclaveStatus)

		matchingEnclaves[enclaveNamespaceName] = enclaveObj
	}

	return matchingEnclaves, nil
}

func (backend *KubernetesKurtosisBackend) getAllEnclavePods(ctx context.Context, enclaveNamespaceName string, enclaveId enclave.EnclaveID) ([]apiv1.Pod, error) {
	matchingPods := []apiv1.Pod{}

	matchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():     label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
	}

	foundPods, err := backend.kubernetesManager.GetPodsByLabels(ctx, enclaveNamespaceName, matchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting pods by labels '%+v' in namespace '%v'", matchLabels, enclaveNamespaceName)
	}

	if foundPods.Items != nil {
		matchingPods = foundPods.Items
	}
	return matchingPods, nil
}

func getEnclaveStatusFromEnclavePods(enclavePods []apiv1.Pod) (enclave.EnclaveStatus, error) {
	resultEnclaveStatus := enclave.EnclaveStatus_Stopped
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
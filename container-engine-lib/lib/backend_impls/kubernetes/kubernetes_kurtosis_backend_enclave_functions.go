package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	apiv1 "k8s.io/api/core/v1"
	"os"
	"path"
	"strings"
)

const (
	// Name to give the file that we'll write for storing specs of pods, containers, etc.
	podSpecFilename             = "spec.json"
	containerLogsFilenameSuffix = ".log"

	shouldFetchStoppedContainersWhenDumpingEnclave = true

	// Permissions for the files & directories we create as a result of the dump
	createdDirPerms  os.FileMode = 0755
	createdFilePerms os.FileMode = 0644

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
		backend.volumeSizePerEnclaveInMegabytes,
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
	namespace, err := backend.getEnclaveNamespace(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave namespace for enclave ID '%v'", enclaveId)
	}
	namespaceName := namespace.Name

	// Create output directory
	if _, err := os.Stat(outputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", outputDirpath)
	}
	if err := os.Mkdir(outputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory at '%v'", outputDirpath)
	}

	enclavePods, err := backend.getAllEnclavePods(ctx, namespaceName, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting pods in enclave '%v'", enclaveId)
	}

	workerPool := workerpool.New(numPodsToDumpAtOnce)
	resultErrsChan := make(chan error, len(enclavePods))
	for _, pod := range enclavePods {
		/*
			!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			It's VERY important that the actual `func()` job function get created inside a helper function!!
			This is because variables declared inside for-loops are created BY REFERENCE rather than by-value, which
				means that if we inline the `func() {....}` creation here then all the job functions would get a REFERENCE to
				any variables they'd use.
			This means that by the time the job functions were run in the worker pool (long after the for-loop finished)
				then all the job functions would be using a reference from the last iteration of the for-loop.

			For more info, see the "Variables declared in for loops are passed by reference" section of:
				https://www.calhoun.io/gotchas-and-common-mistakes-with-closures-in-go/ for more details
			!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		*/
		jobToSubmit := createDumpPodJob(
			ctx,
			backend.kubernetesManager,
			namespaceName,
			pod,
			outputDirpath,
			resultErrsChan,
		)
		workerPool.Submit(jobToSubmit)
	}
	workerPool.StopWait()
	close(resultErrsChan)

	allResultErrStrs := []string{}
	for resultErr := range resultErrsChan {
		allResultErrStrs = append(allResultErrStrs, resultErr.Error())
	}

	if len(allResultErrStrs) > 0 {
		allIndexedResultErrStrs := []string{}
		for idx, resultErrStr := range allResultErrStrs {
			indexedResultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR %v <<<<<<<<<<<<<<<<<\n%v", idx, resultErrStr)
			allIndexedResultErrStrs = append(allIndexedResultErrStrs, indexedResultErrStr)
		}

		// NOTE: We don't use stacktrace here because the actual stacktraces we care about are the ones from the threads!
		return errors.New(fmt.Sprintf(
			"The following errors occurred when trying to dump information about enclave '%v':\n%v",
			enclaveId,
			strings.Join(allIndexedResultErrStrs, "\n\n"),
		))
	}
	return nil
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

func createDumpPodJob(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespaceName string,
	pod apiv1.Pod,
	enclaveOutputDirpath string,
	resultErrsChan chan error,
) func() {
	return func() {
		if err := dumpPodInfo(ctx, kubernetesManager, namespaceName, pod, enclaveOutputDirpath); err != nil {
			resultErrsChan <- stacktrace.Propagate(
				err,
				"An error occurred dumping info for pod '%v'",
				pod.Name,
			)
		}
	}
}

func dumpPodInfo(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespaceName string,
	pod apiv1.Pod,
	enclaveOutputDirpath string,
) error {
	podName := pod.Name

	// Make pod output directory
	podOutputDirpath := path.Join(enclaveOutputDirpath, podName)
	if err := os.Mkdir(podOutputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating directory '%v' to hold the output of pod with name '%v'",
			podOutputDirpath,
			podName,
		)
	}

	jsonSerializedPodSpecBytes, err := json.MarshalIndent(pod.Spec, enclaveDumpJsonSerializationPrefix, enclaveDumpJsonSerializationIndent)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing the spec of pod '%v' to JSON", podName)
	}
	podSpecOutputFilepath := path.Join(podOutputDirpath, podSpecFilename)
	if err := ioutil.WriteFile(podSpecOutputFilepath, jsonSerializedPodSpecBytes, createdFilePerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred writing the spec of pod '%v' to file '%v'",
			podName,
			podSpecOutputFilepath,
		)
	}

	for _, container := range pod.Spec.Containers {
		containerName := container.Name

		// Make container output directory
		containerLogsFilepath := path.Join(podOutputDirpath, containerName + containerLogsFilenameSuffix)
		containerLogsOutputFp, err := os.Create(containerLogsFilepath)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred creating file '%v' to hold the logs of container with name '%v' in pod '%v'",
				containerLogsFilepath,
				containerName,
				podName,
			)
		}
		defer containerLogsOutputFp.Close()

		containerLogReadCloser, err := kubernetesManager.GetContainerLogs(
			ctx,
			namespaceName,
			podName,
			containerName,
			shouldFollowPodLogsWhenDumping,
			shouldAddTimestampsWhenDumpingPodLogs,
		)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred dumping logs of container '%v' in pod '%v' in namespace '%v'",
				containerName,
				podName,
				namespaceName,
			)
		}
		defer containerLogReadCloser.Close()

		if _, err := io.Copy(containerLogsOutputFp, containerLogReadCloser); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred writing logs of container '%v' in pod '%v' to file '%v'",
				containerName,
				podName,
				containerLogsFilepath,
			)
		}
	}

	return nil
}
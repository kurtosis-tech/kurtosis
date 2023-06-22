package kubernetes_kurtosis_backend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	apiv1 "k8s.io/api/core/v1"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"os"
	"path"
	"time"
)

const (
	// Name to give the file that we'll write for storing specs of pods, containers, etc.
	podSpecFilename             = "spec.json"
	containerLogsFilenameSuffix = ".log"

	// Permissions for the files & directories we create as a result of the dump
	createdDirPerms  os.FileMode = 0755
	createdFilePerms os.FileMode = 0644

	numPodsToDumpAtOnce = 20

	shouldFollowPodLogsWhenDumping        = false
	shouldAddTimestampsWhenDumpingPodLogs = true

	enclaveDumpJsonSerializationIndent = "  "
	enclaveDumpJsonSerializationPrefix = ""

	dumpPodErrorTitle = "Pod"
)

// TODO: MIGRATE THIS FOLDER TO USE STRUCTURE OF USER_SERVICE_FUNCTIONS MODULE

// Any of these values being nil indicates that the resource doesn't exist
type enclaveKubernetesResources struct {
	// Will never be nil because enclaves are defined by namespaces
	namespace *apiv1.Namespace

	// Not technically resources that define an enclave, but we need them both
	// to StopEnclave and to return an EnclaveStatus
	pods []apiv1.Pod

	// Not technically resources that define an enclave, but we need them for
	// StopEnclave
	services []apiv1.Service
}

type dumpPodResult struct {
	podName string
	err     error
}

// ====================================================================================================
//
//	Enclave CRUD Methods
//
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) CreateEnclave(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	enclaveName string,
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
		label_key_consts.AppIDKubernetesLabelKey.GetString():       label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.EnclaveUUIDKubernetesLabelKey.GetString(): string(enclaveUuid),
	}
	namespaceList, err := backend.kubernetesManager.GetNamespacesByLabels(ctx, searchNamespaceLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list namespaces from Kubernetes, so can not verify if enclave '%v' already exists.", enclaveUuid)
	}
	if len(namespaceList.Items) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave with ID '%v' because an enclave with ID '%v' already exists", enclaveUuid, enclaveName)
	}

	creationTime := time.Now()

	// Make Enclave attributes provider
	enclaveObjAttrsProvider := backend.objAttrsProvider.ForEnclave(enclaveUuid)
	enclaveNamespaceAttrs, err := enclaveObjAttrsProvider.ForEnclaveNamespace(isPartitioningEnabled, creationTime, enclaveName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave network attributes for the enclave with ID '%v'", enclaveUuid)
	}

	enclaveNamespaceName := enclaveNamespaceAttrs.GetName().GetString()
	enclaveNamespaceLabels := shared_helpers.GetStringMapFromLabelMap(enclaveNamespaceAttrs.GetLabels())
	enclaveAnnotationsStrs := shared_helpers.GetStringMapFromAnnotationMap(enclaveNamespaceAttrs.GetAnnotations())

	enclaveNamespace, err := backend.kubernetesManager.CreateNamespace(ctx, enclaveNamespaceName, enclaveNamespaceLabels, enclaveAnnotationsStrs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%v' for enclave '%v'", enclaveNamespaceName, enclaveUuid)
	}
	shouldDeleteNamespace := true
	defer func() {
		if shouldDeleteNamespace {
			if err := backend.kubernetesManager.RemoveNamespace(teardownContext, enclaveNamespace); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete namespace '%v' that we created but an error was thrown:\n%v", enclaveNamespaceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove namespace with name '%v'!!!!!!!", enclaveNamespaceName)
			}
		}
	}()

	enclaveResources := &enclaveKubernetesResources{
		namespace: enclaveNamespace,
		pods:      []apiv1.Pod{},
		services:  nil,
	}
	enclaveObjsById, err := getEnclaveObjectsFromKubernetesResources(map[enclave.EnclaveUUID]*enclaveKubernetesResources{
		enclaveUuid: enclaveResources,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the new enclave's Kubernetes resources to enclave objects")
	}
	resultEnclave, found := enclaveObjsById[enclaveUuid]
	if !found {
		return nil, stacktrace.NewError("Successfully converted the new enclave's Kubernetes resources to an enclave object, but the resulting map didn't have an entry for enclave UUID '%v'", enclaveUuid)
	}

	shouldDeleteNamespace = false
	return resultEnclave, nil
}

func (backend *KubernetesKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveUUID]*enclave.Enclave,
	error,
) {
	matchingEnclaves, _, err := backend.getMatchingEnclaveObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves matching the following filters: %+v", filters)
	}
	return matchingEnclaves, nil
}

func (backend *KubernetesKurtosisBackend) RenameEnclave(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	newName string,
) error {

	enclave, kubernetesResources, err := backend.getSingleEnclaveAndKubernetesResources(ctx, enclaveUuid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave object and Kubernetes resources for enclave ID '%v'", enclaveUuid)
	}
	namespace := kubernetesResources.namespace
	if namespace == nil {
		return stacktrace.NewError("Cannot rename enclave '%v' because no Kubernetes namespace exists for it", enclaveUuid)
	}

	updatedAnnotations := map[string]string{
		kubernetes_annotation_key_consts.EnclaveNameAnnotationKey.GetString(): newName,
	}

	namespaceApplyConfigurator := func(namespaceApplyConfig *applyconfigurationsv1.NamespaceApplyConfiguration) {
		namespaceApplyConfig.WithAnnotations(updatedAnnotations)
	}

	if _, err := backend.kubernetesManager.UpdateNamespace(ctx, namespace.GetName(), namespaceApplyConfigurator); err != nil {
		return stacktrace.Propagate(err, "An error occurred renaming enclave with UUID '%v', renaming from '%s' to '%s'", enclaveUuid, enclave.GetName(), newName)
	}

	return nil
}

func (backend *KubernetesKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveUUID]bool,
	map[enclave.EnclaveUUID]error,
	error,
) {

	_, matchingKubernetesResources, err := backend.getMatchingEnclaveObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclaves and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveUUID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveUUID]error{}
	for enclaveId, resources := range matchingKubernetesResources {
		namespaceName := resources.namespace.GetName()

		// Pods
		if resources.pods != nil {
			errorsByPodName := map[string]error{}
			for _, pod := range resources.pods {
				podName := pod.GetName()
				if err := backend.kubernetesManager.RemovePod(ctx, &pod); err != nil {
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

		// Services
		if resources.services != nil {
			errorsByServiceName := map[string]error{}
			for _, service := range resources.services {
				serviceName := service.GetName()
				updateConfigurator := func(updatesToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
					specUpdates := applyconfigurationsv1.ServiceSpec().WithSelector(nil)
					updatesToApply.WithSpec(specUpdates)
				}
				if _, err := backend.kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updateConfigurator); err != nil {
					errorsByServiceName[service.Name] = err
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

		successfulEnclaveIds[enclaveId] = true
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

func (backend *KubernetesKurtosisBackend) DumpEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, outputDirpath string) error {
	_, kubernetesResources, err := backend.getSingleEnclaveAndKubernetesResources(ctx, enclaveUuid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave object and Kubernetes resources for enclave ID '%v'", enclaveUuid)
	}
	namespace := kubernetesResources.namespace
	if namespace == nil {
		return stacktrace.NewError("Cannot dump enclave '%v' because no Kubernetes namespace exists for it", enclaveUuid)
	}

	// Create output directory
	if _, err := os.Stat(outputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", outputDirpath)
	}
	if err := os.Mkdir(outputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory at '%v'", outputDirpath)
	}

	podsToDump := kubernetesResources.pods
	if podsToDump == nil {
		podsToDump = []apiv1.Pod{}
	}

	workerPool := workerpool.New(numPodsToDumpAtOnce)
	resultErrsChan := make(chan dumpPodResult, len(podsToDump))
	for _, pod := range podsToDump {
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
			namespace.Name,
			pod,
			outputDirpath,
			resultErrsChan,
		)
		workerPool.Submit(jobToSubmit)
	}
	workerPool.StopWait()
	close(resultErrsChan)

	resultErrorsByPodName := map[string]error{}
	for podResult := range resultErrsChan {
		resultErrorsByPodName[podResult.podName] = podResult.err
	}

	if len(resultErrorsByPodName) > 0 {
		combinedErr := buildCombinedError(resultErrorsByPodName, dumpPodErrorTitle)
		return combinedErr
	}
	return nil
}

func (backend *KubernetesKurtosisBackend) DestroyEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveUUID]bool,
	map[enclave.EnclaveUUID]error,
	error,
) {
	_, matchingResources, err := backend.getMatchingEnclaveObjectsAndKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave Kubernetes resources matching filters: %+v", filters)
	}

	successfulEnclaveIds := map[enclave.EnclaveUUID]bool{}
	erroredEnclaveIds := map[enclave.EnclaveUUID]error{}
	for enclaveId, resources := range matchingResources {
		// Remove the namespace
		if resources.namespace != nil {
			namespaceName := resources.namespace.Name
			if err := backend.kubernetesManager.RemoveNamespace(ctx, resources.namespace); err != nil {
				erroredEnclaveIds[enclaveId] = stacktrace.Propagate(
					err,
					"An error occurred removing namespace '%v' for enclave '%v'",
					namespaceName,
					enclaveId,
				)
				continue
			}
		}

		successfulEnclaveIds[enclaveId] = true
	}

	return successfulEnclaveIds, erroredEnclaveIds, nil
}

// ====================================================================================================
//
//	Private Helper Methods
//
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingEnclaveObjectsAndKubernetesResources(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveUUID]*enclave.Enclave,
	map[enclave.EnclaveUUID]*enclaveKubernetesResources,
	error,
) {
	matchingResources, err := backend.getMatchingEnclaveKubernetesResources(ctx, filters.UUIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave Kubernetes resources matching UUIDs: %+v", filters.UUIDs)
	}

	enclaveObjects, err := getEnclaveObjectsFromKubernetesResources(matchingResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave objects from Kubernetes resources")
	}

	// Finally, apply the filters
	resultEnclaveObjs := map[enclave.EnclaveUUID]*enclave.Enclave{}
	resultKubernetesResources := map[enclave.EnclaveUUID]*enclaveKubernetesResources{}
	for enclaveId, enclaveObj := range enclaveObjects {
		if filters.UUIDs != nil && len(filters.UUIDs) > 0 {
			if _, found := filters.UUIDs[enclaveObj.GetUUID()]; !found {
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

func (backend *KubernetesKurtosisBackend) getSingleEnclaveAndKubernetesResources(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (*enclave.Enclave, *enclaveKubernetesResources, error) {
	enclaveSearchFilters := &enclave.EnclaveFilters{
		UUIDs: map[enclave.EnclaveUUID]bool{
			enclaveUuid: true,
		},
		Statuses: nil,
	}
	matchingEnclaveObjects, matchingKubernetesResources, err := backend.getMatchingEnclaveObjectsAndKubernetesResources(ctx, enclaveSearchFilters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave objects and Kubernetes resources matching enclave '%v'", enclaveUuid)
	}
	if len(matchingEnclaveObjects) == 0 || len(matchingKubernetesResources) == 0 {
		return nil, nil, stacktrace.NewError("Didn't find enclave objects and Kubernetes resources for enclave '%v'", enclaveUuid)
	}
	if len(matchingEnclaveObjects) > 1 || len(matchingKubernetesResources) > 1 {
		return nil, nil, stacktrace.NewError("Found more than one enclave objects/Kubernetes resources for enclave '%v'", enclaveUuid)
	}

	enclaveObject, found := matchingEnclaveObjects[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError("No enclave object exists for enclave '%v'", enclaveUuid)
	}

	kubernetesResources, found := matchingKubernetesResources[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError("No Kubernetes resources object exists for enclave '%v'", enclaveUuid)
	}

	return enclaveObject, kubernetesResources, nil
}

// Get back any and all enclave's Kubernetes resources matching the given enclave IDs, where a nil or empty map == "match all enclave IDs"
func (backend *KubernetesKurtosisBackend) getMatchingEnclaveKubernetesResources(ctx context.Context, enclaveUuids map[enclave.EnclaveUUID]bool) (
	map[enclave.EnclaveUUID]*enclaveKubernetesResources,
	error,
) {
	enclaveMatchLabels := getEnclaveMatchLabels()

	enclaveIdsStrSet := map[string]bool{}
	for enclaveId, booleanValue := range enclaveUuids {
		enclaveIdStr := string(enclaveId)
		enclaveIdsStrSet[enclaveIdStr] = booleanValue
	}

	// Namespaces
	namespaces, err := kubernetes_resource_collectors.CollectMatchingNamespaces(
		ctx,
		backend.kubernetesManager,
		enclaveMatchLabels,
		label_key_consts.EnclaveUUIDKubernetesLabelKey.GetString(),
		enclaveIdsStrSet,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespaces matching UUIDs '%+v'", enclaveIdsStrSet)
	}

	// Per-namespace objects
	result := map[enclave.EnclaveUUID]*enclaveKubernetesResources{}
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

		namespaceName := namespace.GetName()
		enclaveWithIDMatchLabels := map[string]string{
			label_key_consts.AppIDKubernetesLabelKey.GetString():       label_value_consts.AppIDKubernetesLabelValue.GetString(),
			label_key_consts.EnclaveUUIDKubernetesLabelKey.GetString(): enclaveIdStr,
		}

		// Pods
		podsList, err := backend.kubernetesManager.GetPodsByLabels(
			ctx,
			namespaceName,
			enclaveWithIDMatchLabels,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting pods matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespace.GetName())
		}
		pods := []apiv1.Pod{}
		pods = append(pods, podsList.Items...)

		// Services
		servicesList, err := backend.kubernetesManager.GetServicesByLabels(
			ctx,
			namespaceName,
			enclaveWithIDMatchLabels,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting services matching enclave ID '%v' in namespace '%v'", enclaveIdStr, namespace.GetName())
		}
		var services []apiv1.Service
		services = append(services, servicesList.Items...)

		enclaveResources := &enclaveKubernetesResources{
			namespace: namespace,
			pods:      pods,
			services:  services,
		}
		result[enclave.EnclaveUUID(enclaveIdStr)] = enclaveResources
	}

	return result, nil
}

func getEnclaveObjectsFromKubernetesResources(
	allResources map[enclave.EnclaveUUID]*enclaveKubernetesResources,
) (
	map[enclave.EnclaveUUID]*enclave.Enclave,
	error,
) {
	result := map[enclave.EnclaveUUID]*enclave.Enclave{}

	for enclaveId, resourcesForEnclaveId := range allResources {

		if resourcesForEnclaveId.namespace == nil {
			return nil, stacktrace.NewError("Cannot create an enclave object '%v' when no Kubernetes namespace exists", enclaveId)
		}

		enclaveStatus, err := getEnclaveStatusFromEnclavePods(resourcesForEnclaveId.pods)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting enclave status from enclave pods '%+v'", resourcesForEnclaveId.pods)
		}

		enclaveCreationTime, err := getEnclaveCreationTimeFromEnclaveNamespace(resourcesForEnclaveId.namespace)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the enclave's creation time from the enclave's namespace '%+v'", resourcesForEnclaveId.namespace)
		}

		enclaveName := getEnclaveNameFromEnclaveNamespace(resourcesForEnclaveId.namespace)

		enclaveObj := enclave.NewEnclave(
			enclaveId,
			enclaveName,
			enclaveStatus,
			enclaveCreationTime,
		)

		result[enclaveId] = enclaveObj
	}
	return result, nil
}

func getEnclaveStatusFromEnclavePods(enclavePods []apiv1.Pod) (enclave.EnclaveStatus, error) {
	resultEnclaveStatus := enclave.EnclaveStatus_Empty
	if len(enclavePods) > 0 {
		resultEnclaveStatus = enclave.EnclaveStatus_Stopped
	}
	for _, enclavePod := range enclavePods {
		podStatus, err := shared_helpers.GetContainerStatusFromPod(&enclavePod)
		if err != nil {
			return 0, stacktrace.Propagate(err, "An error occurred getting status from pod '%v'", enclavePod.Name)
		}
		// An enclave is considered running if we found at least one pod running
		if podStatus == container_status.ContainerStatus_Running {
			resultEnclaveStatus = enclave.EnclaveStatus_Running
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
	resultChan chan dumpPodResult,
) func() {
	return func() {
		if err := dumpPodInfo(ctx, kubernetesManager, namespaceName, pod, enclaveOutputDirpath); err != nil {
			result := dumpPodResult{
				podName: pod.Name,
				err:     err,
			}
			resultChan <- result
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
		containerLogsFilepath := path.Join(podOutputDirpath, containerName+containerLogsFilenameSuffix)
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

func getEnclaveCreationTimeFromEnclaveNamespace(namespace *apiv1.Namespace) (*time.Time, error) {
	namespaceAnnotations := namespace.Annotations

	enclaveCreationTimeStr, found := namespaceAnnotations[kubernetes_annotation_key_consts.EnclaveCreationTimeAnnotationKey.GetString()]
	if !found {
		//Handling retro-compatibility, enclaves that did not track enclave's creation time
		return nil, nil //TODO remove this return after 2023-01-01
		//TODO uncomment this after 2023-01-01 when we are sure that there is not any old enclave created with the creation time annotation
		//return nil, stacktrace.NewError("Expected to find namespace's annotation with key '%v' but none was found", kubernetes_annotation_key_consts.EnclaveCreationTimeAnnotationKey.GetString())
	}

	enclaveCreationTime, err := time.Parse(time.RFC3339, enclaveCreationTimeStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing enclave creation time '%v' using this format '%v'", enclaveCreationTimeStr, time.RFC3339)
	}

	return &enclaveCreationTime, nil
}

func getEnclaveNameFromEnclaveNamespace(namespace *apiv1.Namespace) string {
	namespaceAnnotations := namespace.Annotations

	enclaveCreationTimeStr, found := namespaceAnnotations[kubernetes_annotation_key_consts.EnclaveNameAnnotationKey.GetString()]
	if !found {
		//Handling retro-compatibility, enclaves that did not track enclave's name
		return ""
	}

	return enclaveCreationTimeStr
}

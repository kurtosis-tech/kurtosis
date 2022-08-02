package shared_helpers

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

const (
	megabytesToBytesFactor = 1_000_000
)

// !!!WARNING!!!
// This files contains functions that are shared by multiple DockerKurtosisBackend functions.
// Generally, we want to prevent long utils folders with functionality that is difficult to find, so be careful
// when adding functionality in this folder.
// Things to think about: Could this function be a private helper function that's scope is smaller than you think?
// Eg. only used by start user services functions thus could go in start_user_services.go

// Kubernetes doesn't provide public IP or port information; this is instead handled by the Kurtosis gateway that the user uses
// to connect to Kubernetes
var servicePublicIp net.IP = nil
var servicePublicPorts map[string]*port_spec.PortSpec = nil

// This maps a Kubernetes pod's phase to a binary "is the pod considered running?" determiner
// Its completeness is enforced via unit test
var isPodRunningDeterminer = map[apiv1.PodPhase]bool{
	apiv1.PodPending: true,
	apiv1.PodRunning: true,
	apiv1.PodSucceeded: false,
	apiv1.PodFailed: false,
	apiv1.PodUnknown: false, //We cannot say that a pod is not running if we don't know the real state
}

// TODO Remove this once we split apart the KubernetesKurtosisBackend into multiple backends (which we can only
//  do once the CLI no longer makes any calls directly to the KurtosisBackend, and instead makes all its calls through
//  the API container & engine APIs)
type CliModeArgs struct {
	// No CLI mode args needed for now
}

type ApiContainerModeArgs struct {
	OwnEnclaveId enclave.EnclaveID

	OwnNamespaceName string

	storageClassName string

	// TODO make this more dynamic - maybe guess based on the files artifact size?
	filesArtifactExpansionVolumeSizeInMegabytes uint
}

type EngineServerModeArgs struct {
	/*
		StorageClass name to be used for volumes in the cluster
		StorageClasses must be defined by a cluster administrator.
		passes this in when starting Kurtosis with Kubernetes.
	*/
	storageClassName string

	/*
		Enclave availability must be set and defined by a cluster administrator.
		The user passes this in when starting Kurtosis with Kubernetes.
	*/
	enclaveDataVolumeSizeInMegabytes uint
}

type UserServiceObjectsAndKubernetesResources struct {
	// Should never be nil because 1 Kubernetes service = 1 Kurtosis service registration
	ServiceRegistration *service.ServiceRegistration

	// May be nil if no pod has been started yet
	Service *service.Service

	// Will never be nil
	KubernetesResources *UserServiceKubernetesResources
}

// Any of these fields can be nil if they don't exist in Kubernetes, though at least
// one field will be present (else this struct won't exist)
type UserServiceKubernetesResources struct {
	// This can be nil if the user manually deleted the Kubernetes service (e.g. using the Kubernetes dashboard)
	Service *apiv1.Service

	// This can be nil if the user hasn't started a pod for the service yet, or if the pod was deleted
	Pod *apiv1.Pod
}

func GetEnclaveNamespaceName(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	cliModeArgs *CliModeArgs,
	apiContainerModeArgs *ApiContainerModeArgs,
	engineServerModeArgs *EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager) (string, error) {
	// TODO This is a big janky hack that results from KubernetesKurtosisBackend containing functions for all of API containers, engines, and CLIs
	//  We want to fix this by splitting the KubernetesKurtosisBackend into a bunch of different backends, one per user, but we can only
	//  do this once the CLI no longer uses API container functionality (e.g. GetServices)
	// CLIs and engines can list namespaces so they'll be able to use the regular list-namespaces-and-find-the-one-matching-the-enclave-ID
	// API containers can't list all namespaces due to being namespaced objects themselves (can only view their own namespace, so
	// they can only check if the requested enclave matches the one they have stored
	var namespaceName string
	if cliModeArgs != nil || engineServerModeArgs != nil {
		matchLabels := getEnclaveMatchLabels()
		matchLabels[label_key_consts.EnclaveIDKubernetesLabelKey.GetString()] = string(enclaveId)

		namespaces, err := kubernetesManager.GetNamespacesByLabels(ctx, matchLabels)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred getting the enclave namespace using labels '%+v'", matchLabels)
		}

		numOfNamespaces := len(namespaces.Items)
		if numOfNamespaces == 0 {
			return "", stacktrace.NewError("No namespace matching labels '%+v' was found", matchLabels)
		}
		if numOfNamespaces > 1 {
			return "", stacktrace.NewError("Expected to find only one enclave namespace matching enclave ID '%v', but found '%v'; this is a bug in Kurtosis", enclaveId, numOfNamespaces)
		}

		namespaceName = namespaces.Items[0].Name
	} else if apiContainerModeArgs != nil {
		if enclaveId != apiContainerModeArgs.OwnEnclaveId {
			return "", stacktrace.NewError(
				"Received a request to get namespace for enclave '%v', but the Kubernetes Kurtosis backend is running in an API " +
					"container in a different enclave '%v' (so Kubernetes would throw a permission error)",
				enclaveId,
				apiContainerModeArgs.OwnEnclaveId,
			)
		}
		namespaceName = apiContainerModeArgs.OwnNamespaceName
	} else {
		return "", stacktrace.NewError("Received a request to get an enclave namespace's name, but the Kubernetes Kurtosis backend isn't in any recognized mode; this is a bug in Kurtosis")
	}

	return namespaceName, nil
}

func GetMatchingUserServiceObjectsAndKubernetesResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	cliModeArgs *CliModeArgs,
	apiContainerModeArgs *ApiContainerModeArgs,
	engineServerModeArgs *EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceGUID]*UserServiceObjectsAndKubernetesResources,
	error,
) {
	allResources, err := GetUserServiceKubernetesResourcesMatchingGuids(ctx, enclaveId, filters.GUIDs, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service Kubernetes resources matching GUIDs: %+v", filters.GUIDs)
	}

	for serviceGuid, serviceResources := range allResources {
		logrus.Tracef("Found resources for service '%v': %+v", serviceGuid, serviceResources)
	}

	allObjectsAndResources, err := getUserServiceObjectsFromKubernetesResources(enclaveId, allResources)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service objects from Kubernetes resources")
	}

	// Sanity check
	if len(allResources) != len(allObjectsAndResources) {
		return nil, stacktrace.NewError(
			"Transformed %v Kubernetes resource objects into %v Kurtosis objects; this is a bug in Kurtosis",
			len(allResources),
			len(allObjectsAndResources),
		)
	}

	// Filter the results down to the requested filters
	results := map[service.ServiceGUID]*UserServiceObjectsAndKubernetesResources{}
	for serviceGuid, objectsAndResources := range allObjectsAndResources {
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[serviceGuid]; !found {
				continue
			}
		}

		registration := objectsAndResources.ServiceRegistration
		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[registration.GetID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			kubernetesService := objectsAndResources.Service

			// If status isn't specified, return registered-only objects; if not, remove them all
			if kubernetesService == nil {
				continue
			}

			if _, found := filters.Statuses[kubernetesService.GetStatus()]; !found {
				continue
			}
		}

		results[serviceGuid] = objectsAndResources
	}

	return results, nil
}

func GetUserServiceKubernetesResourcesMatchingGuids(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuids map[service.ServiceGUID]bool,
	cliModeArgs *CliModeArgs,
	apiContainerModeArgs *ApiContainerModeArgs,
	engineServerModeArgs *EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceGUID]*UserServiceKubernetesResources,
	error,
) {
	namespaceName, err := GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	// TODO switch to properly-typed KubernetesLabelValue object!!!
	postFilterLabelValues := map[string]bool{}
	for serviceGuid := range serviceGuids {
		postFilterLabelValues[string(serviceGuid)] = true
	}

	kubernetesResourceSearchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.EnclaveIDKubernetesLabelKey.GetString():            string(enclaveId),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}

	results := map[service.ServiceGUID]*UserServiceKubernetesResources{}

	// Get k8s services
	matchingKubernetesServices, err := kubernetes_resource_collectors.CollectMatchingServices(
		ctx,
		kubernetesManager,
		namespaceName,
		kubernetesResourceSearchLabels,
		label_key_consts.GUIDKubernetesLabelKey.GetString(),
		postFilterLabelValues,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes services matching service GUIDs: %+v", serviceGuids)
	}
	for serviceGuidStr, kubernetesServicesForGuid := range matchingKubernetesServices {
		logrus.Tracef("Found Kubernetes services for GUID '%v': %+v", serviceGuidStr, kubernetesServicesForGuid)
		serviceGuid := service.ServiceGUID(serviceGuidStr)

		numServicesForGuid := len(kubernetesServicesForGuid)
		if numServicesForGuid == 0 {
			// This would indicate a bug in our service retrieval logic because we shouldn't even have a map entry if there's nothing matching it
			return nil, stacktrace.NewError("Got entry of result services for service GUID '%v', but no Kubernetes services were returned; this is a bug in Kurtosis", serviceGuid)
		}
		if numServicesForGuid > 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes services associated with service GUID '%v'; this is a bug in Kurtosis", numServicesForGuid, serviceGuid)
		}
		kubernetesService := kubernetesServicesForGuid[0]

		resultObj, found := results[serviceGuid]
		if !found {
			resultObj = &UserServiceKubernetesResources{}
		}
		resultObj.Service = kubernetesService
		results[serviceGuid] = resultObj
	}

	// Get k8s pods
	matchingKubernetesPods, err := kubernetes_resource_collectors.CollectMatchingPods(
		ctx,
		kubernetesManager,
		namespaceName,
		kubernetesResourceSearchLabels,
		label_key_consts.GUIDKubernetesLabelKey.GetString(),
		postFilterLabelValues,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes pods matching service GUIDs: %+v", serviceGuids)
	}
	for serviceGuidStr, kubernetesPodsForGuid := range matchingKubernetesPods {
		logrus.Tracef("Found Kubernetes pods for GUID '%v': %+v", serviceGuidStr, kubernetesPodsForGuid)
		serviceGuid := service.ServiceGUID(serviceGuidStr)

		numPodsForGuid := len(kubernetesPodsForGuid)
		if numPodsForGuid == 0 {
			// This would indicate a bug in our pod retrieval logic because we shouldn't even have a map entry if there's nothing matching it
			return nil, stacktrace.NewError("Got entry of result pods for service GUID '%v', but no Kubernetes pods were returned; this is a bug in Kurtosis", serviceGuid)
		}
		if numPodsForGuid > 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes pods associated with service GUID '%v'; this is a bug in Kurtosis", numPodsForGuid, serviceGuid)
		}
		kubernetesPod := kubernetesPodsForGuid[0]

		resultObj, found := results[serviceGuid]
		if !found {
			resultObj = &UserServiceKubernetesResources{}
		}
		resultObj.Pod = kubernetesPod
		results[serviceGuid] = resultObj
	}

	return results, nil
}

// If no expected-ports list is passed in, no validation is done and all the ports are passed back as-is
func GetPrivatePortsAndValidatePortExistence(kubernetesService *apiv1.Service, expectedPortIds map[string]bool) (map[string]*port_spec.PortSpec, error) {
	portSpecsStr, found := kubernetesService.GetAnnotations()[kubernetes_annotation_key_consts.PortSpecsKubernetesAnnotationKey.GetString()]
	if !found {
		return nil, stacktrace.NewError(
			"Couldn't find expected port specs annotation key '%v' on the Kubernetes service",
			kubernetes_annotation_key_consts.PortSpecsKubernetesAnnotationKey.GetString(),
		)
	}
	privatePortSpecs, err := kubernetes_port_spec_serializer.DeserializePortSpecs(portSpecsStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing private port specs string '%v'", privatePortSpecs)
	}

	if expectedPortIds != nil && len(expectedPortIds) > 0 {
		for portId := range expectedPortIds {
			if _, found := privatePortSpecs[portId]; !found {
				return nil, stacktrace.NewError("Missing private port with ID '%v' in the private ports", portId)
			}
		}
	}
	return privatePortSpecs, nil
}

func GetContainerStatusFromPod(pod *apiv1.Pod) (container_status.ContainerStatus, error) {
	// TODO Rename this; this shouldn't be called "ContainerStatus" since there's no longer a 1:1 mapping between container:kurtosis_object
	status := container_status.ContainerStatus_Stopped

	if pod != nil {
		podPhase := pod.Status.Phase
		isPodRunning, found := isPodRunningDeterminer[podPhase]
		if !found {
			// This should never happen because we enforce completeness in a unit test
			return status, stacktrace.NewError("No is-pod-running determination found for pod phase '%v' on pod '%v'; this is a bug in Kurtosis", podPhase, pod.Name)
		}
		if isPodRunning {
			status = container_status.ContainerStatus_Running
		}
	}
	return status, nil
}

func GetSingleUserServiceObjectsAndResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	cliModeArgs *CliModeArgs,
	apiContainerModeArgs *ApiContainerModeArgs,
	engineServerModeArgs *EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	) (*UserServiceObjectsAndKubernetesResources, error) {
	searchFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}
	searchResults, err := GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, searchFilters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finding services matching GUID '%v'", serviceGuid)
	}
	if len(searchResults) == 0 {
		return nil, stacktrace.NewError("No services matched GUID '%v'", serviceGuid)
	}
	if len(searchResults) > 1 {
		return nil, stacktrace.NewError("Expected one service to match GUID '%v' but found %v", serviceGuid, len(searchResults))
	}
	result, found := searchResults[serviceGuid]
	if !found {
		return nil, stacktrace.NewError("Got results from searching for service with GUID '%v', but no results by the GUID we searched for; this is a bug in Kurtosis", serviceGuid)
	}
	return result, nil
}

func getEnclaveMatchLabels() map[string]string {
	matchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return matchLabels
}

func getUserServiceObjectsFromKubernetesResources(
	enclaveId enclave.EnclaveID,
	allKubernetesResources map[service.ServiceGUID]*UserServiceKubernetesResources,
) (map[service.ServiceGUID]*UserServiceObjectsAndKubernetesResources, error) {
	results := map[service.ServiceGUID]*UserServiceObjectsAndKubernetesResources{}
	for serviceGuid, resources := range allKubernetesResources {
		results[serviceGuid] = &UserServiceObjectsAndKubernetesResources{
			KubernetesResources: resources,
			// The other fields will get filled in below
		}
	}

	for serviceGuid, resultObj := range results {
		resourcesToParse := resultObj.KubernetesResources
		kubernetesService := resourcesToParse.Service
		kubernetesPod := resourcesToParse.Pod

		if kubernetesService == nil {
			return nil, stacktrace.NewError(
				"Service with GUID '%v' doesn't have a Kubernetes service; this indicates either a bug in Kurtosis or that the user manually deleted the Kubernetes service",
				serviceGuid,
			)
		}

		serviceLabels := kubernetesService.Labels
		idLabelStr, found := serviceLabels[label_key_consts.IDKubernetesLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find label '%v' on the Kubernetes service but none was found", label_key_consts.IDKubernetesLabelKey.GetString())
		}
		serviceId := service.ServiceID(idLabelStr)

		serviceIpStr := kubernetesService.Spec.ClusterIP
		privateIp := net.ParseIP(serviceIpStr)
		if privateIp == nil {
			return nil, stacktrace.NewError("An error occurred parsing service private IP string '%v' to an IP address object", serviceIpStr)
		}

		serviceRegistrationObj := service.NewServiceRegistration(serviceId, serviceGuid, enclaveId, privateIp)
		resultObj.ServiceRegistration = serviceRegistrationObj

		// A service with no ports annotation means that no pod has yet consumed the registration
		if _, found := kubernetesService.Annotations[kubernetes_annotation_key_consts.PortSpecsKubernetesAnnotationKey.GetString()]; !found {
			// If we're using the unbound port, no actual user ports have been set yet so there's no way we can
			// return a service
			resultObj.Service = nil
			continue
		}

		// From this point onwards, we're guaranteed that a pod was started at _some_ point; it may or may not still be running
		// Therefore, we know that there will be services registered

		// The empty map means "don't validate any port existence"
		privatePorts, err := GetPrivatePortsAndValidatePortExistence(kubernetesService, map[string]bool{})
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing private ports from the user service's Kubernetes service")
		}

		if kubernetesPod == nil {
			// No pod here means that a) a Service had private ports but b) now has no Pod
			// This means that there  used to be a Pod but it was stopped/removed
			resultObj.Service = service.NewService(
				serviceRegistrationObj,
				container_status.ContainerStatus_Stopped,
				privatePorts,
				servicePublicIp,
				servicePublicPorts,
			)
			continue
		}

		containerStatus, err := GetContainerStatusFromPod(resourcesToParse.Pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container status from Kubernetes pod '%+v'", resourcesToParse.Pod)
		}

		resultObj.Service = service.NewService(
			serviceRegistrationObj,
			containerStatus,
			privatePorts,
			servicePublicIp,
			servicePublicPorts,
		)
	}

	return results, nil
}

func ConvertMegabytesToBytes(value uint64) uint64 {
	return value * megabytesToBytesFactor
}
package shared_helpers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net"
	"strings"
	"time"
)

// !!!WARNING!!!
// This files contains functions that are shared by multiple DockerKurtosisBackend functions.
// Generally, we want to prevent long utils folders with functionality that is difficult to find, so be careful
// when adding functionality in this folder.
// Things to think about: Could this function be a private helper function that's scope is smaller than you think?
// Eg. only used by start user services functions thus could go in start_user_services.go

const (
	netstatSuccessExitCode = 0
)

// Kubernetes doesn't provide public IP or port information; this is instead handled by the Kurtosis gateway that the user uses
// to connect to Kubernetes
var servicePublicIp net.IP = nil
var servicePublicPorts map[string]*port_spec.PortSpec = nil

// TODO Remove this once we split apart the KubernetesKurtosisBackend into multiple backends (which we can only
//  do once the CLI no longer makes any calls directly to the KurtosisBackend, and instead makes all its calls through
//  the API container & engine APIs)

// The *Args structs SHOULD be private and in kubernetes_kurtosis_backend because they are closely tied to the KubernetesKurtosisBackend
// however, due to the way that GetEnclaveNamespaceName() works, these functions must exist here and made
// public so that they're accessible to that function
type CliModeArgs struct {
	// No CLI mode args needed for now
}

type ApiContainerModeArgs struct {
	ownEnclaveId enclave.EnclaveUUID

	ownNamespaceName string

	storageClassName string

	// TODO make this more dynamic - maybe guess based on the files artifact size?
	filesArtifactExpansionVolumeSizeInMegabytes uint
}

func NewApiContainerModeArgs(
	ownEnclaveId enclave.EnclaveUUID,
	ownNamespaceName string) *ApiContainerModeArgs {
	return &ApiContainerModeArgs{
		ownEnclaveId:     ownEnclaveId,
		ownNamespaceName: ownNamespaceName,
		storageClassName: "",
		filesArtifactExpansionVolumeSizeInMegabytes: 0,
	}
}

func (apiContainerModeArgs *ApiContainerModeArgs) GetOwnEnclaveId() enclave.EnclaveUUID {
	return apiContainerModeArgs.ownEnclaveId
}

func (apiContainerModeArgs *ApiContainerModeArgs) GetOwnNamespaceName() string {
	return apiContainerModeArgs.ownNamespaceName
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
	enclaveId enclave.EnclaveUUID,
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
		matchLabels[label_key_consts.EnclaveUUIDKubernetesLabelKey.GetString()] = string(enclaveId)

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
		if enclaveId != apiContainerModeArgs.ownEnclaveId {
			return "", stacktrace.NewError(
				"Received a request to get namespace for enclave '%v', but the Kubernetes Kurtosis backend is running in an API "+
					"container in a different enclave '%v' (so Kubernetes would throw a permission error)",
				enclaveId,
				apiContainerModeArgs.ownEnclaveId,
			)
		}
		namespaceName = apiContainerModeArgs.ownNamespaceName
	} else {
		return "", stacktrace.NewError("Received a request to get an enclave namespace's name, but the Kubernetes Kurtosis backend isn't in any recognized mode; this is a bug in Kurtosis")
	}

	return namespaceName, nil
}

func GetMatchingUserServiceObjectsAndKubernetesResources(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	cliModeArgs *CliModeArgs,
	apiContainerModeArgs *ApiContainerModeArgs,
	engineServerModeArgs *EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceUUID]*UserServiceObjectsAndKubernetesResources,
	error,
) {
	allResources, err := GetUserServiceKubernetesResourcesMatchingGuids(ctx, enclaveId, filters.UUIDs, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service Kubernetes resources matching UUIDs: %+v", filters.UUIDs)
	}

	for serviceUuid, serviceResources := range allResources {
		logrus.Tracef("Found resources for service '%v': %+v", serviceUuid, serviceResources)
	}

	allObjectsAndResources, err := GetUserServiceObjectsFromKubernetesResources(enclaveId, allResources)
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
	results := map[service.ServiceUUID]*UserServiceObjectsAndKubernetesResources{}
	for serviceUuid, objectsAndResources := range allObjectsAndResources {
		if filters.UUIDs != nil && len(filters.UUIDs) > 0 {
			if _, found := filters.UUIDs[serviceUuid]; !found {
				continue
			}
		}

		registration := objectsAndResources.ServiceRegistration
		if filters.Names != nil && len(filters.Names) > 0 {
			if _, found := filters.Names[registration.GetName()]; !found {
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

		results[serviceUuid] = objectsAndResources
	}

	return results, nil
}

func GetUserServiceKubernetesResourcesMatchingGuids(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuids map[service.ServiceUUID]bool,
	cliModeArgs *CliModeArgs,
	apiContainerModeArgs *ApiContainerModeArgs,
	engineServerModeArgs *EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceUUID]*UserServiceKubernetesResources,
	error,
) {
	namespaceName, err := GetEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	// TODO switch to properly-typed KubernetesLabelValue object!!!
	postFilterLabelValues := map[string]bool{}
	for serviceUuid := range serviceUuids {
		postFilterLabelValues[string(serviceUuid)] = true
	}

	kubernetesResourceSearchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.EnclaveUUIDKubernetesLabelKey.GetString():          string(enclaveId),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}

	results := map[service.ServiceUUID]*UserServiceKubernetesResources{}

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
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes services matching service UUIDs: %+v", serviceUuids)
	}
	for serviceGuidStr, kubernetesServicesForGuid := range matchingKubernetesServices {
		logrus.Tracef("Found Kubernetes services for GUID '%v': %+v", serviceGuidStr, kubernetesServicesForGuid)
		serviceUuid := service.ServiceUUID(serviceGuidStr)

		numServicesForGuid := len(kubernetesServicesForGuid)
		if numServicesForGuid == 0 {
			// This would indicate a bug in our service retrieval logic because we shouldn't even have a map entry if there's nothing matching it
			return nil, stacktrace.NewError("Got entry of result services for service GUID '%v', but no Kubernetes services were returned; this is a bug in Kurtosis", serviceUuid)
		}
		if numServicesForGuid > 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes services associated with service GUID '%v'; this is a bug in Kurtosis", numServicesForGuid, serviceUuid)
		}
		kubernetesService := kubernetesServicesForGuid[0]

		resultObj, found := results[serviceUuid]
		if !found {
			resultObj = &UserServiceKubernetesResources{
				Service: nil,
				Pod:     nil,
			}
		}
		resultObj.Service = kubernetesService
		results[serviceUuid] = resultObj
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
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes pods matching service UUIDs: %+v", serviceUuids)
	}
	for serviceGuidStr, kubernetesPodsForGuid := range matchingKubernetesPods {
		logrus.Tracef("Found Kubernetes pods for GUID '%v': %+v", serviceGuidStr, kubernetesPodsForGuid)
		serviceUuid := service.ServiceUUID(serviceGuidStr)

		numPodsForGuid := len(kubernetesPodsForGuid)
		if numPodsForGuid == 0 {
			// This would indicate a bug in our pod retrieval logic because we shouldn't even have a map entry if there's nothing matching it
			return nil, stacktrace.NewError("Got entry of result pods for service GUID '%v', but no Kubernetes pods were returned; this is a bug in Kurtosis", serviceUuid)
		}
		if numPodsForGuid > 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes pods associated with service GUID '%v'; this is a bug in Kurtosis", numPodsForGuid, serviceUuid)
		}
		kubernetesPod := kubernetesPodsForGuid[0]

		resultObj, found := results[serviceUuid]
		if !found {
			resultObj = &UserServiceKubernetesResources{
				Service: nil,
				Pod:     nil,
			}
		}
		resultObj.Pod = kubernetesPod
		results[serviceUuid] = resultObj
	}

	return results, nil
}

func GetUserServiceObjectsFromKubernetesResources(
	enclaveId enclave.EnclaveUUID,
	allKubernetesResources map[service.ServiceUUID]*UserServiceKubernetesResources,
) (map[service.ServiceUUID]*UserServiceObjectsAndKubernetesResources, error) {
	results := map[service.ServiceUUID]*UserServiceObjectsAndKubernetesResources{}
	for serviceUuid, resources := range allKubernetesResources {
		results[serviceUuid] = &UserServiceObjectsAndKubernetesResources{
			ServiceRegistration: nil,
			Service:             nil,
			KubernetesResources: resources,
		}
	}

	for serviceUuid, resultObj := range results {
		resourcesToParse := resultObj.KubernetesResources
		kubernetesService := resourcesToParse.Service
		kubernetesPod := resourcesToParse.Pod

		if kubernetesService == nil {
			return nil, stacktrace.NewError(
				"Service with GUID '%v' doesn't have a Kubernetes service; this indicates either a bug in Kurtosis or that the user manually deleted the Kubernetes service",
				serviceUuid,
			)
		}

		serviceLabels := kubernetesService.Labels
		idLabelStr, found := serviceLabels[label_key_consts.IDKubernetesLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find label '%v' on the Kubernetes service but none was found", label_key_consts.IDKubernetesLabelKey.GetString())
		}
		serviceId := service.ServiceName(idLabelStr)

		serviceIpStr := kubernetesService.Spec.ClusterIP
		privateIp := net.ParseIP(serviceIpStr)
		if privateIp == nil {
			return nil, stacktrace.NewError("An error occurred parsing service private IP string '%v' to an IP address object", serviceIpStr)
		}

		serviceRegistrationObj := service.NewServiceRegistration(
			serviceId,
			serviceUuid,
			enclaveId,
			privateIp,
			kubernetesService.GetName(), // Kubernetes automatically set hostname = Kubernetes Service Name
		)
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

func GetSingleUserServiceObjectsAndResources(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	cliModeArgs *CliModeArgs,
	apiContainerModeArgs *ApiContainerModeArgs,
	engineServerModeArgs *EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*UserServiceObjectsAndKubernetesResources, error) {
	searchFilters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		Statuses: nil,
	}
	searchResults, err := GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, searchFilters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finding services matching GUID '%v'", serviceUuid)
	}
	if len(searchResults) == 0 {
		return nil, stacktrace.NewError("No services matched GUID '%v'", serviceUuid)
	}
	if len(searchResults) > 1 {
		return nil, stacktrace.NewError("Expected one service to match GUID '%v' but found %v", serviceUuid, len(searchResults))
	}
	result, found := searchResults[serviceUuid]
	if !found {
		return nil, stacktrace.NewError("Got results from searching for service with UUID '%v', but no results by the GUID we searched for; this is a bug in Kurtosis", serviceUuid)
	}
	return result, nil
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
		isPodRunning, found := consts.IsPodRunningDeterminer[podPhase]
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

func GetStringMapFromLabelMap(labelMap map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue) map[string]string {
	strMap := map[string]string{}
	for labelKey, labelValue := range labelMap {
		strMap[labelKey.GetString()] = labelValue.GetString()
	}
	return strMap
}

func GetStringMapFromAnnotationMap(labelMap map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue) map[string]string {
	strMap := map[string]string{}
	for labelKey, labelValue := range labelMap {
		strMap[labelKey.GetString()] = labelValue.GetString()
	}
	return strMap
}

func GetKubernetesServicePortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ServicePort, error) {
	result := []apiv1.ServicePort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetTransportProtocol()
		kubernetesProtocol, found := consts.KurtosisTransportProtocolToKubernetesTransportProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ServicePort{
			Name:     portId,
			Protocol: kubernetesProtocol,
			// TODO Specify this!!! Will make for a really nice user interface (e.g. "https")
			AppProtocol: nil,
			// Safe to cast because max uint16 < int32
			Port: int32(portSpec.GetNumber()),
			TargetPort: intstr.IntOrString{
				Type:   0,
				IntVal: 0,
				StrVal: "",
			},
			NodePort: 0,
		}
		result = append(result, kubernetesPortObj)
	}
	return result, nil
}

func GetKubernetesContainerPortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ContainerPort, error) {
	result := []apiv1.ContainerPort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetTransportProtocol()
		kubernetesProtocol, found := consts.KurtosisTransportProtocolToKubernetesTransportProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ContainerPort{
			Name:     portId,
			HostPort: 0,
			// Safe to do because max uint16 < int32
			ContainerPort: int32(portSpec.GetNumber()),
			Protocol:      kubernetesProtocol,
			HostIP:        "",
		}
		result = append(result, kubernetesPortObj)
	}
	return result, nil
}

func WaitForPortAvailabilityUsingNetstat(
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespaceName string,
	podName string,
	containerName string,
	portSpec *port_spec.PortSpec,
	maxRetries uint,
	timeBetweenRetries time.Duration,
) error {
	commandStr := fmt.Sprintf(
		"[ -n \"$(netstat -anp %v | grep LISTEN | grep %v)\" ]",
		strings.ToLower(portSpec.GetTransportProtocol().String()),
		portSpec.GetNumber(),
	)
	execCmd := []string{
		"sh",
		"-c",
		commandStr,
	}
	for i := uint(0); i < maxRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		concurrentBuffer := concurrent_writer.NewConcurrentWriter(outputBuffer)
		exitCode, err := kubernetesManager.RunExecCommand(
			namespaceName,
			podName,
			containerName,
			execCmd,
			concurrentBuffer,
			concurrentBuffer,
		)
		if err == nil {
			if exitCode == netstatSuccessExitCode {
				return nil
			}
			logrus.Debugf(
				"Netstat availability-waiting command '%v' returned without a Kubernetes error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				netstatSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"Netstat availability-waiting command '%v' experienced a Kubernetes error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxRetries {
			time.Sleep(timeBetweenRetries)
		}
	}

	return stacktrace.NewError(
		"The port didn't become available (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxRetries,
		timeBetweenRetries,
	)
}

func GetMatchingUserServiceObjectsAndKubernetesResourcesByServiceName(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	cliModeArgs *CliModeArgs,
	apiContainerModeArgs *ApiContainerModeArgs,
	engineServerModeArgs *EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceName]*UserServiceObjectsAndKubernetesResources,
	error,
) {
	matchesByGUID, err := GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting matching user service objects and Kubernetes resources by service ID")
	}
	matchesByServiceName := map[service.ServiceName]*UserServiceObjectsAndKubernetesResources{}
	for _, userServiceObjectsAndKubernetesResource := range matchesByGUID {
		serviceName := userServiceObjectsAndKubernetesResource.ServiceRegistration.GetName()
		matchesByServiceName[serviceName] = userServiceObjectsAndKubernetesResource
	}
	return matchesByServiceName, nil
}

// ====================================================================================================
//
//	Private Helper Methods
//
// ====================================================================================================
func getEnclaveMatchLabels() map[string]string {
	matchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return matchLabels
}

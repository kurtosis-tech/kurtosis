package shared_helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/stacktrace"

	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// !!!WARNING!!!
// This files contains functions that are shared by multiple DockerKurtosisBackend functions.
// Generally, we want to prevent long utils folders with functionality that is difficult to find, so be careful
// when adding functionality in this folder.
// Things to think about: Could this function be a private helper function that's scope is smaller than you think?
// Eg. only used by start user services functions thus could go in start_user_services.go

const (
	netstatSuccessExitCode = 0
	httpSuccessExitCode    = 0

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

	emptyImageName = ""
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

type dumpPodResult struct {
	podName string
	err     error
}

func NewApiContainerModeArgs(
	ownEnclaveId enclave.EnclaveUUID,
	ownNamespaceName string, storageClassName string) *ApiContainerModeArgs {
	return &ApiContainerModeArgs{
		ownEnclaveId:     ownEnclaveId,
		ownNamespaceName: ownNamespaceName,
		storageClassName: storageClassName,
		filesArtifactExpansionVolumeSizeInMegabytes: 0,
	}
}

func (apiContainerModeArgs *ApiContainerModeArgs) GetOwnEnclaveId() enclave.EnclaveUUID {
	return apiContainerModeArgs.ownEnclaveId
}

func (apiContainerModeArgs *ApiContainerModeArgs) GetOwnNamespaceName() string {
	return apiContainerModeArgs.ownNamespaceName
}

// EngineServerModeArgs TODO(victor.colombo): Can we remove this?
type EngineServerModeArgs struct{}

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

	// This can be nil if the user hasn't started the service yet, or if the ingress was deleted
	Ingress *netv1.Ingress
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
		matchLabels[kubernetes_label_key.EnclaveUUIDKubernetesLabelKey.GetString()] = string(enclaveId)

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
		if len(filters.UUIDs) > 0 {
			if _, found := filters.UUIDs[serviceUuid]; !found {
				continue
			}
		}

		registration := objectsAndResources.ServiceRegistration
		if len(filters.Names) > 0 {
			if _, found := filters.Names[registration.GetName()]; !found {
				continue
			}
		}

		if len(filters.Statuses) > 0 {
			kubernetesService := objectsAndResources.Service

			// If status isn't specified, return registered-only objects; if not, remove them all
			if kubernetesService == nil {
				continue
			}

			if _, found := filters.Statuses[kubernetesService.GetContainer().GetStatus()]; !found {
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
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		kubernetes_label_key.EnclaveUUIDKubernetesLabelKey.GetString():          string(enclaveId),
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}

	results := map[service.ServiceUUID]*UserServiceKubernetesResources{}

	// Get k8s services
	matchingKubernetesServices, err := kubernetes_resource_collectors.CollectMatchingServices(
		ctx,
		kubernetesManager,
		namespaceName,
		kubernetesResourceSearchLabels,
		kubernetes_label_key.GUIDKubernetesLabelKey.GetString(),
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
				Ingress: nil,
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
		kubernetes_label_key.GUIDKubernetesLabelKey.GetString(),
		postFilterLabelValues,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes pods matching service UUIDs: %+v", serviceUuids)
	}
	for serviceGuidStr, kubernetesPodsForGuid := range matchingKubernetesPods {
		logrus.Tracef("Found Kubernetes pods for GUID '%v': %+v", serviceGuidStr, kubernetesPodsForGuid)
		serviceUuid := service.ServiceUUID(serviceGuidStr)

		numPodsForGuid := len(kubernetesPodsForGuid)
		if numPodsForGuid > 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes pods associated with service GUID '%v'; this is a bug in Kurtosis", numPodsForGuid, serviceUuid)
		} else if numPodsForGuid == 1 {
			kubernetesPod := kubernetesPodsForGuid[0]

			numContainersForPod := len(kubernetesPod.Spec.Containers)
			if numContainersForPod != 1 {
				return nil, stacktrace.NewError("Found %v containers associated with service GUID '%v'; this is a bug in Kurtosis", numContainersForPod, serviceUuid)
			}

			resultObj, found := results[serviceUuid]
			if !found {
				resultObj = &UserServiceKubernetesResources{
					Service: nil,
					Pod:     nil,
					Ingress: nil,
				}
			}
			resultObj.Pod = kubernetesPod
			results[serviceUuid] = resultObj
		}
	}

	// Get k8s ingresses
	matchingKubernetesIngresses, err := kubernetes_resource_collectors.CollectMatchingIngresses(
		ctx,
		kubernetesManager,
		namespaceName,
		kubernetesResourceSearchLabels,
		kubernetes_label_key.GUIDKubernetesLabelKey.GetString(),
		postFilterLabelValues,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes ingresses matching service UUIDs: %+v", serviceUuids)
	}
	for serviceGuidStr, kubernetesIngressForGuid := range matchingKubernetesIngresses {
		logrus.Tracef("Found Kubernetes ingress for GUID '%v': %+v", serviceGuidStr, kubernetesIngressForGuid)
		serviceUuid := service.ServiceUUID(serviceGuidStr)

		numIngressesForGuid := len(kubernetesIngressForGuid)
		if numIngressesForGuid != 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes ingresses associated with service GUID '%v', but number of ingresses should be exactly 1; this is a bug in Kurtosis", numIngressesForGuid, serviceUuid)
		}
		kubernetesIngress := kubernetesIngressForGuid[0]

		resultObj, found := results[serviceUuid]
		if !found {
			resultObj = &UserServiceKubernetesResources{
				Service: nil,
				Pod:     nil,
				Ingress: nil,
			}
		}
		resultObj.Ingress = kubernetesIngress
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
		idLabelStr, found := serviceLabels[kubernetes_label_key.IDKubernetesLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find label '%v' on the Kubernetes service but none was found", kubernetes_label_key.IDKubernetesLabelKey.GetString())
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
				privatePorts,
				servicePublicIp,
				servicePublicPorts,
				container.NewContainer(
					container.ContainerStatus_Stopped,
					emptyImageName,
					nil,
					nil,
					nil),
			)
			continue
		}

		containerStatus, err := GetContainerStatusFromPod(resourcesToParse.Pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container status from Kubernetes pod '%+v'", resourcesToParse.Pod)
		}

		podContainer := resourcesToParse.Pod.Spec.Containers[0]
		podContainerEnvVars := map[string]string{}
		for _, env := range podContainer.Env {
			podContainerEnvVars[env.Name] = env.Value
		}

		resultObj.Service = service.NewService(
			serviceRegistrationObj,
			privatePorts,
			servicePublicIp,
			servicePublicPorts,
			container.NewContainer(
				containerStatus,
				podContainer.Image,
				podContainer.Command,
				podContainer.Args,
				podContainerEnvVars,
			),
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

	if len(expectedPortIds) > 0 {
		for portId := range expectedPortIds {
			if _, found := privatePortSpecs[portId]; !found {
				return nil, stacktrace.NewError("Missing private port with ID '%v' in the private ports", portId)
			}
		}
	}
	return privatePortSpecs, nil
}

func GetContainerStatusFromPod(pod *apiv1.Pod) (container.ContainerStatus, error) {
	// TODO Rename this; this shouldn't be called "ContainerStatus" since there's no longer a 1:1 mapping between container:kurtosis_object
	status := container.ContainerStatus_Stopped

	if pod != nil {
		podPhase := pod.Status.Phase
		isPodRunning, found := consts.IsPodRunningDeterminer[podPhase]
		if !found {
			// This should never happen because we enforce completeness in a unit test
			return status, stacktrace.NewError("No is-pod-running determination found for pod phase '%v' on pod '%v'; this is a bug in Kurtosis", podPhase, pod.Name)
		}
		if isPodRunning {
			status = container.ContainerStatus_Running
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
		if i < maxRetries-1 {
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

// DumpNamespacePods dump pods from the same namespace
func DumpNamespacePods(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespace *apiv1.Namespace,
	podsToDump []apiv1.Pod,
	outputDirpath string,
) error {
	// Create output directory
	if _, err := os.Stat(outputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", outputDirpath)
	}
	if err := os.Mkdir(outputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory at '%v'", outputDirpath)
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
			kubernetesManager,
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
		combinedErr := BuildCombinedError(resultErrorsByPodName, dumpPodErrorTitle)
		return combinedErr
	}
	return nil
}

// This is a helper function that will take multiple errors, each identified by an ID, and format them together
// If no errors are returned, this function returns nil
func BuildCombinedError(errorsById map[string]error, titleStr string) error {
	allErrorStrs := []string{}
	for errorId, stopErr := range errorsById {
		errorFormatStr := ">>>>>>>>>>>>> %v %v <<<<<<<<<<<<<\n" +
			"%v\n" +
			">>>>>>>>>>>>> END %v %v <<<<<<<<<<<<<"
		errorStr := fmt.Sprintf(
			errorFormatStr,
			strings.ToUpper(titleStr),
			errorId,
			stopErr.Error(),
			strings.ToUpper(titleStr),
			errorId,
		)
		allErrorStrs = append(allErrorStrs, errorStr)
	}

	if len(allErrorStrs) > 0 {
		// NOTE: This is one of the VERY rare cases where we don't want to use stacktrace.Propagate, because
		// attaching stack information for this method (which simply combines errors) just isn't useful. The
		// expected behaviour is that the caller of this function will use stacktrace.Propagate
		return errors.New(strings.Join(
			allErrorStrs,
			"\n\n",
		))
	}

	return nil
}

// ====================================================================================================
//
//	Private Helper Methods
//
// ====================================================================================================
func getEnclaveMatchLabels() map[string]string {
	matchLabels := map[string]string{
		kubernetes_label_key.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return matchLabels
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
	if err := os.WriteFile(podSpecOutputFilepath, jsonSerializedPodSpecBytes, createdFilePerms); err != nil {
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
				"An error occurred getting logs of container '%v' in pod '%v' in namespace '%v'",
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

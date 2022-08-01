package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_functions"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

const (
	// Kubernetes doesn't allow us to create services without ports exposed, but we might not have ports in the following situations:
	//  1) we've registered a service but haven't started a container yet (so ports are yet to come)
	//  2) we've started a container that doesn't listen on any ports
	// In these cases, we use these notional unbound ports
	unboundPortName   = "nonexistent-port"
	unboundPortNumber = 1
)

// This maps a Kubernetes pod's phase to a binary "is the pod considered running?" determiner
// Its completeness is enforced via unit test
var isPodRunningDeterminer = map[apiv1.PodPhase]bool{
	apiv1.PodPending: true,
	apiv1.PodRunning: true,
	apiv1.PodSucceeded: false,
	apiv1.PodFailed: false,
	apiv1.PodUnknown: false, //We cannot say that a pod is not running if we don't know the real state
}

// Kubernetes doesn't provide public IP or port information; this is instead handled by the Kurtosis gateway that the user uses
// to connect to Kubernetes
var servicePublicIp net.IP = nil
var servicePublicPorts map[string]*port_spec.PortSpec = nil

type userServiceObjectsAndKubernetesResources struct {
	// Should never be nil because 1 Kubernetes service = 1 Kurtosis service registration
	serviceRegistration *service.ServiceRegistration

	// May be nil if no pod has been started yet
	service *service.Service

	// Will never be nil
	kubernetesResources *userServiceKubernetesResources
}

// Any of these fields can be nil if they don't exist in Kubernetes, though at least
// one field will be present (else this struct won't exist)
type userServiceKubernetesResources struct {
	// This can be nil if the user manually deleted the Kubernetes service (e.g. using the Kubernetes dashboard)
	service *apiv1.Service

	// This can be nil if the user hasn't started a pod for the service yet, or if the pod was deleted
	pod *apiv1.Pod
}

func RegisterUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceID,
	cliModeArgs *shared_functions.CliModeArgs,
	apiContainerModeArgs *shared_functions.ApiContainerModeArgs,
	engineServerModeArgs *shared_functions.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager) (*service.ServiceRegistration, error) {
	namespaceName, err := getEnclaveNamespaceName(ctx, enclaveId, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	serviceGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID to use for the service GUID")
	}
	serviceGuid := service.ServiceGUID(serviceGuidStr)

	objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveId)

	serviceAttributes, err := enclaveObjAttributesProvider.ForUserServiceService(serviceGuid, serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting attributes for the Kubernetes service for user service '%v'", serviceId)
	}

	serviceNameStr := serviceAttributes.GetName().GetString()

	serviceLabelsStrs := getStringMapFromLabelMap(serviceAttributes.GetLabels())
	serviceAnnotationsStrs := getStringMapFromAnnotationMap(serviceAttributes.GetAnnotations())

	// Set up the labels that the pod will match (i.e. the labels of the pod-to-be)
	// WARNING: We *cannot* use the labels of the Service itself because we're not guaranteed that the labels
	//  between the two will be identical!
	serviceGuidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(serviceGuid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the service GUID '%v'", serviceGuid)
	}
	enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(enclaveId))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the enclave ID '%v'", enclaveId)
	}
	matchedPodLabels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.AppIDKubernetesLabelKey:     label_value_consts.AppIDKubernetesLabelValue,
		label_key_consts.EnclaveIDKubernetesLabelKey: enclaveIdLabelValue,
		label_key_consts.GUIDKubernetesLabelKey:      serviceGuidLabelValue,
	}
	matchedPodLabelStrs := getStringMapFromLabelMap(matchedPodLabels)

	// Kubernetes doesn't allow us to create services without any ports, so we need to set this to a notional value
	// until the user calls StartService
	notionalServicePorts := []apiv1.ServicePort{
		{
			Name: unboundPortName,
			Port: unboundPortNumber,
		},
	}

	createdService, err := kubernetesManager.CreateService(
		ctx,
		namespaceName,
		serviceNameStr,
		serviceLabelsStrs,
		serviceAnnotationsStrs,
		matchedPodLabelStrs,
		apiv1.ServiceTypeClusterIP,
		notionalServicePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes service in enclave '%v' with ID '%v'", enclaveId, serviceId)
	}
	shouldDeleteService := true
	defer func() {
		if shouldDeleteService {
			if err := kubernetesManager.RemoveService(ctx, createdService); err != nil {
				logrus.Errorf("Registering service '%v' didn't complete successfully so we tried to remove the Kubernetes service we created but doing so threw an error:\n%v", serviceId, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to remove service '%v' in namespace '%v' manually!!!", createdService.Name, namespaceName)
			}
		}
	}()

	kubernetesResources := map[service.ServiceGUID]*userServiceKubernetesResources{
		serviceGuid: {
			service: createdService,
			pod:     nil, // No pod yet
		},
	}

	convertedObjects, err := getUserServiceObjectsFromKubernetesResources(enclaveId, kubernetesResources)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a service registration object from Kubernetes service")
	}
	objectsAndResources, found := convertedObjects[serviceGuid]
	if !found {
		return nil, stacktrace.NewError(
			"Successfully converted the Kubernetes service representing registered service with GUID '%v' to a "+
				"Kurtosis object, but couldn't find that key in the resulting map; this is a bug in Kurtosis",
			serviceGuid,
		)
	}
	serviceRegistration := objectsAndResources.serviceRegistration

	shouldDeleteService = false
	return serviceRegistration, nil
}

// Registers a user service for each given serviceId, allocating each an IP and ServiceGUID
func RegisterUserServices(ctx context.Context, enclaveId enclave.EnclaveID, serviceIds map[service.ServiceID]bool, ) (map[service.ServiceID]*service.ServiceRegistration, map[service.ServiceID]error, error){
	return nil, nil, stacktrace.NewError("REGISTER USER SERVICES METHOD IS UNIMPLEMENTED. DON'T USE IT")
}

func getEnclaveNamespaceName(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	cliModeArgs *shared_functions.CliModeArgs,
	apiContainerModeArgs *shared_functions.ApiContainerModeArgs,
	engineServerModeArgs *shared_functions.EngineServerModeArgs,
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

func getStringMapFromLabelMap(labelMap map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue) map[string]string {
	strMap := map[string]string{}
	for labelKey, labelValue := range labelMap {
		strMap[labelKey.GetString()] = labelValue.GetString()
	}
	return strMap
}

func getStringMapFromAnnotationMap(labelMap map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue) map[string]string {
	strMap := map[string]string{}
	for labelKey, labelValue := range labelMap {
		strMap[labelKey.GetString()] = labelValue.GetString()
	}
	return strMap
}

func getUserServiceObjectsFromKubernetesResources(
	enclaveId enclave.EnclaveID,
	allKubernetesResources map[service.ServiceGUID]*userServiceKubernetesResources,
) (map[service.ServiceGUID]*userServiceObjectsAndKubernetesResources, error) {
	results := map[service.ServiceGUID]*userServiceObjectsAndKubernetesResources{}
	for serviceGuid, resources := range allKubernetesResources {
		results[serviceGuid] = &userServiceObjectsAndKubernetesResources{
			kubernetesResources: resources,
			// The other fields will get filled in below
		}
	}

	for serviceGuid, resultObj := range results {
		resourcesToParse := resultObj.kubernetesResources
		kubernetesService := resourcesToParse.service
		kubernetesPod := resourcesToParse.pod

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
		resultObj.serviceRegistration = serviceRegistrationObj

		// A service with no ports annotation means that no pod has yet consumed the registration
		if _, found := kubernetesService.Annotations[kubernetes_annotation_key_consts.PortSpecsKubernetesAnnotationKey.GetString()]; !found {
			// If we're using the unbound port, no actual user ports have been set yet so there's no way we can
			// return a service
			resultObj.service = nil
			continue
		}

		// From this point onwards, we're guaranteed that a pod was started at _some_ point; it may or may not still be running
		// Therefore, we know that there will be services registered

		// The empty map means "don't validate any port existence"
		privatePorts, err := getPrivatePortsAndValidatePortExistence(kubernetesService, map[string]bool{})
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing private ports from the user service's Kubernetes service")
		}

		if kubernetesPod == nil {
			// No pod here means that a) a Service had private ports but b) now has no Pod
			// This means that there  used to be a Pod but it was stopped/removed
			resultObj.service = service.NewService(
				serviceRegistrationObj,
				container_status.ContainerStatus_Stopped,
				privatePorts,
				servicePublicIp,
				servicePublicPorts,
			)
			continue
		}

		containerStatus, err := getContainerStatusFromPod(resourcesToParse.pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container status from Kubernetes pod '%+v'", resourcesToParse.pod)
		}

		resultObj.service = service.NewService(
			serviceRegistrationObj,
			containerStatus,
			privatePorts,
			servicePublicIp,
			servicePublicPorts,
		)
	}

	return results, nil
}

// If no expected-ports list is passed in, no validation is done and all the ports are passed back as-is
func getPrivatePortsAndValidatePortExistence(kubernetesService *apiv1.Service, expectedPortIds map[string]bool) (map[string]*port_spec.PortSpec, error) {
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

func getContainerStatusFromPod(pod *apiv1.Pod) (container_status.ContainerStatus, error) {
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

func getEnclaveMatchLabels() map[string]string {
	matchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString():                label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}
	return matchLabels
}
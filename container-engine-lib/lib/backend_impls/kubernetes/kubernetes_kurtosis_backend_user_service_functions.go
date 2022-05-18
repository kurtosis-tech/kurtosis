package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	apiv1 "k8s.io/api/core/v1"
	"net"
	"time"
)

/*
      *************************************************************************************************
      * See the documentation on the KurtosisBackend interface for the Kurtosis service state diagram *
      *************************************************************************************************

                                  KUBERNETES SERVICES IMPLEMENTATION

States:
- REGISTERED: a ServiceGUID has been generated, a Kubernetes Service has been created with that ServiceGUID as a label,
	the Kubernetes Service gets it selector set to look for pods with that ServiceGUID (of which there will be none at
	first), and the ClusterIP of the Service is returned to the caller.
- RUNNING: one or more Pods match the ServiceGUID selector of the Service.
- STOPPED = the Service's selectors are set to empty (this is to distinguish between "no Pods exist because they haven't
	been started yet", i.e. REGISTERED, and "no Pods exist because they were stopped", i.e. STOPPED).
- DESTROYED = no Service or Pod for the ServiceGUID exists.

State transitions:
- RegisterService: a Service is created with the ServiceGUID label (for finding later), the Service's selectors are set
	to look for Pods with the ServiceGUID, and the Service's ClusterIP is returned to the caller.
- StartService: a Pod is created with the ServiceGUID, which will hook up the Service to the Pod.
- StopServices: The Pod is destroyed (because Kubernetes doesn't have a way to stop Pods - only remove them) and the Service's
	selectors are set to empty. If we want to keep Kubernetes logs around after a Pod is destroyed, we'll need to implement
	our own logstore.
- DestroyServices: the Service is destroyed, as is any Pod that may have been started.

Implementation notes:
- Kubernetes has no concept of stopping a pod without removing it; if we want Pod logs we'll need to implement our own logstore.
- The IP that the container gets as its "own IP" is technically the IP of the Service. This *should* be fine, but we'll need
	to keep an eye out for it.
*/

const (
	userServiceContainerName = "user-service-container"

	shouldMountVolumesAsReadOnly = false

	// Our user services don't need service accounts
	userServiceServiceAccountName = ""
)

// Any of these fields can be nil if they don't exist in Kubernetes, though at least
// one field will be present (else this struct won't exist)
type userServiceKubernetesResources struct {
	// This can be nil if the service has a pod and the user manually deleted the
	service *apiv1.Service

	// This can be nil if the user hasn't started a pod for the service yet, or if the pod was deleted
	pod *apiv1.Pod
}

func (backend *KubernetesKurtosisBackend) RegisterUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceId service.ServiceID) (*service.ServiceRegistration, error) {
	enclaveNamespace, err := backend.getEnclaveNamespace(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace matching enclave ID '%v'", enclaveId)
	}
	namespaceName := enclaveNamespace.Name

	// TODO Switch this, and all other GUIDs, to a UUID??
	serviceGuid := service.ServiceGUID(fmt.Sprintf(
		"%v-%v",
		serviceId,
		time.Now().Unix(),
	))

	/*
	preexistingServiceFilters := &service.ServiceFilters{
		IDs: map[service.ServiceID]bool{
			serviceId: true,
		},
	}
	preexistingRegistrations, _, preexistingResources, err := backend.getMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, preexistingServiceFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for preexisting services in enclave '%v' matching service ID '%v'", enclaveId, serviceId)
	}
	if len(preexistingRegistrations) > 0 {
		return nil, stacktrace.NewError("Found preexisting registration in enclave '%v' with ID '%v'", enclaveId, serviceId)
	}
	 */

	objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveId)

	serviceAttributes, err := enclaveObjAttributesProvider.ForUserServiceService(serviceGuid, serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting attributes for the Kubernetes service for user service '%v'", serviceId)
	}

	serviceNameStr := serviceAttributes.GetName().GetString()

	serviceLabelsStrs := getStringMapFromLabelMap(serviceAttributes.GetLabels())
	serviceAnnotationsStrs := getStringMapFromAnnotationMap(serviceAttributes.GetAnnotations())

	serviceGuidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(serviceGuid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes label value for the service GUID '%v'", serviceGuid)
	}
	matchedPodLabels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.GUIDKubernetesLabelKey: serviceGuidLabelValue,
	}
	matchedPodLabelStrs := getStringMapFromLabelMap(matchedPodLabels)

	createdService, err := backend.kubernetesManager.CreateService(
		ctx,
		enclaveNamespace.Name,
		serviceNameStr,
		serviceLabelsStrs,
		serviceAnnotationsStrs,
		matchedPodLabelStrs,
		apiv1.ServiceTypeClusterIP,
		[]apiv1.ServicePort{},	// This will be filled out when the user starts a pod
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes service in enclave '%v' with ID '%v'", enclaveId, serviceId)
	}
	shouldDeleteService := true
	defer func() {
		if shouldDeleteService {
			if err := backend.kubernetesManager.RemoveService(ctx, namespaceName, createdService.Name); err != nil {
				logrus.Errorf("Registering service '%v' didn't complete successfully so we tried to remove the Kubernetes service we created but doing so threw an error:\n%v", serviceId, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to remove service '%v' in namespace '%v' manually!!!", createdService.Name, namespaceName)
			}
		}
	}()

	serviceRegistration, err := backend.getServiceRegistrationObjectFromKubernetesService(createdService)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a service registration object from Kubernetes service")
	}

	shouldDeleteService = false
	return serviceRegistration, nil
}

func (backend *KubernetesKurtosisBackend) StartUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	containerImageName string,
	privatePorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	// TODO This needs a redo - at minimum it should be by files_artifact_expansion_volume_guid
	filesArtifactVolumeMountDirpaths map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]string,
) (
	newUserService *service.Service,
	resultErr error,
) {
	namespace, err := backend.getEnclaveNamespace(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace for enclave '%v'", enclaveId)
	}
	namespaceName := namespace.Name

	serviceSearchLabels := map[string]string{
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue.GetString(),
		label_key_consts.EnclaveIDKubernetesLabelKey.GetString(): string(enclaveId),
		label_key_consts.GUIDKubernetesLabelKey.GetString(): string(serviceGuid),
	}
	matchingServicesList, err := backend.kubernetesManager.GetServicesByLabels(ctx, namespaceName, serviceSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting preexisting service registrations matching GUID '%v'", serviceGuid)
	}
	matchingServices := matchingServicesList.Items
	if len(matchingServices) == 0 {
		return nil, stacktrace.NewError("Couldn't find any service registrations matching GUID '%v'", serviceGuid)
	}
	if len(matchingServices) > 1 {
		return nil, stacktrace.NewError("Found multiple service registrations matching GUID '%v'", serviceGuid)
	}
	kubernetesService := matchingServices[0]

	if len(kubernetesService.Spec.Selector) == 0 {
		return nil, stacktrace.NewError("Cannot start service with GUID '%v' because the service has already been stopped")
	}

	serviceRegistration, err := backend.getServiceRegistrationObjectFromKubernetesService(&kubernetesService)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a service registration object from the Kubernetes service")
	}

	// TODO Validate that no pods are attached to the Service already!!!

	objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveId)
	podAttributes, err := enclaveObjAttributesProvider.ForUserServicePod(
		serviceRegistration.GetGUID(),
		serviceRegistration.GetID(),
		privatePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting attributes for the Kubernetes service for user service '%v'", serviceRegistration.GetGUID())
	}

	podLabelsStrs := getStringMapFromLabelMap(podAttributes.GetLabels())
	podAnnotationsStrs := getStringMapFromAnnotationMap(podAttributes.GetAnnotations())

	podContainers, err := getUserServicePodContainerSpecs(
		containerImageName,
		entrypointArgs,
		cmdArgs,
		envVars,
		privatePorts,
		filesArtifactVolumeMountDirpaths,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the container specs for the user service pod with image '%v'", containerImageName)
	}

	podVolumes, err := backend.getUserServicePodVolumesFromFilesArtifactMountpoints(
		ctx,
		namespaceName,
		enclaveId,
		serviceGuid,
		filesArtifactVolumeMountDirpaths,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting pod volumes from files artifact mountpoints: %+v", filesArtifactVolumeMountDirpaths)
	}

	podName := podAttributes.GetName().GetString()
	pod, err := backend.kubernetesManager.CreatePod(
		ctx,
		namespaceName,
		podName,
		podLabelsStrs,
		podAnnotationsStrs,
		podContainers,
		podVolumes,
		userServiceServiceAccountName,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating pod '%v' using image '%v'", podName, containerImageName)
	}
	shouldDestroyPod := true
	defer func() {
		if shouldDestroyPod {
			if err := backend.kubernetesManager.RemovePod(ctx, namespaceName, podName); err != nil {
				logrus.Errorf("Starting service didn't complete successfully so we tried to remove the pod we created but doing so threw an error:\n%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to remove pod '%v' in '%v' manually!!!", podName, namespaceName)
			}
		}
	}()


	// TODO update service




	shouldDestroyPod = false
	return nil, stacktrace.NewError("TODO IMPLEMENT!")
}

func (backend *KubernetesKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (successfulUserServices map[service.ServiceGUID]*service.Service, resultError error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (successfulUserServiceLogs map[service.ServiceGUID]io.ReadCloser, erroredUserServiceGuids map[service.ServiceGUID]error, resultError error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceGUID,
) error {
	return stacktrace.NewError("Cannot pause service '%v' in enclave '%v' because pausing is not supported by Kubernetes", serviceId, enclaveId)
}

func (backend *KubernetesKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceGUID,
) error {
	return stacktrace.NewError("Cannot pause service '%v' in enclave '%v' because unpausing is not supported by Kubernetes", serviceId, enclaveId)
}

func (backend *KubernetesKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
) (
	succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) GetConnectionWithUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceGUID service.ServiceGUID) (resultConn net.Conn, resultErr error) {
	// TODO IMPLEMENT
	return nil, stacktrace.NewError("Getting a connection with a user service isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) CopyFromUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceGuid service.ServiceGUID, srcPath string) (resultReadCloser io.ReadCloser, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) StopUserServices(ctx context.Context, enclaveId enclave.EnclaveID, filters *service.ServiceFilters) (successfulUserServiceGuids map[service.ServiceGUID]bool, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}

func (backend *KubernetesKurtosisBackend) DestroyUserServices(ctx context.Context, enclaveId enclave.EnclaveID, filters *service.ServiceFilters) (successfulUserServiceGuids map[service.ServiceGUID]bool, erroredUserServiceGuids map[service.ServiceGUID]error, resultErr error) {
	// TODO Destroy persistent volume claims (???)

	//TODO implement me
	panic("implement me")
}


// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingUserServiceKubernetesResources(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*userServiceKubernetesResources,
	error,
) {

	// TODO TODO CHANGE
	return nil, stacktrace.NewError("TODO IMPLEMENT!!")
}

func (backend *KubernetesKurtosisBackend) getUserServiceObjectsFromKubernetesResources(
	resources map[service.ServiceGUID]*userServiceKubernetesResources,
) (
	map[service.ServiceGUID]*service.ServiceRegistration,
	map[service.ServiceGUID]*service.Service,
	error,
){

	// TODO TODO CHANGE
	return nil, stacktrace.NewError("TODO IMPLEMENT!!")
}

func (backend *KubernetesKurtosisBackend) getMatchingUserServiceObjectsAndKubernetesResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.ServiceRegistration,
	map[service.ServiceGUID]*service.Service,
	map[service.ServiceGUID]*userServiceKubernetesResources,
	error,
) {
	_, err := backend.getMatchingUserServiceKubernetesResources(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engine Kubernetes resources matching IDs: %+v", filters.IDs)
	}


	// TODO TODO CHANGE
	return nil, nil, stacktrace.NewError("TODO IMPLEMENT!!")
}

func (backend *KubernetesKurtosisBackend) getServiceRegistrationObjectFromKubernetesService(kubernetesService *apiv1.Service) (*service.ServiceRegistration, error) {
	labels := kubernetesService.Labels
	guidLabelStr, found := labels[label_key_consts.GUIDKubernetesLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find label '%v' on the Kubernetes service but none was found", label_key_consts.GUIDKubernetesLabelKey.GetString())
	}
	guid := service.ServiceGUID(guidLabelStr)

	idLabelStr, found := labels[label_key_consts.IDKubernetesLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find label '%v' on the Kubernetes service but none was found", label_key_consts.IDKubernetesLabelKey.GetString())
	}
	id := service.ServiceID(idLabelStr)

	enclaveIdLabelStr, found := labels[label_key_consts.EnclaveIDKubernetesLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find label '%v' on the Kubernetes service but none was found", label_key_consts.EnclaveIDKubernetesLabelKey.GetString())
	}
	enclaveId := enclave.EnclaveID(enclaveIdLabelStr)

	serviceIpStr := kubernetesService.Spec.ClusterIP
	serviceIp := net.ParseIP(serviceIpStr)
	if serviceIp == nil {
		return nil, stacktrace.NewError("An error occurred parsing service IP string '%v' to an IP address object", serviceIpStr)
	}

	return service.NewServiceRegistration(id, guid, enclaveId, serviceIp), nil
}

func getUserServicePodContainerSpecs(
	image string,
	entrypointArgs []string,
	cmdArgs []string,
	envVarStrs map[string]string,
	privatePorts map[string]*port_spec.PortSpec,
	filesArtifactMountpoints map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]string,
) (
	[]apiv1.Container,
	error,
){
	var containerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVarStrs {
		envVar := apiv1.EnvVar{
			Name:  varName,
			Value: varValue,
		}
		containerEnvVars = append(containerEnvVars, envVar)
	}

	kubernetesContainerPorts, err := getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes container ports from the private port specs map")
	}

	volumeMountsList := []apiv1.VolumeMount{}
	for volumeName, mountpoint := range filesArtifactMountpoints {
		volumeMountObj := apiv1.VolumeMount{
			Name:             string(volumeName),
			ReadOnly:         shouldMountVolumesAsReadOnly,
			MountPath:        mountpoint,
		}
		volumeMountsList = append(volumeMountsList, volumeMountObj)
	}

	// TODO create networking sidecars here
	containers := []apiv1.Container{
		{
			Name:                     userServiceContainerName,
			Image:                    image,
			// Yes, even though this is called "command" it actually corresponds to the Docker ENTRYPOINT
			Command:                  entrypointArgs,
			Args:                     cmdArgs,
			Ports:                    kubernetesContainerPorts,
			Env:                      containerEnvVars,
			VolumeMounts:             volumeMountsList,

			// NOTE: There are a bunch of other interesting Container options that we omitted for now but might
			// want to specify in the future
		},
	}

	return containers, nil
}

func (backend *KubernetesKurtosisBackend) getUserServicePodVolumesFromFilesArtifactMountpoints(
	ctx context.Context,
	namespaceName string,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	filesArtifactVolumeMountpoints map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]string,
) ([]apiv1.Volume, error) {
	filesArtifactExpansionVolumeSearchLabels := map[string]string{
		label_key_consts.EnclaveIDKubernetesLabelKey.GetString(): string(enclaveId),
		label_key_consts.KurtosisVolumeTypeKubernetesLabelKey.GetString(): label_value_consts.FilesArtifactExpansionVolumeTypeKubernetesLabelValue.GetString(),
		label_key_consts.UserServiceGUIDKubernetesLabelKey.GetString(): string(serviceGuid),
	}
	persistentVolumeClaimsList, err := backend.kubernetesManager.GetPersistentVolumeClaimsByLabels(ctx, namespaceName, filesArtifactExpansionVolumeSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact expansion volumes for service '%v' matching labels: %+v", filesArtifactExpansionVolumeSearchLabels)
	}
	persistentVolumeClaims := persistentVolumeClaimsList.Items

	if len(persistentVolumeClaims) != len(filesArtifactVolumeMountpoints) {
		return nil, stacktrace.Propagate(
			err,
			"Received request to start service with %v files artifact mountpoints '%+v', but only found %v persistent volume claims prepared for the service",
			len(filesArtifactVolumeMountpoints),
			filesArtifactVolumeMountpoints,
			len(persistentVolumeClaims),
		)
	}

	result := []apiv1.Volume{}
	for _, claim := range persistentVolumeClaims {
		claimName := claim.Name

		volumeObj := apiv1.Volume{
			Name:         claim.Spec.VolumeName,
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
				},
			},
		}

		result = append(result, volumeObj)
	}

	return result, nil
}
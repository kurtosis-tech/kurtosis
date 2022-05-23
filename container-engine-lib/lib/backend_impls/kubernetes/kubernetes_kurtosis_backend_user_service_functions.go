package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	apiv1 "k8s.io/api/core/v1"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"net"
	"strconv"
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
- StartService:
	1. a Pod is created with the ServiceGUID
	2. the Kubernetes Service gets its ports updated with the requested ports
	3. the Kubernetes Service gets the port specs serialized to JSON and stored as an annotation
	4. the Service's ServiceGUID selector will mean that the Service will automatically connect to the Pod.
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
	userServiceContainerName               = "user-service-container"
	filesArtifactExpanderInitContainerName = "files-artifact-expander"

	shouldMountVolumesAsReadOnly         = false
	shouldAddTimestampsToUserServiceLogs = false
	// Our user services don't need service accounts
	userServiceServiceAccountName = ""

	// Kubernetes doesn't allow us to create services without ports exposed, but we might not have ports in the following situations:
	//  1) we've registered a service but haven't started a container yet (so ports are yet to come)
	//  2) we've started a container that doesn't listen on any ports
	// In these cases, we use these notional unbound ports
	unboundPortName = "nonexistent-port"
	unboundPortNumber = 1

	tarSuccessExitCode = 0

	isFilesArtifactExpansionVolumeReadOnly = false
)

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

	// This may be nil if no files artifacts were expanded
	filesArtifactExpansionPersistentVolumeClaim *apiv1.PersistentVolumeClaim
}

func (backend *KubernetesKurtosisBackend) RegisterUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceId service.ServiceID) (*service.ServiceRegistration, error) {
	namespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
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
		label_key_consts.AppIDKubernetesLabelKey: label_value_consts.AppIDKubernetesLabelValue,
		label_key_consts.EnclaveIDKubernetesLabelKey: enclaveIdLabelValue,
		label_key_consts.GUIDKubernetesLabelKey: serviceGuidLabelValue,
	}
	matchedPodLabelStrs := getStringMapFromLabelMap(matchedPodLabels)

	// Kubernetes doesn't allow us to create services without any ports, so we need to set this to a notional value
	// until the user calls StartService
	notionalServicePorts := []apiv1.ServicePort{
		{
			Name:        unboundPortName,
			Port:        unboundPortNumber,
		},
	}

	createdService, err := backend.kubernetesManager.CreateService(
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
			if err := backend.kubernetesManager.RemoveService(ctx, createdService); err != nil {
				logrus.Errorf("Registering service '%v' didn't complete successfully so we tried to remove the Kubernetes service we created but doing so threw an error:\n%v", serviceId, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to remove service '%v' in namespace '%v' manually!!!", createdService.Name, namespaceName)
			}
		}
	}()

	kubernetesResources := map[service.ServiceGUID]*userServiceKubernetesResources{
		serviceGuid: {
			service: createdService,
			pod:     nil, // No pod yet
			filesArtifactExpansionPersistentVolumeClaim: nil, // No PVC yet
		},
	}

	convertedObjects, err := getUserServiceObjectsFromKubernetesResources(enclaveId, kubernetesResources)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a service registration object from Kubernetes service")
	}
	objectsAndResources, found := convertedObjects[serviceGuid]
	if !found {
		return nil, stacktrace.NewError(
			"Successfully converted the Kubernetes service representing registered service with GUID '%v' to a " +
				"Kurtosis object, but couldn't find that key in the resulting map; this is a bug in Kurtosis",
			serviceGuid,
		)
	}
	serviceRegistration := objectsAndResources.serviceRegistration

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
	filesArtifactsExpansion *backend_interface.FilesArtifactsExpansion,
) (
	resultUserService *service.Service,
	resultErr error,
) {
	preexistingServiceFilters := &service.ServiceFilters{
		GUIDs:    map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}
	preexistingObjectsAndResources, err := backend.getMatchingUserServiceObjectsAndKubernetesResources(
		ctx,
		enclaveId,
		preexistingServiceFilters,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service objects and Kubernetes resources matching service GUID '%v'", serviceGuid)
	}
	if len(preexistingObjectsAndResources) == 0 {
		return nil, stacktrace.NewError("Couldn't find any service registrations matching service GUID '%v'", serviceGuid)
	}
	if len(preexistingObjectsAndResources) > 1 {
		// Should never happen because service GUIDs should be unique
		return nil, stacktrace.NewError("Found more than one service registration matching service GUID '%v'; this is a bug in Kurtosis", serviceGuid)
	}
	matchingObjectAndResources, found := preexistingObjectsAndResources[serviceGuid]
	if !found {
		return nil, stacktrace.NewError("Even though we pulled back some Kubernetes resources, no Kubernetes resources were available for requested service GUID '%v'; this is a bug in Kurtosis", serviceGuid)
	}
	kubernetesService := matchingObjectAndResources.kubernetesResources.service
	serviceObj := matchingObjectAndResources.service
	if serviceObj != nil {
		return nil, stacktrace.NewError("Cannot start service with GUID '%v' because the service has already been started previously", serviceGuid)
	}

	namespaceName := kubernetesService.GetNamespace()
	serviceRegistrationObj := matchingObjectAndResources.serviceRegistration

	objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveId)

	var filesArtifactsExpansionPvc *apiv1.PersistentVolumeClaim
	var podInitContainers []apiv1.Container
	var podVolumes []apiv1.Volume
	var userServiceContainerVolumeMounts []apiv1.VolumeMount
	shouldDeleteExpansionPvc := true
	if filesArtifactsExpansion != nil {
		filesArtifactsExpansionPvc, podVolumes, userServiceContainerVolumeMounts, podInitContainers, err = backend.prepareFilesArtifactsExpansionResources(
			ctx,
			namespaceName,
			serviceGuid,
			enclaveObjAttributesProvider,
			filesArtifactsExpansion.ExpanderImage,
			filesArtifactsExpansion.ExpanderEnvVars,
			filesArtifactsExpansion.ExpanderDirpathsToServiceDirpaths,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred preparing the files artifacts expansion resources")
		}
		defer func() {
			if shouldDeleteExpansionPvc {
				// Use background context so we delete these even if input context was cancelled
				if err := backend.kubernetesManager.RemovePersistentVolumeClaim(context.Background(), filesArtifactsExpansionPvc); err != nil {
					logrus.Errorf(
						"Starting service '%v' didn't complete successfully so we tried to delete files artifact expansion PVC '%v' that we " +
							"created, but doing so threw an error:\n%v",
						serviceGuid,
						filesArtifactsExpansionPvc.Name,
						err,
					)
					logrus.Errorf("You'll need to delete PVC '%v' manually!", filesArtifactsExpansionPvc.Name)
				}
			}
		}()
	}

	// Create the pod
	podAttributes, err := enclaveObjAttributesProvider.ForUserServicePod(serviceGuid, serviceRegistrationObj.GetID(), privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting attributes for new pod for service with GUID '%v'", serviceGuid)
	}
	podLabelsStrs := getStringMapFromLabelMap(podAttributes.GetLabels())
	podAnnotationsStrs := getStringMapFromAnnotationMap(podAttributes.GetAnnotations())

	podContainers, err := getUserServicePodContainerSpecs(
		containerImageName,
		entrypointArgs,
		cmdArgs,
		envVars,
		privatePorts,
		userServiceContainerVolumeMounts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the container specs for the user service pod with image '%v'", containerImageName)
	}

	podName := podAttributes.GetName().GetString()
	createdPod, err := backend.kubernetesManager.CreatePod(
		ctx,
		namespaceName,
		podName,
		podLabelsStrs,
		podAnnotationsStrs,
		podInitContainers,
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
			if err := backend.kubernetesManager.RemovePod(ctx, createdPod); err != nil {
				logrus.Errorf("Starting service didn't complete successfully so we tried to remove the pod we created but doing so threw an error:\n%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to remove pod '%v' in '%v' manually!!!", podName, namespaceName)
			}
		}
	}()

	updatedService, undoServiceUpdateFunc, err := backend.updateServiceWhenContainerStarted(ctx, namespaceName, kubernetesService, privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred updating service '%v' to reflect its new ports: %+v", kubernetesService.GetName(), privatePorts)
	}
	shouldUndoServiceUpdate := true
	defer func() {
		if shouldUndoServiceUpdate {
			undoServiceUpdateFunc()
		}
	}()

	logrus.Debugf("Updated user service service after updating ports: %+v", updatedService)

	kubernetesResources := map[service.ServiceGUID]*userServiceKubernetesResources{
		serviceGuid: {
			service: updatedService,
			pod:     createdPod,
			filesArtifactExpansionPersistentVolumeClaim: filesArtifactsExpansionPvc,
		},
	}

	logrus.Debugf("Kubernetes resources that will be converted into a Service object: %+v", kubernetesResources)

	convertedObjects, err := getUserServiceObjectsFromKubernetesResources(enclaveId, kubernetesResources)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a service object from the Kubernetes service and newly-created pod")
	}
	objectsAndResources, found := convertedObjects[serviceGuid]
	if !found {
		return nil, stacktrace.NewError(
			"Successfully converted the Kubernetes service + pod representing a running service with GUID '%v' to a " +
				"Kurtosis object, but couldn't find that key in the resulting map; this is a bug in Kurtosis",
			serviceGuid,
		)
	}

	logrus.Debugf("Post-conversion objects & Kubernetes resources: %+v", objectsAndResources)

	shouldDeleteExpansionPvc = false
	shouldDestroyPod = false
	shouldUndoServiceUpdate = false
	return objectsAndResources.service, nil
}

func (backend *KubernetesKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (successfulUserServices map[service.ServiceGUID]*service.Service, resultError error) {
	allObjectsAndResources, err := backend.getMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting user services in enclave '%v' matching filters: %+v",
			enclaveId,
			filters,
		)
	}
	result := map[service.ServiceGUID]*service.Service{}
	for guid, serviceObjsAndResources := range allObjectsAndResources {
		serviceObj := serviceObjsAndResources.service
		if serviceObj == nil {
			// Indicates a registration-only service; skip
			continue
		}
		result[guid] = serviceObj
	}
	return result, nil
}

func (backend *KubernetesKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (successfulUserServiceLogs map[service.ServiceGUID]io.ReadCloser, erroredUserServiceGuids map[service.ServiceGUID]error, resultError error) {
	serviceObjectsAndResources, err := backend.getMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Expected to be able to get user services and Kubernetes resources, instead a non nil error was returned")
	}
	userServiceLogs := map[service.ServiceGUID]io.ReadCloser{}
	erredServiceLogs := map[service.ServiceGUID]error{}
	shouldCloseLogStreams := true
	for _, serviceObjectAndResource := range serviceObjectsAndResources {
		serviceGuid := serviceObjectAndResource.service.GetRegistration().GetGUID()
		servicePod := serviceObjectAndResource.kubernetesResources.pod
		if servicePod == nil {
			erredServiceLogs[serviceGuid] = stacktrace.NewError("Expected to find a pod for Kurtosis service with GUID '%v', instead no pod was found", serviceGuid)
			continue
		}
		serviceNamespaceName := serviceObjectAndResource.kubernetesResources.service.GetNamespace()
		// Get logs
		logReadCloser, err := backend.kubernetesManager.GetContainerLogs(ctx, serviceNamespaceName, servicePod.Name, userServiceContainerName, shouldFollowLogs, shouldAddTimestampsToUserServiceLogs)
		if err != nil {
			erredServiceLogs[serviceGuid] = stacktrace.Propagate(err, "Expected to be able to call Kubernetes to get logs for service with GUID '%v', instead a non-nil error was returned", serviceGuid)
			continue
		}
		defer func() {
			if shouldCloseLogStreams {
				logReadCloser.Close()
			}
		}()

		userServiceLogs[serviceGuid] = logReadCloser
	}

	shouldCloseLogStreams = false
	return userServiceLogs, erredServiceLogs, nil
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

// TODO Switch these to streaming methods, so that huge command outputs don't blow up the memory of the API container
func (backend *KubernetesKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
) (
	succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	namespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	requestedGuids := map[service.ServiceGUID]bool{}
	for guid := range userServiceCommands {
		requestedGuids[guid] = true
	}
	matchingServicesFilters := &service.ServiceFilters{
		GUIDs: requestedGuids,
	}
	matchingObjectsAndResources, err := backend.getMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, matchingServicesFilters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching the requested GUIDs: %+v", requestedGuids)
	}

	for guid, commandArgs := range userServiceCommands {
		objectsAndResources, found := matchingObjectsAndResources[guid]
		if !found {
			return nil, nil, stacktrace.NewError(
				"Requested to execute command '%+v' on service '%v', but the service does not exist",
				commandArgs,
				guid,
			)
		}
		serviceObj := objectsAndResources.service
		if serviceObj == nil {
			return nil, nil, stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because the service is not started yet",
				commandArgs,
				guid,
			)
		}
		if serviceObj.GetStatus() != container_status.ContainerStatus_Running {
			return nil, nil, stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because the service status is '%v'",
				commandArgs,
				guid,
				serviceObj.GetStatus().String(),
			)
		}
	}

	// TODO Parallelize for perf
	userServiceExecSuccess := map[service.ServiceGUID]*exec_result.ExecResult{}
	userServiceExecErr := map[service.ServiceGUID]error{}
	for serviceGuid, serviceCommand := range userServiceCommands {
		userServiceObjectAndResources, found := matchingObjectsAndResources[serviceGuid]
		if !found {
			// Should never happen because we validate that the object exists earlier
			return nil, nil, stacktrace.NewError("Validated that service '%v' has Kubernetes resources, but couldn't find them when we need to run the exec", serviceGuid)
		}
		// Don't need to validate that this is non-nil because we did so before we started executing
		userServicePod := userServiceObjectAndResources.kubernetesResources.pod
		userServicePodName := userServicePod.Name

		outputBuffer := &bytes.Buffer{}
		concurrentBuffer := concurrent_writer.NewConcurrentWriter(outputBuffer)
		exitCode, err := backend.kubernetesManager.RunExecCommand(
			namespaceName,
			userServicePodName,
			userServiceContainerName,
			serviceCommand,
			concurrentBuffer,
			concurrentBuffer,
		)
		if err != nil {
			userServiceExecErr[serviceGuid] = stacktrace.Propagate(
				err,
				"Expected to be able to execute command '%+v' in user service container '%v' in Kubernetes pod '%v' " +
					"for Kurtosis service with guid '%v', instead a non-nil error was returned",
				serviceCommand,
				userServiceContainerName,
				userServicePodName,
				serviceGuid,
			)
			continue
		}
		userServiceExecSuccess[serviceGuid] = exec_result.NewExecResult(exitCode, outputBuffer.String())
	}
	return userServiceExecSuccess, userServiceExecErr, nil
}

func (backend *KubernetesKurtosisBackend) GetConnectionWithUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceGUID service.ServiceGUID) (resultConn net.Conn, resultErr error) {
	// See https://github.com/kubernetes/client-go/issues/912
	/*
		in := streams.NewIn(os.Stdin)
		if err := in.SetRawTerminal(); err != nil{
	                 // handle err
		}
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:             in,
			Stdout:           stdout,
			Stderr:            stderr,
	        }
	 */

	// TODO IMPLEMENT
	return nil, stacktrace.NewError("Getting a connection with a user service isn't yet implemented on Kubernetes")
}

func (backend *KubernetesKurtosisBackend) CopyFilesFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	srcPath string,
	output io.Writer,
) error {
	namespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	objectAndResources, err := backend.getSingleUserServiceObjectsAndResources(ctx, enclaveId, serviceGuid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service object & Kubernetes resources for service '%v' in enclave '%v'", serviceGuid, enclaveId)
	}
	pod := objectAndResources.kubernetesResources.pod
	if pod == nil {
		return stacktrace.NewError(
			"Cannot copy path '%v' on service '%v' in enclave '%v' because no pod exists for the service",
			srcPath,
			serviceGuid,
			enclaveId,
		)
	}
	if pod.Status.Phase != apiv1.PodRunning {
		return stacktrace.NewError(
			"Cannot copy path '%v' on service '%v' in enclave '%v' because the pod isn't running",
			srcPath,
			serviceGuid,
			enclaveId,
		)
	}

	commandToRun := fmt.Sprintf(
		`if command -v 'tar' > /dev/null; then tar cf - '%v'; else echo "Cannot copy files from path '%v' because the tar binary doesn't exist on the machine" >&2; exit 1; fi`,
		srcPath,
		srcPath,
	)
	shWrappedCommandToRun := []string{
		"sh",
		"-c",
		commandToRun,
	}

	// NOTE: If we hit problems with very large files and connections breaking before they do, 'kubectl cp' implements a retry
	// mechanism that we could draw inspiration from:
	// https://github.com/kubernetes/kubectl/blob/335090af6913fb1ebf4a1f9e2463c46248b3e68d/pkg/cmd/cp/cp.go#L345
	stdErrOutput := &bytes.Buffer{}
	exitCode, err := backend.kubernetesManager.RunExecCommand(
		 namespaceName,
		 pod.Name,
		 userServiceContainerName,
		 shWrappedCommandToRun,
		 output,
		 stdErrOutput,
	)
	if err != nil {
		 return stacktrace.Propagate(
			 err,
			 "An error occurred running command '%v' on pod '%v' for service '%v' in namespace '%v'",
			 commandToRun,
			 pod.Name,
			 serviceGuid,
			 namespaceName,
		 )
	}
	if exitCode != tarSuccessExitCode {
		return stacktrace.NewError(
			"Command '%v' exited with non-%v exit code %v and the following STDERR:\n%v",
			commandToRun,
			tarSuccessExitCode,
			exitCode,
			stdErrOutput.String(),
		)
	}

	return nil
}

func (backend *KubernetesKurtosisBackend) StopUserServices(ctx context.Context, enclaveId enclave.EnclaveID, filters *service.ServiceFilters) (resultSuccessfulGuids map[service.ServiceGUID]bool, resultErroredGuids map[service.ServiceGUID]error, resultErr error) {
	namespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	allObjectsAndResources, err := backend.getMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services in enclave '%v' matching filters: %+v", enclaveId, filters)
	}

	successfulGuids := map[service.ServiceGUID]bool{}
	erroredGuids := map[service.ServiceGUID]error{}
	for serviceGuid, serviceObjsAndResources := range allObjectsAndResources {
		resources := serviceObjsAndResources.kubernetesResources

		pod := resources.pod
		if pod != nil {
			if err := backend.kubernetesManager.RemovePod(ctx, pod); err != nil {
				erroredGuids[serviceGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing Kubernetes pod '%v' in namespace '%v'",
					pod.Name,
					namespaceName,
				)
				continue
			}
		}

		kubernetesService := resources.service
		serviceName := kubernetesService.Name
		updateConfigurator := func(updatesToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
			specUpdates := applyconfigurationsv1.ServiceSpec().WithSelector(nil)
			updatesToApply.WithSpec(specUpdates)
		}
		if _, err := backend.kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updateConfigurator); err != nil {
			erroredGuids[serviceGuid] = stacktrace.Propagate(
				err,
				"An error occurred updating service '%v' in namespace '%v' to reflect that it's no longer running",
				serviceName,
				namespaceName,
			)
			continue
		}

		successfulGuids[serviceGuid] = true
	}
	return successfulGuids, erroredGuids, nil
}

func (backend *KubernetesKurtosisBackend) DestroyUserServices(ctx context.Context, enclaveId enclave.EnclaveID, filters *service.ServiceFilters) (resultSuccessfulGuids map[service.ServiceGUID]bool, resultErroredGuids map[service.ServiceGUID]error, resultErr error) {
	namespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	allObjectsAndResources, err := backend.getMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred getting user services in enclave '%v' matching filters: %+v",
			enclaveId,
			filters,
		)
	}

	successfulGuids := map[service.ServiceGUID]bool{}
	erroredGuids := map[service.ServiceGUID]error{}
	for serviceGuid, serviceObjsAndResources := range allObjectsAndResources {
		resources := serviceObjsAndResources.kubernetesResources

		pod := resources.pod
		if pod != nil {
			if err := backend.kubernetesManager.RemovePod(ctx, pod); err != nil {
				erroredGuids[serviceGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing Kubernetes pod '%v' in namespace '%v'",
					pod.Name,
					namespaceName,
				)
				continue
			}
		}

		pvc := resources.filesArtifactExpansionPersistentVolumeClaim
		if pvc != nil {
			if err := backend.kubernetesManager.RemovePersistentVolumeClaim(ctx, pvc); err != nil {
				erroredGuids[serviceGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing files artifact expansion persistent volume claim for service '%v'",
					serviceGuid,
				)
				continue
			}
		}

		// Canonical resource; this must be deleted last!
		kubernetesService := resources.service
		if kubernetesService != nil {
			if err := backend.kubernetesManager.RemoveService(ctx, kubernetesService); err != nil {
				erroredGuids[serviceGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing Kubernetes service '%v' in namespace '%v'",
					kubernetesService.Name,
					namespaceName,
				)
				continue
			}
		}
		// WARNING: DO NOT ADD ANYTHING HERE! The Service must be deleted as the very last thing as it's the canonical resource for the Kurtosis Service
		successfulGuids[serviceGuid] = true
	}
	return successfulGuids, erroredGuids, nil
}


// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func (backend *KubernetesKurtosisBackend) getMatchingUserServiceObjectsAndKubernetesResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*userServiceObjectsAndKubernetesResources,
	error,
) {
	allResources, err := backend.getUserServiceKubernetesResourcesMatchingGuids(ctx, enclaveId, filters.GUIDs)
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
	results := map[service.ServiceGUID]*userServiceObjectsAndKubernetesResources{}
	for serviceGuid, objectsAndResources := range allObjectsAndResources {
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[serviceGuid]; !found {
				continue
			}
		}

		registration := objectsAndResources.serviceRegistration
		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[registration.GetID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			kubernetesService := objectsAndResources.service

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

func (backend *KubernetesKurtosisBackend) getSingleUserServiceObjectsAndResources(ctx context.Context, enclaveId enclave.EnclaveID, serviceGuid service.ServiceGUID) (*userServiceObjectsAndKubernetesResources, error) {
	searchFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}
	searchResults, err := backend.getMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveId, searchFilters)
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

func (backend *KubernetesKurtosisBackend) getUserServiceKubernetesResourcesMatchingGuids(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuids map[service.ServiceGUID]bool,
) (
	map[service.ServiceGUID]*userServiceKubernetesResources,
	error,
) {
	namespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	// TODO switch to properly-typed KubernetesLabelValue object!!!
	postFilterLabelValues := map[string]bool{}
	for serviceGuid := range serviceGuids {
		postFilterLabelValues[string(serviceGuid)] = true
	}

	kubernetesResourceSearchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString(): label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.EnclaveIDKubernetesLabelKey.GetString(): string(enclaveId),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}

	results := map[service.ServiceGUID]*userServiceKubernetesResources{}

	// Get k8s services
	matchingKubernetesServices, err := kubernetes_resource_collectors.CollectMatchingServices(
		ctx,
		backend.kubernetesManager,
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
			resultObj = &userServiceKubernetesResources{}
		}
		resultObj.service = kubernetesService
		results[serviceGuid] = resultObj
	}

	// Get k8s pods
	matchingKubernetesPods, err := kubernetes_resource_collectors.CollectMatchingPods(
		ctx,
		backend.kubernetesManager,
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
			resultObj = &userServiceKubernetesResources{}
		}
		resultObj.pod = kubernetesPod
		results[serviceGuid] = resultObj
	}

	// Get files artifact expansion persistent volume claims
	matchingkubernetesPvcs, err := kubernetes_resource_collectors.CollectMatchingPersistentVolumeClaims(
		ctx,
		backend.kubernetesManager,
		namespaceName,
		kubernetesResourceSearchLabels,
		label_key_consts.GUIDKubernetesLabelKey.GetString(),
		postFilterLabelValues,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes PVCs matching service GUIDs: %+v", serviceGuids)
	}
	for serviceGuidStr, kubernetesPvcsForGuid := range matchingkubernetesPvcs {
		logrus.Tracef("Found Kubernetes PVCs for GUID '%v': %+v", serviceGuidStr, kubernetesPvcsForGuid)
		serviceGuid := service.ServiceGUID(serviceGuidStr)

		numPvcsForGuid := len(kubernetesPvcsForGuid)
		if numPvcsForGuid == 0 {
			// This would indicate a bug in our PVC retrieval logic because we shouldn't even have a map entry if there's nothing matching it
			return nil, stacktrace.NewError("Got entry of result PVCs for service GUID '%v', but no Kubernetes PVCs were returned; this is a bug in Kurtosis", serviceGuid)
		}
		if numPvcsForGuid > 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes PVCs associated with service GUID '%v'; this is a bug in Kurtosis", numPvcsForGuid, serviceGuid)
		}
		kubernetesPvc := kubernetesPvcsForGuid[0]

		resultObj, found := results[serviceGuid]
		if !found {
			resultObj = &userServiceKubernetesResources{}
		}
		resultObj.filesArtifactExpansionPersistentVolumeClaim = kubernetesPvc
		results[serviceGuid] = resultObj
	}

	return results, nil
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

		// Selectors but no pod means that the registration is open but no pod has yet been started to consume it
		stillUsingUnboundPort := false
		for _, servicePort := range kubernetesService.Spec.Ports {
			if servicePort.Name == unboundPortName {
				stillUsingUnboundPort = true
				break
			}
		}
		if stillUsingUnboundPort {
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

// NOTE: If this function succeeds, it is the user's responsibility to take care of the PersistentVolumeClaim that was created
func (backend *KubernetesKurtosisBackend) prepareFilesArtifactsExpansionResources(
	ctx context.Context,
	namespaceName string,
	serviceGuid service.ServiceGUID,
	enclaveObjAttributesProvider object_attributes_provider.KubernetesEnclaveObjectAttributesProvider,
	expanderImage string,
	expanderEnvVars map[string]string,
	expanderDirpathsToUserServiceDirpaths map[string]string,
) (
	resultFilesArtifactExpansionPvc *apiv1.PersistentVolumeClaim,
	resultPodVolumes []apiv1.Volume,
	resultUserServiceContainerVolumeMounts []apiv1.VolumeMount,
	resultPodInitContainers []apiv1.Container,
	resultErr error,
) {
	apiContainerArgs := backend.apiContainerModeArgs
	if apiContainerArgs == nil {
		return nil, nil, nil, nil, stacktrace.NewError(
			"Received request to start service '%v' with files artifact expansions, but no API container mode " +
				"args were defined which are necessary for creating the files artifacts expansion volume",
			serviceGuid,
		)
	}
	storageClass := apiContainerArgs.storageClassName
	expansionVolumeSizeMb := apiContainerArgs.filesArtifactExpansionVolumeSizeInMegabytes


	pvc, err := backend.createFilesArtifactsExpansionPersistentVolumeClaim(
		ctx,
		namespaceName,
		expansionVolumeSizeMb,
		storageClass,
		serviceGuid,
		enclaveObjAttributesProvider,
	)
	if err != nil {
		return nil, nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating files artifact expansion persistent volume claim for service '%v'",
			serviceGuid,
		)
	}
	shouldDeleteExpansionPvc := true
	defer func() {
		if shouldDeleteExpansionPvc {
			// Use background context so we delete these even if input context was cancelled
			if err := backend.kubernetesManager.RemovePersistentVolumeClaim(context.Background(), pvc); err != nil {
				logrus.Errorf(
					"Running files artifact expansion didn't complete successfully so we tried to delete files artifact expansion PVC '%v' that we " +
						"created, but doing so threw an error:\n%v",
					pvc.Name,
					err,
				)
				logrus.Errorf("You'll need to delete PVC '%v' manually!", pvc.Name)
			}
		}
	}()

	underlyingVolumeName := pvc.Spec.VolumeName

	podVolumes := []apiv1.Volume{
		{
			Name: underlyingVolumeName,
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
					ReadOnly:  isFilesArtifactExpansionVolumeReadOnly,
				},
			},
		},
	}

	volumeMountsOnExpanderContainer := []apiv1.VolumeMount{}
	volumeMountsOnUserServiceContainer := []apiv1.VolumeMount{}
	volumeSubdirIndex := 0
	for requestedExpanderDirpath, requestedUserServiceDirpath := range expanderDirpathsToUserServiceDirpaths {
		subdirName := strconv.Itoa(volumeSubdirIndex)

		expanderContainerMount := apiv1.VolumeMount{
			Name:             underlyingVolumeName,
			ReadOnly:         isFilesArtifactExpansionVolumeReadOnly,
			MountPath:        requestedExpanderDirpath,
			SubPath:          subdirName,
		}
		volumeMountsOnExpanderContainer = append(volumeMountsOnExpanderContainer, expanderContainerMount)

		userServiceContainerMount := apiv1.VolumeMount{
			Name:             underlyingVolumeName,
			ReadOnly:         isFilesArtifactExpansionVolumeReadOnly,
			MountPath:        requestedUserServiceDirpath,
			SubPath:          subdirName,
		}
		volumeMountsOnUserServiceContainer = append(volumeMountsOnUserServiceContainer, userServiceContainerMount)

		volumeSubdirIndex = volumeSubdirIndex + 1
	}

	filesArtifactExpansionInitContainer := getFilesArtifactExpansionInitContainerSpecs(
		expanderImage,
		expanderEnvVars,
		volumeMountsOnExpanderContainer,
	)

	podInitContainers := []apiv1.Container{
		filesArtifactExpansionInitContainer,
	}

	shouldDeleteExpansionPvc = false
	return pvc, podVolumes, volumeMountsOnUserServiceContainer, podInitContainers, nil
}

func getFilesArtifactExpansionInitContainerSpecs(
	image string,
	envVars map[string]string,
	volumeMounts []apiv1.VolumeMount,
) apiv1.Container {
	expanderEnvVars := []apiv1.EnvVar{}
	for key, value := range envVars {
		envVar := apiv1.EnvVar{
			Name:      key,
			Value: value,
		}
		expanderEnvVars = append(expanderEnvVars, envVar)
	}

	filesArtifactExpansionInitContainer := apiv1.Container{
		Name:          filesArtifactExpanderInitContainerName,
		Image:         image,
		Env:           expanderEnvVars,
		VolumeMounts:  volumeMounts,
	}

	return filesArtifactExpansionInitContainer

}

func getUserServicePodContainerSpecs(
	image string,
	entrypointArgs []string,
	cmdArgs []string,
	envVarStrs map[string]string,
	privatePorts map[string]*port_spec.PortSpec,
	containerMounts []apiv1.VolumeMount,
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
			VolumeMounts:             containerMounts,

			// NOTE: There are a bunch of other interesting Container options that we omitted for now but might
			// want to specify in the future
		},
	}

	return containers, nil
}

// Update the service to:
// - Set the service ports appropriately
// - Irrevocably record that a pod is bound to the service (so that even if the pod is deleted, the service won't
// 	 be usable again
func (backend *KubernetesKurtosisBackend) updateServiceWhenContainerStarted(
	ctx context.Context,
	namespaceName string,
	kubernetesService *apiv1.Service,
	privatePorts map[string]*port_spec.PortSpec,
) (
	*apiv1.Service,
	func(), // function to undo the update
	error,
) {
	serviceName := kubernetesService.GetName()

	serializedPortSpecs, err := kubernetes_port_spec_serializer.SerializePortSpecs(privatePorts)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred serializing the following private port specs: %+v", privatePorts)
	}

	// We only need to modify the ports from the default (unbound) ports if the user actually declares ports
	newServicePorts := kubernetesService.Spec.Ports
	if len(privatePorts) > 0 {
		candidateNewServicePorts, err := getKubernetesServicePortsFromPrivatePortSpecs(privatePorts)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes service ports for the following private port specs: %+v", privatePorts)
		}
		newServicePorts = candidateNewServicePorts
	}

	newAnnotations := kubernetesService.Annotations
	if newAnnotations == nil {
		newAnnotations = map[string]string{}
	}
	newAnnotations[kubernetes_annotation_key_consts.PortSpecsKubernetesAnnotationKey.GetString()] = serializedPortSpecs.GetString()

	updatingConfigurator := func(updatesToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
		specUpdateToApply := applyconfigurationsv1.ServiceSpec()
		for _, newServicePort := range newServicePorts {
			// You'd *think* that we could just feed in &newServicePort below, since the struct below needs pointer
			// args. However, this would be a bug: Go for loop iteration variables are updated in-place (probably to
			// save on memory), so if you pass around references to the for loop iteration variable then all dereferences
			// done after the loop will get the same value (the value of the last iteration). Therefore, we copy the variable
			// on each loop so that we have a fixed moment-in-time value.
			newServicePortCopy := newServicePort
			portUpdateToApply := &applyconfigurationsv1.ServicePortApplyConfiguration{
				Name:        &newServicePortCopy.Name,
				Protocol:    &newServicePortCopy.Protocol,
				// TODO fill this out for an app port!
				AppProtocol: nil,
				Port:        &newServicePortCopy.Port,
			}
			specUpdateToApply.WithPorts(portUpdateToApply)
		}
		updatesToApply.WithSpec(specUpdateToApply)

		updatesToApply.WithAnnotations(newAnnotations)
	}

	undoUpdateFunc := func() {
		undoingConfigurator := func(reversionToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
			specReversionToApply := applyconfigurationsv1.ServiceSpec()
			for _, oldServicePort := range newServicePorts {
				portUpdateToApply := &applyconfigurationsv1.ServicePortApplyConfiguration{
					Name:        &oldServicePort.Name,
					Protocol:    &oldServicePort.Protocol,
					// TODO fill this out for an app port!
					AppProtocol: nil,
					Port:        &oldServicePort.Port,
				}
				specReversionToApply.WithPorts(portUpdateToApply)
			}
			reversionToApply.WithSpec(specReversionToApply)

			reversionToApply.WithAnnotations(kubernetesService.Annotations)
		}
		if _, err := backend.kubernetesManager.UpdateService(ctx, namespaceName, serviceName, undoingConfigurator); err != nil {
			logrus.Errorf(
				"An error occurred updating Kubernetes service '%v' in namespace '%v' to open the service ports " +
					"and add the serialized private port specs annotation so we tried to revert the service to " +
					"its values before the update, but an error occurred; this means the service is likely in " +
					"an inconsistent state:\n%v",
				serviceName,
				namespaceName,
				err,
			)
			// Nothing we can really do here to recover
		}
	}

	updatedService, err := backend.kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updatingConfigurator)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred updating Kubernetes service '%v' in namespace '%v' to open the service ports and add " +
				"the serialized private port specs annotation",
			serviceName,
			namespaceName,
		)
	}
	shouldUndoUpdate := true
	defer func() {
		if shouldUndoUpdate {
			undoUpdateFunc()
		}
	}()

	shouldUndoUpdate = false
	return updatedService, undoUpdateFunc, nil
}

// Creates a persistent volume claim that the expander job will write all its expansions to
func (backend *KubernetesKurtosisBackend) createFilesArtifactsExpansionPersistentVolumeClaim(
	ctx context.Context,
	namespaceName string,
	expansionVolumeSizeInMegabytes uint,
	storageClass string,
	serviceGuid service.ServiceGUID,
	enclaveObjAttrsProvider object_attributes_provider.KubernetesEnclaveObjectAttributesProvider,
) (*apiv1.PersistentVolumeClaim, error) {
	pvcAttrs, err := enclaveObjAttrsProvider.ForFilesArtifactsExpansionPersistentVolumeClaim(serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating files artifact expansion persistent volume claim for service '%v'", serviceGuid)
	}
	pvcNameStr := pvcAttrs.GetName().GetString()
	pvcLabelsStr := getStringMapFromLabelMap(pvcAttrs.GetLabels())
	pvcAnnotationsStrs := getStringMapFromAnnotationMap(pvcAttrs.GetAnnotations())

	pvc, err := backend.kubernetesManager.CreatePersistentVolumeClaim(
		ctx,
		namespaceName,
		pvcNameStr,
		pvcLabelsStr,
		pvcAnnotationsStrs,
		expansionVolumeSizeInMegabytes,
		storageClass,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating files artifact expansion persistent volume claim for service '%v'",
			serviceGuid,
		)
	}
	return pvc, nil
}
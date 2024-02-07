package user_services_functions

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_user"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

const (
	userServiceContainerName = "user-service-container"
	// Our user services don't need service accounts
	userServiceServiceAccountName = ""

	megabytesToBytesFactor = 1_000_000

	// Kubernetes doesn't allow us to create services without ports exposed, but we might not have ports in the following situations:
	//  1) we've registered a service but haven't started a container yet (so ports are yet to come)
	//  2) we've started a container that doesn't listen on any ports
	// In these cases, we use these notional unbound ports
	unboundPortName   = "nonexistent-port"
	unboundPortNumber = 1

	unlimitedReplacements = -1
)

// Completeness enforced via unit test
var kurtosisTransportProtocolToKubernetesTransportProtocolTranslator = map[port_spec.TransportProtocol]apiv1.Protocol{
	port_spec.TransportProtocol_TCP:  apiv1.ProtocolTCP,
	port_spec.TransportProtocol_UDP:  apiv1.ProtocolUDP,
	port_spec.TransportProtocol_SCTP: apiv1.ProtocolSCTP,
}

func RegisterUserServices(
	ctx context.Context,
	enclaveUUID enclave.EnclaveUUID,
	servicesToRegister map[service.ServiceName]bool,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceName]*service.ServiceRegistration,
	map[service.ServiceName]error,
	error,
) {
	successfulServicesPool := map[service.ServiceName]*service.ServiceRegistration{}
	failedServicesPool := map[service.ServiceName]error{}

	// Check whether any services have been provided at all
	if len(servicesToRegister) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	successfulRegistrations, failedRegistrations, err := registerUserServices(ctx, enclaveUUID, servicesToRegister, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services with Names '%v'", servicesToRegister)
	}
	return successfulRegistrations, failedRegistrations, nil
}

func UnregisterUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	servicesToUnregister map[service.ServiceUUID]bool,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	successfullyUnregisteredServices := map[service.ServiceUUID]bool{}
	failedServices := map[service.ServiceUUID]error{}

	if len(servicesToUnregister) == 0 {
		return successfullyUnregisteredServices, failedServices, nil
	}
	userServiceFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    servicesToUnregister,
		Statuses: nil,
	}
	successfullyDestroyedServices, failedToDestroyUUIDs, err := DestroyUserServices(ctx, enclaveUuid, userServiceFilters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Attempted to destroy all services with UUIDs '%v' together but had no success. You must manually destroy the services!", servicesToUnregister)
	}
	return successfullyDestroyedServices, failedToDestroyUUIDs, nil
}

func StartRegisteredUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	services map[service.ServiceUUID]*service.ServiceConfig,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	restartPolicy apiv1.RestartPolicy,
) (
	map[service.ServiceUUID]*service.Service,
	map[service.ServiceUUID]error,
	error,
) {
	successfulServicesPool := map[service.ServiceUUID]*service.Service{}
	failedServicesPool := map[service.ServiceUUID]error{}

	// Check whether any services have been provided at all
	if len(services) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	// Sanity check for port bindings on all services
	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	for _, config := range services {
		publicPorts := config.GetPublicPorts()
		if len(publicPorts) > 0 {
			logrus.Warn("The Kubernetes Kurtosis backend doesn't support defining static ports for services; the public ports will be ignored.")
		}
	}

	// Get existing objects by UUID. We use UUIDs and not Names as Names can match services being deleted or services in other Enclaves.
	serviceUUIDsToFilter := map[service.ServiceUUID]bool{}
	for serviceUuid := range services {
		serviceUUIDsToFilter[serviceUuid] = true
	}
	existingObjectsAndResourcesFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    serviceUUIDsToFilter,
		Statuses: nil,
	}
	// This is safe, as there's an N -> 1 mapping between UUID and ID and the UUIDs that we filter on don't have any matching IDs
	existingObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(ctx, enclaveUuid, existingObjectsAndResourcesFilters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user service objects and Kubernetes resources matching service UUIDs '%v'", serviceUUIDsToFilter)
	}
	if len(existingObjectsAndResources) > len(serviceUUIDsToFilter) {
		// Should never happen because service UUIDs should be unique
		return nil, nil, stacktrace.NewError("Found more than one service registration matching service UUIDs; this is a bug in Kurtosis")
	}
	serviceRegisteredThatCanBeStarted := map[service.ServiceUUID]*service.ServiceConfig{}
	for serviceUuid, serviceConfig := range services {
		if _, found := existingObjectsAndResources[serviceUuid]; !found {
			failedServicesPool[serviceUuid] = stacktrace.NewError("Couldn't find any service registrations for service UUID '%v'. This is a bug in Kurtosis.", serviceUuid)
			continue
		}
		if serviceConfig.GetPrivateIPAddrPlaceholder() == "" {
			failedServicesPool[serviceUuid] = stacktrace.NewError("Service with UUID '%v' has an empty private IP Address placeholder. Expect this to be of length greater than zero.", serviceUuid)
			continue
		}
		serviceRegisteredThatCanBeStarted[serviceUuid] = serviceConfig
	}

	successfulStarts, failedStarts, err := runStartServiceOperationsInParallel(
		ctx,
		enclaveUuid,
		serviceRegisteredThatCanBeStarted,
		existingObjectsAndResources,
		kubernetesManager,
		restartPolicy)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to start services in parallel.")
	}

	// Add operations to their respective pools
	for serviceUuid, serviceSuccessfullyStarted := range successfulStarts {
		successfulServicesPool[serviceUuid] = serviceSuccessfullyStarted
	}

	for serviceUuid, serviceFailed := range failedStarts {
		failedServicesPool[serviceUuid] = serviceFailed
	}

	logrus.Debugf("Started services '%v' successfully while '%v' failed", successfulServicesPool, failedServicesPool)
	return successfulServicesPool, failedServicesPool, nil
}

func RemoveRegisteredUserServiceProcesses(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	services map[service.ServiceUUID]bool,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	// in Kubernetes, removing service process is equivalent to stopping it. It sets its number of pod to zero
	removeServiceProcessesFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    services,
		Statuses: nil,
	}
	successfullyRemovedService, failedRemovedService, err := StopUserServices(
		ctx,
		enclaveUuid,
		removeServiceProcessesFilters,
		cliModeArgs,
		apiContainerModeArgs,
		engineServerModeArgs,
		kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error removing service processes")
	}
	return successfullyRemovedService, failedRemovedService, nil
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func runStartServiceOperationsInParallel(
	ctx context.Context,
	enclaveUUID enclave.EnclaveUUID,
	services map[service.ServiceUUID]*service.ServiceConfig,
	servicesObjectsAndResources map[service.ServiceUUID]*shared_helpers.UserServiceObjectsAndKubernetesResources,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	restartPolicy apiv1.RestartPolicy,
) (
	map[service.ServiceUUID]*service.Service,
	map[service.ServiceUUID]error,
	error,
) {
	startServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for serviceName, config := range services {
		startServiceOperations[operation_parallelizer.OperationID(serviceName)] = createStartServiceOperation(
			ctx,
			serviceName,
			config,
			servicesObjectsAndResources,
			enclaveUUID,
			kubernetesManager,
			restartPolicy)
	}

	successfulServiceObjs, failedOperations := operation_parallelizer.RunOperationsInParallel(startServiceOperations)

	successfulServices := map[service.ServiceUUID]*service.Service{}
	failedServices := map[service.ServiceUUID]error{}

	for guid, data := range successfulServiceObjs {
		serviceUuid := service.ServiceUUID(guid)
		serviceObjectPtr, ok := data.(*service.Service)
		if !ok {
			return nil, nil, stacktrace.NewError(
				"An error occurred downcasting data returned from the start user service operation for service with UUID: %v. "+
					"This is a Kurtosis bug. Make sure the desired type is actually being returned in the corresponding Operation.", serviceUuid)
		}
		successfulServices[serviceUuid] = serviceObjectPtr
	}

	for guid, err := range failedOperations {
		serviceUuid := service.ServiceUUID(guid)
		failedServices[serviceUuid] = err
	}

	return successfulServices, failedServices, nil
}

func createStartServiceOperation(
	ctx context.Context,
	serviceUuid service.ServiceUUID,
	serviceConfig *service.ServiceConfig,
	servicesObjectsAndResources map[service.ServiceUUID]*shared_helpers.UserServiceObjectsAndKubernetesResources,
	enclaveUuid enclave.EnclaveUUID,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	restartPolicy apiv1.RestartPolicy) operation_parallelizer.Operation {

	return func() (interface{}, error) {
		filesArtifactsExpansion := serviceConfig.GetFilesArtifactsExpansion()
		persistentDirectories := serviceConfig.GetPersistentDirectories()
		containerImageName := serviceConfig.GetContainerImageName()
		privatePorts := serviceConfig.GetPrivatePorts()
		entrypointArgs := serviceConfig.GetEntrypointArgs()
		cmdArgs := serviceConfig.GetCmdArgs()
		envVars := serviceConfig.GetEnvVars()
		cpuAllocationMillicpus := serviceConfig.GetCPUAllocationMillicpus()
		memoryAllocationMegabytes := serviceConfig.GetMemoryAllocationMegabytes()
		privateIPAddrPlaceholder := serviceConfig.GetPrivateIPAddrPlaceholder()
		minCpuAllocationMilliCpus := serviceConfig.GetMinCPUAllocationMillicpus()
		minMemoryAllocationMegabytes := serviceConfig.GetMinMemoryAllocationMegabytes()
		user := serviceConfig.GetUser()
		tolerations := serviceConfig.GetTolerations()

		matchingObjectAndResources, found := servicesObjectsAndResources[serviceUuid]
		if !found {
			return nil, stacktrace.NewError("Even though we pulled back some Kubernetes resources, no Kubernetes resources were available for requested service UUID '%v'; this is a bug in Kurtosis", serviceUuid)
		}
		kubernetesService := matchingObjectAndResources.KubernetesResources.Service

		// We replace the placeholder value with the actual private IP address
		privateIPAddr := matchingObjectAndResources.ServiceRegistration.GetPrivateIP().String()
		for index := range entrypointArgs {
			entrypointArgs[index] = strings.Replace(entrypointArgs[index], privateIPAddrPlaceholder, privateIPAddr, unlimitedReplacements)
		}
		for index := range cmdArgs {
			cmdArgs[index] = strings.Replace(cmdArgs[index], privateIPAddrPlaceholder, privateIPAddr, unlimitedReplacements)
		}
		for key := range envVars {
			envVars[key] = strings.Replace(envVars[key], privateIPAddrPlaceholder, privateIPAddr, unlimitedReplacements)
		}

		namespaceName := kubernetesService.GetNamespace()
		serviceRegistrationObj := matchingObjectAndResources.ServiceRegistration
		serviceName := serviceRegistrationObj.GetName()

		objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
		enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveUuid)

		var podInitContainers []apiv1.Container
		var podVolumes []apiv1.Volume
		var err error
		var userServiceContainerVolumeMounts []apiv1.VolumeMount
		if filesArtifactsExpansion != nil {
			podVolumes, userServiceContainerVolumeMounts, podInitContainers, err = prepareFilesArtifactsExpansionResources(
				filesArtifactsExpansion.ExpanderImage,
				filesArtifactsExpansion.ExpanderEnvVars,
				filesArtifactsExpansion.ExpanderDirpathsToServiceDirpaths,
			)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating the volumes necessary to perform file artifact expansion for service '%s'", serviceName)
			}
		}

		shouldDestroyPersistentVolumesAndClaims := true
		createVolumesWithClaims := map[string]*kubernetesVolumeWithClaim{}
		if persistentDirectories != nil {
			createVolumesWithClaims, err = preparePersistentDirectoriesResources(
				ctx,
				namespaceName,
				enclaveObjAttributesProvider,
				persistentDirectories.ServiceDirpathToPersistentDirectory,
				kubernetesManager)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating the persistent volumes and claims requested for service '%s'", serviceName)
			}
			for serviceMountDirPath, volumeAndClaim := range createVolumesWithClaims {
				podVolumes = append(podVolumes, *volumeAndClaim.GetVolume())
				userServiceContainerVolumeMounts = append(userServiceContainerVolumeMounts, *volumeAndClaim.GetVolumeMount(serviceMountDirPath))
			}
		}
		defer func() {
			if !shouldDestroyPersistentVolumesAndClaims {
				return
			}
			for _, volumeAndClaim := range createVolumesWithClaims {
				volumeClaimName := volumeAndClaim.VolumeClaimName
				if err := kubernetesManager.RemovePersistentVolumeClaim(ctx, namespaceName, volumeClaimName); err != nil {
					logrus.Errorf("Starting service didn't complete successfully so we tried to remove the persistent volume claim we created but doing so threw an error:\n%v", err)
					logrus.Errorf("ACTION REQUIRED: You'll need to remove persistent volume claim '%v' in '%v' manually!!!", volumeClaimName, namespaceName)
				}
			}
		}()

		// Create the pod
		podAttributes, err := enclaveObjAttributesProvider.ForUserServicePod(serviceUuid, serviceName, privatePorts, serviceConfig.GetLabels())
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting attributes for new pod for service with UUID '%v'", serviceUuid)
		}
		podLabelsStrs := shared_helpers.GetStringMapFromLabelMap(podAttributes.GetLabels())
		podAnnotationsStrs := shared_helpers.GetStringMapFromAnnotationMap(podAttributes.GetAnnotations())

		podContainers, err := getUserServicePodContainerSpecs(
			containerImageName,
			entrypointArgs,
			cmdArgs, envVars,
			privatePorts,
			userServiceContainerVolumeMounts,
			cpuAllocationMillicpus,
			memoryAllocationMegabytes,
			minCpuAllocationMilliCpus,
			minMemoryAllocationMegabytes,
			user,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the container specs for the user service pod with image '%v'", containerImageName)
		}

		podName := podAttributes.GetName().GetString()
		createdPod, err := kubernetesManager.CreatePod(
			ctx,
			namespaceName,
			podName,
			podLabelsStrs,
			podAnnotationsStrs,
			podInitContainers,
			podContainers,
			podVolumes,
			userServiceServiceAccountName,
			restartPolicy,
			tolerations,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating pod '%v' using image '%v'", podName, containerImageName)
		}
		shouldDestroyPod := true
		defer func() {
			if !shouldDestroyPod {
				return
			}
			if err := kubernetesManager.RemovePod(ctx, createdPod); err != nil {
				logrus.Errorf("Starting service didn't complete successfully so we tried to remove the pod we created but doing so threw an error:\n%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to remove pod '%v' in '%v' manually!!!", podName, namespaceName)
			}
		}()

		// Create the ingress for the reverse proxy
		ingressAttributes, err := enclaveObjAttributesProvider.ForUserServiceIngress(serviceUuid, serviceName, privatePorts)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting attributes for new ingress for service with UUID '%v'", serviceUuid)
		}
		ingressLabelsStrs := shared_helpers.GetStringMapFromLabelMap(ingressAttributes.GetLabels())
		ingressAnnotationsStrs := shared_helpers.GetStringMapFromAnnotationMap(ingressAttributes.GetAnnotations())

		ingressRules, err := getUserServiceIngressRules(serviceRegistrationObj, privatePorts)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the user service ingress rules for service with UUID '%v'", serviceUuid)
		}

		shouldDestroyIngress := false
		if ingressRules != nil {
			ingressName := string(serviceName)
			createdIngress, err := kubernetesManager.CreateIngress(
				ctx,
				namespaceName,
				ingressName,
				ingressLabelsStrs,
				ingressAnnotationsStrs,
				ingressRules,
			)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating ingress for service with UUID '%v'", serviceUuid)
			}
			shouldDestroyIngress = true
			defer func() {
				if !shouldDestroyIngress {
					return
				}
				if err := kubernetesManager.RemoveIngress(ctx, createdIngress); err != nil {
					logrus.Errorf("Starting service didn't complete successfully so we tried to remove the ingress we created but doing so threw an error:\n%v", err)
					logrus.Errorf("ACTION REQUIRED: You'll need to remove ingress '%v' in '%v' manually!!!", ingressName, namespaceName)
				}
			}()
		}

		updatedService, undoServiceUpdateFunc, err := updateServiceWhenContainerStarted(ctx, namespaceName, kubernetesService, privatePorts, kubernetesManager)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred updating service '%v' to reflect its new ports: %+v", kubernetesService.GetName(), privatePorts)
		}
		shouldUndoServiceUpdate := true
		defer func() {
			if shouldUndoServiceUpdate {
				undoServiceUpdateFunc()
			}
		}()

		kubernetesResources := map[service.ServiceUUID]*shared_helpers.UserServiceKubernetesResources{
			serviceUuid: {
				Service: updatedService,
				Pod:     createdPod,
			},
		}

		convertedObjects, err := shared_helpers.GetUserServiceObjectsFromKubernetesResources(enclaveUuid, kubernetesResources)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting a service object from the Kubernetes service and newly-created pod")
		}
		objectsAndResources, found := convertedObjects[serviceUuid]
		if !found {
			return nil, stacktrace.NewError(
				"Successfully converted the Kubernetes service + pod representing a running service with UUID '%v' to a "+
					"Kurtosis object, but couldn't find that key in the resulting map; this is a bug in Kurtosis",
				serviceUuid,
			)
		}

		shouldDestroyPod = false
		shouldDestroyIngress = false
		shouldUndoServiceUpdate = false
		shouldDestroyPersistentVolumesAndClaims = false
		return objectsAndResources.Service, nil
	}
}

// Update the service to:
//   - Set the service ports appropriately
//   - Irrevocably record that a pod is bound to the service (so that even if the pod is deleted, the service won't
//     be usable again
func updateServiceWhenContainerStarted(
	ctx context.Context,
	namespaceName string,
	kubernetesService *apiv1.Service,
	privatePorts map[string]*port_spec.PortSpec,
	kubernetesManager *kubernetes_manager.KubernetesManager,
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
				AppProtocol: newServicePortCopy.AppProtocol,
				Port:        &newServicePortCopy.Port,
				TargetPort:  nil,
				NodePort:    nil,
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
					AppProtocol: oldServicePort.AppProtocol,
					Port:        &oldServicePort.Port,
					TargetPort:  nil,
					NodePort:    nil,
				}
				specReversionToApply.WithPorts(portUpdateToApply)
			}
			reversionToApply.WithSpec(specReversionToApply)

			reversionToApply.WithAnnotations(kubernetesService.Annotations)
		}
		if _, err := kubernetesManager.UpdateService(ctx, namespaceName, serviceName, undoingConfigurator); err != nil {
			logrus.Errorf(
				"An error occurred updating Kubernetes service '%v' in namespace '%v' to open the service ports "+
					"and add the serialized private port specs annotation so we tried to revert the service to "+
					"its values before the update, but an error occurred; this means the service is likely in "+
					"an inconsistent state:\n%v",
				serviceName,
				namespaceName,
				err,
			)
			// Nothing we can really do here to recover
		}
	}

	updatedService, err := kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updatingConfigurator)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred updating Kubernetes service '%v' in namespace '%v' to open the service ports and add "+
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

func getUserServicePodContainerSpecs(
	image string,
	entrypointArgs []string,
	cmdArgs []string,
	envVarStrs map[string]string,
	privatePorts map[string]*port_spec.PortSpec,
	containerMounts []apiv1.VolumeMount,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
	minCpuAllocationMilliCpus uint64,
	minMemoryAllocationMegabytes uint64,
	user *service_user.ServiceUser,
) ([]apiv1.Container, error) {

	var containerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVarStrs {
		envVar := apiv1.EnvVar{
			Name:      varName,
			Value:     varValue,
			ValueFrom: nil,
		}
		containerEnvVars = append(containerEnvVars, envVar)
	}

	kubernetesContainerPorts, err := getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes container ports from the private port specs map")
	}

	resourceLimitsList := apiv1.ResourceList{}
	resourceRequestsList := apiv1.ResourceList{}
	// 0 is considered the empty value (meaning the field was never set), so if either fields are 0, that resource is left unbounded
	if cpuAllocationMillicpus != 0 {
		resourceLimitsList[apiv1.ResourceCPU] = *resource.NewMilliQuantity(int64(cpuAllocationMillicpus), resource.DecimalSI)
	}

	if minCpuAllocationMilliCpus != 0 {
		resourceRequestsList[apiv1.ResourceCPU] = *resource.NewMilliQuantity(int64(minCpuAllocationMilliCpus), resource.DecimalSI)
	}

	if memoryAllocationMegabytes != 0 {
		memoryAllocationInBytes := convertMegabytesToBytes(memoryAllocationMegabytes)
		resourceLimitsList[apiv1.ResourceMemory] = *resource.NewQuantity(int64(memoryAllocationInBytes), resource.DecimalSI)
	}

	if minMemoryAllocationMegabytes != 0 {
		minMemoryAllocationInBytes := convertMegabytesToBytes(minMemoryAllocationMegabytes)
		resourceRequestsList[apiv1.ResourceMemory] = *resource.NewQuantity(int64(minMemoryAllocationInBytes), resource.DecimalSI)
	}

	resourceRequirements := apiv1.ResourceRequirements{ //nolint:exhaustruct
		Limits:   resourceLimitsList,
		Requests: resourceRequestsList,
	}

	containers := []apiv1.Container{
		{
			Name:  userServiceContainerName,
			Image: image,
			// Yes, even though this is called "command" it actually corresponds to the Docker ENTRYPOINT
			Command:      entrypointArgs,
			Args:         cmdArgs,
			Ports:        kubernetesContainerPorts,
			Env:          containerEnvVars,
			VolumeMounts: containerMounts,
			Resources:    resourceRequirements,

			// NOTE: There are a bunch of other interesting Container options that we omitted for now but might
			// want to specify in the future
		},
	}

	if user != nil {
		uid := int64(user.GetUID())
		// nolint: exhaustruct
		securityContext := &apiv1.SecurityContext{
			RunAsUser: &uid,
		}

		gid, gidIsSet := user.GetGID()
		if gidIsSet {
			gidAsInt64 := int64(gid)
			securityContext.RunAsGroup = &gidAsInt64
		}

		containers[0].SecurityContext = securityContext
	}

	return containers, nil
}

func getKubernetesServicePortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ServicePort, error) {
	result := []apiv1.ServicePort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetTransportProtocol()
		kubernetesProtocol, found := kurtosisTransportProtocolToKubernetesTransportProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ServicePort{
			Name:        portId,
			Protocol:    kubernetesProtocol,
			AppProtocol: portSpec.GetMaybeApplicationProtocol(),
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

func getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ContainerPort, error) {
	result := []apiv1.ContainerPort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetTransportProtocol()
		kubernetesProtocol, found := kurtosisTransportProtocolToKubernetesTransportProtocolTranslator[kurtosisProtocol]
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

func convertMegabytesToBytes(value uint64) uint64 {
	return value * megabytesToBytesFactor
}

func registerUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceNames map[service.ServiceName]bool,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager) (map[service.ServiceName]*service.ServiceRegistration, map[service.ServiceName]error, error) {
	successfulServicesPool := map[service.ServiceName]*service.ServiceRegistration{}
	failedServicesPool := map[service.ServiceName]error{}

	// If no services were passed in to register, return empty maps
	if len(serviceNames) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveUuid, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveUuid)
	}

	objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveUuid)

	registerServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for serviceName := range serviceNames {
		registerServiceOperations[operation_parallelizer.OperationID(serviceName)] = createRegisterUserServiceOperation(
			ctx,
			enclaveUuid,
			serviceName,
			namespaceName,
			enclaveObjAttributesProvider,
			kubernetesManager)
	}

	successfulRegistrations, failedRegistrations := operation_parallelizer.RunOperationsInParallel(registerServiceOperations)

	for id, data := range successfulRegistrations {
		serviceName := service.ServiceName(id)
		serviceRegistration, ok := data.(*service.ServiceRegistration)
		if !ok {
			return nil, nil, stacktrace.NewError(
				"An error occurred downcasting data returned from the register user service operation for service with id: %v. "+
					"This is a Kurtosis bug. Make sure the desired type is actually being returned in the corresponding Operation.", serviceName)
		}
		successfulServicesPool[serviceName] = serviceRegistration
	}

	for opID, err := range failedRegistrations {
		failedServicesPool[service.ServiceName(opID)] = err
	}

	return successfulServicesPool, failedServicesPool, nil
}

func createRegisterUserServiceOperation(
	ctx context.Context,
	enclaveID enclave.EnclaveUUID,
	serviceName service.ServiceName,
	namespaceName string,
	enclaveObjAttributesProvider object_attributes_provider.KubernetesEnclaveObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) operation_parallelizer.Operation {
	return func() (interface{}, error) {
		serviceUuidStr, err := uuid_generator.GenerateUUIDString()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred generating a UUID to use for the service UUID")
		}
		serviceUuid := service.ServiceUUID(serviceUuidStr)
		serviceAttributes, err := enclaveObjAttributesProvider.ForUserServiceService(serviceUuid, serviceName)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting attributes for the Kubernetes service for user service '%v'", serviceName)
		}

		serviceNameStr := serviceAttributes.GetName().GetString()

		serviceLabelsStrs := shared_helpers.GetStringMapFromLabelMap(serviceAttributes.GetLabels())
		serviceAnnotationsStrs := shared_helpers.GetStringMapFromAnnotationMap(serviceAttributes.GetAnnotations())

		// Set up the labels that the pod will match (i.e. the labels of the pod-to-be)
		// WARNING: We *cannot* use the labels of the Service itself because we're not guaranteed that the labels
		//  between the two will be identical!
		serviceUuidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(serviceUuid))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the service UUID '%v'", serviceUuid)
		}
		enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(enclaveID))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the enclave ID '%v'", enclaveID)
		}
		matchedPodLabels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
			kubernetes_label_key.AppIDKubernetesLabelKey:       label_value_consts.AppIDKubernetesLabelValue,
			kubernetes_label_key.EnclaveUUIDKubernetesLabelKey: enclaveIdLabelValue,
			kubernetes_label_key.GUIDKubernetesLabelKey:        serviceUuidLabelValue,
		}
		matchedPodLabelStrs := shared_helpers.GetStringMapFromLabelMap(matchedPodLabels)

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
			return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes service in enclave '%v' with ID '%v'", enclaveID, serviceName)
		}
		shouldDeleteService := true
		defer func() {
			if shouldDeleteService {
				if err := kubernetesManager.RemoveService(ctx, createdService); err != nil {
					logrus.Errorf("Registering service '%v' didn't complete successfully so we tried to remove the Kubernetes service we created but doing so threw an error:\n%v", serviceName, err)
					logrus.Errorf("ACTION REQUIRED: You'll need to remove service '%v' in namespace '%v' manually!!!", createdService.Name, namespaceName)
				}
			}
		}()

		kubernetesResources := map[service.ServiceUUID]*shared_helpers.UserServiceKubernetesResources{
			serviceUuid: {
				Service: createdService,
				Pod:     nil, // No pod yet
			},
		}

		convertedObjects, err := shared_helpers.GetUserServiceObjectsFromKubernetesResources(enclaveID, kubernetesResources)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting a service registration object from Kubernetes service")
		}
		objectsAndResources, found := convertedObjects[serviceUuid]
		if !found {
			return nil, stacktrace.NewError(
				"Successfully converted the Kubernetes service representing registered service with UUID '%v' to a "+
					"Kurtosis object, but couldn't find that key in the resulting map; this is a bug in Kurtosis",
				serviceUuid,
			)
		}

		shouldDeleteService = false
		return objectsAndResources.ServiceRegistration, nil
	}
}

func getUserServiceIngressRules(
	serviceRegistration *service.ServiceRegistration,
	privatePorts map[string]*port_spec.PortSpec,
) ([]netv1.IngressRule, error) {
	var ingressRules []netv1.IngressRule
	enclaveShortUuid := uuid_generator.ShortenedUUIDString(string(serviceRegistration.GetEnclaveID()))
	serviceShortUuid := uuid_generator.ShortenedUUIDString(string(serviceRegistration.GetUUID()))
	for _, portSpec := range privatePorts {
		maybeApplicationProtocol := ""
		if portSpec.GetMaybeApplicationProtocol() != nil {
			maybeApplicationProtocol = *portSpec.GetMaybeApplicationProtocol()
		}
		if maybeApplicationProtocol == consts.HttpApplicationProtocol {
			host := fmt.Sprintf("%d-%s-%s", portSpec.GetNumber(), serviceShortUuid, enclaveShortUuid)
			ingressRule := netv1.IngressRule{
				Host: host,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{
							{
								Path:     consts.IngressRulePathAllPaths,
								PathType: &consts.IngressRulePathTypePrefix,
								Backend: netv1.IngressBackend{
									Service: &netv1.IngressServiceBackend{
										Name: string(serviceRegistration.GetName()),
										Port: netv1.ServiceBackendPort{
											Name:   "",
											Number: int32(portSpec.GetNumber()),
										},
									},
									Resource: nil,
								},
							},
						},
					},
				},
			}
			ingressRules = append(ingressRules, ingressRule)
		}
	}
	return ingressRules, nil
}

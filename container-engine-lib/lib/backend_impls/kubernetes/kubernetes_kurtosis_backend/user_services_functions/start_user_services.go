package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"strings"
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
var kurtosisPortProtocolToKubernetesPortProtocolTranslator = map[port_spec.PortProtocol]apiv1.Protocol{
	port_spec.PortProtocol_TCP:  apiv1.ProtocolTCP,
	port_spec.PortProtocol_UDP:  apiv1.ProtocolUDP,
	port_spec.PortProtocol_SCTP: apiv1.ProtocolSCTP,
}

func StartUserServices(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	services map[service.ServiceID]*service.ServiceConfig,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceID]*service.Service,
	map[service.ServiceID]error,
	error,
) {
	successfulServicesPool := map[service.ServiceID]*service.Service{}
	failedServicesPool := map[service.ServiceID]error{}

	// Check whether any services have been provided at all
	if len(services) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	var serviceIDsToRegister []service.ServiceID
	for serviceID, config := range services {
		if config.GetPrivateIPAddrPlaceholder() == "" {
			failedServicesPool[serviceID] = stacktrace.NewError("Service with ID '%v' has an empty private IP Address placeholder. Expect this to be of length greater than zero.", serviceID)
			continue
		}
		serviceIDsToRegister = append(serviceIDsToRegister, serviceID)
	}
	if len(serviceIDsToRegister) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	successfulRegistrations, failedRegistrations, err := registerUserServices(ctx, enclaveID, serviceIDsToRegister, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services with IDs '%v'", serviceIDsToRegister)
	}
	// Defer an undo to all the successful registrations in case an error occurs in future phases
	serviceIDsToRemove := map[service.ServiceID]bool{}
	for serviceID, _ := range successfulRegistrations {
		serviceIDsToRemove[serviceID] = true
	}
	defer func() {
		userServiceFilters := &service.ServiceFilters{
			IDs: serviceIDsToRemove,
		}
		_, failedToDestroyGUIDs, err := DestroyUserServices(ctx, enclaveID, userServiceFilters, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
		if err != nil {
			for serviceID, _ := range serviceIDsToRemove {
				failedServicesPool[serviceID] = stacktrace.Propagate(err, "Attempted to destroy all services with IDs '%v' together but had no success. You must manually destroy the service '%v'!", serviceIDsToRemove, serviceID)
			}
			return
		}
		if len(failedToDestroyGUIDs) == 0 {
			return
		}
		for serviceID, registration := range successfulRegistrations {
			destroyErr, found := failedToDestroyGUIDs[registration.GetGUID()]
			if !found {
				continue
			}
			failedServicesPool[serviceID] = stacktrace.Propagate(destroyErr, "Failed to destroy the service '%v' after it failed to start. You must manually destroy the service!", serviceID)
		}
	}()
	for serviceID, registrationError := range failedRegistrations {
		failedServicesPool[serviceID] = stacktrace.Propagate(registrationError, "Failed to register service with ID '%v'", serviceID)
	}

	serviceConfigsToStart := map[service.ServiceID]*service.ServiceConfig{}
	for serviceID, serviceConfig := range services {
		if _, found := successfulRegistrations[serviceID]; !found {
			continue
		}
		serviceConfigsToStart[serviceID] = serviceConfig
	}

	// If no services had successful registrations, return immediately
	// This is to prevent an empty filter being used to query for matching objects and resources, returning all services
	// and causing logic to break (eg. check for duplicate service GUIDs)
	// Making this check allows us to eject early and maintain a guarantee that objects and resources returned
	// are 1:1 with serviceGUIDs
	if len(serviceConfigsToStart) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	// Sanity check for port bindings on all services
	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	for _, config := range serviceConfigsToStart {
		publicPorts := config.GetPublicPorts()
		if publicPorts != nil && len(publicPorts) > 0 {
			logrus.Warn("The Kubernetes Kurtosis backend doesn't support defining static ports for services; the public ports will be ignored.")
		}
	}

	// Get existing objects
	serviceIDsToFilter := map[service.ServiceID]bool{}
	for serviceID := range serviceConfigsToStart {
		serviceIDsToFilter[serviceID] = true
	}
	existingObjectsAndResourcesFilters := &service.ServiceFilters{
		IDs: serviceIDsToFilter,
	}
	existingObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(
		ctx,
		enclaveID,
		existingObjectsAndResourcesFilters,
		cliModeArgs,
		apiContainerModeArgs,
		engineServerModeArgs,
		kubernetesManager,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user service objects and Kubernetes resources matching service IDs '%v'", serviceIDsToFilter)
	}
	for serviceID, _ := range serviceIDsToFilter {
		registration, found := successfulRegistrations[serviceID]
		guid := registration.GetGUID()
		if !found {
			failedServicesPool[serviceID] = stacktrace.NewError("Couldn't find a service registration for service ID '%v'. This is a bug in Kurtosis.", serviceID)
			delete(serviceConfigsToStart, serviceID)
		}
		if _, found := existingObjectsAndResources[guid]; !found {
			failedServicesPool[serviceID] = stacktrace.NewError("Couldn't find any service registrations for service GUID '%v' for service ID '%v'. This is a bug in Kurtosis.", guid, serviceID)
			delete(serviceConfigsToStart, serviceID)
		}
	}
	if len(existingObjectsAndResources) > len(serviceIDsToFilter) {
		// Should never happen because service GUIDs should be unique
		return nil, nil, stacktrace.NewError("Found more than one service registration matching service GUIDs; this is a bug in Kurtosis")
	}

	successfulStarts, failedStarts, err := runStartServiceOperationsInParallel(
		ctx,
		enclaveID,
		serviceConfigsToStart,
		existingObjectsAndResources,
		kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to start services in parallel.")
	}

	// Add operations to their respective pools
	for serviceID, service := range successfulStarts {
		successfulServicesPool[serviceID] = service
	}

	for serviceID, serviceErr := range failedStarts {
		failedServicesPool[serviceID] = serviceErr
	}

	// Do not remove services that were started successfully
	for serviceID, _ := range successfulServicesPool {
		delete(serviceIDsToRemove, serviceID)
	}
	return successfulServicesPool, failedServicesPool, nil
}

// ====================================================================================================
//                       Private helper functions
// ====================================================================================================
func runStartServiceOperationsInParallel(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	services map[service.ServiceID]*service.ServiceConfig,
	servicesObjectsAndResources map[service.ServiceGUID]*shared_helpers.UserServiceObjectsAndKubernetesResources,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceID]*service.Service,
	map[service.ServiceID]error,
	error,
) {
	startServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	servicesObjectsAndResourcesByServiceID := map[service.ServiceID]*shared_helpers.UserServiceObjectsAndKubernetesResources{}
	for _, servicesObjectsAndResource := range servicesObjectsAndResources {
		serviceID := servicesObjectsAndResource.ServiceRegistration.GetID()
		servicesObjectsAndResourcesByServiceID[serviceID] = servicesObjectsAndResource
	}
	for serviceID, config := range services {
		startServiceOperations[operation_parallelizer.OperationID(serviceID)] = createStartServiceOperation(
			ctx,
			serviceID,
			config,
			servicesObjectsAndResourcesByServiceID,
			enclaveID,
			kubernetesManager)
	}

	successfulServiceObjs, failedOperations := operation_parallelizer.RunOperationsInParallel(startServiceOperations)

	successfulServices := map[service.ServiceID]*service.Service{}
	failedServices := map[service.ServiceID]error{}

	for id, data := range successfulServiceObjs {
		serviceID := service.ServiceID(id)
		serviceObjectPtr, ok := data.(*service.Service)
		if !ok {
			return nil, nil, stacktrace.NewError(
				"An error occurred downcasting data returned from the start user service operation for service with ID: %v. "+
					"This is a Kurtosis bug. Make sure the desired type is actually being returned in the corresponding Operation.", serviceID)
		}
		successfulServices[serviceID] = serviceObjectPtr
	}

	for id, err := range failedOperations {
		serviceID := service.ServiceID(id)
		failedServices[serviceID] = err
	}

	return successfulServices, failedServices, nil
}

func createStartServiceOperation(
	ctx context.Context,
	serviceID service.ServiceID,
	serviceConfig *service.ServiceConfig,
	servicesObjectsAndResources map[service.ServiceID]*shared_helpers.UserServiceObjectsAndKubernetesResources,
	enclaveID enclave.EnclaveID,
	kubernetesManager *kubernetes_manager.KubernetesManager) operation_parallelizer.Operation {

	return func() (interface{}, error) {
		filesArtifactsExpansion := serviceConfig.GetFilesArtifactsExpansion()
		containerImageName := serviceConfig.GetContainerImageName()
		privatePorts := serviceConfig.GetPrivatePorts()
		entrypointArgs := serviceConfig.GetEntrypointArgs()
		cmdArgs := serviceConfig.GetCmdArgs()
		envVars := serviceConfig.GetEnvVars()
		cpuAllocationMillicpus := serviceConfig.GetCPUAllocationMillicpus()
		memoryAllocationMegabytes := serviceConfig.GetMemoryAllocationMegabytes()
		privateIPAddrPlaceholder := serviceConfig.GetPrivateIPAddrPlaceholder()

		matchingObjectAndResources, found := servicesObjectsAndResources[serviceID]
		if !found {
			return nil, stacktrace.NewError("Even though we pulled back some Kubernetes resources, no Kubernetes resources were available for requested service ID '%v'; this is a bug in Kurtosis", serviceID)
		}
		kubernetesService := matchingObjectAndResources.KubernetesResources.Service
		serviceObj := matchingObjectAndResources.Service
		if serviceObj != nil {
			return nil, stacktrace.NewError("Cannot start service with ID '%v' because the service has already been started previously", serviceID)
		}

		// We replace the placeholder value with the actual private IP address
		privateIPAddr := matchingObjectAndResources.ServiceRegistration.GetPrivateIP().String()
		for index, _ := range entrypointArgs {
			entrypointArgs[index] = strings.Replace(entrypointArgs[index], privateIPAddrPlaceholder, privateIPAddr, unlimitedReplacements)
		}
		for index, _ := range cmdArgs {
			cmdArgs[index] = strings.Replace(cmdArgs[index], privateIPAddrPlaceholder, privateIPAddr, unlimitedReplacements)
		}
		for key, _ := range envVars {
			envVars[key] = strings.Replace(envVars[key], privateIPAddrPlaceholder, privateIPAddr, unlimitedReplacements)
		}

		namespaceName := kubernetesService.GetNamespace()
		serviceRegistrationObj := matchingObjectAndResources.ServiceRegistration
		serviceGUID := serviceRegistrationObj.GetGUID()

		var podInitContainers []apiv1.Container
		var podVolumes []apiv1.Volume
		var userServiceContainerVolumeMounts []apiv1.VolumeMount
		if filesArtifactsExpansion != nil {
			podVolumes, userServiceContainerVolumeMounts, podInitContainers, _ = prepareFilesArtifactsExpansionResources(
				filesArtifactsExpansion.ExpanderImage,
				filesArtifactsExpansion.ExpanderEnvVars,
				filesArtifactsExpansion.ExpanderDirpathsToServiceDirpaths,
			)
		}

		objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
		enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveID)

		// Create the pod
		podAttributes, err := enclaveObjAttributesProvider.ForUserServicePod(serviceGUID, serviceID, privatePorts)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting attributes for new pod for service with ID '%v'", serviceID)
		}
		podLabelsStrs := shared_helpers.GetStringMapFromLabelMap(podAttributes.GetLabels())
		podAnnotationsStrs := shared_helpers.GetStringMapFromAnnotationMap(podAttributes.GetAnnotations())

		podContainers, err := getUserServicePodContainerSpecs(
			containerImageName,
			entrypointArgs,
			cmdArgs,
			envVars,
			privatePorts,
			userServiceContainerVolumeMounts,
			cpuAllocationMillicpus,
			memoryAllocationMegabytes,
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
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating pod '%v' using image '%v'", podName, containerImageName)
		}
		shouldDestroyPod := true
		defer func() {
			if shouldDestroyPod {
				if err := kubernetesManager.RemovePod(ctx, createdPod); err != nil {
					logrus.Errorf("Starting service didn't complete successfully so we tried to remove the pod we created but doing so threw an error:\n%v", err)
					logrus.Errorf("ACTION REQUIRED: You'll need to remove pod '%v' in '%v' manually!!!", podName, namespaceName)
				}
			}
		}()

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

		kubernetesResources := map[service.ServiceGUID]*shared_helpers.UserServiceKubernetesResources{
			serviceGUID: {
				Service: updatedService,
				Pod:     createdPod,
			},
		}

		convertedObjects, err := shared_helpers.GetUserServiceObjectsFromKubernetesResources(enclaveID, kubernetesResources)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting a service object from the Kubernetes service and newly-created pod")
		}
		objectsAndResources, found := convertedObjects[serviceGUID]
		if !found {
			return nil, stacktrace.NewError(
				"Successfully converted the Kubernetes service + pod representing a running service with GUID '%v' to a "+
					"Kurtosis object, but couldn't find that key in the resulting map; this is a bug in Kurtosis",
				serviceGUID,
			)
		}

		shouldDestroyPod = false
		shouldUndoServiceUpdate = false
		return objectsAndResources.Service, nil
	}
}

// Update the service to:
// - Set the service ports appropriately
// - Irrevocably record that a pod is bound to the service (so that even if the pod is deleted, the service won't
// 	 be usable again
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
				Name:     &newServicePortCopy.Name,
				Protocol: &newServicePortCopy.Protocol,
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
					Name:     &oldServicePort.Name,
					Protocol: &oldServicePort.Protocol,
					// TODO fill this out for an app port!
					AppProtocol: nil,
					Port:        &oldServicePort.Port,
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
) (
	[]apiv1.Container,
	error,
) {
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

	resourceLimitsList := apiv1.ResourceList{}
	resourceRequestsList := apiv1.ResourceList{}
	// 0 is considered the empty value (meaning the field was never set), so if either fields are 0, that resource is left unbounded
	if cpuAllocationMillicpus != 0 {
		resourceLimitsList[apiv1.ResourceCPU] = *resource.NewMilliQuantity(int64(cpuAllocationMillicpus), resource.DecimalSI)
		resourceRequestsList[apiv1.ResourceCPU] = *resource.NewMilliQuantity(int64(cpuAllocationMillicpus), resource.DecimalSI)
	}
	if memoryAllocationMegabytes != 0 {
		memoryAllocationInBytes := convertMegabytesToBytes(memoryAllocationMegabytes)
		resourceLimitsList[apiv1.ResourceMemory] = *resource.NewQuantity(int64(memoryAllocationInBytes), resource.DecimalSI)
		resourceRequestsList[apiv1.ResourceMemory] = *resource.NewQuantity(int64(memoryAllocationInBytes), resource.DecimalSI)
	}
	resourceRequirements := apiv1.ResourceRequirements{
		Limits: resourceLimitsList,
		Requests: resourceRequestsList,
	}

	// TODO create networking sidecars here
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

	return containers, nil
}

func getKubernetesServicePortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ServicePort, error) {
	result := []apiv1.ServicePort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetProtocol()
		kubernetesProtocol, found := kurtosisPortProtocolToKubernetesPortProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ServicePort{
			Name:        portId,
			Protocol:    kubernetesProtocol,
			// TODO Specify this!!! Will make for a really nice user interface (e.g. "https")
			AppProtocol: nil,
			// Safe to cast because max uint16 < int32
			Port:        int32(portSpec.GetNumber()),
		}
		result = append(result, kubernetesPortObj)
	}
	return result, nil
}

func getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ContainerPort, error) {
	result := []apiv1.ContainerPort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetProtocol()
		kubernetesProtocol, found := kurtosisPortProtocolToKubernetesPortProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ContainerPort{
			Name:          portId,
			// Safe to do because max uint16 < int32
			ContainerPort: int32(portSpec.GetNumber()),
			Protocol:      kubernetesProtocol,
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
	enclaveID enclave.EnclaveID,
	serviceIDs []service.ServiceID,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager) (map[service.ServiceID]*service.ServiceRegistration, map[service.ServiceID]error, error) {
	successfulServicesPool := map[service.ServiceID]*service.ServiceRegistration{}
	failedServicesPool := map[service.ServiceID]error{}

	// If no services were passed in to register, return empty maps
	if len(serviceIDs) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	namespaceName, err := shared_helpers.GetEnclaveNamespaceName(ctx, enclaveID, cliModeArgs, apiContainerModeArgs, engineServerModeArgs, kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveID)
	}

	objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
	enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveID)

	registerServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for _, serviceID := range serviceIDs {
		registerServiceOperations[operation_parallelizer.OperationID(serviceID)] = createRegisterUserServiceOperation(
			ctx,
			enclaveID,
			serviceID,
			namespaceName,
			enclaveObjAttributesProvider,
			kubernetesManager)
	}

	successfulRegistrations, failedRegistrations := operation_parallelizer.RunOperationsInParallel(registerServiceOperations)

	for id, data := range successfulRegistrations {
		serviceID := service.ServiceID(id)
		serviceRegistration, ok := data.(*service.ServiceRegistration)
		if !ok {
			return nil, nil, stacktrace.NewError(
				"An error occurred downcasting data returned from the register user service operation for service with id: %v. "+
					"This is a Kurtosis bug. Make sure the desired type is actually being returned in the corresponding Operation.", serviceID)
		}
		successfulServicesPool[serviceID] = serviceRegistration
	}

	for opID, err := range failedRegistrations {
		failedServicesPool[service.ServiceID(opID)] = err
	}

	return successfulServicesPool, failedServicesPool, nil
}

func createRegisterUserServiceOperation(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	serviceID service.ServiceID,
	namespaceName string,
	enclaveObjAttributesProvider object_attributes_provider.KubernetesEnclaveObjectAttributesProvider,
	kubernetesManager *kubernetes_manager.KubernetesManager) operation_parallelizer.Operation {
	return func() (interface{}, error) {
		serviceGuidStr, err := uuid_generator.GenerateUUIDString()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred generating a UUID to use for the service GUID")
		}
		serviceGuid := service.ServiceGUID(serviceGuidStr)
		serviceAttributes, err := enclaveObjAttributesProvider.ForUserServiceService(serviceGuid, serviceID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting attributes for the Kubernetes service for user service '%v'", serviceID)
		}

		serviceNameStr := serviceAttributes.GetName().GetString()

		serviceLabelsStrs := shared_helpers.GetStringMapFromLabelMap(serviceAttributes.GetLabels())
		serviceAnnotationsStrs := shared_helpers.GetStringMapFromAnnotationMap(serviceAttributes.GetAnnotations())

		// Set up the labels that the pod will match (i.e. the labels of the pod-to-be)
		// WARNING: We *cannot* use the labels of the Service itself because we're not guaranteed that the labels
		//  between the two will be identical!
		serviceGuidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(serviceGuid))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the service GUID '%v'", serviceGuid)
		}
		enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(enclaveID))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes pod match label value for the enclave ID '%v'", enclaveID)
		}
		matchedPodLabels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
			label_key_consts.AppIDKubernetesLabelKey:     label_value_consts.AppIDKubernetesLabelValue,
			label_key_consts.EnclaveIDKubernetesLabelKey: enclaveIdLabelValue,
			label_key_consts.GUIDKubernetesLabelKey:      serviceGuidLabelValue,
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
			return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes service in enclave '%v' with ID '%v'", enclaveID, serviceID)
		}
		shouldDeleteService := true
		defer func() {
			if shouldDeleteService {
				if err := kubernetesManager.RemoveService(ctx, createdService); err != nil {
					logrus.Errorf("Registering service '%v' didn't complete successfully so we tried to remove the Kubernetes service we created but doing so threw an error:\n%v", serviceID, err)
					logrus.Errorf("ACTION REQUIRED: You'll need to remove service '%v' in namespace '%v' manually!!!", createdService.Name, namespaceName)
				}
			}
		}()

		kubernetesResources := map[service.ServiceGUID]*shared_helpers.UserServiceKubernetesResources{
			serviceGuid: {
				Service: createdService,
				Pod:     nil, // No pod yet
			},
		}

		convertedObjects, err := shared_helpers.GetUserServiceObjectsFromKubernetesResources(enclaveID, kubernetesResources)
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

		shouldDeleteService = false
		return objectsAndResources.ServiceRegistration, nil
	}
}

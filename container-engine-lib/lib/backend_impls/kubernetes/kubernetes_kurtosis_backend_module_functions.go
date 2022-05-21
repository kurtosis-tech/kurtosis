package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_resource_collectors"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	apiv1 "k8s.io/api/core/v1"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"net"
	"time"
)

const (
	kurtosisModuleContainerName = "kurtosis-module-container"

	kurtosisModulePortProtocol = port_spec.PortProtocol_TCP

	// Our module don't need service accounts
	moduleServiceAccountName = ""

	maxWaitForModuleContainerAvailabilityRetries         = 30
	timeBetweenWaitForModuleContainerAvailabilityRetries = 1 * time.Second

	shouldAddTimestampsToModuleLogs = false
)

type moduleObjectsAndKubernetesResources struct {
	module *module.Module
	kubernetesResources *moduleKubernetesResources
}

// Any of these values being nil indicates that the resource doesn't exist
type moduleKubernetesResources struct {

	//This should never be nil as the canonical reference of a module is its service
	service *apiv1.Service

	//This can be nil if the pod wasn't started yet, or has been removed already
	pod *apiv1.Pod
}

// ====================================================================================================
//                                     		Module CRUD Methods
// ====================================================================================================

func (backend *KubernetesKurtosisBackend) CreateModule(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	id module.ModuleID,
	grpcPortNum uint16,
	envVars map[string]string,
) (
	newModule *module.Module,
	resultErr error,
) {
	guidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID string for module with ID '%v'", id)
	}
	guid := module.ModuleGUID(guidStr)

	enclaveAttributesProvider:= backend.objAttrsProvider.ForEnclave(enclaveId)

	enclaveNamespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave namespace name for enclave with ID '%v'", enclaveId)
	}

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, kurtosisModulePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the module's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			kurtosisModulePortProtocol.String(),
		)
	}

	privatePorts := map[string]*port_spec.PortSpec{
		kurtosisInternalContainerGrpcPortSpecId: privateGrpcPortSpec,
	}

	// Get Pod Attributes so that we can select them with the Service
	modulePodAttributes, err := enclaveAttributesProvider.ForModulePod(guid, id, privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get module Kubernetes pod attributes with module id '%v', guid '%v' and private ports '%+v', instead got a non-nil error",
			id,
			guid,
			privatePorts,
		)
	}
	modulePodName := modulePodAttributes.GetName().GetString()
	modulePodLabels := getStringMapFromLabelMap(modulePodAttributes.GetLabels())
	modulePodAnnotations := getStringMapFromAnnotationMap(modulePodAttributes.GetAnnotations())

	// Get Service Attributes
	moduleServiceAttributes, err := enclaveAttributesProvider.ForModuleService(
		guid,
		id,
		privatePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the module service attributes using private grpc port spec '%+v'",
			privateGrpcPortSpec,
		)
	}
	moduleServiceName := moduleServiceAttributes.GetName().GetString()
	moduleServiceLabels := getStringMapFromLabelMap(moduleServiceAttributes.GetLabels())
	moduleServiceAnnotations := getStringMapFromAnnotationMap(moduleServiceAttributes.GetAnnotations())

	// Define service ports. These hook up to ports on the containers running in the module pod
	// Kubernetes will assign a public port number to them
	servicePorts, err := getKubernetesServicePortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes service ports from the private port specs map '%+v'", privatePorts)
	}

	moduleService, err := backend.kubernetesManager.CreateService(
		ctx,
		enclaveNamespaceName,
		moduleServiceName,
		moduleServiceLabels,
		moduleServiceAnnotations,
		modulePodLabels,
		apiv1.ServiceTypeClusterIP,
		servicePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the service with name '%s', labels '%+v', annotations '%+v' " +
				", ports '%+v' and pod labels '%+v' in namespace '%s'",
			moduleServiceName,
			moduleServiceLabels,
			moduleServiceAnnotations,
			servicePorts,
			modulePodLabels,
			enclaveNamespaceName,
		)
	}
	var shouldRemoveService = true
	defer func() {
		if shouldRemoveService {
			if err := backend.kubernetesManager.RemoveService(ctx, enclaveNamespaceName, moduleServiceName); err != nil {
				logrus.Errorf("Creating the module didn't complete successfully, so we tried to delete Kubernetes service '%v' that we created but an error was thrown:\n%v", moduleServiceName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes service with name '%v'!!!!!!!", moduleServiceName)
			}
		}
	}()

	// Create the Pod
	containerPorts, err := getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes container ports from the private port specs map '%+v'", privatePorts)
	}

	moduleContainers := getModuleContainers(image, containerPorts, envVars)

	// Create pod with module containers and not volumes in Kubernetes
	modulePod, err := backend.kubernetesManager.CreatePod(
		ctx,
		enclaveNamespaceName,
		modulePodName,
		modulePodLabels,
		modulePodAnnotations,
		moduleContainers,
		nil,
		moduleServiceAccountName,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the pod with name '%s', labels '%+v', annotations '%+v'," +
				" containers '%v' in namespace '%s' using image '%s'",
			modulePodName,
			modulePodLabels,
			modulePodAnnotations,
			moduleContainers,
			enclaveNamespaceName,
			image,
		)
	}
	var shouldRemovePod = true
	defer func() {
		if shouldRemovePod {
			if err := backend.kubernetesManager.RemovePod(ctx, enclaveNamespaceName, modulePodName); err != nil {
				logrus.Errorf("Creating the module didn't complete successfully, so we tried to delete Kubernetes pod '%v' that we created but an error was thrown:\n%v", modulePodName, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove Kubernetes pod with name '%v'!!!!!!!", modulePodName)
			}
		}
	}()

	moduleResources := &moduleKubernetesResources{
		service:          moduleService,
		pod:              modulePod,
	}
	moduleObjsById, err := getModuleObjectsFromKubernetesResources(enclaveId, map[module.ModuleGUID]*moduleKubernetesResources{
		guid: moduleResources,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the new module's Kubernetes resources to module objects")
	}
	resultModuleObjectAndResources, found := moduleObjsById[guid]
	if !found {
		return nil, stacktrace.NewError("Successfully converted the new module's Kubernetes resources to an module object, but the resulting map didn't have an entry for GUID '%v'", guid)
	}

	resultModule := resultModuleObjectAndResources.module

	if err := waitForPortAvailabilityUsingNetstat(
		backend.kubernetesManager,
		enclaveNamespaceName,
		modulePodName,
		kurtosisModuleContainerName,
		privateGrpcPortSpec,
		maxWaitForModuleContainerAvailabilityRetries,
		timeBetweenWaitForModuleContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the module grpc port '%v/%v' to become available", privateGrpcPortSpec.GetProtocol(), privateGrpcPortSpec.GetNumber())
	}

	shouldRemovePod = false
	shouldRemoveService = false
	return resultModule, nil
}

func (backend *KubernetesKurtosisBackend) GetModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	map[module.ModuleGUID]*module.Module,
	error,
) {

	allObjectsAndResources, err := backend.getMatchingModuleObjectsAndKubernetesResources(ctx, enclaveId, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting modules and Kubernetes resources matching filters '%+v'", filters)
	}

	results := map[module.ModuleGUID]*module.Module{}
	for guid, moduleObjsAndResources := range allObjectsAndResources {
		results[guid] = moduleObjsAndResources.module
	}
	return results, nil
}

func (backend *KubernetesKurtosisBackend) GetModuleLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
	shouldFollowLogs bool,
) (
	map[module.ModuleGUID]io.ReadCloser,
	map[module.ModuleGUID]error,
	error,
) {
	moduleObjectsAndResources, err := backend.getMatchingModuleObjectsAndKubernetesResources(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Expected to be able to get modules and Kubernetes resources, instead a non nil error was returned")
	}
	moduleLogs := map[module.ModuleGUID]io.ReadCloser{}
	erroredModuleLogs := map[module.ModuleGUID]error{}
	shouldCloseLogStreams := true
	for _, moduleObjectAndResource := range moduleObjectsAndResources {
		moduleGuid := moduleObjectAndResource.module.GetGUID()
		modulePod := moduleObjectAndResource.kubernetesResources.pod
		if modulePod == nil {
			erroredModuleLogs[moduleGuid] = stacktrace.NewError("Expected to find a pod for Kurtosis module with GUID '%v', instead no pod was found", moduleGuid)
			continue
		}
		enclaveNamespaceName := moduleObjectAndResource.kubernetesResources.service.GetNamespace()
		// Get logs
		logStream, err := backend.kubernetesManager.GetContainerLogs(
			ctx,
			enclaveNamespaceName,
			modulePod.Name,
			kurtosisModuleContainerName,
			shouldFollowLogs,
			shouldAddTimestampsToModuleLogs,
		)
		if err != nil {
			erroredModuleLogs[moduleGuid] = stacktrace.Propagate(err, "Expected to be able to call Kubernetes to get logs for module with GUID '%v', instead a non-nil error was returned", moduleGuid)
			continue
		}
		defer func() {
			if shouldCloseLogStreams {
				logStream.Close()
			}
		}()

		moduleLogs[moduleGuid] = logStream
	}

	shouldCloseLogStreams = false
	return moduleLogs, erroredModuleLogs, nil
}

func (backend *KubernetesKurtosisBackend) StopModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	map[module.ModuleGUID]bool,
	map[module.ModuleGUID]error,
	error,
) {
	allObjectsAndResources, err := backend.getMatchingModuleObjectsAndKubernetesResources(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting modules and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulModuleGUIDs := map[module.ModuleGUID]bool{}
	erroredModuleGUIDs := map[module.ModuleGUID]error{}
	for moduleGuid, moduleObjsAndResources := range allObjectsAndResources {
		kubernetesService := moduleObjsAndResources.kubernetesResources.service
		if kubernetesService != nil {
			namespaceName := kubernetesService.GetNamespace()
			serviceName := kubernetesService.GetName()
			updateConfigurator := func(updatesToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
				specUpdates := applyconfigurationsv1.ServiceSpec().WithSelector(nil)
				updatesToApply.WithSpec(specUpdates)
			}
			if _, err := backend.kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updateConfigurator); err != nil {
				erroredModuleGUIDs[moduleGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing selectors from service '%v' in namespace '%v' for module with GUID '%v'",
					serviceName,
					namespaceName,
					moduleGuid,
				)
				continue
			}
		}

		kubernetesPod := moduleObjsAndResources.kubernetesResources.pod
		if kubernetesPod != nil {
			podName := kubernetesPod.GetName()
			namespaceName := kubernetesPod.GetNamespace()
			if err := backend.kubernetesManager.RemovePod(ctx, namespaceName, podName); err != nil {
				erroredModuleGUIDs[moduleGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing pod '%v' in namespace '%v' for module with GUID '%v'",
					podName,
					namespaceName,
					moduleGuid,
				)
				continue
			}
		}

		successfulModuleGUIDs[moduleGuid] = true
	}

	return successfulModuleGUIDs, erroredModuleGUIDs, nil
}

func (backend *KubernetesKurtosisBackend) DestroyModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	map[module.ModuleGUID]bool,
	map[module.ModuleGUID]error,
	error,
) {
	allObjectsAndResources, err := backend.getMatchingModuleObjectsAndKubernetesResources(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting modules and Kubernetes resources matching filters '%+v'", filters)
	}

	successfulModuleGUIDs := map[module.ModuleGUID]bool{}
	erroredModuleGUIDs := map[module.ModuleGUID]error{}
	for moduleGuid, moduleObjsAndResources := range allObjectsAndResources {

		// Remove Pod
		kubernetesPod := moduleObjsAndResources.kubernetesResources.pod
		if kubernetesPod != nil {
			podName := kubernetesPod.GetName()
			namespaceName := kubernetesPod.GetNamespace()
			if err := backend.kubernetesManager.RemovePod(ctx, podName, namespaceName); err != nil {
				erroredModuleGUIDs[moduleGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing pod '%v' in namespace '%v' for module with GUID '%v'",
					podName,
					namespaceName,
					moduleGuid,
				)
				continue
			}
		}

		// Remove Service
		kubernetesService := moduleObjsAndResources.kubernetesResources.service
		if kubernetesService != nil {
			serviceName := kubernetesService.GetName()
			namespaceName := kubernetesService.GetNamespace()
			if err := backend.kubernetesManager.RemoveService(ctx, serviceName, namespaceName); err != nil {
				erroredModuleGUIDs[moduleGuid] = stacktrace.Propagate(
					err,
					"An error occurred removing service '%v' in namespace '%v' for module with GUID '%v'",
					serviceName,
					namespaceName,
					moduleGuid,
				)
				continue
			}
		}

		successfulModuleGUIDs[moduleGuid] = true
	}
	return successfulModuleGUIDs, erroredModuleGUIDs, nil
}


// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func getModuleContainers(
	containerImageAndTag string,
	containerPorts []apiv1.ContainerPort,
	envVars map[string]string,
) (
	resultContainers []apiv1.Container,
) {

	var containerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVars {
		envVar := apiv1.EnvVar{
			Name:  varName,
			Value: varValue,
		}
		containerEnvVars = append(containerEnvVars, envVar)
	}
	containers := []apiv1.Container{
		{
			Name:  kurtosisModuleContainerName,
			Image: containerImageAndTag,
			Env:   containerEnvVars,
			Ports: containerPorts,

		},
	}

	return containers
}

func (backend *KubernetesKurtosisBackend) getMatchingModuleObjectsAndKubernetesResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	map[module.ModuleGUID]*moduleObjectsAndKubernetesResources,
	error,
) {
	matchingResources, err := backend.getModuleKubernetesResourcesMatchingGuids(ctx, enclaveId, filters.GUIDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting module Kubernetes resources matching module GUIDs: %+v", filters.GUIDs)
	}

	moduleObjectsAndResources, err := getModuleObjectsFromKubernetesResources(enclaveId, matchingResources)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting module objects from Kubernetes resources")
	}

	// Sanity check
	if len(matchingResources) != len(moduleObjectsAndResources) {
		return nil, stacktrace.NewError(
			"Transformed %v Kubernetes resource objects into %v Kurtosis objects; this is a bug in Kurtosis",
			len(matchingResources),
			len(moduleObjectsAndResources),
		)
	}

	// Finally, apply the filters
	results := map[module.ModuleGUID]*moduleObjectsAndKubernetesResources{}
	for moduleGuid, objectsAndResources := range moduleObjectsAndResources {
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[objectsAndResources.module.GetGUID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[objectsAndResources.module.GetStatus()]; !found {
				continue
			}
		}

		results[moduleGuid] = objectsAndResources
	}

	return results, nil
}

// Get back any and all module's Kubernetes resources matching the given GUIDs, where a nil or empty map == "match all GUIDs"
func (backend *KubernetesKurtosisBackend) getModuleKubernetesResourcesMatchingGuids(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	moduleGuids map[module.ModuleGUID]bool,
) (
	map[module.ModuleGUID]*moduleKubernetesResources,
	error,
) {
	namespaceName, err := backend.getEnclaveNamespaceName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting namespace name for enclave '%v'", enclaveId)
	}

	// TODO switch to properly-typed KubernetesLabelValue object!!!
	postFilterLabelValues := map[string]bool{}
	for moduleGuid := range moduleGuids {
		postFilterLabelValues[string(moduleGuid)] = true
	}

	kubernetesResourceSearchLabels := map[string]string{
		label_key_consts.AppIDKubernetesLabelKey.GetString(): label_value_consts.AppIDKubernetesLabelValue.GetString(),
		label_key_consts.EnclaveIDKubernetesLabelKey.GetString(): string(enclaveId),
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey.GetString(): label_value_consts.ModuleKurtosisResourceTypeKubernetesLabelValue.GetString(),
	}

	results := map[module.ModuleGUID]*moduleKubernetesResources{}

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
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes services matching module GUIDs: %+v", moduleGuids)
	}
	for moduleGuidStr, kubernetesServicesForGuid := range matchingKubernetesServices {
		moduleGuid := module.ModuleGUID(moduleGuidStr)

		numServicesForGuid := len(kubernetesServicesForGuid)
		if numServicesForGuid == 0 {
			// This would indicate a bug in our service retrieval logic because we shouldn't even have a map entry if there's nothing matching it
			return nil, stacktrace.NewError("Got entry of result services for module GUID '%v', but no Kubernetes services were returned; this is a bug in Kurtosis", moduleGuid)
		}
		if numServicesForGuid > 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes services associated with module GUID '%v'; this is a bug in Kurtosis", numServicesForGuid, moduleGuid)
		}
		kubernetesService := kubernetesServicesForGuid[0]

		resultObj, found := results[moduleGuid]
		if !found {
			resultObj = &moduleKubernetesResources{}
		}
		resultObj.service = kubernetesService
		results[moduleGuid] = resultObj
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
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes pods matching module GUIDs: %+v", moduleGuids)
	}
	for moduleGuidStr, kubernetesPodsForGuid := range matchingKubernetesPods {
		moduleGuid := module.ModuleGUID(moduleGuidStr)

		numPodsForGuid := len(kubernetesPodsForGuid)
		if numPodsForGuid == 0 {
			// This would indicate a bug in our pod retrieval logic because we shouldn't even have a map entry if there's nothing matching it
			return nil, stacktrace.NewError("Got entry of result pods for module GUID '%v', but no Kubernetes pods were returned; this is a bug in Kurtosis", moduleGuid)
		}
		if numPodsForGuid > 1 {
			return nil, stacktrace.NewError("Found %v Kubernetes pods associated with module GUID '%v'; this is a bug in Kurtosis", numPodsForGuid, moduleGuid)
		}
		kubernetesPod := kubernetesPodsForGuid[0]

		resultObj, found := results[moduleGuid]
		if !found {
			resultObj = &moduleKubernetesResources{}
		}
		resultObj.pod = kubernetesPod
		results[moduleGuid] = resultObj
	}

	return results, nil
}

func getModuleObjectsFromKubernetesResources(
	enclaveId enclave.EnclaveID,
	allResources map[module.ModuleGUID]*moduleKubernetesResources,
) (
	map[module.ModuleGUID]*moduleObjectsAndKubernetesResources,
	error,
) {
	results := map[module.ModuleGUID]*moduleObjectsAndKubernetesResources{}

	for moduleGuid, resources := range allResources {
		results[moduleGuid] = &moduleObjectsAndKubernetesResources{
			kubernetesResources: resources,
			// The other fields will get filled in below
		}
	}

	for moduleGuid, resultObj := range results {
		resourcesToParse := resultObj.kubernetesResources

		kubernetesService := resourcesToParse.service

		if kubernetesService == nil {
			return nil, stacktrace.NewError("Cannot create a module object if no Kubernetes service for module '%v' exists", moduleGuid)
		}

		serviceLabels := kubernetesService.Labels
		idLabelStr, found := serviceLabels[label_key_consts.IDKubernetesLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find label '%v' on the Kubernetes service but none was found", label_key_consts.IDKubernetesLabelKey.GetString())
		}
		id := module.ModuleID(idLabelStr)

		guidLabelStr, found := serviceLabels[label_key_consts.GUIDKubernetesLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find label '%v' on the Kubernetes service but none was found", label_key_consts.GUIDKubernetesLabelKey.GetString())
		}
		guid := module.ModuleGUID(guidLabelStr)

		privateIpAddr := net.ParseIP(kubernetesService.Spec.ClusterIP)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the module service, instead parsing the cluster ip of service '%v' returned nil", kubernetesService.Name)
		}

		privatePorts, err := getPrivatePortsAndValidatePortExistence(
			kubernetesService,
			map[string]bool{
				kurtosisInternalContainerGrpcPortSpecId: true,
			},
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting private module ports and validating gRPC port existence from the module service")
		}
		privateGrpcPortSpec := privatePorts[kurtosisInternalContainerGrpcPortSpecId]

		// NOTE: We set these to nil because in Kubernetes we have no way of knowing what the public info is!
		var publicIpAddr net.IP = nil
		var publicGrpcPortSpec *port_spec.PortSpec = nil

		kubernetesPod := resourcesToParse.pod
		if kubernetesPod == nil {
			// No pod here means that a) a Module had private ports but b) now has no Pod
			// This means that there used to be a Pod, but it was stopped/removed
			resultObj.module = module.NewModule(
				enclaveId,
				id,
				guid,
				container_status.ContainerStatus_Stopped,
				privateIpAddr,
				privateGrpcPortSpec,
				publicIpAddr,
				publicGrpcPortSpec,
			)
			continue
		}

		status, err := getContainerStatusFromPod(resourcesToParse.pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container status from Kubernetes pod '%+v'", resourcesToParse.pod)
		}

		moduleObj := module.NewModule(
			enclaveId,
			id,
			guid,
			status,
			privateIpAddr,
			privateGrpcPortSpec,
			publicIpAddr,
			publicGrpcPortSpec,
		)

		resultObj.module = moduleObj
	}
	return results, nil
}

package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/object_name_constants"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

const (
	kurtosisModulePortProtocol = port_spec.PortProtocol_TCP

	// The location where the enclave data volume will be mounted
	//  on the module container
	enclaveDataVolumeDirpathOnModuleContainer = "/kurtosis-data"

	// Our module don't need service accounts
	moduleServiceAccountName = ""
)

// Any of these values being nil indicates that the resource doesn't exist
type moduleKubernetesResources struct {

	serviceAccount *apiv1.ServiceAccount

	service *apiv1.Service

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
	guid module.ModuleGUID,
	grpcPortNum uint16,
	envVars map[string]string,
) (
	newModule *module.Module,
	resultErr error,
) {
	//TODO This validation is the same for Docker and for Kubernetes because we are using kurtBackend
	//TODO we could move this to a top layer for validations, perhaps

	// Verify no module container with the given GUID already exists in the enclave
	preexistingModuleFilters := &module.ModuleFilters{
		GUIDs: map[module.ModuleGUID]bool{
			guid: true,
		},
	}
	preexistingModules, err := backend.GetModules(ctx, preexistingModuleFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting preexisting modules with GUID '%v'", guid)
	}
	if len(preexistingModules) > 0 {
		return nil, stacktrace.NewError("Found existing module container(s) in with GUID '%v'; cannot start a new one", guid)
	}

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

	// Create the Pod
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

	enclaveDataPersistentVolumeClaim, err := backend.getEnclaveDataPersistentVolumeClaim(ctx, enclaveNamespaceName, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data persistent volume claim for enclave '%v' in namespace '%v'", enclaveId, enclaveNamespaceName)
	}

	grpcPortInt32 := int32(grpcPortNum)

	containerPorts := []apiv1.ContainerPort{
		{
			Name:          object_name_constants.KurtosisInternalContainerGrpcPortName.GetString(),
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: grpcPortInt32,
		},
	}

	moduleContainers, moduleVolumes := getModuleContainersAndVolumes(guid, image, containerPorts, envVars, enclaveDataPersistentVolumeClaim)

	// Create pod with module containers and volumes in Kubernetes
	modulePod, err := backend.kubernetesManager.CreatePod(
		ctx,
		enclaveNamespaceName,
		modulePodName,
		modulePodLabels,
		modulePodAnnotations,
		moduleContainers,
		moduleVolumes,
		moduleServiceAccountName,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the pod with name '%s', labels '%+v', annotations '%+v'," +
				" containers '%v' and volumes '%+v' in namespace '%s' using image '%s'",
			modulePodName,
			modulePodLabels,
			modulePodAnnotations,
			moduleContainers,
			moduleVolumes,
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
	servicePorts := []apiv1.ServicePort{
		{
			Name:     object_name_constants.KurtosisInternalContainerGrpcPortName.GetString(),
			Protocol: apiv1.ProtocolTCP,
			Port:     grpcPortInt32,
		},
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
	resultModule, found := moduleObjsById[guid]
	if !found {
		return nil, stacktrace.NewError("Successfully converted the new module's Kubernetes resources to an module object, but the resulting map didn't have an entry for GUID '%v'", guid)
	}

	//TODO implement GRPC availability Checker

	shouldRemovePod = false
	shouldRemoveService = false
	return resultModule, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func getModuleContainersAndVolumes(
	moduleGuid module.ModuleGUID,
	containerImageAndTag string,
	containerPorts []apiv1.ContainerPort,
	envVars map[string]string,
	enclaveDataPersistentVolumeClaim *apiv1.PersistentVolumeClaim,
) (
	resultContainers []apiv1.Container,
	resultVolumes []apiv1.Volume,
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
			Name:  string(moduleGuid),
			Image: containerImageAndTag,
			Env:   containerEnvVars,
			Ports: containerPorts,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      enclaveDataPersistentVolumeClaim.Spec.VolumeName,
					MountPath: enclaveDataVolumeDirpathOnModuleContainer,
				},
			},
		},
	}

	volumes := []apiv1.Volume{
		{
			Name: enclaveDataPersistentVolumeClaim.Spec.VolumeName,
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: enclaveDataPersistentVolumeClaim.GetName(),
				},
			},
		},
	}

	return containers, volumes
}

func getModuleObjectsFromKubernetesResources(
	enclaveId enclave.EnclaveID,
	allResources map[module.ModuleGUID]*moduleKubernetesResources,
) (
	map[module.ModuleGUID]*module.Module,
	error,
) {
	result := map[module.ModuleGUID]*module.Module{}

	for moduleGuid, resourcesForModuleGuid := range allResources {

		kubernetesService := resourcesForModuleGuid.service

		if kubernetesService == nil {
			return nil, stacktrace.NewError("Can not create a module object if there is not a module's Kubernetes service for module '%v'", moduleGuid)
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

		status, err := getContainerStatusFromPod(resourcesForModuleGuid.pod)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container status from Kubernetes pod '%+v'", resourcesForModuleGuid.pod)
		}

		privateIpAddr := net.ParseIP(resourcesForModuleGuid.service.Spec.ClusterIP)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the module service, instead parsing the cluster ip of service '%v' returned nil", resourcesForModuleGuid.service.Name)
		}

		serviceAnnotations := kubernetesService.Annotations
		portSpecsStr, found := serviceAnnotations[annotation_key_consts.PortSpecsAnnotationKey.GetString()]
		if !found {
			// If the service doesn't have a private port specs annotation, it means a pod was never started so there's nothing more to do
			continue
		}
		privatePorts, err := kubernetes_port_spec_serializer.DeserializePortSpecs(portSpecsStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing port specs string '%v'", portSpecsStr)
		}

		privateGrpcPortSpec, found := privatePorts[kurtosisInternalContainerGrpcPortSpecId]
		if !found {
			return nil, stacktrace.NewError("Expected to find port spec '%v' after deserializing the port spec string '%v' stored in the service's labels with GUID '%v', but was not found", kurtosisInternalContainerGrpcPortSpecId, portSpecsStr, guid)
		}

		// NOTE: We set these to nil because in Kubernetes we have no way of knowing what the public info is!
		var publicIpAddr net.IP = nil
		var publicGrpcPortSpec *port_spec.PortSpec = nil

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

		result[moduleGuid] = moduleObj
	}
	return result, nil
}

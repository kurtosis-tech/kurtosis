package kubernetes

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/object_name_constants"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	apiv1 "k8s.io/api/core/v1"
	"net"
	"strconv"
	"time"
)

const (
	// The location where the engine data directory (on the Docker host machine) will be bind-mounted
	//  on the engine server
	// This is the same for kubernetes as it is for docker
	engineDataDirpathOnEngineServerContainer = "/engine-data"

	kurtosisEngineNamespace = "default"
	// The ID of the GRPC port for Kurtosis-internal containers (e.g. API container, engine, modules, etc.) which will
	//  be stored in the port spec label
	kurtosisInternalContainerGrpcPortSpecId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	// need a proxy  that will translate grpc-web requests before they hit the main GRPC server
	kurtosisInternalContainerGrpcProxyPortSpecId = "grpcProxy"

	// The engine server uses gRPC so MUST listen on TCP (no other protocols are supported), which also
	// means that its grpc-proxy must listen on TCP
	enginePortProtocol = port_spec.PortProtocol_TCP

	maxWaitForEngineAvailabilityRetries         = 10
	timeBetweenWaitForEngineAvailabilityRetries = 1 * time.Second

	// We leave a relatively short timeout so that the engine gets a chance to gracefully clean up, but the
	// user isn't stuck waiting on a long-running operation when they tell the engine to stop
	engineStopTimeout = 1 * time.Second

	numKurtosisEngineReplicas = 1
	storageClass              = "standard"
	defaultQuantity           = "10Gi"
	externalServiceType       = "LoadBalancer"

	// Engine container port number string parsing constants
	hostMachinePortNumStrParsingBase = 10
	hostMachinePortNumStrParsingBits = 16
)

// ====================================================================================================
//                                     Engine CRUD Methods
// ====================================================================================================

func (backend *KubernetesKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	engineDataDirpathOnHostMachine string,
	envVars map[string]string,
) (
	*engine.Engine,
	error,
) {

	containerStartTimeUnixSecs := time.Now().Unix()
	engineIdStr := fmt.Sprintf("%v", containerStartTimeUnixSecs)

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, enginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			enginePortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, enginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc proxy port spec object using number '%v' and protocol '%v'",
			grpcProxyPortNum,
			enginePortProtocol.String(),
		)
	}
	engineAttributesProvider, err := backend.objAttrsProvider.ForEngine(engineIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the engine attributes using id '%v', grpc port num '%v', and "+
				"grpc proxy port num '%v'",
			engineIdStr,
			grpcPortNum,
			grpcProxyPortNum,
		)
	}

	// Get Deployment Attributes Attributes
	engineDeploymentAttributes, err := engineAttributesProvider.ForEngineDeployment()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get attributes for a kubernetes deployment for engine with  id '%v', instead got a non-nil error",
			engineIdStr,
		)
	}

	// Get Pod Attributes
	enginePodAttributes, err := engineAttributesProvider.ForEnginePod()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Expected to be able to get attributes for a kubernetes pod for engine with  id '%v', instead got a non-nil error",
			engineIdStr,
		)
	}
	engineDeploymentName := engineDeploymentAttributes.GetName().GetString()
	engineDeploymentLabels := getStringMapFromLabelMap(engineDeploymentAttributes.GetLabels())
	enginePodLabels := getStringMapFromLabelMap(enginePodAttributes.GetLabels())
	// Define Containers in our Engine Pod and hook them up to our Engine Volumes
	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		imageOrgAndRepo,
		imageVersionTag,
	)
	engineContainers, engineVolumes := getEngineContainersAndVolumes(containerImageAndTag, envVars)

	// Deploy pods with engine containers and volumes to kubernetes
	_, err = backend.kubernetesManager.CreateDeployment(ctx,
		kurtosisEngineNamespace, engineDeploymentName,
		engineDeploymentLabels, enginePodLabels, numKurtosisEngineReplicas, engineContainers, engineVolumes)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the deployment with name '%s' in namespace '%s' with image '%s'", engineDeploymentName, kurtosisEngineNamespace, containerImageAndTag)
	}

	// Get Service Attributes
	engineServiceAttributes, err := engineAttributesProvider.ForEngineService(kurtosisInternalContainerGrpcPortSpecId, privateGrpcPortSpec, kurtosisInternalContainerGrpcProxyPortSpecId, privateGrpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get attributes for a kubernetes service, instead a non-nil err was returned")
	}
	engineServiceName := engineServiceAttributes.GetName().GetString()
	engineServiceLabels := getStringMapFromLabelMap(engineServiceAttributes.GetLabels())
	engineServiceAnnotations := getStringMapFromAnnotationMap(engineServiceAttributes.GetAnnotations())
	grpcPortInt32 := int32(grpcPortNum)
	grpcProxyPortInt32 := int32(grpcProxyPortNum)
	// Define service ports. These hook up to ports on the containers running in the engine pod
	// Kubernetes will assign a public port number to them
	servicePorts := []apiv1.ServicePort{
		{
			Name:     object_name_constants.KurtosisInternalContainerGrpcPortName.GetString(),
			Protocol: apiv1.ProtocolTCP,
			Port:     grpcPortInt32,
		},
		{
			Name:     object_name_constants.KurtosisInternalContainerGrpcProxyPortName.GetString(),
			Protocol: apiv1.ProtocolTCP,
			Port:     grpcProxyPortInt32,
		},
	}

	// Create Service
	service, err := backend.kubernetesManager.CreateService(ctx, kurtosisEngineNamespace, engineServiceName, engineServiceLabels, engineServiceAnnotations, enginePodLabels, externalServiceType, servicePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service with name '%s' in namespace '%s' with ports '%v' and '%v'", engineServiceName, kurtosisEngineNamespace, grpcPortInt32, grpcProxyPortInt32)
	}

	// TODO implement retry in case load balancer doesn't have ingress set up in time for our engine service
	time.Sleep(2 * time.Second)
	service, err = backend.kubernetesManager.GetServiceByName(ctx, kurtosisEngineNamespace, service.Name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service with name '%v' in namespace '%v'", service.Name, kurtosisEngineNamespace)
	}

	// Determine our public ip address from the loadbalancer ingress
	loadBalancerIps := service.Status.LoadBalancer.Ingress
	if len(loadBalancerIps) != 1 {
		return nil, stacktrace.NewError("Expected there to be exactly 1 load balancer IP for our engine service, instead found '%v'", len(loadBalancerIps))
	}
	publicIpAddress := net.ParseIP(loadBalancerIps[0].IP)

	publicGrpcPort, publicGrpcProxyPort, err := getEngineGrpcPortSpecsFromServicePorts(service.Spec.Ports)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to determine kurtosis port specs from kubernetes service '%v', instead a non-nil err was returned")
	}

	resultEngine := engine.NewEngine(
		engineIdStr,
		container_status.ContainerStatus_Running,
		publicIpAddress, publicGrpcPort, publicGrpcProxyPort)

	return resultEngine, nil
}

func (backend *KubernetesKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[string]*engine.Engine, error) {
	matchingEngines, err := backend.getMatchingEngines(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	matchingEnginesByEngineId := map[string]*engine.Engine{}
	for _, engineObj := range matchingEngines {
		matchingEnginesByEngineId[engineObj.GetID()] = engineObj
	}

	return matchingEnginesByEngineId, nil
}

func (backend *KubernetesKurtosisBackend) StopEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineIds map[string]bool,
	resultErroredEngineIds map[string]error,
	resultErr error,
) {
	matchingEnginesByServiceName, err := backend.getMatchingEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines matching filters '%+v'", filters)
	}

	servicesToRemovePortsFrom := map[string]bool{}
	for serviceName := range matchingEnginesByServiceName {
		servicesToRemovePortsFrom[serviceName] = true
	}

	successfulServiceNames, erroredServiceNames := backend.removeSelectorsFromEngineServices(ctx, servicesToRemovePortsFrom)

	successfulEngineIds := map[string]bool{}
	erroredEngineIds := map[string]error{}

	for serviceName := range successfulServiceNames {
		engineObj, found := matchingEnginesByServiceName[serviceName]
		if !found {
			return nil, nil, stacktrace.NewError("Successfully removed ports from with kubernetes service '%v' that wasn't requested; this is a bug in Kurtosis!", serviceName)
		}
		successfulEngineIds[engineObj.GetID()] = true
	}

	for serviceName, err := range erroredServiceNames {
		engineObj, found := matchingEnginesByServiceName[serviceName]
		if !found {
			return nil, nil, stacktrace.NewError("An error occurred removing engine pod selectors for engine with ID '%v' that wasn't requested; this is a bug in Kurtosis!", serviceName)
		}
		wrappedErr := stacktrace.Propagate(err, "An error occurred removing engine selectors from kubernetes service for kurtosis engine with ID '%v' and kubernetes service name '%v'", engineObj.GetID(), serviceName)
		erroredEngineIds[engineObj.GetID()] = wrappedErr
	}

	return successfulEngineIds, erroredEngineIds, nil
}

func (backend *KubernetesKurtosisBackend) DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	successfulEngineIds map[string]bool,
	erroredEngineIds map[string]error,
	resultErr error,
) {
	//TODO implement me
	panic("implement me")

	return nil, nil, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets engines matching the search filters, indexed by their service name
func (backend *KubernetesKurtosisBackend) getMatchingEngines(ctx context.Context, filters *engine.EngineFilters) (map[string]*engine.Engine, error) {
	matchingEngines := map[string]*engine.Engine{}
	engineMatchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():        label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ResourceTypeLabelKey.GetString(): label_value_consts.EngineResourceTypeLabelValue.GetString(),
	}

	serviceList, err := backend.kubernetesManager.GetServicesByLabels(ctx, kurtosisEngineNamespace, engineMatchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engine services using labels: %+v", engineMatchLabels)
	}

	for _, service := range serviceList.Items {
		engineObj, err := getEngineObjectFromKubernetesService(service)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to get a kurtosis engine object service from kubernetes service '%v', instead a non-nil error was returned", service.Name)
		}
		// If the ID filter is specified, drop engines not matching it
		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[engineObj.GetID()]; !found {
				continue
			}
		}

		// If status filter is specified, drop engines not matching it
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[engineObj.GetStatus()]; !found {
				continue
			}
		}

		matchingEngines[service.Name] = engineObj
	}

	return matchingEngines, nil
}

// TODO parallelize to improve performance
func (backend *KubernetesKurtosisBackend) removeSelectorsFromEngineServices(ctx context.Context, serviceNames map[string]bool) (map[string]bool, map[string]error) {
	successfulServices := map[string]bool{}
	failedServices := map[string]error{}
	for serviceName, _ := range serviceNames {
		if err := backend.kubernetesManager.RemoveSelectorsFromService(ctx, kurtosisEngineNamespace, serviceName); err != nil {
			failedServices[serviceName] = err
		} else {
			successfulServices[serviceName] = true
		}
	}

	return successfulServices, failedServices
}

/*
func (backend *KubernetesKurtosisBackend) destroyEngineResources(ctx context.Context, engineId string) {
	engineObjAttrsProvider, err := backend.objAttrsProvider.ForEngine(engineId)
	engineVolumeAttributes, err := engineObjAttrsProvider.ForEngineVolume()
	enginePodAttributes, err := engineObjAttrsProvider.ForEnginePod()

	// Remove Deployment
	if err := backend.kubernetesManager.RemoveDeployment(ctx, kurtosisEngineNamespace, enginePodAttributes.GetName().GetString()); err != nil {

	}
	// Destroy Service ?

	// Destroy Persistent Volume Claim
	backend.kubernetesManager.RemovePersistentVolumeClaim(ctx, kurtosisEngineNamespace, engineVolumeAttributes.GetName().GetString())

	// Destroy Volume (maybe
}
*/

func getEngineObjectFromKubernetesService(service apiv1.Service) (*engine.Engine, error) {
	engineId, isFound := service.Labels[label_key_consts.IDLabelKey.GetString()]
	if isFound == false {
		return nil, stacktrace.NewError("Expected to be able to find label describing the engine id on service '%v' with label key '%v', but was unable to", service.Name, label_key_consts.IDLabelKey.GetString())
	}
	// the ContainerStatus naming is confusing
	engineStatus := getKurtosisStatusFromKubernetesService(service)
	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	var publicGrpcProxyPortSpec *port_spec.PortSpec
	if engineStatus == container_status.ContainerStatus_Running {
		serviceExternalIps := service.Status.LoadBalancer.Ingress
		if len(serviceExternalIps) != 1 {
			return nil, stacktrace.NewError("Expected there to be exactly 1 external IP for our engine service '%v', instead found '%v'", service.Name, len(serviceExternalIps))
		}
		publicIpAddr = net.ParseIP(serviceExternalIps[0].IP)
		var portSpecError error
		publicGrpcPortSpec, publicGrpcProxyPortSpec, portSpecError = getEngineGrpcPortSpecsFromServicePorts(service.Spec.Ports)
		if portSpecError != nil {
			return nil, stacktrace.Propagate(portSpecError, "Expected to be able to determine engine grpc port specs from kubernetes service ports for engine '%v', instead a non-nil error was returned", engineId)
		}
	}

	return engine.NewEngine(engineId, engineStatus, publicIpAddr, publicGrpcPortSpec, publicGrpcProxyPortSpec), nil

}
func getKurtosisStatusFromKubernetesService(service apiv1.Service) container_status.ContainerStatus {
	// If a Kubernetes Service has open ports, then we assume the engine is reachable, and thus not stopped
	// see stopEngineService for how we stop the engine
	// label keys and values used to determine pods this service routes traffic too
	// TODO Better determination of if the engine is reachable? Check that there are two ports with names we expect them to have?
	servicePorts := service.Spec.Selector
	if len(servicePorts) == 0 {
		return container_status.ContainerStatus_Stopped
	}
	return container_status.ContainerStatus_Running
}

func getEngineContainersAndVolumes(containerImageAndTag string, engineEnvVars map[string]string) (resultContainers []apiv1.Container, resultVolumes []apiv1.Volume) {
	// Define Engine Container and Volumes with Kubernetes API
	//containerVolumeName := "kurtosis-engine-container-volume"
	containerName := "kurtosis-engine-container"
	/*
		volumes := []apiv1.Volume{
			{
				Name: containerVolumeName,
				VolumeSource: apiv1.VolumeSource{
					PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
						ClaimName: persistentVolumeClaimName,
					},
				},
			},
		}
		// TODO determine EngineDataDirpathOnEngineServerContainer
		volumeMounts := []apiv1.VolumeMount{
			{
				Name:      containerVolumeName,
				MountPath: engineDataDirpathOnEngineServerContainer,
			},
		}
	*/

	var engineContainerEnvVars []apiv1.EnvVar
	for varName, varValue := range engineEnvVars {
		envVar := apiv1.EnvVar{
			Name:  varName,
			Value: varValue,
		}
		engineContainerEnvVars = append(engineContainerEnvVars, envVar)
	}
	containers := []apiv1.Container{
		{
			Name:  containerName,
			Image: containerImageAndTag,
			Env:   engineContainerEnvVars,
		},
	}

	return containers, nil
}

func getEngineGrpcPortSpecsFromServicePorts(servicePorts []apiv1.ServicePort) (resultGrpcPortSpec *port_spec.PortSpec, resultGrpcProxyPortSpec *port_spec.PortSpec, resultErr error) {
	var publicGrpcPort *port_spec.PortSpec
	var publicGrpcProxyPort *port_spec.PortSpec
	for _, servicePort := range servicePorts {
		servicePortName := servicePort.Name
		switch servicePortName {
		case kurtosisInternalContainerGrpcPortSpecId:
			{
				publicGrpcPortSpec, err := getPublicPortSpecFromServicePort(servicePort, enginePortProtocol)
				if err != nil {
					return nil, nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing an engine's public grpc port from kubernetes service port '%v', instead a non nil error was returned", servicePortName)
				}
				publicGrpcPort = publicGrpcPortSpec
			}
		case kurtosisInternalContainerGrpcProxyPortSpecId:
			{
				publicGrpcProxyPortSpec, err := getPublicPortSpecFromServicePort(servicePort, enginePortProtocol)
				if err != nil {
					return nil, nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing an engine's public grpc proxy port from kubernetes service port '%v', instead a non nil error was returned", servicePortName)
				}
				publicGrpcProxyPort = publicGrpcProxyPortSpec
			}
		}
	}

	if publicGrpcPort == nil || publicGrpcProxyPort == nil {
		return nil, nil, stacktrace.NewError("Expected to get public port specs from kubernetes service ports, instead got a nil pointer")
	}

	return publicGrpcPort, publicGrpcProxyPort, nil

}

// getPublicPortSpecFromServicePort returns a port_spec representing a kurtosis port spec for a service port in kubernetes
func getPublicPortSpecFromServicePort(servicePort apiv1.ServicePort, portProtocol port_spec.PortProtocol) (*port_spec.PortSpec, error) {
	publicPortNumStr := strconv.FormatInt(int64(servicePort.Port), 10)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, hostMachinePortNumStrParsingBase, hostMachinePortNumStrParsingBits)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			hostMachinePortNumStrParsingBase,
			hostMachinePortNumStrParsingBits,
		)
	}
	publicPortNum := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command
	publicGrpcPort, err := port_spec.NewPortSpec(publicPortNum, portProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a public port on a kubernetes node using number '%v' and protocol '%v', instead a non nil error was returned", publicPortNum, portProtocol)
	}

	return publicGrpcPort, nil
}

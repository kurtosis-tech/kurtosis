package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	apiv1 "k8s.io/api/core/v1"
	"net"
)

// ====================================================================================================
//                                     API Container CRUD Methods
// ====================================================================================================

/*
func (backend *KubernetesKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	grpcPortNum uint16,
	grpcProxyPortNum uint16, // TODO remove when we switch fully to enclave data volume
	enclaveDataDirpathOnHostMachine string, // The dirpath on the API container where the enclave data volume should be mounted
	enclaveDataVolumeDirpath string,
	envVars map[string]string,
) (*api_container.APIContainer, error) {

}*/

func (backend *KubernetesKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveID]*api_container.APIContainer, error) {
matchingApiContainers, err := backend.getMatchingApiContainers(ctx, filters)
if err != nil {
return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching the following filters: %+v", filters)
}

matchingApiContainersByEnclaveID := map[enclave.EnclaveID]*api_container.APIContainer{}
for _, apicObj := range matchingApiContainers {
matchingApiContainersByEnclaveID[apicObj.GetEnclaveID()] = apicObj
}

return matchingApiContainersByEnclaveID, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets api containers matching the search filters, indexed by their [namespace][service name]
func (backend *KubernetesKurtosisBackend) getMatchingApiContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[string]map[string]*api_container.APIContainer, error) {
	matchingApiContainers := map[string]map[string]*api_container.APIContainer{}
	apiContainersMatchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():                label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.KurtosisResourceTypeLabelKey.GetString(): label_value_consts.APIContainerContainerTypeLabelValue.GetString(),
	}

	for enclaveId := range filters.EnclaveIDs {
		apiContainerAttributesProvider, err := backend.objAttrsProvider.ForApiContainer(enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the api container attributes provider using enclave ID '%v'", enclaveId)
		}

		// Get Namespace Attributes
		apiContainerNamespaceAttributes, err := apiContainerAttributesProvider.ForApiContainerNamespace()
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"Expected to be able to get attributes for a kubernetes namespace for api container in enclave with ID '%v', instead got a non-nil error",
				enclaveId,
			)
		}
		apiContainerNamespaceName := apiContainerNamespaceAttributes.GetName().GetString()

		serviceList, err := backend.kubernetesManager.GetServicesByLabels(ctx, apiContainerNamespaceName, apiContainersMatchLabels)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting api container services using labels: %+v in namespace '%v'", apiContainersMatchLabels, apiContainerNamespaceName)
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

			matchingApiContainers[apiContainerNamespaceName][service.Name] = engineObj
		}
	}



	return matchingApiContainers, nil
}

func getApiContainerObjectFromKubernetesService(service apiv1.Service) (*api_container.APIContainer, error) {
	enclaveId, found := service.Labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to be able to find label describing the enclave ID on service '%v' with label key '%v', but was unable to", service.Name, label_key_consts.EnclaveIDLabelKey.GetString())
	}

	engineStatus := getKurtosisStatusFromKubernetesService(service)
	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	var publicGrpcProxyPortSpec *port_spec.PortSpec
	if engineStatus == container_status.ContainerStatus_Running {
		publicIpAddr = net.ParseIP(service.Spec.ClusterIP)
		if publicIpAddr == nil {
			return nil, stacktrace.NewError("Expected to be able to get the cluster ip of the engine service, instead parsing the cluster ip of service '%v' returned nil", service.Name)
		}
		var portSpecError error
		publicGrpcPortSpec, publicGrpcProxyPortSpec, portSpecError = getEngineGrpcPortSpecsFromServicePorts(service.Spec.Ports)
		if portSpecError != nil {
			return nil, stacktrace.Propagate(portSpecError, "Expected to be able to determine engine grpc port specs from kubernetes service ports for engine '%v', instead a non-nil error was returned", engineId)
		}
	}

	return engine.NewEngine(engineId, engineStatus, publicIpAddr, publicGrpcPortSpec, publicGrpcProxyPortSpec), nil

}
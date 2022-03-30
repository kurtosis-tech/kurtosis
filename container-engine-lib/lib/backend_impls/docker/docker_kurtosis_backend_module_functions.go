package docker

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

const (

)

func (backendCore *DockerKurtosisBackend) CreateModule(
	ctx context.Context,
	id module.ModuleID,
	guid module.ModuleGUID,
	containerImageName string,
	serializedParams string,
)(
	newModule *module.Module,
	resultErr error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) GetModules(
	ctx context.Context,
	filters *module.ModuleFilters,
) (
	map[string]*module.Module,
	error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) DestroyModules(
	ctx context.Context,
	filters *module.ModuleFilters,
) (
	successfulModuleIds map[module.ModuleGUID]bool,
	erroredModuleIds map[module.ModuleGUID]error,
	resultErr error,
) {
	panic("Implement me")
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets engines matching the search filters, indexed by their container ID
func (backendCore *DockerKurtosisBackend) getMatchingModules(ctx context.Context, filters *module.ModuleFilters) (map[string]*module.Module, error) {
	moduleContainerSearchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():         label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.ModuleContainerTypeLabelValue.GetString(),
	}
	matchingModuleContainers, err := backendCore.dockerManager.GetContainersByLabels(ctx, moduleContainerSearchLabels, shouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching module containers using labels: %+v", moduleContainerSearchLabels)
	}

	matchingModuleObjects := map[string]*module.Module{}
	for _, moduleContainer := range matchingModuleContainers {
		containerId := moduleContainer.GetId()
		moduleObj, err := getModuleObjectFromContainerInfo(
			containerId,
			moduleContainer.GetLabels(),
			moduleContainer.GetStatus(),
			moduleContainer.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with ID '%v' into a module object", moduleContainer.GetId())
		}

		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[moduleObj.GetEnclaveID()]; !found {
				continue
			}
		}

		// If the ID filter is specified, drop engines not matching it
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[moduleObj.GetGUID()]; !found {
				continue
			}
		}

		// If status filter is specified, drop engines not matching it
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[moduleObj.GetStatus()]; !found {
				continue
			}
		}

		matchingModuleObjects[containerId] = moduleObj
	}

	return matchingModuleObjects, nil
}

func getModuleObjectFromContainerInfo(
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (*module.Module, error) {
	enclaveId, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the module's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDLabelKey.GetString())
	}

	id, found := labels[label_key_consts.IDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find module ID label key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
	}

	guid, found := labels[label_key_consts.GUIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find module GUID label key '%v' but none was found", label_key_consts.GUIDLabelKey.GetString())
	}

	privateIpAddrStr, found := labels[label_key_consts.PrivateIPLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find module private IP label key '%v' but none was found", label_key_consts.PrivateIPLabelKey.GetString())
	}
	privateIpAddr := net.ParseIP(privateIpAddrStr)
	if privateIpAddr == nil {
		return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
	}

	privateGrpcPortSpec, err := getPrivateModulePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the module container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := isContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for module container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var moduleStatus container_status.ContainerStatus
	if isContainerRunning {
		moduleStatus = container_status.ContainerStatus_Running
	} else {
		moduleStatus = container_status.ContainerStatus_Stopped
	}

	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	if moduleStatus == container_status.ContainerStatus_Running {
		publicGrpcPortIpAddr, candidatePublicGrpcPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The module is running, but an error occurred getting the public port spec for the module's grpc private port spec")
		}
		publicGrpcPortSpec = candidatePublicGrpcPortSpec
		publicIpAddr = publicGrpcPortIpAddr
	}

	result := module.NewModule(
		enclave.EnclaveID(enclaveId),
		module.ModuleID(id),
		module.ModuleGUID(guid),
		moduleStatus,
		privateIpAddr,
		privateGrpcPortSpec,
		publicIpAddr,
		publicGrpcPortSpec,
	)

	return result, nil
}

func getPrivateModulePorts(containerLabels map[string]string) (
	resultGrpcPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing port specs string '%v'", serializedPortSpecs)
	}

	grpcPortSpec, foundGrpcPort := portSpecs[kurtosisInternalContainerGrpcPortId]
	if !foundGrpcPort {
		return nil, stacktrace.NewError("No grpc port with ID '%v' found in the port specs", kurtosisInternalContainerGrpcPortId)
	}

	return grpcPortSpec, nil
}

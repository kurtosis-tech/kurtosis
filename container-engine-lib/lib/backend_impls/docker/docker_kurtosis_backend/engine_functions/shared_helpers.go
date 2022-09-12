package engine_functions

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"strconv"
	"strings"
)

const (
	shouldShowStoppedLogsDatabaseContainers = true
)

// Gets engines matching the search filters, indexed by their container ID
func getMatchingEngines(ctx context.Context, filters *engine.EngineFilters, dockerManager *docker_manager.DockerManager) (map[string]*engine.Engine, error) {
	engineContainerSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.EngineContainerTypeDockerLabelValue.GetString(),
		// NOTE: we do NOT use the engine GUID label here, and instead do postfiltering, because Docker has no way to do disjunctive search!
	}
	allEngineContainers, err := dockerManager.GetContainersByLabels(ctx, engineContainerSearchLabels, consts.ShouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching engine containers using labels: %+v", engineContainerSearchLabels)
	}

	allMatchingEnginesByContainerId := map[string]*engine.Engine{}
	for _, engineContainer := range allEngineContainers {
		containerId := engineContainer.GetId()
		engineObj, err := getEngineObjectFromContainerInfo(
			containerId,
			engineContainer.GetLabels(),
			engineContainer.GetStatus(),
			engineContainer.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with GUID '%v' into an engine object", engineContainer.GetId())
		}

		// If the GUID filter is specified, drop engines not matching it
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[engineObj.GetGUID()]; !found {
				continue
			}
		}

		// If status filter is specified, drop engines not matching it
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[engineObj.GetStatus()]; !found {
				continue
			}
		}

		allMatchingEnginesByContainerId[containerId] = engineObj
	}

	return allMatchingEnginesByContainerId, nil
}

func getEngineObjectFromContainerInfo(
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (*engine.Engine, error) {
	engineGuidStr, found := labels[label_key_consts.GUIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError(
			"Expected a '%v' label on engine container with GUID '%v', but none was found",
			label_key_consts.GUIDDockerLabelKey.GetString(),
			containerId,
		)
	}
	engineGuid := engine.EngineGUID(engineGuidStr)

	privateGrpcPortSpec, privateGrpcProxyPortSpec, err := getEnginePrivatePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for engine container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var engineStatus container_status.ContainerStatus
	if isContainerRunning {
		engineStatus = container_status.ContainerStatus_Running
	} else {
		engineStatus = container_status.ContainerStatus_Stopped
	}

	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	var publicGrpcProxyPortSpec *port_spec.PortSpec
	if engineStatus == container_status.ContainerStatus_Running {
		publicGrpcPortIpAddr, candidatePublicGrpcPortSpec, err := shared_helpers.GetPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The engine is running, but an error occurred getting the public port spec for the engine's grpc private port spec")
		}
		publicGrpcPortSpec = candidatePublicGrpcPortSpec

		// TODO REMOVE THIS CONDITIONAL AFTER 2022-05-03 WHEN WE'RE CONFIDENT NOBODY IS USING ENGINES WITHOUT A PROXY
		if privateGrpcProxyPortSpec != nil {
			publicGrpcProxyPortIpAddr, candidatePublicGrpcProxyPortSpec, err := shared_helpers.GetPublicPortBindingFromPrivatePortSpec(privateGrpcProxyPortSpec, allHostMachinePortBindings)
			if err != nil {
				return nil, stacktrace.Propagate(err, "The engine is running, but an error occurred getting the public port spec for the engine's grpc private port spec")
			}
			publicGrpcProxyPortSpec = candidatePublicGrpcProxyPortSpec

			if publicGrpcPortIpAddr.String() != publicGrpcProxyPortIpAddr.String() {
				return nil, stacktrace.NewError(
					"Expected the engine's grpc port public IP address '%v' and grpc-proxy port public IP address '%v' to be the same, but they were different",
					publicGrpcPortIpAddr.String(),
					publicGrpcProxyPortIpAddr.String(),
				)
			}
		}
		publicIpAddr = publicGrpcPortIpAddr
	}

	result := engine.NewEngine(
		engineGuid,
		engineStatus,
		publicIpAddr,
		publicGrpcPortSpec,
		publicGrpcProxyPortSpec,
	)

	return result, nil
}

func getEnginePrivatePorts(containerLabels map[string]string) (
	resultGrpcPortSpec *port_spec.PortSpec,
	resultGrpcProxyPortSpec *port_spec.PortSpec,
	resultErr error,
) {

	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsDockerLabelKey.GetString())
	}

	portSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		// TODO AFTER 2022-05-02 SWITCH THIS TO A PLAIN ERROR WHEN WE'RE SURE NOBODY WILL BE USING THE OLD PORT SPEC STRING!
		oldPortSpecs, err := deserialize_pre_2022_03_02_PortSpecs(serializedPortSpecs)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v' even when trying the old method", serializedPortSpecs)
		}
		portSpecs = oldPortSpecs
	}

	grpcPortSpec, foundGrpcPort := portSpecs[consts.KurtosisInternalContainerGrpcPortId]
	if !foundGrpcPort {
		return nil, nil, stacktrace.NewError("No engine grpc port with GUID '%v' found in the engine server port specs", consts.KurtosisInternalContainerGrpcPortId)
	}

	grpcProxyPortSpec, foundGrpcProxyPort := portSpecs[consts.KurtosisInternalContainerGrpcProxyPortId]
	if !foundGrpcProxyPort {
		grpcProxyPortSpec = nil
		// TODO AFTER 2022-05-02 SWITCH THIS TO AN ERROR WHEN WE'RE SURE NOBODY WILL HAVE AN ENGINE WITHOUT THE PROXY
		// return nil, nil, stacktrace.NewError("No engine grpc-proxy port with GUID '%v' found in the engine server port specs", kurtosisInternalContainerGrpcProxyPortId)
	}

	return grpcPortSpec, grpcProxyPortSpec, nil
}

// TODO DELETE THIS AFTER 2022-05-02, WHEN WE'RE CONFIDENT NO ENGINES WILL BE USING THE OLD PORT SPEC!
func deserialize_pre_2022_03_02_PortSpecs(specsStr string) (map[string]*port_spec.PortSpec, error) {
	const (
		portIdAndInfoSeparator      = "."
		portNumAndProtocolSeparator = "-"
		portSpecsSeparator          = "_"

		expectedNumPortIdAndSpecFragments      = 2
		expectedNumPortNumAndProtocolFragments = 2
		portUintBase                           = 10
		portUintBits                           = 16
	)

	result := map[string]*port_spec.PortSpec{}
	portIdAndSpecStrs := strings.Split(specsStr, portSpecsSeparator)
	for _, portIdAndSpecStr := range portIdAndSpecStrs {
		portIdAndSpecFragments := strings.Split(portIdAndSpecStr, portIdAndInfoSeparator)
		numPortIdAndSpecFragments := len(portIdAndSpecFragments)
		if numPortIdAndSpecFragments != expectedNumPortIdAndSpecFragments {
			return nil, stacktrace.NewError(
				"Expected splitting port GUID & spec string '%v' to yield %v fragments but got %v",
				portIdAndSpecStr,
				expectedNumPortIdAndSpecFragments,
				numPortIdAndSpecFragments,
			)
		}
		portId := portIdAndSpecFragments[0]
		portSpecStr := portIdAndSpecFragments[1]

		portNumAndProtocolFragments := strings.Split(portSpecStr, portNumAndProtocolSeparator)
		numPortNumAndProtocolFragments := len(portNumAndProtocolFragments)
		if numPortNumAndProtocolFragments != expectedNumPortNumAndProtocolFragments {
			return nil, stacktrace.NewError(
				"Expected splitting port num & protocol string '%v' to yield %v fragments but got %v",
				portSpecStr,
				expectedNumPortIdAndSpecFragments,
				numPortIdAndSpecFragments,
			)
		}
		portNumStr := portNumAndProtocolFragments[0]
		portProtocolStr := portNumAndProtocolFragments[1]

		portNumUint64, err := strconv.ParseUint(portNumStr, portUintBase, portUintBits)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred parsing port num string '%v' to uint with base %v and %v bits",
				portNumStr,
				portUintBase,
				portUintBits,
			)
		}
		portNumUint16 := uint16(portNumUint64)
		portProtocol, err := port_spec.PortProtocolString(portProtocolStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting port protocol string '%v' to a port protocol enum", portProtocolStr)
		}

		portSpec, err := port_spec.NewPortSpec(portNumUint16, portProtocol)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating port spec object from GUID & spec string '%v'",
				portIdAndSpecStr,
			)
		}

		result[portId] = portSpec
	}
	return result, nil
}

func extractEngineGuidFromUncastedEngineObj(uncastedEngineObj interface{}) (string, error) {
	castedObj, ok := uncastedEngineObj.(*engine.Engine)
	if !ok {
		return "", stacktrace.NewError("An error occurred downcasting the engine object")
	}
	return string(castedObj.GetGUID()), nil
}

func getLogsDatabaseContainer(ctx context.Context, dockerManager *docker_manager.DockerManager) (*types.Container, error) {

	logsDatabaseContainerSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.LogsDatabaseTypeDockerLabelValue.GetString(),
	}

	matchingLogsDatabaseContainers, err := dockerManager.GetContainersByLabels(ctx, logsDatabaseContainerSearchLabels, shouldShowStoppedLogsDatabaseContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching logs database containers using labels: %+v", logsDatabaseContainerSearchLabels)
	}
	if len(matchingLogsDatabaseContainers) == 0 {
		return nil, stacktrace.NewError("Didn't find any logs database Docker container matching labels '%+v'; this is a bug in Kurtosis", logsDatabaseContainerSearchLabels)
	}
	if len(matchingLogsDatabaseContainers) > 1 {
		return nil, stacktrace.NewError("Found more than one logs database Docker container matching labels '%+v'; this is a bug in Kurtosis", logsDatabaseContainerSearchLabels)
	}

	logsDatabaseContainer := matchingLogsDatabaseContainers[0]

	return logsDatabaseContainer, nil
}

func removeLogsComponentsGracefully(ctx context.Context, dockerManager *docker_manager.DockerManager) error {

	logsCollectorContainer, err := shared_helpers.GetLogsCollectorContainer(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs collector container")
	}

	if err := dockerManager.StopContainer(ctx, logsCollectorContainer.GetId(), stopLogsComponentsContainersTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs collector container with ID '%v'", logsCollectorContainer.GetId())
	}

	if err := dockerManager.RemoveContainer(ctx, logsCollectorContainer.GetId()); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the logs collector container with ID '%v'", logsCollectorContainer.GetId())
	}

	logsDatabaseContainer, err := getLogsDatabaseContainer(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs database container")
	}

	if err := dockerManager.StopContainer(ctx, logsDatabaseContainer.GetId(), stopLogsComponentsContainersTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs database container with ID '%v'", logsDatabaseContainer.GetId())
	}

	if err := dockerManager.RemoveContainer(ctx, logsDatabaseContainer.GetId()); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the logs database container with ID '%v'", logsDatabaseContainer.GetId())
	}

	return nil
}
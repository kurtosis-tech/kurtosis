package engine_functions

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"net"
	"strconv"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const emptyApplicationProtocol = ""

// Gets engines matching the search filters, indexed by their container ID
func getMatchingEngines(ctx context.Context, filters *engine.EngineFilters, dockerManager *docker_manager.DockerManager) (map[string]*engine.Engine, error) {
	engineContainerSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.ContainerTypeDockerLabelKey.GetString(): label_value_consts.EngineContainerTypeDockerLabelValue.GetString(),
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
		if len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[engineObj.GetGUID()]; !found {
				continue
			}
		}

		// If status filter is specified, drop engines not matching it
		if len(filters.Statuses) > 0 {
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
	engineGuidStr, found := labels[docker_label_key.GUIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError(
			"Expected a '%v' label on engine container with GUID '%v', but none was found",
			docker_label_key.GUIDDockerLabelKey.GetString(),
			containerId,
		)
	}
	engineGuid := engine.EngineGUID(engineGuidStr)

	privateGrpcPortSpec, err := getEnginePrivatePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for engine container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var engineStatus container.ContainerStatus
	if isContainerRunning {
		engineStatus = container.ContainerStatus_Running
	} else {
		engineStatus = container.ContainerStatus_Stopped
	}

	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	if engineStatus == container.ContainerStatus_Running {
		publicGrpcPortIpAddr, candidatePublicGrpcPortSpec, err := shared_helpers.GetPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The engine is running, but an error occurred getting the public port spec for the engine's grpc private port spec")
		}
		publicGrpcPortSpec = candidatePublicGrpcPortSpec
		publicIpAddr = publicGrpcPortIpAddr
	}

	result := engine.NewEngine(
		engineGuid,
		engineStatus,
		publicIpAddr,
		publicGrpcPortSpec,
	)

	return result, nil
}

func getEnginePrivatePorts(containerLabels map[string]string) (
	resultGrpcPortSpec *port_spec.PortSpec,
	resultErr error,
) {

	serializedPortSpecs, found := containerLabels[docker_label_key.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", docker_label_key.PortSpecsDockerLabelKey.GetString())
	}

	portSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		// TODO AFTER 2022-05-02 SWITCH THIS TO A PLAIN ERROR WHEN WE'RE SURE NOBODY WILL BE USING THE OLD PORT SPEC STRING!
		oldPortSpecs, err := deserialize_pre_2022_03_02_PortSpecs(serializedPortSpecs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v' even when trying the old method", serializedPortSpecs)
		}
		portSpecs = oldPortSpecs
	}

	grpcPortSpec, foundGrpcPort := portSpecs[consts.KurtosisInternalContainerGrpcPortId]
	if !foundGrpcPort {
		return nil, stacktrace.NewError("No engine grpc port with GUID '%v' found in the engine server port specs", consts.KurtosisInternalContainerGrpcPortId)
	}

	return grpcPortSpec, nil
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
		portProtocol, err := port_spec.TransportProtocolString(portProtocolStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting port protocol string '%v' to a port protocol enum", portProtocolStr)
		}

		//TODO we are passing nil wait so far because the wait's serialization/deserialization logic is not added yet because
		//TODO the port wait feature is in the design stage and its name and fields could change,
		//TODO we will include this in a next PR
		portSpec, err := port_spec.NewPortSpec(portNumUint16, portProtocol, emptyApplicationProtocol, nil, consts.EmptyApplicationURL)
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

func extractEngineGuidFromEngine(engine *engine.Engine) string {
	return string(engine.GetGUID())
}

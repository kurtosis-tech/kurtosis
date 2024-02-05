package consts

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
)

const (
	// The Docker API's default is to return just containers whose status is "running"
	// However, we'd rather do our own filtering on what "running" means (because, e.g., "restarting"
	// should also be considered as running)
	ShouldFetchAllContainersWhenRetrievingContainers = true

	// The ID of the GRPC port for Kurtosis-internal containers (e.g. API container, engine, etc.) which will
	//  be stored in the port spec label
	KurtosisInternalContainerGrpcPortId = "grpc"

	// The ID of the REST API port.
	KurtosisInternalContainerRESTAPIPortId = "rest-api"

	// The engine server uses gRPC so MUST listen on TCP (no other protocols are supported)
	EngineTransportProtocol = port_spec.TransportProtocol_TCP

	// This needs to be bind-mounted into the engine & API containers so they can manipulate Docker
	DockerSocketFilepath = "/var/run/podman/podman.sock"

	// The host engine config directory to mount and its local mapping
	HostEngineConfigDirToMount = "/root/engine_config"
	EngineConfigLocalDir       = "/run/engine"

	// The Docker network name where all the containers in the engine and logs service context will be added
	HttpApplicationProtocol                             = "http"
	NameOfNetworkToStartEngineAndLogServiceContainersIn = "podman"
)

// This maps a Docker container's status to a binary "is the container considered running?" determiner
// Its completeness is enforced via unit test
var IsContainerRunningDeterminer = map[types.ContainerStatus]bool{
	types.ContainerStatus_Paused:     false,
	types.ContainerStatus_Restarting: true,
	types.ContainerStatus_Running:    true,
	types.ContainerStatus_Removing:   false,
	types.ContainerStatus_Dead:       false,
	types.ContainerStatus_Created:    false,
	types.ContainerStatus_Exited:     false,
}

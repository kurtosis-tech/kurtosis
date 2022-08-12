package consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
)

const (
	// The Docker API's default is to return just containers whose status is "running"
	// However, we'd rather do our own filtering on what "running" means (because, e.g., "restarting"
	// should also be considered as running)
	ShouldFetchAllContainersWhenRetrievingContainers = true

	// The ID of the GRPC port for Kurtosis-internal containers (e.g. API container, engine, modules, etc.) which will
	//  be stored in the port spec label
	KurtosisInternalContainerGrpcPortId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	// need a proxy  that will translate grpc-web requests before they hit the main GRPC server
	KurtosisInternalContainerGrpcProxyPortId = "grpcProxy"

	// The engine server uses gRPC so MUST listen on TCP (no other protocols are supported), which also
	// means that its grpc-proxy must listen on TCP
	EnginePortProtocol = port_spec.PortProtocol_TCP

	// This needs to be bind-mounted into the engine & API containers so they can manipulate Docker
	DockerSocketFilepath = "/var/run/docker.sock"
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


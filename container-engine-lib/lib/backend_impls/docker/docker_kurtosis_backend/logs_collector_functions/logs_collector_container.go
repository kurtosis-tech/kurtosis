package logs_collector_functions

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
)

type LogsCollectorContainer interface {
	CreateAndStart(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		logsDatabaseHost string,
		logsDatabasePort uint16,
		tcpPortNumber uint16,
		httpPortNumber uint16,
		logsCollectorTcpPortId string,
		logsCollectorHttpPortId string,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (string, map[string]string, map[nat.Port]*nat.PortBinding, func(), error)

	// GetHttpHealthCheckEndpoint returns endpoint for verifying the availability of the logs collector application on container
	GetHttpHealthCheckEndpoint() string
}

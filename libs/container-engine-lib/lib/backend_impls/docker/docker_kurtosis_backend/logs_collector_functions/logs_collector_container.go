package logs_collector_functions

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
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
	) (
		resultContainerId string,
		resultContainerLabels map[string]string,
		restulHostMachinePortBindings map[nat.Port]*nat.PortBinding,
		resultRemoveLogsCollectorContainerFunc func(),
		resultErr error,
	)
}

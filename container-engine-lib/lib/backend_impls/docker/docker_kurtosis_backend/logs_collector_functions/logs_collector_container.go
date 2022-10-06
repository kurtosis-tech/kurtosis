package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
)

type LogsCollectorContainer interface {
	CreateAndStart(
		ctx context.Context,
		logsDatabaseHost string,
		logsDatabasePort uint16,
		httpPortNumber uint16,
		logsCollectorTcpPortId string,
		logsCollectorHttpPortId string,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (
		resultContainerId  string,
		resultContainerLabels map[string]string,
		resultRemoveLogsCollectorContainerFunc func(),
		resultErr error,
	)
}

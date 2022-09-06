package logs_components

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
)

type LogsCollectorContainer interface {
	CreateAndStart(
		ctx context.Context,
		logsDatabaseHost string,
		logsDatabasePort uint16,
		httpPortNumber uint16,
		logsCollectorTcpPortId string,
		logsCollectorHttpPortId string,
		engineGuid engine.EngineGUID,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (
		resultKillLogsCollectorContainerFunc func(),
		resultErr error,
	)
}

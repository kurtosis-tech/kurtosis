package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
)

type LogsAggregatorContainer interface {
	CreateAndStart(
		ctx context.Context,
		// This is the port that this LogsAggregatorContainer will listen for logs on
		// LogsCollectors should forward logs to this port
		logsListeningPort uint16,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (string, map[string]string, func(), error)

	// GetLogsBaseDirPath returns the base directory path where all logs will be output on its container
	GetLogsBaseDirPath() string
}

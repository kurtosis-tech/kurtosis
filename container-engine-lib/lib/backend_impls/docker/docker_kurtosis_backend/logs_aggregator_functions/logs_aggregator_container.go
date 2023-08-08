package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
)

type LogsAggregatorContainer interface {
	CreateAndStart(
		ctx context.Context,
		// This is the port that this LogAggregatorContainer will listen for forward logs on
		// LogCollectors should forward logs to this port
		logListeningPort uint16,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (string, map[string]string, func(), error)
}

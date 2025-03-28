package logs_aggregator_functions

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
)

type LogsAggregatorContainer interface {
	CreateAndStart(
		ctx context.Context,
		// This is the port that this LogsAggregatorContainer will listen for logs on
		// LogsCollectors should forward logs to this port
		logsListeningPort uint16,
		sinks logs_aggregator.Sinks,
		httpPortNumber uint16,
		logsAggregatorHttpPortId string,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (string, map[string]string, func(), error)

	// GetLogsBaseDirPath returns the base directory path where all logs will be output on its container
	GetLogsBaseDirPath() string

	// GetHttpHealthCheckEndpoint returns endpoint for verifying the availability of the logs aggregator application on container
	GetHttpHealthCheckEndpoint() string

	// Clean removes any resources the logs aggregator creates for durability of logs in the case of crashes (e.g. disk buffers)
	Clean(
		ctx context.Context,
		logsAggregator *types.Container,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) error
}

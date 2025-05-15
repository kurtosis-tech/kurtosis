package logs_aggregator_functions

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
)

func GetLogsAggregator(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (*logs_aggregator.LogsAggregator, error) {

	maybeLogsAggregatorObject, _, err := getLogsAggregatorObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator")
	}

	return maybeLogsAggregatorObject, nil
}

package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

//If nothing is found returns nil
func GetLogsCollector(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (
	resultMaybeLogsCollector *logs_collector.LogsCollector,
	resultErr error,
){

	maybeLogsCollectorObject, _, err := getLogsCollectorObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	return maybeLogsCollectorObject, nil
}

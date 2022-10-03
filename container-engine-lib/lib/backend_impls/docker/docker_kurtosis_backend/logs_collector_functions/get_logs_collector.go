package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

func GetLogsCollector(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (*logs_collector.LogsCollector, error){

	logsCollectorObject, _, err := getLogsCollectorObjectAndContainerIdMatching(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	return logsCollectorObject, nil
}

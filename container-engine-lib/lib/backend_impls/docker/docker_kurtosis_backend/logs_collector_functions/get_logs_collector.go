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

	allLogsCollectorContainers, err := getAllLogsCollectorContainers(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting all logs collector containers")
	}
	if len(allLogsCollectorContainers) == 0 {
		return nil, nil
	}
	if len(allLogsCollectorContainers) > 1 {
		return nil, stacktrace.NewError("Found more than one logs collector Docker container'; this is a bug in Kurtosis")
	}

	logsCollectorContainer := allLogsCollectorContainers[0]

	logsCollectorObject, err := getLogsCollectorObjectFromContainerInfo(
		ctx,
		logsCollectorContainer.GetId(),
		logsCollectorContainer.GetLabels(),
		logsCollectorContainer.GetStatus(),
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs collector object using container ID '%v', labels '%+v' and the status '%v'", logsCollectorContainer.GetId(), logsCollectorContainer.GetLabels(), logsCollectorContainer.GetStatus())
	}

	return logsCollectorObject, nil
}

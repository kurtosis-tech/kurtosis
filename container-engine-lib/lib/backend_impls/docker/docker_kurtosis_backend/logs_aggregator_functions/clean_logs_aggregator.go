package logs_aggregator_functions

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func CleanLogsAggregator(
	ctx context.Context,
	logsAggregatorContainer LogsAggregatorContainer,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) error {
	logsAggregator, found, err := getLogsAggregatorContainer(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting logs aggregator container.")
	}
	if !found {
		logrus.Debugf("No logs aggregator container was found, skip cleaning.")
		return nil
	}

	if err := logsAggregatorContainer.Clean(ctx, logsAggregator, objAttrsProvider, dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred cleaning logs aggregator container")
	}

	return nil
}

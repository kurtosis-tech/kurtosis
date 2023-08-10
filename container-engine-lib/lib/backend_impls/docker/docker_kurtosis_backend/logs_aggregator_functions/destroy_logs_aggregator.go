package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	stopLogsAggregatorContainerTimeout = 1 * time.Minute
)

// Returns nil if logs aggregator container is successfully destroyed or no logs aggregator container was found
func DestroyLogsAggregator(ctx context.Context, dockerManager *docker_manager.DockerManager) error {
	_, maybeLogsAggregatorContainerId, err := getLogsAggregatorObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		logrus.Warnf("Attempted to destroy logs aggregator but no logs aggregator container was found.")
		return nil
	}

	if maybeLogsAggregatorContainerId == "" {
		return nil
	}

	if err := dockerManager.StopContainer(ctx, maybeLogsAggregatorContainerId, stopLogsAggregatorContainerTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs aggregator container with ID '%v'", maybeLogsAggregatorContainerId)
	}

	if err := dockerManager.RemoveContainer(ctx, maybeLogsAggregatorContainerId); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the logs aggregator container with ID '%v'", maybeLogsAggregatorContainerId)
	}

	return nil
}

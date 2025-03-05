package logs_aggregator_functions

import (
	"context"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	stopLogsAggregatorContainerTimeout = 2 * time.Second
)

// Destroys logs aggregator idempotently, returns nil if no logs aggregator logs aggregator container was found
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

	maybeLogsCollectorVolumeName, err := getLogsAggregatorVolumeName(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting logs aggregator volume")
	}

	if maybeLogsCollectorVolumeName == "" {
		return nil
	}

	err = dockerManager.RemoveVolume(ctx, maybeLogsCollectorVolumeName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing the logs aggregator volume '%v'", maybeLogsCollectorVolumeName)
	}

	return nil
}

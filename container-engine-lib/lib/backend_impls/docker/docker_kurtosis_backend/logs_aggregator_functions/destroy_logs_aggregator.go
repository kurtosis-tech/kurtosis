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
func DestroyLogsAggregator(ctx context.Context, dockerManager *docker_manager.DockerManager, usePodmanBridgeNetwork bool) error {
	_, maybeLogsAggregatorContainerId, err := getLogsAggregatorObjectAndContainerId(ctx, dockerManager, usePodmanBridgeNetwork)
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

	// We only destroy logs aggregator config volume and not data volume because we may want to restart the logs aggregator
	// and use the data from previous execution

	maybeLogsAggregatorConfigVolumeName, err := getLogsAggregatorConfigVolumeName(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting logs aggregator config volume")
	}

	if maybeLogsAggregatorConfigVolumeName == "" {
		return nil
	}

	err = dockerManager.RemoveVolume(ctx, maybeLogsAggregatorConfigVolumeName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing the logs aggregator config volume '%v'", maybeLogsAggregatorConfigVolumeName)
	}

	return nil
}

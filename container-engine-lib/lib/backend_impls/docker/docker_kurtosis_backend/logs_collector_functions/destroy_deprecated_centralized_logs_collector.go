package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	deprecatedLogsCollectorVolumeName    = "kurtosis-logs-collector-vol"
	deprecatedLogsCollectorContainerName = "kurtosis-logs-collector"
)

// TODO(centralized-logs-collector-deprecation) remove this entire function after enough people are on > 0.66.0
func DestroyDeprecatedCentralizedLogsCollector(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) error {

	maybeLogsCollectorContainerId, err := getDeprecatedLogsCollectorContainerId(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	if maybeLogsCollectorContainerId != "" {
		if err := dockerManager.StopContainer(ctx, maybeLogsCollectorContainerId, stopLogsCollectorContainersTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the logs collector container with ID '%v'", maybeLogsCollectorContainerId)
		}

		if err := dockerManager.RemoveContainer(ctx, maybeLogsCollectorContainerId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing the logs collector container with ID '%v'", maybeLogsCollectorContainerId)
		}
	}

	volumes, err := dockerManager.GetVolumesByName(ctx, deprecatedLogsCollectorVolumeName)
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to fetch volumes to get the volume of the deprecated centralized logs collector but failed")
	}

	foundExactMatch := false
	for _, volumeName := range volumes {
		if volumeName == deprecatedLogsCollectorVolumeName {
			foundExactMatch = true
			break
		}
	}

	// we found some matching volumes; but they weren't the old volume
	if !foundExactMatch {
		return nil
	}

	if err := dockerManager.RemoveVolume(ctx, deprecatedLogsCollectorVolumeName); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the deprecated centralized logs collector volume with volume name '%v'", deprecatedLogsCollectorVolumeName)
	}

	return nil
}

func getDeprecatedLogsCollectorContainerId(ctx context.Context, dockerManager *docker_manager.DockerManager) (string, error) {
	deprecatedLogsCollectorContainerId, err := dockerManager.GetContainerIdByExactName(ctx, deprecatedLogsCollectorContainerName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting all logs collector containers")
	}

	return deprecatedLogsCollectorContainerId, nil
}

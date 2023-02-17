package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	deprecatedLogsCollectorVolumeName    = "kurtosis-logs-collector-vol"
	deprecatedLogsCollectorContainerName = "kurtosis-logs-collector"
)

// TODO(centralized-logs-collector-deprecation) remove this entire function after enough people are on > 0.66.0
func DestroyDeprecatedCentralizedLogsCollectors(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) error {

	logsCollectorContainerSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.LogsCollectorTypeDockerLabelValue.GetString(),
	}

	matchingLogsCollectorContainers, err := dockerManager.GetContainersByLabels(ctx, logsCollectorContainerSearchLabels, shouldShowStoppedLogsCollectorContainers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred fetching logs collector containers using labels: %+v", logsCollectorContainerSearchLabels)
	}

	for _, logsCollectorContainer := range matchingLogsCollectorContainers {
		if err := dockerManager.StopContainer(ctx, logsCollectorContainer.GetId(), stopLogsCollectorContainersTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the logs collector container with ID '%v'", maybeLogsCollectorContainerId)
		}

		if err := dockerManager.RemoveContainer(ctx, logsCollectorContainer.GetId()); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing the logs collector container with ID '%v'", maybeLogsCollectorContainerId)
		}
	}

	//This removes the old main volume
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

func destroyContainer()

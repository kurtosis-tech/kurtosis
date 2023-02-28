package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	deprecatedLogsCollectorVolumeNameSuffix = "kurtosis-logs-collector-vol"
)

// TODO(centralized-logs-collectors-deprecation) remove this entire function after enough people are on > 0.68.0
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
			return stacktrace.Propagate(err, "An error occurred stopping the logs collector container with ID '%v'", logsCollectorContainer.GetId())
		}

		if err := dockerManager.RemoveContainer(ctx, logsCollectorContainer.GetId()); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing the logs collector container with ID '%v'", logsCollectorContainer.GetId())
		}
	}

	volumes, err := dockerManager.GetVolumesByName(ctx, deprecatedLogsCollectorVolumeNameSuffix)
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to fetch volumes to get the volumes of the deprecated centralized logs collectors but failed")
	}

	for _, volumeName := range volumes {
		if err := dockerManager.RemoveVolume(ctx, volumeName); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing the deprecated centralized logs collector volume with volume name '%v'", volumeName)
		}
	}

	return nil
}

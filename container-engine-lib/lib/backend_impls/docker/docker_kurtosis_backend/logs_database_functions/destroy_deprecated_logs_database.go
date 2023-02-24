package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	deprecatedLogsDatabaseVolumeName = "kurtosis-logs-db-vol"
)

// TODO(centralized-logs-database-deprecation) remove this entire function after enough people are on > 0.68.0
func DestroyDeprecatedCentralizedLogsDatabase(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) error {

	//It stops and removes the container
	if err := DestroyLogsDatabase(ctx, dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred while destroying the deprecated centralized logs database")
	}

	volumes, err := dockerManager.GetVolumesByName(ctx, deprecatedLogsDatabaseVolumeName)
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to fetch volumes to get the volumes of the deprecated centralized logs database but failed")
	}

	for _, volumeName := range volumes {
		if err := dockerManager.RemoveVolume(ctx, volumeName); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing the deprecated centralized logs database volume with volume name '%v'", volumeName)
		}
	}

	return nil
}

package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
)

//If nothing is found returns nil
func GetLogsDatabase(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (
	resultMaybeLogsDatabase *logs_database.LogsDatabase,
	resultErr error,
){

	maybeLogsDatabaseObject, _, err := getLogsDatabaseObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database")
	}

	return maybeLogsDatabaseObject, nil
}

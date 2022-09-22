package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
)

func GetLogsDatabase(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (*logs_database.LogsDatabase, error){

	allLogsDatabaseContainers, err := getAllLogsDatabaseContainers(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting all logs database containers")
	}
	if len(allLogsDatabaseContainers) == 0 {
		return nil, nil
	}
	if len(allLogsDatabaseContainers) > 1 {
		return nil, stacktrace.NewError("Found more than one logs database Docker container'; this is a bug in Kurtosis")
	}

	logsDatabaseContainer := allLogsDatabaseContainers[0]

	logsDatabaseObject, err := getLogsDatabaseObjectFromContainerInfo(
		ctx,
		logsDatabaseContainer.GetId(),
		logsDatabaseContainer.GetLabels(),
		logsDatabaseContainer.GetStatus(),
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs database object using container ID '%v', labels '%+v' and the status '%v'", logsDatabaseContainer.GetId(), logsDatabaseContainer.GetLabels(), logsDatabaseContainer.GetStatus())
	}

	return logsDatabaseObject, nil
}

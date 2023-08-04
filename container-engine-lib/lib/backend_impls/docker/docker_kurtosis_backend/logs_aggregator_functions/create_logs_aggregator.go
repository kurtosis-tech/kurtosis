package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
)

func CreateLogsAggregator(
	ctx context.Context,
	portNumber uint16,
	logsAggregatorContainer LogsAggregatorContainer,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*logs_aggregator.LogsAggregator,
	error,
) {

	// check if aggregator already exists(not needed
	// if so don't start another one

	// get the network this aggregator should connect to (needed)

	// start the aggregator container (needed)

	// if something goes wrong defer an undo (not needed)

	// get the resulting object from container info (not needed but do it)

	// return the object
	return nil, nil
}

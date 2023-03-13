package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
)

type LogsDatabaseContainer interface {
	CreateAndStart(
		ctx context.Context,
		httpPortId string,
	//TODO now the httpPortNumber is configured from the client, because this will be published to the host machine until
	//TODO we productize logs search, tracked by this issue: https://github.com/kurtosis-tech/kurtosis/issues/340
	//TODO remove this parameter when we do not publish the port again
		httpPortNumber uint16,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (
		resultContainerId string,
		resultContainerLabels map[string]string,
		resultRemoveLogsDatabaseContainerFunc func(),
		resultErr error,
	)
}

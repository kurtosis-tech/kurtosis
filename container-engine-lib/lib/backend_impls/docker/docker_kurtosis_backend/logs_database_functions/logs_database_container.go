package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
)

type LogsDatabaseContainer interface {
	CreateAndStart(
		ctx context.Context,
		httpPortId string,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (
		resultContainerId  string,
		resultContainerLabels map[string]string,
		resultRemoveLogsDatabaseContainerFunc func(),
		resultErr error,
	)
}

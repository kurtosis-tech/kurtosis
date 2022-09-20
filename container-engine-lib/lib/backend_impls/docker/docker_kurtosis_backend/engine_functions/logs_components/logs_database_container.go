package logs_components

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
)

type LogsDatabaseContainer interface {
	CreateAndStart(
		ctx context.Context,
		logsDatabaseHttpPortId string,
		engineGuid engine.EngineGUID,
		targetNetworkId string,
		targetNetworkName string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (
		resultLogsDatabasePrivateHost string,
		resultLogsDatabasePrivatePort uint16,
		resultKillLogsDatabaseContainerFunc func(),
		resultErr error,
	)
}

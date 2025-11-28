package reverse_proxy_functions

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
)

type ReverseProxyContainer interface {
	CreateAndStart(
		ctx context.Context,
		engineGuid engine.EngineGUID,
		httpPort uint16,
		dashboardPort uint16,
		targetNetworkId string,
		objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
		dockerManager *docker_manager.DockerManager,
	) (string, map[string]string, func(), error)
}

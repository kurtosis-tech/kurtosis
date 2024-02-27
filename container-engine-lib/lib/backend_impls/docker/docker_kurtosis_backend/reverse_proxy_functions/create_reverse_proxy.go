package reverse_proxy_functions

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultContainerStatusForNewReverseProxyContainer = types.ContainerStatus_Running
)

// Create reverse proxy idempotently, if existing reverse proxy is found, then it is returned
func CreateReverseProxy(
	ctx context.Context,
	engineGuid engine.EngineGUID,
	reverseProxyContainer ReverseProxyContainer,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*reverse_proxy.ReverseProxy,
	func(),
	error,
) {
	proxyDockerContainer, found, err := getReverseProxyContainer(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting reverse proxy container.")
	}
	if found {
		logrus.Debugf("Found existing reverse proxy; cannot start a new one.")
		reverseProxyObj, proxyContainerId, getProxyObjErr := getReverseProxyObjectAndContainerId(ctx, dockerManager)
		if getProxyObjErr == nil {
			removeProxyFunc := func() {
				removeCtx := context.Background()
				if err := dockerManager.RemoveContainer(removeCtx, proxyContainerId); err != nil {
					logrus.Errorf("an error occurred to remove the current reverse proxy container with ID '%s' failed. Error was:\n%v", proxyContainerId, err)
					logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the reverse proxy server with Docker container ID '%v'!!!!!!", proxyContainerId)
				}
			}
			return reverseProxyObj, removeProxyFunc, nil
		}
		logrus.Debugf("Something failed while trying to create the reverse proxy object using container with ID '%s'. Error was:\n%s", proxyDockerContainer.GetId(), getProxyObjErr.Error())
		logrus.Debugf("Destroying the failing reverse proxy to create a new one...")
		if destroyProxyContainerErr := destroyReverseProxyWithContainerId(ctx, dockerManager, proxyDockerContainer.GetId()); destroyProxyContainerErr != nil {
			return nil, nil, stacktrace.Propagate(destroyProxyContainerErr, "an error occurred destroying the current reverse proxy that was failing to create a new one")
		}
		logrus.Debugf("... current reverse proxy successfully destroyed, starting a new one now.")
	}

	reverseProxyNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the reverse proxy network.")
	}
	targetNetworkId := reverseProxyNetwork.GetId()

	containerId, containerLabels, removeReverseProxyContainerFunc, err := reverseProxyContainer.CreateAndStart(
		ctx,
		engineGuid,
		defaultReverseProxyHttpPortNum,
		defaultReverseProxyDashboardPortNum,
		targetNetworkId,
		objAttrsProvider,
		dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the reverse proxy container in Docker network with ID '%v'",
			targetNetworkId,
		)
	}
	shouldRemoveReverseProxyContainer := true
	defer func() {
		if shouldRemoveReverseProxyContainer {
			removeReverseProxyContainerFunc()
		}
	}()

	reverseProxy, err := getReverseProxyObjectFromContainerInfo(
		ctx,
		containerId,
		defaultContainerStatusForNewReverseProxyContainer,
		dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting reverse proxy object using container ID '%v', labels '%+v', status '%v'.", containerId, containerLabels, defaultContainerStatusForNewReverseProxyContainer)
	}

	removeReverseProxyFunc := func() {
		removeReverseProxyContainerFunc()
	}

	shouldRemoveReverseProxyContainer = false
	return reverseProxy, removeReverseProxyFunc, nil
}

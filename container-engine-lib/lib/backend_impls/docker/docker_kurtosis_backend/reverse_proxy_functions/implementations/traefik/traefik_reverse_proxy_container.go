package traefik

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type traefikReverseProxyContainer struct{}

func NewTraefikReverseProxyContainer() *traefikReverseProxyContainer {
	return &traefikReverseProxyContainer{}
}

func (traefikContainer *traefikReverseProxyContainer) CreateAndStart(
	ctx context.Context,
	engineGuid engine.EngineGUID,
	httpPort uint16,
	dashboardPort uint16,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (string, map[string]string, func(), error) {
	traefikContainerConfigProviderObj := createTraefikContainerConfigProvider(httpPort, dashboardPort, targetNetworkId)

	reverseProxyAttrs, err := objAttrsProvider.ForReverseProxy(engineGuid)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the reverse proxy container attributes.")
	}
	containerName := reverseProxyAttrs.GetName().GetString()
	containerLabelStrs := map[string]string{}
	for labelKey, labelValue := range reverseProxyAttrs.GetLabels() {
		containerLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	createAndStartArgs, err := traefikContainerConfigProviderObj.GetContainerArgs(containerName, containerLabelStrs, httpPort, dashboardPort, targetNetworkId)
	if err != nil {
		return "", nil, nil, err
	}

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the reverse proxy container with these args '%+v'", createAndStartArgs)
	}
	removeContainerFunc := func() {
		removeCtx := context.Background()

		if err := dockerManager.RemoveContainer(removeCtx, containerId); err != nil {
			logrus.Errorf(
				"Launching the reverse proxy server with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the reverse proxy server with Docker container ID '%v'!!!!!!", containerId)
		}
	}
	shouldRemoveReverseProxyContainer := true
	defer func() {
		if shouldRemoveReverseProxyContainer {
			removeContainerFunc()
		}
	}()

	shouldRemoveReverseProxyContainer = false
	return containerId, containerLabelStrs, removeContainerFunc, nil
}

package reverse_proxy_functions

import (
	"context"
	"net"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	shouldShowStoppedReverseProxyContainers = true
)

func getReverseProxyObjectAndContainerId(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	usePodmanBridgeNetwork bool,
) (*reverse_proxy.ReverseProxy, string, error) {
	reverseProxyContainer, found, err := getReverseProxyContainer(ctx, dockerManager)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting the reverse proxy container")
	}
	if !found {
		return nil, "", nil
	}

	reverseProxyContainerID := reverseProxyContainer.GetId()

	reverseProxyObject, err := getReverseProxyObjectFromContainerInfo(
		ctx,
		reverseProxyContainerID,
		reverseProxyContainer.GetStatus(),
		dockerManager,
		usePodmanBridgeNetwork,
	)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting the reverse proxy object using container ID '%v', labels '%+v' and the status '%v'", reverseProxyContainer.GetId(), reverseProxyContainer.GetLabels(), reverseProxyContainer.GetStatus())
	}

	return reverseProxyObject, reverseProxyContainerID, nil
}

func getReverseProxyContainer(ctx context.Context, dockerManager *docker_manager.DockerManager) (*types.Container, bool, error) {
	reverseProxyContainerSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.ContainerTypeDockerLabelKey.GetString(): label_value_consts.ReverseProxyTypeDockerLabelValue.GetString(),
	}

	matchingReverseProxyContainers, err := dockerManager.GetContainersByLabels(ctx, reverseProxyContainerSearchLabels, shouldShowStoppedReverseProxyContainers)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "An error occurred fetching the reverse proxy container using labels: %+v", reverseProxyContainerSearchLabels)
	}

	if len(matchingReverseProxyContainers) == 0 {
		return nil, false, nil
	}
	if len(matchingReverseProxyContainers) > 1 {
		return nil, false, stacktrace.NewError("Found more than one reverse proxy Docker container'; this is a bug in Kurtosis")
	}
	return matchingReverseProxyContainers[0], true, nil
}

func getReverseProxyObjectFromContainerInfo(
	ctx context.Context,
	containerId string,
	containerStatus types.ContainerStatus,
	dockerManager *docker_manager.DockerManager,
	usePodmanBridgeNetwork bool,
) (*reverse_proxy.ReverseProxy, error) {
	var privateIpAddr net.IP
	var enclaveNetworksIpAddress map[string]net.IP

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for reverse proxy container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}

	var reverseProxyStatus container.ContainerStatus
	if isContainerRunning {
		reverseProxyStatus = container.ContainerStatus_Running

		bridgeNetworkName := shared_helpers.GetBridgeNetworkName(usePodmanBridgeNetwork)
		privateIpAddrStr, err := dockerManager.GetContainerIP(ctx, bridgeNetworkName, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the private IP address of container '%v' in network '%v'", containerId, bridgeNetworkName)
		}
		privateIpAddr = net.ParseIP(privateIpAddrStr)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
		}

		networksIpAddressStr, err := dockerManager.GetContainerIps(ctx, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the networks private IP address of container '%v'", containerId)
		}
		enclaveNetworksIpAddress = map[string]net.IP{}
		for networkId, networkIpAddressStr := range networksIpAddressStr {
			if networkIpAddressStr != privateIpAddrStr {
				networkIpAddress := net.ParseIP(networkIpAddressStr)
				if networkIpAddress == nil {
					return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", networkIpAddressStr)
				}
				enclaveNetworksIpAddress[networkId] = networkIpAddress
			}
		}
		logrus.Debugf("Enclave networks: '%v'", enclaveNetworksIpAddress)
	} else {
		reverseProxyStatus = container.ContainerStatus_Stopped
	}

	reverseProxyObj := reverse_proxy.NewReverseProxy(
		reverseProxyStatus,
		privateIpAddr,
		enclaveNetworksIpAddress,
		defaultReverseProxyHttpPortNum,
		defaultReverseProxyDashboardPortNum,
	)

	return reverseProxyObj, nil
}

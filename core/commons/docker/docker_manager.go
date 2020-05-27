package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

// TODO TODO TODO - do we ever need to handle different local host IPs?
const LOCAL_HOST_IP = "0.0.0.0"


type DockerManager struct {
	dockerCtx           context.Context
	dockerClient        *client.Client
	freeHostPortTracker *FreeHostPortTracker
	subnetMask string
}

func NewDockerManager(
	dockerCtx context.Context,
	dockerClient *client.Client,
	subnetMask string,
	hostPortRangeStart int,
	hostPortRangeEnd int) (dockerManager *DockerManager, err error) {

	freeHostPortTracker, err := NewFreeHostPortTracker(hostPortRangeStart, hostPortRangeEnd)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return &DockerManager{
		dockerCtx:           dockerCtx,
		dockerClient:        dockerClient,
		freeHostPortTracker: freeHostPortTracker,
		subnetMask: subnetMask,
	}, nil
}

func (manager DockerManager) CreateAndStartContainerForService(
	// TODO This arg is a hack that will go away as soon as Gecko removes the --public-ip command!
	dockerImage string,
	staticIp string,
	dockerNetwork string,
	usedPorts map[int]bool,
	startCmdArgs []string) (containerIpAddr string, containerId string, err error) {

	imageExistsLocally, err := manager.isImageAvailableLocally(dockerImage)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to check for image availability.")
	}

	if !imageExistsLocally {
		err = manager.pullImage(dockerImage)
		if err != nil {
			return "", "", stacktrace.Propagate(err, "Failed to pull Docker image.")
		}
	}

	networkExistsLocally, err := manager.isNetworkAvailableLocally(dockerNetwork)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to check for network availability.")
	}

	if !networkExistsLocally {
		_, err := manager.createNetwork(dockerNetwork, manager.subnetMask)
		if err != nil {
			return "", "", stacktrace.Propagate(err, "Failed to create Docker network.")
		}
	}

	// TODO this relies on serviceId being incremental, and is a total hack until --public-ips flag is gone from Gecko!
	containerConfigPtr, err := manager.getContainerCfgFromServiceCfg(dockerImage, usedPorts, startCmdArgs)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to configure container from service.")
	}
	containerHostConfigPtr, err := manager.getContainerHostConfig(dockerNetwork, usedPorts)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to configure host to container mappings from service.")
	}
	// TODO probably use a UUID for the network name (and maybe include test name too)
	resp, err := manager.dockerClient.ContainerCreate(manager.dockerCtx, containerConfigPtr, containerHostConfigPtr, nil, "")
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Could not create Docker container from image %v.", dockerImage)
	}
	containerId = resp.ID
	if err := manager.dockerClient.ContainerStart(manager.dockerCtx, containerId, types.ContainerStartOptions{}); err != nil {
		return "", "", stacktrace.Propagate(err, "Could not start Docker container from image %v.", dockerImage)
	}
	err = manager.connectToNetwork(dockerNetwork, containerId, staticIp)
	if err != nil {
		return "","", stacktrace.Propagate(err, "Failed to connect container %s to network.", containerId)
	}
	return staticIp, containerId, nil
}

func (manager DockerManager) createNetwork(name string, subnetMask string) (id string, err error)  {
	ipamConfig := []network.IPAMConfig{{Subnet: subnetMask}}
	resp, err := manager.dockerClient.NetworkCreate(manager.dockerCtx, name, types.NetworkCreate{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Config: ipamConfig,
		},
	})
	if err != nil {
		return "", stacktrace.Propagate( err, "")
	}
	return resp.ID, nil
}

func (manager DockerManager) getFreePort() (freePort *nat.Port, err error) {
	freePortInt, err := manager.freeHostPortTracker.GetFreePort()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	port, err := nat.NewPort("tcp", strconv.Itoa(freePortInt))
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return &port, nil
}

func (manager DockerManager) getLocalHostIp() string {
	return LOCAL_HOST_IP
}

func (manager DockerManager) isImageAvailableLocally(imageName string) (isAvailable bool, err error) {
	referenceArg := filters.Arg("reference", imageName)
	filters := filters.NewArgs(referenceArg)
	images, err := manager.dockerClient.ImageList(
		context.Background(),
		types.ImageListOptions{
			All: true,
			Filters: filters,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to list images.")
	}
	return len(images) > 0, nil
}

func (manager DockerManager) isNetworkAvailableLocally(networkName string) (isAvailable bool, err error) {
	referenceArg := filters.Arg("name", networkName)
	filters := filters.NewArgs(referenceArg)
	networks, err := manager.dockerClient.NetworkList(
		context.Background(),
		types.NetworkListOptions{
			Filters: filters,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to list networks.")
	}
	return len(networks) > 0, nil
}

func (manager DockerManager) getNetworkId(networkName string) (networkId string, err error) {
	referenceArg := filters.Arg("name", networkName)
	filters := filters.NewArgs(referenceArg)
	networks, err := manager.dockerClient.NetworkList(
		context.Background(),
		types.NetworkListOptions{
			Filters: filters,
		})
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to list networks.")
	}
	if len(networks) == 0 {
		return "", stacktrace.NewError("Network does not exist.")
	}
	return networks[0].ID, nil
}

func (manager DockerManager) getIPAddr(networkName string, containerId string) (ipAddr string, err error) {
	networkId, err := manager.getNetworkId(networkName)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	networkJson, err := manager.dockerClient.NetworkInspect(
		context.Background(),
		networkId,
		types.NetworkInspectOptions{})
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to inspect container %s.", containerId)
	}
	//log.Printf("NetworkJson: %+v", networkJson)
	ipAddrCIDR := networkJson.Containers[containerId].IPv4Address
	ipAddr = strings.Split(ipAddrCIDR, "/")[0]
	return ipAddr, nil
}

func (manager DockerManager) connectToNetwork(networkName string, containerId string, staticIpAddr string) (err error) {
	networkId, err := manager.getNetworkId(networkName)
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	err = manager.dockerClient.NetworkConnect(
		context.Background(),
		networkId,
		containerId,
		&network.EndpointSettings{
			IPAddress: staticIpAddr,
		})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to connect container %s to network %s.", containerId, networkName)
	}
	return nil
}

func (manager DockerManager) pullImage(imageName string) (err error) {
	log.Printf("Pulling image %s...", imageName)
	out, err := manager.dockerClient.ImagePull(manager.dockerCtx, imageName, types.ImagePullOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)
	return nil
}

// Creates a Docker-Container-To-Host Port mapping, defining how a Container's JSON RPC and service-specific ports are
// mapped to the host ports
func (manager *DockerManager) getContainerHostConfig(networkName string, usedPorts map[int]bool) (hostConfig *container.HostConfig, err error) {
	portMap := nat.PortMap{}
	for port, _ := range usedPorts {
		portObj, err := nat.NewPort("tcp", strconv.Itoa(port))
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not create port object fro port '%v'", port)
		}

		freeHostPort, err := manager.getFreePort()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not get a free host port!")
		}

		portMap[portObj] = []nat.PortBinding{
			{
				HostIP: manager.getLocalHostIp(),
				HostPort: freeHostPort.Port(),
			},
		}
	}
	containerHostConfigPtr := &container.HostConfig{
		PortBindings: portMap,
		NetworkMode: container.NetworkMode("default"),
	}
	return containerHostConfigPtr, nil
}

// TODO should I actually be passing sorta-complex objects like JsonRpcServiceConfig by value???
// Creates a more generalized Docker Container configuration for Gecko, with a 5-parameter initialization command.
// Gecko HTTP and Staking ports inside the Container are the standard defaults.
func (manager *DockerManager) getContainerCfgFromServiceCfg(
			dockerImage string,
			usedPorts map[int]bool,
			startCmdArgs []string) (config *container.Config, err error) {
	portSet := nat.PortSet{}
	for port, _ := range usedPorts {
		otherPort, err := nat.NewPort("tcp", strconv.Itoa(port))
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not parse port int.")
		}
		portSet[otherPort] = struct{}{}
	}

	nodeConfigPtr := &container.Config{
		Image: dockerImage,
		// TODO allow modifying of protocol at some point
		ExposedPorts: portSet,
		Cmd: startCmdArgs,
		Tty: false,
	}
	return nodeConfigPtr, nil
}
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
)

const (
	LOCAL_HOST_IP = "127.0.0.1"
	DOCKER_NETWORK_NAME ="kurtosis-bridge"
)

type DockerManager struct {
	dockerCtx           context.Context
	dockerClient        *client.Client
	freeHostPortTracker *FreeHostPortTracker
}

func NewDockerManager(
	dockerCtx context.Context,
	dockerClient *client.Client,
	hostPortRangeStart int,
	hostPortRangeEnd int) (dockerManager *DockerManager, err error) {

	freeHostPortTracker, err := NewFreeHostPortTracker(hostPortRangeStart, hostPortRangeEnd)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get a free port.")
	}
	return &DockerManager{
		dockerCtx:           dockerCtx,
		dockerClient:        dockerClient,
		freeHostPortTracker: freeHostPortTracker,
	}, nil
}

func (manager DockerManager) CreateNetwork(subnetMask string) (id string, err error)  {
	networkId, ok, err := manager.getNetworkId(DOCKER_NETWORK_NAME)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to check for network existence.")
	}
	// Network already exists - return existing id.
	if ok {
		return networkId, nil
	}
	ipamConfig := []network.IPAMConfig{{
		Subnet: subnetMask,
	}}
	resp, err := manager.dockerClient.NetworkCreate(manager.dockerCtx, DOCKER_NETWORK_NAME, types.NetworkCreate{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Config: ipamConfig,
		},
	})
	if err != nil {
		return "", stacktrace.Propagate( err, "Failed to create network %s with subnet %s", DOCKER_NETWORK_NAME, subnetMask)
	}
	return resp.ID, nil
}

func (manager DockerManager) CreateAndStartControllerContainer(
		dockerImage string,
		staticIp string,
		testName string,
		volumeName string) (containerIpAddr string, containerId string, err error){

	// TODO uncomment me
	/*
	volumeName := fmt.Sprintf("volume-%v-%v", testName, instanceUuid.String())
	volume.VolumeCreateBody{
		Driver:     "",
		DriverOpts: nil,
		Labels:     nil,
		Name:       volumeName,
	}
	manager.dockerClient.VolumeCreate()
	*/

	startCmdArgs := []string{
		testName,
	}

	controllerIp, controllerContainerId, err := manager.CreateAndStartContainer(dockerImage, staticIp, make(map[int]bool), startCmdArgs)
	return controllerIp, controllerContainerId, err
}

func (manager DockerManager) CreateAndStartContainer(
	dockerImage string,
	staticIp string,
	usedPorts map[int]bool,
	startCmdArgs []string) (containerIpAddr string, containerId string, err error) {

	manager.ensureImageExistsLocally(dockerImage)

	// TODO replace with configurable network
	_, networkExistsLocally, err := manager.getNetworkId(DOCKER_NETWORK_NAME)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to check for network availability.")
	}
	if !networkExistsLocally {
		return "", "", stacktrace.NewError("Kurtosis Docker network was never created before trying to launch containers. Please call DockerManager.CreateNetwork first.")
	}

	// TODO this relies on serviceId being incremental, and is a total hack until --public-ips flag is gone from Gecko!
	containerConfigPtr, err := manager.getContainerCfg(dockerImage, usedPorts, startCmdArgs)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "Failed to configure container from service.")
	}
	containerHostConfigPtr, err := manager.getContainerHostConfig(usedPorts)
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
	err = manager.connectToNetwork(DOCKER_NETWORK_NAME, containerId, staticIp)
	if err != nil {
		return "","", stacktrace.Propagate(err, "Failed to connect container %s to network.", containerId)
	}
	return staticIp, containerId, nil
}

func (manager DockerManager) ensureImageExistsLocally(dockerImage string) error {
	imageExistsLocally, err := manager.isImageAvailableLocally(dockerImage)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to check for image availability.")
	}

	if !imageExistsLocally {
		err = manager.pullImage(dockerImage)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to pull Docker image.")
		}
	}
	return nil
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

func (manager DockerManager) getNetworkId(networkName string) (networkId string, ok bool, err error) {
	referenceArg := filters.Arg("name", networkName)
	filters := filters.NewArgs(referenceArg)
	networks, err := manager.dockerClient.NetworkList(
		context.Background(),
		types.NetworkListOptions{
			Filters: filters,
		})
	if err != nil {
		return "", false, stacktrace.Propagate(err, "Failed to list networks.")
	}
	if len(networks) == 0 {
		return "", false, nil
	}
	return networks[0].ID, true, nil
}

func (manager DockerManager) connectToNetwork(networkName string, containerId string, staticIpAddr string) (err error) {
	networkId, ok, err := manager.getNetworkId(networkName)
	if err != nil || !ok {
		return stacktrace.Propagate(err, "Failed to get network id for %s", networkName)
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
		return stacktrace.Propagate(err, "Failed to pull image %s", imageName)
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)
	return nil
}

// Creates a Docker-Container-To-Host Port mapping, defining how a Container's JSON RPC and service-specific ports are
// mapped to the host ports
func (manager *DockerManager) getContainerHostConfig(usedPorts map[int]bool) (hostConfig *container.HostConfig, err error) {
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

// Creates a Docker container representing a service that will listen on ports in the network
func (manager *DockerManager) getContainerCfg(
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

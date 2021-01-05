/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package commons

// TODO TODO TODO Move everything in this file, and the auxiliary files, to "docker_manager" package

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"time"
)

/*
WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING

This manager is used on a per-test basis. Because tests can run in parallel but we need to pretty-print
each test's logs in a single block, we need to have a seprate logger per test. As such, this class takes in a
logrus.Logger, and *all log messages should be sent through this logger rather than the systemwide logger!!!*

No logrus.Info, logrus.Debug, etc. calls should happen in this file - only manager.log.Info, manager.log.Debug, etc.!

WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
 */

const (
	// We use a bridge network because, as of 2020-08-01, we're only running locally; however, this may need to change
	//  at some point in the future
	DOCKER_NETWORK_DRIVER = "bridge"
)

/*
A handle to interacting with the Docker environment running a test.
 */
type DockerManager struct {
	// The logger that all log messages will be written to
	log *logrus.Logger // NOTE: This log should be used for all log statements - the system-wide logger should NOT be used!

	// The underlying Docker client that will be used to modify the Docker environment
	dockerClient        *client.Client
}

/*
Creates a new Docker manager for manipulating the Docker engine using the given client.

Args:
	log: The logger that this Docker manager will write all its log messages to.
	dockerClient: The Docker client that will be used when interacting with the underlying Docker engine the Docker engine.
*/
func NewDockerManager(log *logrus.Logger, dockerClient *client.Client) (dockerManager *DockerManager, err error) {
	return &DockerManager{
		log: log,
		dockerClient:        dockerClient,
	}, nil
}

/*
Creates a new Docker network with the given parameters; does nothing if a network with the given name already exists.

Args:
	context: The Context that this request is running in (useful for cancellation)
	name: The name to give the new Docker network
	subnetMask: The subnet mask defining allowed IPs for the Docker network
	gatewayIP: The IP to give the network gateway

Returns:
	id: The Docker-managed ID of the network
 */
func (manager DockerManager) CreateNetwork(context context.Context, name string, subnetMask string, gatewayIP net.IP) (id string, err error)  {
	networkIds, err := manager.GetNetworkIdsByName(context, name)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred checking for existence of network with name %v", name)
	}
	if len(networkIds) != 0 {
		// We throw an error if the network already exists because we don't know what settings that network was created
		//  with - likely a completely different subnetMask and gatewayIP
		return "", stacktrace.NewError("Network with name %v cannot be created because one or more networks with that name already exist", name)
	}
	ipamConfig := []network.IPAMConfig{{
		Subnet: subnetMask,
		Gateway: gatewayIP.String(),
	}}
	resp, err := manager.dockerClient.NetworkCreate(context, name, types.NetworkCreate{
		Driver: DOCKER_NETWORK_DRIVER,
		IPAM: &network.IPAM{
			Config: ipamConfig,
		},
	})
	if err != nil {
		return "", stacktrace.Propagate( err, "Failed to create network %s with subnet %s", name, subnetMask)
	}
	return resp.ID, nil
}

/*
Returns the Docker network IDs of the networks matching the given name (if any).
 */
func (manager DockerManager) GetNetworkIdsByName(context context.Context, name string) ([]string, error) {
	networks, err := manager.getNetworksByFilter(context, "name", name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for existence of network with name %v", name)
	}

	result := []string{}
	for _, network := range networks {
		result = append(result, network.ID)
	}
	return result, nil
}


/*
Removes the Docker network with the given id, attempting to stop all containers connected to the network first (because
	otherwise, the remove call will fail)

Args:
	context: The Context that this request is running in (useful for cancellation)
	networkId: ID of Docker network to remove
	containerStopTimeout: How long to wait for containers to stop
 */
func (manager DockerManager) RemoveNetwork(context context.Context, networkId string, containerStopTimeout time.Duration) error {

	inspectResponse, err := manager.dockerClient.NetworkInspect(context, networkId, types.NetworkInspectOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get network information for network with ID %v", networkId)
	}

	for containerId, _ := range inspectResponse.Containers {
		if err := manager.dockerClient.ContainerStop(context, containerId, &containerStopTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping container with ID %v, which prevented the network from being removed", containerId)
		}
	}

	if err := manager.dockerClient.NetworkRemove(context, networkId); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the Docker network with ID %v", networkId)
	}
	return nil
}

/*
Creates a Docker volume identified by the given name.

Args:
	context: The Context that this request is running in (useful for cancellation)
	volumeName: The unique identifier used by Docker to identify this volume (NOTE: at time of writing, Docker doesn't
		even give volumes IDs - this name is all there is)
 */
func (manager DockerManager) CreateVolume(context context.Context, volumeName string) error {
	volumeConfig := volume.VolumeCreateBody{
		Name:       volumeName,
	}

	/*
	We don't use the return value of VolumeCreate because there's not much useful information on there - Docker doesn't
	use UUIDs to identify volumes - only the name - so there's no UUID to retrieve, and the volume's Mountpoint (what you'd
	think would be the path of the volume on the local machine) isn't useful either becuase Docker itself runs inside a VM
	so *this path is only a path inside the Docker VM* (meaning we can't use it to read/write files). AFAICT, the only way
	to read/write data to a volume is to mount it in a container. ~ ktoday, 2020-07-01
	 */
	_, err := manager.dockerClient.VolumeCreate(context, volumeConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Could not create Docker volume for test controller")
	}

	return nil
}


/*
Creates a Docker container with the given args and starts it.

Args:
	context: The Context that this request is running in (useful for cancellation)
	dockerImage: image to start
	networkId: The ID of the Docker network that this container should be attached to
	staticIp: IP the container will be assigned (leave nil to not assign any IP, which only works with the bridge network)
	addedCapabilities: A "set" of capabilities to add to the container, corresponding to the --cap-add Docker flag
		For more info, see the --cap-add section of https://docs.docker.com/engine/reference/run/
	networkMode: When a non-empty string, sets the Docker --network flag to be this given string
	usedPorts: A map of (ports that the container will listen on) -> (an *optional* IP/port on the host that
		will get bound to the container port); set the binding to nil to exclude it
	startCmdArgs: The args that will be used to run the container (leave as nil to run the CMD in the image)
	envVariables: A key-value mapping of Docker environment variables which will be passed to the container during startup
	bindMounts: Mapping of (host file) -> (mountpoint on container) that will be mounted on container startup
	volumeMounts: Mapping of (volume name) -> (mountpoint on container) to mount during container launch

Returns:
	The Docker container ID of the newly-created container
 */
func (manager DockerManager) CreateAndStartContainer(
			context context.Context,
			dockerImage string,
			networkId string,
			staticIp net.IP,
			addedCapabilities map[ContainerCapability]bool,
			networkMode DockerManagerNetworkMode,
			usedPortsWithHostBindings map[nat.Port]*nat.PortBinding,
			startCmdArgs []string,
			envVariables map[string]string,
			bindMounts map[string]string,
			volumeMounts map[string]string) (containerId string, err error) {

	imageExistsLocally, err := manager.isImageAvailableLocally(context, dockerImage)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred checking for local availability of Docker image %v", dockerImage)
	}

	if !imageExistsLocally {
		err = manager.pullImage(context, dockerImage)
		if err != nil {
			return "", stacktrace.Propagate(err, "Failed to pull Docker image %v from remote image repository", dockerImage)
		}
	}

	networks, err := manager.getNetworksByFilter(context, "id", networkId)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred checking for the existence of network with ID %v", networkId)
	}
	if len(networks) == 0 {
		return "", stacktrace.NewError("Kurtosis Docker network with ID %v was never created before trying to launch containers. Please call DockerManager.CreateNetwork first.", networkId)
	} else if len(networks) > 1 {
		return "", stacktrace.NewError("Kurtosis Docker network with ID %v matches several networks!", networkId)
	}

	usedPorts := map[nat.Port]bool{}
	usedPortsWithHostBindingsDeref := map[nat.Port]nat.PortBinding{}
	for usedPort, hostBinding := range usedPortsWithHostBindings {
		usedPorts[usedPort] = true
		if hostBinding != nil {
			usedPortsWithHostBindingsDeref[usedPort] = *hostBinding
		}
	}

	containerConfigPtr, err := manager.getContainerCfg(dockerImage, usedPorts, startCmdArgs, envVariables)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to configure container from service.")
	}
	containerHostConfigPtr, err := manager.getContainerHostConfig(
		addedCapabilities,
		networkMode,
		bindMounts,
		volumeMounts,
		usedPortsWithHostBindingsDeref)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to configure host to container mappings from service.")
	}
	resp, err := manager.dockerClient.ContainerCreate(context, containerConfigPtr, containerHostConfigPtr, nil, "")
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not create Docker container from image %v.", dockerImage)
	}
	containerId = resp.ID

	// If the user doesn't provide an IP, the Docker network will auto-assign one
	if staticIp != nil {
		if err := manager.connectToNetwork(context, networkId, containerId, staticIp); err != nil {
			return "", stacktrace.Propagate(err, "Failed to connect container %s to network.", containerId)
		}
	}

	if err := manager.dockerClient.ContainerStart(context, containerId, types.ContainerStartOptions{}); err != nil {
		return "", stacktrace.Propagate(err, "Could not start Docker container from image %v.", dockerImage)
	}
	return containerId, nil
}

/*
Stops the container with the given container ID, waiting for the provided timeout before forcefully terminating the container

Args:
	context: The context that the stopping runs in (useful for cancellation)
	containerId: ID of Docker container to stop
	timeout: How long to wait for container stoppage before throwing an errorj
 */
func (manager DockerManager) StopContainer(context context.Context, containerId string, timeout time.Duration) error {
	err := manager.dockerClient.ContainerStop(context, containerId, &timeout)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping container with ID '%v'", containerId)
	}
	return nil
}

/*
Blocks until the given container exits or the context is cancelled.

Args:
	context: Context the waiting will run in (useful for cancellation)
	containerId: The ID of the Docker container that should be waited on

Returns:
	exitCode: The exit code of the container if it stopped
	err: The error if an error occurred waiting for exit
 */
func (manager DockerManager) WaitForExit(context context.Context, containerId string) (exitCode int64, err error) {
	statusChannel, errChannel := manager.dockerClient.ContainerWait(context, containerId, container.WaitConditionNotRunning)

	// Blocks until one of the channels returns
	select {
	case channErr := <-errChannel:
		exitCode = 1
		err = stacktrace.Propagate(channErr, "Failed to wait for container to return.")
	case status := <-statusChannel:
		exitCode = status.StatusCode
		err = nil
	}
	return
}

/*
Gets the logs for the given container as a io.ReadCloser. The caller is responsible for closing the ReadCloser!!!

NOTE: These logs have STDOUT and STDERR multiplexed together, and the 'stdcopy' package needs to be used to
	demultiplex them per https://github.com/moby/moby/issues/32794
 */
func (manager DockerManager) GetContainerLogs(context context.Context, containerId string) (io.ReadCloser, error) {
	containerLogOpts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}
	readCloser, err := manager.dockerClient.ContainerLogs(context, containerId, containerLogOpts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs for container ID '%v'", containerId)
	}
	return readCloser, nil
}

/*
Executes the given command inside the container with the given ID, blocking until the command completes

If the command has a non-zero exit code, an error will be returned
 */
func (manager DockerManager) RunExecCommand(
		context context.Context,
		containerId string,
		command []string,
		logOutput io.Writer) error {
	dockerClient := manager.dockerClient
	execConfig := types.ExecConfig{
		Cmd:          command,
	}

	createResp, err := dockerClient.ContainerExecCreate(context, containerId, execConfig)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating the exec process")
	}

	execId := createResp.ID
	if execId == "" {
		return stacktrace.NewError("Got back an empty exec ID when running '%v' on container '%v'", command, containerId)
	}

	execStartConfig := types.ExecStartCheck{
		// Because detach is false, we'll block until the command comes back
		Detach: false,
		Tty:    false,
	}
	if err := dockerClient.ContainerExecStart(context, execId, execStartConfig); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred starting the exec command")
	}

	attachResp, err := dockerClient.ContainerExecAttach(context, execId, execStartConfig)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred attaching to the exec command")
	}
	defer attachResp.Close()

	// NOTE: We have to demultiplex the logs that come back
	// This will keep reading until it receives EOF
	if _, err := stdcopy.StdCopy(logOutput, logOutput, attachResp.Reader); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying the exec command output to the given output writer")
	}

	inspectResponse, err := dockerClient.ContainerExecInspect(context, execId)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred inspecting the exec to get the response code")
	}
	if inspectResponse.Running {
		return stacktrace.NewError("Expected exec to have stopped, but it's still running!")
	}
	if inspectResponse.ExitCode != 0 {
		return stacktrace.NewError("Expected exit code 0 but exec exit code was '%v'", inspectResponse.ExitCode)
	}
	return  nil
}


// =================================================================================================================
//                                          INSTANCE HELPER FUNCTIONS
// =================================================================================================================
func (manager DockerManager) isImageAvailableLocally(ctx context.Context, imageName string) (isAvailable bool, err error) {
	referenceArg := filters.Arg("reference", imageName)
	filters := filters.NewArgs(referenceArg)
	images, err := manager.dockerClient.ImageList(
		ctx,
		types.ImageListOptions{
			All: true,
			Filters: filters,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to list images.")
	}
	return len(images) > 0, nil
}

func (manager DockerManager) getNetworksByFilter(ctx context.Context, filterKey string, filterValue string) ([]types.NetworkResource, error) {
	referenceArg := filters.Arg(filterKey, filterValue)
	filters := filters.NewArgs(referenceArg)
	networks, err := manager.dockerClient.NetworkList(
		ctx,
		types.NetworkListOptions{
			Filters: filters,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list networks while doing a filter search for %v = %v", filterKey, filterValue)
	}
	return networks, nil
}

func (manager DockerManager) connectToNetwork(ctx context.Context, networkId string, containerId string, staticIpAddr net.IP) (err error) {
	manager.log.Tracef(
		"Connecting container ID %v to network ID %v using static IP %v",
		containerId,
		networkId,
		staticIpAddr.String())

	ipamConfig := &network.EndpointIPAMConfig{
		IPv4Address:  staticIpAddr.String(),
	}

	err = manager.dockerClient.NetworkConnect(
		ctx,
		networkId,
		containerId,
		&network.EndpointSettings{
			IPAMConfig: ipamConfig,
		})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to connect container %s to network with ID %s.", containerId, networkId)
	}
	return nil
}

func (manager DockerManager) pullImage(context context.Context, imageName string) (err error) {
	manager.log.Infof("Pulling image '%s'...", imageName)
	out, err := manager.dockerClient.ImagePull(context, imageName, types.ImagePullOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to pull image %s", imageName)
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)
	return nil
}

/*
Creates a Docker-Container-To-Host Port mapping, defining how a Container's JSON RPC and service-specific ports are
mapped to the host ports.

Args:
	bindMounts: Mapping of (host file) -> (mountpoint on container) that will be mounted at container startup (used when
		sharing data between the host filesystem - in our case, the test initializer - and a Docker container)
	volumeMounts: Mapping of (volume name) -> (mountpoint on container) that will be mounted at container startup (used
		when sharing data between containers). This is distinct from a bind mount because the host filesystem can't easily
		read from a Docker volume - you need to be inside a Docker container to do so.
	hostPortBindings: Mapping of (container port) -> (IP/port on host to bind to)
 */
func (manager *DockerManager) getContainerHostConfig(
		addedCapabilities map[ContainerCapability]bool,
		networkMode DockerManagerNetworkMode,
		bindMounts map[string]string,
		volumeMounts map[string]string,
		hostPortBindings map[nat.Port]nat.PortBinding) (hostConfig *container.HostConfig, err error) {

	bindsList := make([]string, 0, len(bindMounts))
	for hostFilepath, containerFilepath := range bindMounts {
		bindsList = append(bindsList, hostFilepath + ":" + containerFilepath)
	}
	for volumeName, containerFilepath := range volumeMounts {
		// Yes, it's SUPER confusing that "volumes" need to be put into the "binds" section because there's
		//  a separate thing called a "bind mount".... blame the Docker API
		bindsList = append(bindsList, volumeName + ":" + containerFilepath)
	}

	manager.log.Debugf("Binds: %v", bindsList)

	portMap := nat.PortMap{}
	for containerPort, hostBinding := range hostPortBindings {
		portMap[containerPort] = []nat.PortBinding{hostBinding}
	}

	addedCapabilitiesSlice := []string{}
	for capability, _ := range addedCapabilities {
		capabilityStr := string(capability)
		addedCapabilitiesSlice = append(addedCapabilitiesSlice, capabilityStr)
	}

	containerHostConfigPtr := &container.HostConfig{
		Binds: bindsList,
		CapAdd: addedCapabilitiesSlice,
		NetworkMode: container.NetworkMode(networkMode),
		PortBindings: portMap,
	}
	return containerHostConfigPtr, nil
}

// Creates a Docker container representing a service that will listen on ports in the network
func (manager *DockerManager) getContainerCfg(
			dockerImage string,
			usedPorts map[nat.Port]bool,
			startCmdArgs []string,
			envVariables map[string]string) (config *container.Config, err error) {
	portSet := nat.PortSet{}
	for port, _ := range usedPorts {
		portSet[port] = struct{}{}
	}

	envVariablesSlice := make([]string, 0, len(envVariables))
	for key, val := range envVariables {
		envVariablesSlice = append(envVariablesSlice, fmt.Sprintf("%v=%v", key, val))
	}

	nodeConfigPtr := &container.Config{
		Tty: false,
		Image: dockerImage,
		ExposedPorts: portSet,
		Cmd: startCmdArgs,
		Env: envVariablesSlice,
	}
	return nodeConfigPtr, nil
}

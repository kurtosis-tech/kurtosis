/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package docker_manager

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
	"github.com/kurtosis-tech/kurtosis/commons/docker_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"net"
	"strings"
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
	dockerNetworkDriver = "bridge"

	// Per https://docs.docker.com/engine/reference/commandline/kill/ , this seems to mean "the default
	//  kill signal"
	dockerKillSignal = "KILL"

	nameFilterKey = "name"

	expectedHostIp = "0.0.0.0"
)

// The dimensions of the TTY that the container should output to when in interactive mode
type InteractiveModeTtySize struct {
	Height uint
	Width uint
}

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
func NewDockerManager(log *logrus.Logger, dockerClient *client.Client) *DockerManager {
	return &DockerManager{
		log: log,
		dockerClient:        dockerClient,
	}
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
		Driver: dockerNetworkDriver,
		IPAM: &network.IPAM{
			Config: ipamConfig,
		},
	})
	if err != nil {
		return "", stacktrace.Propagate( err, "Failed to create network %s with subnet %s", name, subnetMask)
	}
	return resp.ID, nil
}

func (manager DockerManager) ListNetworks(ctx context.Context) ([]types.NetworkResource, error) {
	networks, err := manager.dockerClient.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred listing the Docker networks")
	}
	return networks, nil
}

/*
Returns the Docker network IDs of the networks matching the given name (if any).
 */
func (manager DockerManager) GetNetworkIdsByName(ctx context.Context, name string) ([]string, error) {
	networks, err := manager.getNetworksByFilter(ctx, nameFilterKey, name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for existence of network with name %v", name)
	}

	result := []string{}
	for _, network := range networks {
		result = append(result, network.ID)
	}
	return result, nil
}

func (manager DockerManager) GetContainerIdsConnectedToNetwork(context context.Context, networkId string) ([]string, error) {
	inspectResponse, err := manager.dockerClient.NetworkInspect(context, networkId, types.NetworkInspectOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get network information for network with ID '%v'", networkId)
	}
	result := []string{}
	for containerId := range inspectResponse.Containers {
		result = append(result, containerId)
	}
	return result, nil
}


/*
Removes the Docker network with the given id

NOTE: All containers attached to the network must be shut off or disconnected, else the call will fail!
	otherwise, the remove call will fail)

Args:
	context: The Context that this request is running in (useful for cancellation)
	networkId: ID of Docker network to remove
 */
func (manager DockerManager) RemoveNetwork(context context.Context, networkId string) error {
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
	dockerImage: Image to start
	name: The name to give the container to be created
	interactiveModeTerminalSize: If non-nil, the container will be started in interactive mode, with a container TTY
		set to the specified dimensions
	networkId: The ID of the Docker network that this container should be attached to
	staticIp: IP the container will be assigned (leave nil to not assign any IP, which only works with the bridge network)
	addedCapabilities: A "set" of capabilities to add to the container, corresponding to the --cap-add Docker flag
		For more info, see the --cap-add section of https://docs.docker.com/engine/reference/run/
	networkMode: When a non-empty string, sets the Docker --network flag to be this given string
	usedPorts: A set of ports that the container will listen on
	shouldPublishAllPorts: If true, we'll publish all the exposed ports to the Docker host so that the outside world can connect
		to the container
	entrypointArgs: The args that will be used to override the ENTRYPOINT of the image (leave as nil to not override)
	cmdArgs: The args that will be used to run the container (leave as nil to run the CMD in the image)
	envVariables: A key-value mapping of Docker environment variables which will be passed to the container during startup
	bindMounts: Mapping of (host file) -> (mountpoint on container) that will be mounted on container startup
	volumeMounts: Mapping of (volume name) -> (mountpoint on container) to mount during container launch
	needsAccessToDockerHostMachine: Will provide the container with a magic "host.docker.internal" domain name
		that it can use to access ports of the machine running Docker itself (useful if, e.g., the container
		needs to check the host machine's free ports)

Returns:
	containerId: The Docker container ID of the newly-created container
	containerHostPortBindings: If shouldPublishAllPorts is true, returns the ports on the host container interface where each of the
		container's exposed ports can be found; if shouldPublishAllPorts is false, this will be an empty map
 */
func (manager DockerManager) CreateAndStartContainer(
			context context.Context,
			dockerImage string,
			name string,
			alias string,
			interactiveModeTtySize *InteractiveModeTtySize, // If nil, interactive mode will be disabled; if non-nil, then interactive mode will be enabled
			networkId string,
			staticIp net.IP,
			addedCapabilities map[ContainerCapability]bool,
			networkMode DockerManagerNetworkMode,
			usedPortsSet map[nat.Port]bool,
			shouldPublishAllPorts bool,
			entrypointArgs []string,
			cmdArgs []string,
			envVariables map[string]string,
			bindMounts map[string]string,
			volumeMounts map[string]string,
			needsAccessToDockerHostMachine bool) (containerId string, containerPortHostBindings map[nat.Port]*nat.PortBinding, err error) {

	imageExistsLocally, err := manager.isImageAvailableLocally(context, dockerImage)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred checking for local availability of Docker image %v", dockerImage)
	}

	if !imageExistsLocally {
		err = manager.pullImage(context, dockerImage)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Failed to pull Docker image %v from remote image repository", dockerImage)
		}
	}

	networks, err := manager.getNetworksByFilter(context, "id", networkId)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred checking for the existence of network with ID %v", networkId)
	}
	if len(networks) == 0 {
		return "", nil, stacktrace.NewError("Kurtosis Docker network with ID %v was never created before trying to launch containers. Please call DockerManager.CreateNetwork first.", networkId)
	} else if len(networks) > 1 {
		return "", nil, stacktrace.NewError("Kurtosis Docker network with ID %v matches several networks!", networkId)
	}

	isInteractiveMode := interactiveModeTtySize != nil

	containerConfigPtr, err := manager.getContainerCfg(
		dockerImage,
		isInteractiveMode,
		usedPortsSet,
		entrypointArgs,
		cmdArgs,
		envVariables,
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Failed to configure container from service.")
	}
	containerHostConfigPtr, err := manager.getContainerHostConfig(
		addedCapabilities,
		networkMode,
		bindMounts,
		volumeMounts,
		usedPortsSet,
		shouldPublishAllPorts,
		needsAccessToDockerHostMachine)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Failed to configure host to container mappings from service.")
	}
	resp, err := manager.dockerClient.ContainerCreate(context, containerConfigPtr, containerHostConfigPtr, nil, name)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Could not create Docker container '%v' from image '%v'", name, dockerImage)
	}
	containerId = resp.ID

	// If the user doesn't provide an IP, the Docker network will auto-assign one
	if staticIp != nil {
		if err := manager.ConnectContainerToNetwork(context, networkId, containerId, staticIp, alias); err != nil {
			return "", nil, stacktrace.Propagate(err, "Failed to connect container %s to network.", containerId)
		}
	}

	if err := manager.dockerClient.ContainerStart(context, containerId, types.ContainerStartOptions{}); err != nil {
		return "", nil, stacktrace.Propagate(err, "Could not start Docker container from image %v.", dockerImage)
	}
	functionFinishedSuccessfully := false
	defer func() {
		if !functionFinishedSuccessfully {
			if err := manager.KillContainer(context, containerId); err != nil {
				manager.log.Error("The container creation function didn't finish successfully, meaning we needed to kill the container we created. However, the killing threw an error:")
				fmt.Fprintln(manager.log.Out, err)
				manager.log.Errorf("ACTION NEEDED: You'll need to manually kill this container with ID '%v'", containerId)
			}
		}
	}()

	if isInteractiveMode {
		/*
		Two notes:
		 1) Container resizing must be done after the container is started
		 2) This resize is very important - if we don't do it, then the output will look garbled for
			 lines longer than the user's terminal
		*/
		resizeOpts := types.ResizeOptions{
			Height: interactiveModeTtySize.Height,
			Width:  interactiveModeTtySize.Width,
		}
		if err := manager.dockerClient.ContainerResize(context, containerId, resizeOpts); err != nil {
			return "", nil, stacktrace.Propagate(
				err,
				"An error occurred resizing the new container's TTY size to height %v and width %v to match the user's terminal",
				interactiveModeTtySize.Height,
				interactiveModeTtySize.Width,
			)
		}
	}

	// If the user wanted their ports exposed, Docker will have auto-assigned the ports to ports in the ephemeral range
	//  on the host. We need to look up what those ports are so we can return report them back to the user.
	resultHostPortBindings := map[nat.Port]*nat.PortBinding{}
	if shouldPublishAllPorts {
		resp, err := manager.dockerClient.ContainerInspect(context, containerId)
		if err != nil {
			return "", nil, stacktrace.Propagate(
				err,
				"Publishing all ports was requested, but an error occurred inspecting the newly-started " +
					"container which is necessary for determining which host ports the container's ports were bound to",
			)
		}
		networkSettings := resp.NetworkSettings
		if networkSettings == nil {
			return "", nil, stacktrace.NewError(
				"We got a response from inspecting container '%v' which is necessary for determining the " +
					"exposed host ports, but the network settings object was nil",
				containerId,
			)
		}
		allInterfaceHostPortBindings := networkSettings.Ports
		if allInterfaceHostPortBindings == nil {
			return "", nil, stacktrace.NewError(
				"Pulbishing all ports was requested for container '%v', but the container host port bindings were null",
				containerId,
			)
		}

		portBindingsOnExpectedInterface := map[nat.Port]*nat.PortBinding{}
		for port, allInterfaceBindings := range allInterfaceHostPortBindings {
			for _, interfaceBinding := range allInterfaceBindings {
				if interfaceBinding.HostIP == expectedHostIp {
					portBindingsOnExpectedInterface[port] = &nat.PortBinding{
						HostIP:   interfaceBinding.HostIP,
						HostPort: interfaceBinding.HostPort,
					}
				}
			}
		}

		numUsedPorts := len(usedPortsSet)
		numPortBindingsOnExpectedInterface := len(portBindingsOnExpectedInterface)
		if numUsedPorts != numPortBindingsOnExpectedInterface {
			return "", nil, stacktrace.NewError(
				"Publishing all ports was requested, but there were %v used ports declared while only %v ports got host port bindings on the expected interface '%v'",
				numUsedPorts,
				numPortBindingsOnExpectedInterface,
				expectedHostIp,
			)
		}

		resultHostPortBindings = portBindingsOnExpectedInterface
	}

	functionFinishedSuccessfully = true
	return containerId, resultHostPortBindings, nil
}

// Gets the container's ID on a given network
// NOTE: Yes, it's a testament to how poorly-designed the Docker API is that we need to use network name here even though
//  everywhere else in the Docker API uses network ID
func (manager DockerManager) GetContainerIP(ctx context.Context, networkName string, containerId string) (string, error) {
	resp, err := manager.dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred inspecting container with ID '%v'", containerId)
	}
	allNetworkInfo := resp.NetworkSettings.Networks
	networkInfo, found := allNetworkInfo[networkName]
	if !found {
		return "", stacktrace.NewError("Container ID '%v' isn't connected to network '%v'", containerId, networkName)
	}
	return networkInfo.IPAddress, nil
}

func (manager DockerManager) AttachToContainer(ctx context.Context, containerId string) (types.HijackedResponse, error) {
	attachOpts := types.ContainerAttachOptions{
		Stream:     true,
		Stdin:      true,
		Stdout:     true,
		Stderr:     true,
	}
	hijackedResponse, err := manager.dockerClient.ContainerAttach(ctx, containerId, attachOpts)
	if err != nil {
		return types.HijackedResponse{}, stacktrace.Propagate(err, "An error occurred attaching to container '%v'", containerId)
	}
	return hijackedResponse, nil
}

/*
Stops the container with the given container ID, waiting for the provided timeout before forcefully terminating the container

Args:
	context: The context that the stopping runs in (useful for cancellation)
	containerId: ID of Docker container to stop
	timeout: How long to wait for container stoppage before throwing an error
 */
func (manager DockerManager) StopContainer(context context.Context, containerId string, timeout time.Duration) error {
	err := manager.dockerClient.ContainerStop(context, containerId, &timeout)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping container with ID '%v'", containerId)
	}
	return nil
}

/*
Kills the container with the given ID if it's running, giving it no opportunity to gracefully exit

Args:
	context: The context that the kill runs in
	containerId: ID of Docker container to kill
 */
func (manager DockerManager) KillContainer(context context.Context, containerId string) error {
	err := manager.dockerClient.ContainerKill(context, containerId, dockerKillSignal)
	if err != nil {
		// For some stupid reason, ContainerKill throws an error if the container isn't running (even though
		//  ContainerStop does not)
		if strings.Contains(err.Error(), "is not running") {
			return nil
		}
		return stacktrace.Propagate(err, "An error occurred killing container with ID '%v'", containerId)
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
func (manager DockerManager) GetContainerLogs(context context.Context, containerId string, shouldFollowLogs bool) (io.ReadCloser, error) {
	containerLogOpts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow: shouldFollowLogs,
	}
	readCloser, err := manager.dockerClient.ContainerLogs(context, containerId, containerLogOpts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs for container ID '%v'", containerId)
	}
	return readCloser, nil
}

/*
Executes the given command inside the container with the given ID, blocking until the command completes
 */
func (manager DockerManager) RunExecCommand(context context.Context, containerId string, command []string, logOutput io.Writer) (int32, error) {
	dockerClient := manager.dockerClient
	execConfig := types.ExecConfig{
		Cmd:          command,
		AttachStderr: true,
		AttachStdout: true,
		Detach: false,
	}

	createResp, err := dockerClient.ContainerExecCreate(context, containerId, execConfig)
	if err != nil {
		return 0, stacktrace.Propagate(
			err,
			"An error occurred creating the exec process")
	}

	execId := createResp.ID
	if execId == "" {
		return 0, stacktrace.NewError("Got back an empty exec ID when running '%v' on container '%v'", command, containerId)
	}

	execStartConfig := types.ExecStartCheck{
		// Because detach is false, we'll block until the command comes back
		Detach: false,
	}

	// IMPORTANT NOTE:
	// You'd think that we'd need to call ContainerExecStart separately after this ContainerExecAttach....
	//  ...but ContainerExecAttach **actually starts the exec command!!!!**
	// We used to be doing them both, but then we were hitting this occasional race condition: https://github.com/moby/moby/issues/42408
	// Therefore, we ONLY call Attach, without Start
	attachResp, err := dockerClient.ContainerExecAttach(context, execId, execStartConfig)
	if err != nil {
		return 0, stacktrace.Propagate(
			err,
			"An error occurred starting/attaching to the exec command")
	}
	defer attachResp.Close()

	// NOTE: We have to demultiplex the logs that come back
	// This will keep reading until it receives EOF
	if _, err := stdcopy.StdCopy(logOutput, logOutput, attachResp.Reader); err != nil {
		return 0, stacktrace.Propagate(
			err,
			"An error occurred copying the exec command output to the given output writer")
	}

	inspectResponse, err := dockerClient.ContainerExecInspect(context, execId)
	if err != nil {
		return 0, stacktrace.Propagate(
			err,
			"An error occurred inspecting the exec to get the response code")
	}
	if inspectResponse.Running {
		return 0, stacktrace.NewError("Expected exec to have stopped, but it's still running!")
	}
	unsizedExitCode := inspectResponse.ExitCode
	if unsizedExitCode > math.MaxInt32 || unsizedExitCode < math.MinInt32 {
		return 0, stacktrace.NewError("Could not cast unsized int '%v' to int32 because it does not fit", unsizedExitCode)
	}
	int32ExitCode := int32(unsizedExitCode)
	return int32ExitCode, nil
}

/*
Connects the container with the given container ID to the network with the given network ID, using the given IP address
*/
func (manager DockerManager) ConnectContainerToNetwork(ctx context.Context, networkId string, containerId string, staticIpAddr net.IP, alias string) (err error) {
	manager.log.Tracef(
		"Connecting container ID %v to network ID %v using static IP %v",
		containerId,
		networkId,
		staticIpAddr.String())

	ipamConfig := &network.EndpointIPAMConfig{
		IPv4Address:  staticIpAddr.String(),
	}

	config := &network.EndpointSettings{
		IPAMConfig: ipamConfig,
	}

	if alias != "" {
		config.Aliases = []string{alias}
	}

	err = manager.dockerClient.NetworkConnect(
		ctx,
		networkId,
		containerId,
		config,
	)

	if err != nil {
		return stacktrace.Propagate(err, "Failed to connect container %s to network with ID %s.", containerId, networkId)
	}
	return nil
}

func (manager DockerManager) DisconnectContainerFromNetwork(ctx context.Context, containerId string, networkId string) error {
	if err := manager.dockerClient.NetworkDisconnect(ctx, networkId, containerId, true); err != nil {
		return stacktrace.Propagate(err, "An error occurred disconnecting container '%v' from network '%v'", containerId, networkId)
	}
	return nil
}

func (manager DockerManager) GetContainerIdsByName(ctx context.Context, nameStr string) ([]string, error) {
	filterArg := filters.Arg(nameFilterKey, nameStr)
	nameFilterList := filters.NewArgs(filterArg)
	opts := types.ContainerListOptions{
		Filters: nameFilterList,
	}
	containers, err := manager.dockerClient.ContainerList(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the containers with names matching string '%v'", nameStr)
	}
	result := []string{}
	for _, containerObj := range containers {
		result = append(result, containerObj.ID)
	}
	return result, nil
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
	exposedPorts: Set of container ports to expose
	shouldPublishAllPorts: If true, we'll publish all the exposed ports to the Docker host so that the outside world can connect
		to the container
	needsToAccessDockerHostMachine: If true, adds a "host.docker.internal:host-gateway" extra host binding, which is necessary
		for machines that will need to access the machine hosting Docker itself.

 */
func (manager *DockerManager) getContainerHostConfig(
		addedCapabilities map[ContainerCapability]bool,
		networkMode DockerManagerNetworkMode,
		bindMounts map[string]string,
		volumeMounts map[string]string,
		exposedPorts map[nat.Port]bool,
		shouldPublishAllPorts bool,
		needsToAccessDockerHostMachine bool) (hostConfig *container.HostConfig, err error) {

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
	for containerPort := range exposedPorts {
		portMap[containerPort] = nil
	}

	addedCapabilitiesSlice := []string{}
	for capability, _ := range addedCapabilities {
		capabilityStr := string(capability)
		addedCapabilitiesSlice = append(addedCapabilitiesSlice, capabilityStr)
	}

	extraHosts := []string{}
	if needsToAccessDockerHostMachine {
		// This explicit specification is necessary because in Docker-for-Linux, the magic "host.docker.internal"
		//  domain name isn't automatically available inside a container
		extraHosts = append(
			extraHosts,
			fmt.Sprintf("%v:%v", docker_constants.HostMachineDomainInsideContainer, docker_constants.HostGatewayName),
		)
	}

	containerHostConfigPtr := &container.HostConfig{
		Binds: bindsList,
		CapAdd: addedCapabilitiesSlice,
		NetworkMode: container.NetworkMode(networkMode),
		PortBindings: portMap,
		PublishAllPorts: shouldPublishAllPorts,
		ExtraHosts: extraHosts,
	}
	return containerHostConfigPtr, nil
}

// Creates a Docker container representing a service that will listen on ports in the network
func (manager *DockerManager) getContainerCfg(
			dockerImage string,
			isInteractiveMode bool,
			usedPorts map[nat.Port]bool,
			entrypointArgs []string,
			cmdArgs []string,
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
		AttachStderr: isInteractiveMode,	// Analogous to `-a STDERR` option to `docker run`
		AttachStdin:  isInteractiveMode,	// Analogous to `-a STDIN` option to `docker run`
		AttachStdout: isInteractiveMode,	// Analogous to `-a STDOUT` option to `docker run`
		Tty:          isInteractiveMode,	// Analogous to the `-t` option to `docker run`
		OpenStdin: true,	// Analogous to the `-i` option to `docker run`
		Image:        dockerImage,
		ExposedPorts: portSet,
		Cmd:          cmdArgs,
		Entrypoint:   entrypointArgs,
		Env:          envVariablesSlice,
	}
	return nodeConfigPtr, nil
}

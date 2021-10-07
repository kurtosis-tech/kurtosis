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
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

/*
WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING

This manager is used on a per-test basis. Because tests can run in parallel, but we need to pretty-print
each test's logs in a single block, we need to have a separate logger per test. As such, this class takes in a
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
	labelFilterKey = "label"

	expectedHostIp = "0.0.0.0"

	// Character Docker uses to separate the repo from
	dockerTagSeparatorChar = ":"

	// If no tag is specified for an image, this is the tag Dock
	dockerDefaultTag = "latest"

	// For some reason, when publish-all-ports is requested, Docker will return successfully from starting a
	//  container, but without having bound the host ports
	// See: https://github.com/moby/moby/issues/42860
	// To work around this, we retry a few times
	timeBetweenHostPortBindingChecks = 500 * time.Millisecond
	maxNumHostPortBindingChecks = 4
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

Returns:
	containerId: The Docker container ID of the newly-created container
	containerHostPortBindings: If shouldPublishAllPorts is true, returns the ports on the host container interface where each of the
		container's exposed ports can be found; if shouldPublishAllPorts is false, this will be an empty map
 */
func (manager DockerManager) CreateAndStartContainer(
	ctx context.Context,
	args *CreateAndStartContainerArgs,
) (string, map[nat.Port]*nat.PortBinding, error) {

	// If the user passed in a Docker image that doesn't have a tag separator (indicating no tag was specified), manually append
	//  the Docker default tag so that when we search for the image we're searching for a very specific image
	dockerImage := args.dockerImage
	if !strings.Contains(dockerImage, dockerTagSeparatorChar) {
		dockerImage = dockerImage + dockerTagSeparatorChar + dockerDefaultTag
	}

	manager.log.Tracef("Checking if image '%v' is available locally...", dockerImage)
	imageExistsLocally, err := manager.isImageAvailableLocally(ctx, dockerImage)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred checking for local availability of Docker image %v", dockerImage)
	}
	manager.log.Tracef("Is image available locally?: %v", imageExistsLocally)

	if !imageExistsLocally {
		manager.log.Tracef("Image doesn't exist locally, so attempting to pull it...")
		err = manager.PullImage(ctx, dockerImage)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Failed to pull Docker image %v from remote image repository", dockerImage)
		}
		manager.log.Tracef("Image successfully pulled from remote to local")
	}

	networks, err := manager.getNetworksByFilter(ctx, "id", args.networkId)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred checking for the existence of network with ID %v", args.networkId)
	}
	if len(networks) == 0 {
		return "", nil, stacktrace.NewError(
			"Kurtosis Docker network with ID %v was never created before trying to launch containers. Please call DockerManager.CreateNetwork first.",
			args.networkId,
		)
	} else if len(networks) > 1 {
		return "", nil, stacktrace.NewError("Kurtosis Docker network with ID %v matches several networks!", args.networkId)
	}

	isInteractiveMode := args.interactiveModeTtySize != nil

	containerConfigPtr, err := manager.getContainerCfg(
		dockerImage,
		isInteractiveMode,
		args.usedPortsSet,
		args.entrypointArgs,
		args.cmdArgs,
		args.envVariables,
		args.labels,
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Failed to configure container from service.")
	}
	containerHostConfigPtr, err := manager.getContainerHostConfig(
		args.addedCapabilities,
		args.networkMode,
		args.bindMounts,
		args.volumeMounts,
		args.usedPortsSet,
		args.shouldPublishAllPorts,
		args.needsAccessToDockerHostMachine)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Failed to configure host to container mappings from service.")
	}
	containerCreateResp, err := manager.dockerClient.ContainerCreate(ctx, containerConfigPtr, containerHostConfigPtr, nil, args.name)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Could not create Docker container '%v' from image '%v'", args.name, dockerImage)
	}
	containerId := containerCreateResp.ID
	if containerId == "" {
		return "", nil, stacktrace.NewError(
			"Creation of container '%v' from image '%v' succeeded without error, but we didn't get a container ID back - this is VERY strange!",
			args.name,
			dockerImage,
		)
	}
	manager.log.Debugf("Created container with ID '%v' from image '%v'", containerId, dockerImage)

	// If the user doesn't provide an IP, the Docker network will auto-assign one
	if args.staticIp != nil {
		if err := manager.ConnectContainerToNetwork(ctx, args.networkId, containerId, args.staticIp, args.alias); err != nil {
			return "", nil, stacktrace.Propagate(err, "Failed to connect container %s to network.", containerId)
		}
	}
	// TODO defer a disconnct-from-network if this function doesn't succeed??

	if err := manager.dockerClient.ContainerStart(ctx, containerId, types.ContainerStartOptions{}); err != nil {
		return "", nil, stacktrace.Propagate(err, "Could not start Docker container from image %v.", dockerImage)
	}
	functionFinishedSuccessfully := false
	defer func() {
		if !functionFinishedSuccessfully {
			if err := manager.KillContainer(ctx, containerId); err != nil {
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
			Height: args.interactiveModeTtySize.Height,
			Width:  args.interactiveModeTtySize.Width,
		}
		if err := manager.dockerClient.ContainerResize(ctx, containerId, resizeOpts); err != nil {
			return "", nil, stacktrace.Propagate(
				err,
				"An error occurred resizing the new container's TTY size to height %v and width %v to match the user's terminal",
				args.interactiveModeTtySize.Height,
				args.interactiveModeTtySize.Width,
			)
		}
	}

	// If the user wanted their ports exposed, Docker will have auto-assigned the ports to ports in the ephemeral range
	//  on the host. We need to look up what those ports are so we can return report them back to the user.
	resultHostPortBindings := map[nat.Port]*nat.PortBinding{}
	if args.shouldPublishAllPorts {
		// Thanks to https://github.com/moby/moby/issues/42860, we have to retry several times to get the host port bindings
		//  from Docker
		for i := 0; i < maxNumHostPortBindingChecks; i++ {
			manager.log.Tracef("Trying to get host port bindings (%v previous attempts)...", i)
			containerInspectResp, err := manager.dockerClient.ContainerInspect(ctx, containerId)
			if err != nil {
				return "", nil, stacktrace.Propagate(
					err,
					"Publishing all ports was requested, but an error occurred inspecting the newly-started "+
						"container which is necessary for determining which host ports the container's ports were bound to",
				)
			}
			manager.log.Tracef("Container inspect response: %+v", containerInspectResp)
			networkSettings := containerInspectResp.NetworkSettings
			if networkSettings == nil {
				return "", nil, stacktrace.NewError(
					"We got a response from inspecting container '%v' which is necessary for determining the "+
						"exposed host ports, but the network settings object was nil",
					containerId,
				)
			}
			manager.log.Tracef("Network settings: %+v", networkSettings)
			allInterfaceHostPortBindings := networkSettings.Ports
			if allInterfaceHostPortBindings == nil {
				return "", nil, stacktrace.NewError(
					"Publishing all ports was requested for container '%v', but the container host port bindings were null",
					containerId,
				)
			}
			manager.log.Tracef("Network settings -> ports: %+v", allInterfaceHostPortBindings)

			// This is "candidate" because if Docker is missing ports, it may end up as empty or half-filled (which we won't accept)
			candidatePortBindingsOnExpectedInterface := manager.getHostPortBindingsFromDockerInspectResult(args.usedPortsSet, allInterfaceHostPortBindings)
			if len(candidatePortBindingsOnExpectedInterface) == len(args.usedPortsSet) {
				resultHostPortBindings = candidatePortBindingsOnExpectedInterface
				break
			}
			time.Sleep(timeBetweenHostPortBindingChecks)
		}

		// Final verification that all used ports get a host machine port bindings
		if len(resultHostPortBindings) != len(args.usedPortsSet) {
			return "", nil, stacktrace.NewError(
				"Publishing all ports was requested, but container '%v' never got host port bindings for all the user ports on host machine interface '%v'",
				containerId,
				expectedHostIp,
			)
		}
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
	result, err := manager.getContainerIdsByFilterArgs(ctx, nameFilterList, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the containers with name '%v'", nameStr)
	}
	return result, nil
}

func (manager DockerManager) GetContainersByLabels(ctx context.Context, labels map[string]string, all bool) ([]*Container, error) {
	labelsFilterList := getLabelsFilterList(labels)
	result, err := manager.getContainersByFilterArgs(ctx, labelsFilterList, all)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting containers with labels '%+v'", labelsFilterList)
	}
	return result, nil
}

func (manager DockerManager) PullImage(context context.Context, imageName string) (err error) {
	manager.log.Infof("Pulling image '%s'...", imageName)
	out, err := manager.dockerClient.ImagePull(context, imageName, types.ImagePullOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to pull image %s", imageName)
	}
	defer out.Close()
	io.Copy(ioutil.Discard, out)
	return nil
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
	numMatchingImages := len(images)
	if numMatchingImages > 1 {
		return false, stacktrace.NewError(
			"Searching for Docker images matching image name '%v' returned %v images; " +
				"this indicates a bug because the image name being searched should only reference 0 or 1 images. Images matched:\n%+v",
			imageName,
			numMatchingImages,
			images,
		)
	}
	return numMatchingImages > 0, nil
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
	if shouldPublishAllPorts {
		for containerPort := range exposedPorts {
			portMap[containerPort] = []nat.PortBinding{
				// Leaving this struct empty will cause Docker to automatically choose an interface IP & port on the host machine
				{},
			}
		}
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

	// NOTE: Do NOT use PublishAllPorts here!!!! This will work if a Dockerfile doesn't have an EXPOSE directive, but
	//  if the Dockerfile *does* have an EXPOSE directive then _only_ the ports with EXPOSE will be published
	// See also: https://www.ctl.io/developers/blog/post/docker-networking-rules/
	containerHostConfigPtr := &container.HostConfig{
		Binds: bindsList,
		CapAdd: addedCapabilitiesSlice,
		NetworkMode: container.NetworkMode(networkMode),
		PortBindings: portMap,
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
			envVariables map[string]string,
			labels map[string]string) (config *container.Config, err error) {
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
		OpenStdin:    true,	// Analogous to the `-i` option to `docker run`
		Image:        dockerImage,
		ExposedPorts: portSet,
		Cmd:          cmdArgs,
		Entrypoint:   entrypointArgs,
		Env:          envVariablesSlice,
		Labels:       labels,
	}
	return nodeConfigPtr, nil
}

// Takes in a PortMap (as reported by Docker container inspect) and returns a map of the used ports -> host port binding on the expected interface
// If the given PortMap doesn't have host port bindings for all the usedPortsSet, then len(resultMap) < len(usedPortsSet)
func (manager *DockerManager) getHostPortBindingsFromDockerInspectResult(usedPortsSet map[nat.Port]bool, allInterfaceHostPortBindings nat.PortMap) map[nat.Port]*nat.PortBinding {
	result := map[nat.Port]*nat.PortBinding{}
	for port, allInterfaceBindings := range allInterfaceHostPortBindings {
		// Skip ports that aren't a part of the usedPorts set, so that the portBindings
		//  result will have a 1:1 mapping
		if _, found := usedPortsSet[port]; !found {
			manager.log.Tracef("Port '%v' isn't in used port set, so we're skipping looking for a host port binding for it", port)
			continue
		}

		foundHostPortBinding := false
		for _, interfaceBinding := range allInterfaceBindings {
			manager.log.Tracef(
				"Examining interface binding with host IP '%v' and port '%v' for port '%v'...",
				interfaceBinding.HostIP,
				interfaceBinding.HostPort,
				port,
			)
			if interfaceBinding.HostIP == expectedHostIp {
				manager.log.Tracef("Interface binding matched expected host IP '%v'; registering binding", expectedHostIp)
				result[port] = &nat.PortBinding{
					HostIP:   interfaceBinding.HostIP,
					HostPort: interfaceBinding.HostPort,
				}
				foundHostPortBinding = true
				break
			}
		}
		if !foundHostPortBinding {
			// If we're missing a host port binding, it's likely because of https://github.com/moby/moby/issues/42860
			// Eject, which will result in an incomplete candidate host port bindings map, which will
			//  retry the whole thing again in a little bit
			break
		}
	}
	return result
}

// Returns a list of Container Ids by filter arguments previously set, these containers can be running or stopped
func (manager DockerManager) getContainerIdsByFilterArgs(ctx context.Context, filterArgs filters.Args, all bool) ([]string, error) {
	opts := types.ContainerListOptions{
		Filters: filterArgs,
		All: all,
	}
	containers, err := manager.dockerClient.ContainerList(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the containers with filter args '%+v'", filterArgs)
	}
	result := []string{}
	for _, containerObj := range containers {
		result = append(result, containerObj.ID)
	}
	return result, nil
}

func (manager DockerManager) getContainersByFilterArgs(ctx context.Context, filterArgs filters.Args, all bool) ([]*Container, error) {
	opts := types.ContainerListOptions{
		Filters: filterArgs,
		All: all,
	}
	dockerContainers, err := manager.dockerClient.ContainerList(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the docker containers with filter args '%+v'", filterArgs)
	}
	result := make([]*Container, 0, len(dockerContainers))
	for _, dockerContainer := range dockerContainers {
		containerStatus := getContainerStatusByDockerContainerState(dockerContainer.State)
		containerName, err := getContainerNameByDockerContainerNames(dockerContainer.Names)
		containerHostPortBindings := getContainerHostPortBindingsByContainerPorts(dockerContainer.Ports)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting container name from docker container names '%+v'", dockerContainer.Names)
		}
		container, err := NewContainer(
			dockerContainer.ID,
			containerName,
			dockerContainer.Labels,
			containerStatus,
			containerHostPortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a new container")
		}
		result = append(result, container)
	}

	return result, nil
}

func getContainerStatusByDockerContainerState(dockerContainerState string) Status {
	allStatuses := getAllContainerStatuses()
	for _, status := range allStatuses {
		if status.string() == dockerContainerState {
			return status
		}
	}
	return unknown
}

func getContainerNameByDockerContainerNames(dockerContainerNames []string) (string, error) {
	if len(dockerContainerNames) > 0 {
		containerName := dockerContainerNames[0] //We do this because Docker Container Names is a []strings and the first value is the "actual" container's name. You can check this here: https://github.com/moby/moby/blob/master/integration-cli/docker_api_containers_test.go#L52
		containerName = strings.TrimPrefix(containerName, "/") //Docker container's names contains "/" prefix
		return containerName, nil
	}
	return "", stacktrace.NewError("There is not any docker container name to get")
}

func getContainerHostPortBindingsByContainerPorts(dockerContainerPorts []types.Port) map[nat.Port]*nat.PortBinding {
	hostPortBindings := make(map[nat.Port]*nat.PortBinding, len(dockerContainerPorts))

	for _, port := range dockerContainerPorts {
		natPort := nat.Port(strings.Join(
			[]string{
				strconv.FormatUint(uint64(port.PublicPort), 10),
				port.Type},
				"/"))

		natPortBinding := &nat.PortBinding{
			HostIP: port.IP,
			HostPort: strconv.FormatUint(uint64(port.PublicPort),10),
		}
		hostPortBindings[natPort] = natPortBinding
	}
	return hostPortBindings
}



func getLabelsFilterList(labels map[string]string) filters.Args {
	filtersArgs := []filters.KeyValuePair{}
	for labelsKey, labelsValue := range labels {
		labelFilterValue := strings.Join([]string{labelsKey,  labelsValue}, "=")
		filterArg := filters.Arg(labelFilterKey, labelFilterValue)
		filtersArgs = append(filtersArgs, filterArg)
	}
	labelsFilterList := filters.NewArgs(filtersArgs...)
	return labelsFilterList
}

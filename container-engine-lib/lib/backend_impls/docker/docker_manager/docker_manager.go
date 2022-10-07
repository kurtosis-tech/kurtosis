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
	docker_manager_types "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"net"
	"strings"
	"time"
)

const (
	// We use a bridge network because, as of 2020-08-01, we're only running locally; however, this may need to change
	//  at some point in the future
	dockerNetworkDriver = "bridge"

	// Per https://docs.docker.com/engine/reference/commandline/kill/ , this seems to mean "the default
	//  kill signal"
	dockerKillSignal = "KILL"

	expectedHostIp = "0.0.0.0"

	// When Docker binds a contianer port to the host machine, it binds it to host interface 0.0.0.0
	// Linux machines will use 127.0.0.1 for 0.0.0.0, but Windows machines don't
	// We therefore return 127.0.0.1 to the users rather than 0.0.0.0 so everybody can use them
	hostPortBindingInterfaceForUserConsumption = "127.0.0.1"

	// Character Docker uses to separate the repo from
	dockerTagSeparatorChar = ":"

	// If no tag is specified for an image, this is the tag Docker will use for the image
	dockerDefaultTag = "latest"

	// This is the magic domain name inside a container that Docker will give the host machine running Docker itself
	// This is available by default on Docker for Mac & Windows because they run in VMs, but needs to be specifically
	//  bound in Docker for Linux
	hostMachineDomainInsideContainer = "host.docker.internal"

	// hostGatewayName is the string value that Docker will replace by
	// the value of HostGatewayIP daemon config value
	hostGatewayName = "host-gateway"

	// ------------------ Filter Search Keys ----------------------
	// All these defined in https://docs.docker.com/engine/api/v1.24

	containerNameSearchFilterKey  = "name"
	containerLabelSearchFilterKey = "label"

	volumeNameSearchFilterKey  = "name"
	volumeLabelSearchFilterKey = "label"

	networkNameSearchFilterKey  = "name"
	networkIdSearchFilterKey    = "id"
	networkLabelSearchFilterKey = "label"
	// ---------------- End Filter Search Keys ----------------------

	// For some reason, when publish-all-ports is requested, Docker will return successfully from starting a
	//  container, but without having bound the host ports
	// See: https://github.com/moby/moby/issues/42860
	// To work around this, we retry a few times
	timeBetweenHostPortBindingChecks = 500 * time.Millisecond
	maxNumHostPortBindingChecks      = 4

	// Not sure why we'd ever want 'force' set to false when removing volumes & containers
	shouldForceVolumeRemoval = true

	shouldRemoveAnonymousVolumesWhenRemovingContainers = true
	shouldRemoveLinksWhenRemovingContainers            = false // We don't use container links
	shouldKillContainersWhenRemovingContainers         = true

	shouldFollowContainerLogsWhenGettingFailedContainerLogs = false

	shouldAttachStdinWhenCreatingContainerExec                = true
	shouldAttachStandardStreamsToTtyWhenCreatingContainerExec = true
	shouldAttachStderrWhenCreatingContainerExec               = true
	shouldAttachStdoutWhenCreatingContainerExec               = true
	shouldExecuteInDetachModeWhenCreatingContainerExec        = false

	megabytesToBytesFactor    = 1_000_000
	millicpusToNanoCPUsFactor = 1_000_000

	minMemoryLimit = 6
)

/*
InteractiveModeTtySize
The dimensions of the TTY that the container should output to when in interactive mode
*/
type InteractiveModeTtySize struct {
	Height uint
	Width  uint
}

/*
DockerManager
A handle to interacting with the Docker environment running a test.
*/
type DockerManager struct {
	// The underlying Docker client that will be used to modify the Docker environment
	dockerClient *client.Client
}

/*
NewDockerManager
Creates a new Docker manager for manipulating the Docker engine using the given client.

Args:

	dockerClient: The Docker client that will be used when interacting with the underlying Docker engine the Docker engine.
*/
func NewDockerManager(dockerClient *client.Client) *DockerManager {
	return &DockerManager{
		dockerClient: dockerClient,
	}
}

/*
CreateNetwork
Creates a new Docker network with the given parameters; does nothing if a network with the given name already exists.

Args:

	context: The Context that this request is running in (useful for cancellation)
	name: The name to give the new Docker network
	subnetMask: The subnet mask defining allowed IPs for the Docker network
	gatewayIP: The IP to give the network gateway
	labels: Labels to give the network object

Returns:

	id: The Docker-managed ID of the network
*/
func (manager DockerManager) CreateNetwork(context context.Context, name string, subnetMask string, gatewayIP net.IP, labels map[string]string) (id string, err error) {
	ipamConfig := []network.IPAMConfig{{
		Subnet:  subnetMask,
		Gateway: gatewayIP.String(),
	}}

	resp, err := manager.dockerClient.NetworkCreate(context, name, types.NetworkCreate{
		Driver: dockerNetworkDriver,
		IPAM: &network.IPAM{
			Config: ipamConfig,
		},
		Labels: labels,
	})
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create network %s with subnet %s", name, subnetMask)
	}
	return resp.ID, nil
}

func (manager DockerManager) ListNetworks(ctx context.Context) ([]types.NetworkResource, error) {
	networks, err := manager.dockerClient.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred listing the Docker networks")
	}
	// The Network objects that come back ostensibly, *should*  have the Containers field filled out... but they don't
	// If we ever need that field, we have to call an InspectNetwork, and even then it seems to have some amount of
	// nondeterminism (i.e. brand-new containers won't show up)
	return networks, nil
}

/*
GetNetworksByName
Returns Network list matching the given name (if any).
*/
// TODO Combine with GetNetworksByLabel using a search filter builder
func (manager DockerManager) GetNetworksByName(ctx context.Context, name string) ([]*docker_manager_types.Network, error) {
	nameSearchArgs := filters.NewArgs(filters.KeyValuePair{
		Key:   networkNameSearchFilterKey,
		Value: name,
	})
	dockerNetworks, err := manager.getNetworksByFilterArgs(ctx, nameSearchArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for existence of network with name %v", name)
	}

	networks, err := newNetworkListFromDockerNetworkList(dockerNetworks)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new network list from Docker network list '%+v'", dockerNetworks)
	}

	return networks, nil
}

/*
GetNetworksByLabels
Gets networks matching the given labels
*/
func (manager DockerManager) GetNetworksByLabels(ctx context.Context, labels map[string]string) ([]*docker_manager_types.Network, error) {
	labelsSearchArgs := getLabelsFilterArgs(networkLabelSearchFilterKey, labels)
	dockerNetworks, err := manager.getNetworksByFilterArgs(ctx, labelsSearchArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for existence of network with labels '%+v'", labels)
	}

	networks, err := newNetworkListFromDockerNetworkList(dockerNetworks)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new network list from Docker network list '%+v'", dockerNetworks)
	}

	return networks, nil
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
RemoveNetwork
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
CreateVolume
Creates a Docker volume identified by the given name.

Args:

	context: The Context that this request is running in (useful for cancellation)
	volumeName: The unique identifier used by Docker to identify this volume (NOTE: at time of writing, Docker doesn't
		even give volumes IDs - this name is all there is)
	labels: Labels to attach to the volume object
*/
func (manager DockerManager) CreateVolume(context context.Context, volumeName string, labels map[string]string) error {
	volumeConfig := volume.VolumeCreateBody{
		Name:   volumeName,
		Labels: labels,
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
GetVolumesByName
Searches for volumes whose names match the given one

Args:

	context: The Context that this request is running in (useful for cancellation)
	volumeName: The unique identifier used by Docker to identify this volume (NOTE: at time of writing, Docker doesn't
		even give volumes IDs - this name is all there is)

Returns: A list of names of volumes matching the search term
*/
func (manager *DockerManager) GetVolumesByName(ctx context.Context, volumeName string) ([]string, error) {
	nameFilter := filters.KeyValuePair{
		Key:   volumeNameSearchFilterKey,
		Value: volumeName,
	}
	filterArgs := filters.NewArgs(nameFilter)
	resp, err := manager.dockerClient.VolumeList(ctx, filterArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finding volumes with name matching '%v'", volumeName)
	}

	respNames := []string{}
	for _, foundVolume := range resp.Volumes {
		respNames = append(respNames, foundVolume.Name)
	}
	return respNames, nil
}

/*
GetVolumesByLabels
Gets the volumes matching the given labels
*/
func (manager *DockerManager) GetVolumesByLabels(ctx context.Context, labels map[string]string) ([]*types.Volume, error) {
	labelsFilterArgs := getLabelsFilterArgs(volumeLabelSearchFilterKey, labels)
	resp, err := manager.dockerClient.VolumeList(ctx, labelsFilterArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finding volumes with labels '%+v'", labels)
	}

	result := []*types.Volume{}
	if resp.Volumes != nil {
		result = resp.Volumes
	}

	return result, nil
}

/*
RemoveVolume
Removes a Docker volume identified by the given name, deleting it permanently

Args:

	context: The Context that this request is running in (useful for cancellation)
	volumeName: The unique identifier used by Docker to identify the volume that will get removed
*/
func (manager *DockerManager) RemoveVolume(ctx context.Context, volumeName string) error {
	if err := manager.dockerClient.VolumeRemove(ctx, volumeName, shouldForceVolumeRemoval); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing volume '%v'", volumeName)
	}
	return nil
}

func (manager *DockerManager) InspectContainer(ctx context.Context, containerId string) (types.ContainerJSON, error) {
	result, err := manager.dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		return types.ContainerJSON{}, stacktrace.Propagate(err, "An error occurred inspecting container '%v'", containerId)
	}
	return result, nil
}

/*
CreateAndStartContainer
Creates a Docker container with the given args and starts it.

Returns:

	containerId: The Docker container ID of the newly-created container
	hostMachinePortBindings: For every port in the args' "usedPorts" object that has publishing turned on, an entry
		will be generated in this map with the binding on the host machine where the port can be found
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

	err := manager.FetchImage(ctx, dockerImage)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred fetching image '%v'", dockerImage)
	}

	idFilterArgs := filters.NewArgs(filters.KeyValuePair{
		Key:   networkIdSearchFilterKey,
		Value: args.networkId,
	})
	networks, err := manager.getNetworksByFilterArgs(ctx, idFilterArgs)
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

	usedPortsSet := nat.PortSet{}
	for port := range args.usedPorts {
		usedPortsSet[port] = struct{}{}
	}

	containerConfigPtr, err := manager.getContainerCfg(
		dockerImage,
		isInteractiveMode,
		usedPortsSet,
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
		args.usedPorts,
		args.needsAccessToDockerHostMachine,
		args.cpuAllocationMillicpus,
		args.memoryAllocationMegabytes,
		args.loggingDriverConfig)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Failed to configure host to container mappings from service.")
	}
	containerCreateResp, err := manager.dockerClient.ContainerCreate(ctx, containerConfigPtr, containerHostConfigPtr, nil, nil, args.name)
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
	logrus.Debugf("Created container with ID '%v' from image '%v'", containerId, dockerImage)

	// If the user doesn't provide an IP, the Docker network will auto-assign one
	if args.staticIp != nil {
		if err := manager.ConnectContainerToNetwork(ctx, args.networkId, containerId, args.staticIp, args.alias); err != nil {
			return "", nil, stacktrace.Propagate(err, "Failed to connect container %s to network.", containerId)
		}
	}
	// TODO defer a disconnct-from-network if this function doesn't succeed??

	if err = manager.dockerClient.ContainerStart(ctx, containerId, types.ContainerStartOptions{}); err != nil {
		containerLogs := manager.getFailedContainerLogsOrErrorString(ctx, containerId)
		containerLogsHeader := "\n--------------------- CONTAINER LOGS -----------------------\n"
		containerLogsFooter := "\n------------------- END CONTAINER LOGS --------------------"
		return "", nil, stacktrace.Propagate(err, "Could not start Docker container from image '%v'; logs are below:%v%v%v", dockerImage, containerLogsHeader, containerLogs, containerLogsFooter)
	}

	functionFinishedSuccessfully := false
	defer func() {
		if !functionFinishedSuccessfully {
			if err := manager.KillContainer(ctx, containerId); err != nil {
				logrus.Error("The container creation function didn't finish successfully, meaning we needed to kill the container we created. However, the killing threw an error:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION NEEDED: You'll need to manually kill this container with ID '%v'", containerId)
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

	publishedPortsSet := map[nat.Port]bool{}
	for containerPort, publishSpec := range args.usedPorts {
		if publishSpec.mustBeFoundAfterContainerStart() {
			publishedPortsSet[containerPort] = true
		}
	}
	logrus.Tracef("Published ports set: %+v", publishedPortsSet)

	// If the user wanted their ports exposed, Docker will have auto-assigned the ports to ports in the ephemeral range
	//  on the host. We need to look up what those ports are so we can return report them back to the user.
	resultHostPortBindings := map[nat.Port]*nat.PortBinding{}
	numPublishedPorts := len(publishedPortsSet)
	if numPublishedPorts > 0 {
		// Thanks to https://github.com/moby/moby/issues/42860, we have to retry several times to get the host port bindings
		//  from Docker
		for i := 0; i < maxNumHostPortBindingChecks; i++ {
			logrus.Tracef("Trying to get host machine port bindings (%v previous attempts)...", i)
			containerInspectResp, err := manager.dockerClient.ContainerInspect(ctx, containerId)
			if err != nil {
				return "", nil, stacktrace.Propagate(
					err,
					"%v ports were published to the host machine, but an error occurred inspecting the newly-started "+
						"container which is necessary for determining which host machine ports the container's ports were bound to",
					numPublishedPorts,
				)
			}
			logrus.Tracef("Container inspect response: %+v", containerInspectResp)
			networkSettings := containerInspectResp.NetworkSettings
			if networkSettings == nil {
				return "", nil, stacktrace.NewError(
					"We got a response from inspecting container '%v' which is necessary for determining the "+
						"ports published to the host machine, but the network settings object was nil",
					containerId,
				)
			}
			logrus.Tracef("Network settings: %+v", networkSettings)
			allInterfaceHostPortBindings := networkSettings.Ports
			if allInterfaceHostPortBindings == nil {
				return "", nil, stacktrace.NewError(
					"%v ports on container '%v' were to be published to the host machine, but the container host port bindings were null",
					numPublishedPorts,
					containerId,
				)
			}
			logrus.Tracef("Network settings -> ports: %+v", allInterfaceHostPortBindings)

			allHostPortBindingsOnExpectedInterface := getHostPortBindingsOnExpectedInterface(allInterfaceHostPortBindings)

			// Filter to the ports matching the ports we wanted to publish
			usedHostPortBindingsOnExpectedInterface := map[nat.Port]*nat.PortBinding{}
			for port, hostPortBinding := range allHostPortBindingsOnExpectedInterface {
				if _, found := usedPortsSet[port]; !found {
					logrus.Tracef("Port '%v' isn't in used port set, so we're skipping its host port binding", port)
					continue
				}

				usedHostPortBindingsOnExpectedInterface[port] = hostPortBinding
			}

			// If we're missing a host port binding, it's likely because of https://github.com/moby/moby/issues/42860
			// We'll retry after a sleep
			if len(usedHostPortBindingsOnExpectedInterface) == numPublishedPorts {
				resultHostPortBindings = usedHostPortBindingsOnExpectedInterface
				break
			}
			time.Sleep(timeBetweenHostPortBindingChecks)
		}

		// Final verification that all published ports get a host machine port bindings
		if len(resultHostPortBindings) != numPublishedPorts {
			return "", nil, stacktrace.NewError(
				"%v ports were to be published to the host machine, but container '%v' never got host machine port bindings on interface %v for all published ports even after %v checks with %v between checks",
				numPublishedPorts,
				containerId,
				expectedHostIp,
				maxNumHostPortBindingChecks,
				timeBetweenHostPortBindingChecks,
			)
		}
	}

	functionFinishedSuccessfully = true
	return containerId, resultHostPortBindings, nil
}

/*
GetContainerIP
Gets the container's IP on a given network
NOTE: Yes, it's a testament to how poorly-designed the Docker API is that we need to use network name here even though

	everywhere else in the Docker API uses network ID
*/
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
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	}
	hijackedResponse, err := manager.dockerClient.ContainerAttach(ctx, containerId, attachOpts)
	if err != nil {
		return types.HijackedResponse{}, stacktrace.Propagate(err, "An error occurred attaching to container '%v'", containerId)
	}
	return hijackedResponse, nil
}

/*
StopContainer
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
KillContainer
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
RemoveContainer
Removes the container with the given ID, deleting it permanently

Args:

	context: The context that the removal runs in
	containerId: ID of Docker container to remove
*/
func (manager *DockerManager) RemoveContainer(ctx context.Context, containerId string) error {
	removeOpts := types.ContainerRemoveOptions{
		RemoveVolumes: shouldRemoveAnonymousVolumesWhenRemovingContainers,
		RemoveLinks:   shouldRemoveLinksWhenRemovingContainers,
		Force:         shouldKillContainersWhenRemovingContainers,
	}
	if err := manager.dockerClient.ContainerRemove(ctx, containerId, removeOpts); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing container with ID '%v'", containerId)
	}
	return nil
}

/*
WaitForExit
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
GetContainerLogs gets the logs for the given container as a io.ReadCloser. The caller is responsible for closing the ReadCloser!!!

NOTE: These logs have STDOUT and STDERR multiplexed together, and the 'stdcopy' package needs to be used to

	demultiplex them per https://github.com/moby/moby/issues/32794
*/
func (manager DockerManager) GetContainerLogs(
	ctx context.Context,
	containerId string,
	shouldFollowLogs bool,
) (io.ReadCloser, error) {
	containerLogOpts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     shouldFollowLogs,
	}
	readCloser, err := manager.dockerClient.ContainerLogs(ctx, containerId, containerLogOpts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs for container ID '%v'", containerId)
	}
	return readCloser, nil
}

/*
PauseContainer
Pauses all processes running in the given container, but does not shut it down.
*/
func (manager DockerManager) PauseContainer(context context.Context, containerId string) error {
	err := manager.dockerClient.ContainerPause(context, containerId)
	if err != nil {
		return stacktrace.Propagate(err, "Docker client failed to pause container '%v'", containerId)
	}
	return nil
}

/*
UnpauseContainer
Unpauses all processes running in the given container.
*/
func (manager DockerManager) UnpauseContainer(context context.Context, containerId string) error {
	err := manager.dockerClient.ContainerUnpause(context, containerId)
	if err != nil {
		return stacktrace.Propagate(err, "Docker client failed to unpause container '%v'", containerId)
	}
	return nil
}

/*
RunExecCommand
Executes the given command inside the container with the given ID, blocking until the command completes
*/
func (manager DockerManager) RunExecCommand(context context.Context, containerId string, command []string, logOutput io.Writer) (int32, error) {
	dockerClient := manager.dockerClient
	execConfig := types.ExecConfig{
		Cmd:          command,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
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
	concurrentWriter := concurrent_writer.NewConcurrentWriter(logOutput)
	if _, err := stdcopy.StdCopy(concurrentWriter, concurrentWriter, attachResp.Reader); err != nil {
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
ConnectContainerToNetwork
Connects the container with the given container ID to the network with the given network ID, using the given IP address
*/
func (manager DockerManager) ConnectContainerToNetwork(ctx context.Context, networkId string, containerId string, staticIpAddr net.IP, alias string) (err error) {
	logrus.Tracef(
		"Connecting container ID %v to network ID %v using static IP %v",
		containerId,
		networkId,
		staticIpAddr.String())

	ipamConfig := &network.EndpointIPAMConfig{
		IPv4Address: staticIpAddr.String(),
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

// TODO Refactor this to be GetContainersByName - no need to be so specific now that we have the Container type we can return
func (manager DockerManager) GetContainerIdsByName(ctx context.Context, nameStr string) ([]string, error) {
	filterArg := filters.Arg(containerNameSearchFilterKey, nameStr)
	nameFilterList := filters.NewArgs(filterArg)
	// TODO Make the "should show stopped containers" flag configurable????
	matchingContainers, err := manager.getContainersByFilterArgs(ctx, nameFilterList, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the containers with name '%v'", nameStr)
	}
	result := []string{}
	for _, containerObj := range matchingContainers {
		result = append(result, containerObj.GetId())
	}
	return result, nil
}

func (manager DockerManager) GetContainersByLabels(ctx context.Context, labels map[string]string, shouldShowStoppedContainers bool) ([]*docker_manager_types.Container, error) {
	labelsFilterList := getLabelsFilterArgs(containerLabelSearchFilterKey, labels)
	result, err := manager.getContainersByFilterArgs(ctx, labelsFilterList, shouldShowStoppedContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting containers with labels '%+v'", labelsFilterList)
	}
	return result, nil
}

func (manager DockerManager) FetchImage(ctx context.Context, dockerImage string) error {
	logrus.Tracef("Checking if image '%v' is available locally...", dockerImage)
	imageExistsLocally, err := manager.isImageAvailableLocally(ctx, dockerImage)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking for local availability of Docker image '%v'", dockerImage)
	}
	logrus.Tracef("Is image available locally?: %v", imageExistsLocally)

	if !imageExistsLocally {
		logrus.Tracef("Image doesn't exist locally, so attempting to pull it...")
		err = manager.PullImage(ctx, dockerImage)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to pull Docker image '%v' from remote image repository", dockerImage)
		}
		logrus.Tracef("Image successfully pulled from remote to local")
	}

	return nil
}

func (manager DockerManager) PullImage(context context.Context, imageName string) (err error) {
	logrus.Infof("Pulling image '%s'...", imageName)
	out, err := manager.dockerClient.ImagePull(context, imageName, types.ImagePullOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to pull image %s", imageName)
	}
	defer out.Close()
	_, err = io.Copy(ioutil.Discard, out)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred discarding the output")
	}
	return nil
}

func (manager DockerManager) CreateContainerExec(context context.Context, containerId string, cmd []string) (*types.HijackedResponse, error) {
	config := types.ExecConfig{
		AttachStdin:  shouldAttachStdinWhenCreatingContainerExec,
		Tty:          shouldAttachStandardStreamsToTtyWhenCreatingContainerExec,
		AttachStderr: shouldAttachStderrWhenCreatingContainerExec,
		AttachStdout: shouldAttachStdoutWhenCreatingContainerExec,
		Detach:       shouldExecuteInDetachModeWhenCreatingContainerExec,
		Cmd:          cmd,
	}

	response, err := manager.dockerClient.ContainerExecCreate(context, containerId, config)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while creating the ContainerExec in container with ID '%v'", containerId)
	}

	execID := response.ID
	if execID == "" {
		return nil, stacktrace.NewError("the Exec ID was empty")
	}

	execStartCheck := types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	}

	hijackedResponse, err := manager.dockerClient.ContainerExecAttach(context, execID, execStartCheck)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was an error while attaching connection to the execution process with ID '%v' in container with ID '%v'", execID, containerId)
	}

	return &hijackedResponse, nil
}

// CopyFromContainer returns a io.ReadCloser representing the bytes of the TAR'd files at srcPath
// The caller must close the result
func (manager DockerManager) CopyFromContainer(ctx context.Context, containerId string, srcPath string) (io.ReadCloser, error) {

	tarStreamReadCloser, _, err := manager.dockerClient.CopyFromContainer(
		ctx,
		containerId,
		srcPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred copying content '%v' from container with ID '%v'", srcPath, containerId)
	}

	return tarStreamReadCloser, nil
}

// =================================================================================================================
//
//	INSTANCE HELPER FUNCTIONS
//
// =================================================================================================================
func (manager DockerManager) isImageAvailableLocally(ctx context.Context, imageName string) (isAvailable bool, err error) {
	referenceArg := filters.Arg("reference", imageName)
	filters := filters.NewArgs(referenceArg)
	images, err := manager.dockerClient.ImageList(
		ctx,
		types.ImageListOptions{
			All:     true,
			Filters: filters,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to list images.")
	}
	numMatchingImages := len(images)
	if numMatchingImages > 1 {
		return false, stacktrace.NewError(
			"Searching for Docker images matching image name '%v' returned %v images; "+
				"this indicates a bug because the image name being searched should only reference 0 or 1 images. Images matched:\n%+v",
			imageName,
			numMatchingImages,
			images,
		)
	}
	return numMatchingImages > 0, nil
}

func (manager *DockerManager) getNetworksByFilterArgs(ctx context.Context, args filters.Args) ([]types.NetworkResource, error) {
	// NOTE: Even though this returns a `NetworkResource` object which has a Containers field on it, this is a lie!!
	// For whatever insane reason, Docker doesn't fill this field out when NetworkList is used and there doesn't seem to
	// be a way to get it to do so. Instead we'd have to do an InspectNetwork call.
	networks, err := manager.dockerClient.NetworkList(
		ctx,
		types.NetworkListOptions{
			Filters: args,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list networks while doing a filter search using args '%+v'", args)
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
	usedPortsWithPublishSpec: Ports that are used by the container, with a specification for how they should be published to the
		host machine (if at all)
	needsToAccessDockerHostMachine: If true, adds a "host.docker.internal:host-gateway" extra host binding, which is necessary
		for machines that will need to access the machine hosting Docker itself.
*/
func (manager *DockerManager) getContainerHostConfig(
	addedCapabilities map[ContainerCapability]bool,
	networkMode DockerManagerNetworkMode,
	bindMounts map[string]string,
	volumeMounts map[string]string,
	usedPortsWithPublishSpec map[nat.Port]PortPublishSpec,
	needsToAccessDockerHostMachine bool,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
	loggingDriverConfig LoggingDriver,
) (hostConfig *container.HostConfig, err error) {

	bindsList := make([]string, 0, len(bindMounts))
	for hostFilepath, containerFilepath := range bindMounts {
		bindsList = append(bindsList, hostFilepath+":"+containerFilepath)
	}
	for volumeName, containerFilepath := range volumeMounts {
		// Yes, it's SUPER confusing that "volumes" need to be put into the "binds" section because there's
		//  a separate thing called a "bind mount".... blame the Docker API
		bindsList = append(bindsList, volumeName+":"+containerFilepath)
	}

	logrus.Debugf("Binds: %v", bindsList)

	portMap := nat.PortMap{}
	for containerPort, publishSpec := range usedPortsWithPublishSpec {
		publishSpecType := publishSpec.getType()
		switch publishSpecType {
		case noPublishing:
			continue
		case automaticPublishing:
			portMap[containerPort] = []nat.PortBinding{
				// Leaving this struct empty will cause Docker to automatically choose an interface IP & port on the host machine
				{},
			}
		case manualPublishing:
			manualSpec, ok := publishSpec.(*manuallySpecifiedPortPublishSpec)
			if !ok {
				return nil, stacktrace.NewError(
					"The port publish spec had type '%v', but downcasting it failed; this is very strange!",
					publishSpecType,
				)
			}
			hostMachinePortNumStr := fmt.Sprintf("%v", manualSpec.getHostMachinePortNum())
			portMap[containerPort] = []nat.PortBinding{
				{
					HostIP:   expectedHostIp,
					HostPort: hostMachinePortNumStr,
				},
			}
		default:
			return nil, stacktrace.NewError("Unrecognized port publish spec type '%v'; this is a bug in this library", publishSpecType)
		}
	}

	addedCapabilitiesSlice := []string{}
	for capability := range addedCapabilities {
		capabilityStr := string(capability)
		addedCapabilitiesSlice = append(addedCapabilitiesSlice, capabilityStr)
	}

	extraHosts := []string{}
	if needsToAccessDockerHostMachine {
		// This explicit specification is necessary because in Docker-for-Linux, the magic "host.docker.internal"
		//  domain name isn't automatically available inside a container
		extraHosts = append(
			extraHosts,
			fmt.Sprintf("%v:%v", hostMachineDomainInsideContainer, hostGatewayName),
		)
	}

	resources := container.Resources{}
	if cpuAllocationMillicpus != 0 {
		nanoCPUs := convertMillicpusToNanoCPUs(cpuAllocationMillicpus)
		resources.NanoCPUs = int64(nanoCPUs)
	}
	if memoryAllocationMegabytes != 0 {
		if memoryAllocationMegabytes < minMemoryLimit {
			return nil, stacktrace.NewError("Memory allocation, `%d`, is too low. Docker requires the memory limit to be at least `%d` megabytes.", memoryAllocationMegabytes, minMemoryLimit)
		}
		memoryAllocationBytes := convertMegabytesToBytes(memoryAllocationMegabytes)
		resources.Memory = int64(memoryAllocationBytes)

		// MemorySwap needs to be set to exactly memory to ensure memory is actually limited to memoryAllocationInBytes
		// https://faun.pub/understanding-docker-container-memory-limit-behavior-41add155236c
		resources.MemorySwap = int64(memoryAllocationBytes)
	}

	logConfig := container.LogConfig{}
	if loggingDriverConfig != nil {
		logConfig = loggingDriverConfig.GetLogConfig()
	}

	// NOTE: Do NOT use PublishAllPorts here!!!! This will work if a Dockerfile doesn't have an EXPOSE directive, but
	//  if the Dockerfile *does* have an EXPOSE directive then _only_ the ports with EXPOSE will be published
	// See also: https://www.ctl.io/developers/blog/post/docker-networking-rules/
	containerHostConfigPtr := &container.HostConfig{
		Binds:        bindsList,
		CapAdd:       addedCapabilitiesSlice,
		NetworkMode:  container.NetworkMode(networkMode),
		PortBindings: portMap,
		ExtraHosts:   extraHosts,
		Resources:    resources,
		LogConfig:    logConfig,
	}
	return containerHostConfigPtr, nil
}

// Creates a Docker container representing a service that will listen on ports in the network
func (manager *DockerManager) getContainerCfg(
	dockerImage string,
	isInteractiveMode bool,
	usedPorts nat.PortSet,
	entrypointArgs []string,
	cmdArgs []string,
	envVariables map[string]string,
	labels map[string]string) (config *container.Config, err error) {

	envVariablesSlice := make([]string, 0, len(envVariables))
	for key, val := range envVariables {
		envVariablesSlice = append(envVariablesSlice, fmt.Sprintf("%v=%v", key, val))
	}

	nodeConfigPtr := &container.Config{
		AttachStderr: isInteractiveMode, // Analogous to `-a STDERR` option to `docker run`
		AttachStdin:  isInteractiveMode, // Analogous to `-a STDIN` option to `docker run`
		AttachStdout: isInteractiveMode, // Analogous to `-a STDOUT` option to `docker run`
		Tty:          isInteractiveMode, // Analogous to the `-t` option to `docker run`
		OpenStdin:    true,              // Analogous to the `-i` option to `docker run`
		Image:        dockerImage,
		ExposedPorts: usedPorts,
		Cmd:          cmdArgs,
		Entrypoint:   entrypointArgs,
		Env:          envVariablesSlice,
		Labels:       labels,
	}
	return nodeConfigPtr, nil
}

// Takes in a PortMap (as reported by Docker container inspect) and returns a map of the used ports -> host port binding on the expected interface
// If no bindings for the interface are found, len(output) < len(input)
func getHostPortBindingsOnExpectedInterface(hostPortBindingsOnAllInterfaces nat.PortMap) map[nat.Port]*nat.PortBinding {
	result := map[nat.Port]*nat.PortBinding{}
	for port, allInterfaceBindings := range hostPortBindingsOnAllInterfaces {
		for _, interfaceBinding := range allInterfaceBindings {
			logrus.Tracef(
				"Examining interface binding with host IP '%v' and port '%v' for port '%v'...",
				interfaceBinding.HostIP,
				interfaceBinding.HostPort,
				port,
			)
			if interfaceBinding.HostIP == expectedHostIp {
				logrus.Tracef("Interface binding matched expected host IP '%v'; registering binding", expectedHostIp)
				result[port] = &nat.PortBinding{
					HostIP:   hostPortBindingInterfaceForUserConsumption,
					HostPort: interfaceBinding.HostPort,
				}
				// Finding multiple public ports for the same interface would be silly, and unnecessary, so we break here
				break
			}
		}
	}
	return result
}

func (manager DockerManager) getContainersByFilterArgs(ctx context.Context, filterArgs filters.Args, shouldShowStoppedContainers bool) ([]*docker_manager_types.Container, error) {
	opts := types.ContainerListOptions{
		Filters: filterArgs,
		All:     shouldShowStoppedContainers,
	}
	dockerContainers, err := manager.dockerClient.ContainerList(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the docker containers with filter args '%+v'", filterArgs)
	}
	containers, err := newContainersListFromDockerContainersList(dockerContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new containers list from Docker containers list")
	}

	return containers, nil
}

func newContainersListFromDockerContainersList(dockerContainers []types.Container) ([]*docker_manager_types.Container, error) {
	containers := make([]*docker_manager_types.Container, 0, len(dockerContainers))
	for _, dockerContainer := range dockerContainers {
		container, err := newContainerFromDockerContainer(dockerContainer)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating new container from Docker container '%+v'", dockerContainer)
		}
		containers = append(containers, container)
	}
	return containers, nil
}

func newContainerFromDockerContainer(dockerContainer types.Container) (*docker_manager_types.Container, error) {
	containerStatus, err := getContainerStatusByDockerContainerState(dockerContainer.State)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting ContainerStatus from Docker container state '%v'", dockerContainer.State)
	}
	containerName, err := getContainerNameByDockerContainerNames(dockerContainer.Names)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting container name from Docker container names '%+v'", dockerContainer.Names)
	}

	// Frustratingly, Docker's ContainerList returns ports in a completely different format than ContainerInspect, so we need
	// to process the ports into the same format as ContainerInspect so we can call getHostPortBindingsOnExpectedInterface
	portMap := nat.PortMap{}
	for _, port := range dockerContainer.Ports {
		// It's kinda bad that we use "forbidden knowledge" about how nat.Port represents its internals to create one, but
		// the nat.Port API is so infuriatingly awful to use (how do you even create one??)
		privatePortStr := fmt.Sprintf("%v/%v", port.PrivatePort, port.Type)
		privatePort := nat.Port(privatePortStr)

		bindingsForPort, found := portMap[privatePort]
		if !found {
			bindingsForPort = []nat.PortBinding{}
		}

		hostBinding := nat.PortBinding{
			HostIP:   port.IP,
			HostPort: fmt.Sprintf("%v", port.PublicPort),
		}

		bindingsForPort = append(bindingsForPort, hostBinding)
		portMap[privatePort] = bindingsForPort
	}
	containerHostPortBindings := getHostPortBindingsOnExpectedInterface(portMap)

	newContainer := docker_manager_types.NewContainer(
		dockerContainer.ID,
		containerName,
		dockerContainer.Labels,
		containerStatus,
		containerHostPortBindings,
	)

	return newContainer, nil
}

func getContainerStatusByDockerContainerState(dockerContainerState string) (docker_manager_types.ContainerStatus, error) {
	containerStatus, err := docker_manager_types.ContainerStatusString(dockerContainerState)
	if err != nil {
		return 0, stacktrace.NewError("No container status matches Docker container state '%v'; this is a bug in Kurtosis", dockerContainerState)
	}

	return containerStatus, nil
}

func getContainerNameByDockerContainerNames(dockerContainerNames []string) (string, error) {
	if len(dockerContainerNames) > 0 {
		containerName := dockerContainerNames[0]               //We do this because Docker Container Names is a []strings and the first value is the "actual" container's name. You can check this here: https://github.com/moby/moby/blob/master/integration-cli/docker_api_containers_test.go#L52
		containerName = strings.TrimPrefix(containerName, "/") //Docker container's names contains "/" prefix
		return containerName, nil
	}
	return "", stacktrace.NewError("There is not any Docker container name to get")
}

func getLabelsFilterArgs(searchFilterKey string, labels map[string]string) filters.Args {
	filtersArgs := []filters.KeyValuePair{}
	for labelsKey, labelsValue := range labels {
		labelFilterValue := strings.Join([]string{labelsKey, labelsValue}, "=")
		filterArg := filters.Arg(searchFilterKey, labelFilterValue)
		filtersArgs = append(filtersArgs, filterArg)
	}
	labelsFilterList := filters.NewArgs(filtersArgs...)
	return labelsFilterList
}

func newNetworkListFromDockerNetworkList(dockerNetworks []types.NetworkResource) ([]*docker_manager_types.Network, error) {
	networks := []*docker_manager_types.Network{}

	for _, dockerNetwork := range dockerNetworks {
		dockerManagerNetwork, err := newNetworkFromDockerNetwork(dockerNetwork)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating new network from Docker network with ID '%v'", dockerNetwork.ID)
		}
		networks = append(networks, dockerManagerNetwork)
	}

	return networks, nil
}

func newNetworkFromDockerNetwork(dockerNetwork types.NetworkResource) (*docker_manager_types.Network, error) {
	if len(dockerNetwork.IPAM.Config) == 0 {
		return nil, stacktrace.NewError("Kurtosis Docker network with ID %v does not contains any IPAM config.", dockerNetwork.ID)
	}
	if len(dockerNetwork.IPAM.Config) > 1 {
		return nil, stacktrace.NewError("This is an unexpected error Docker network with ID '%v' shouldn't have more than one IPAM config; this is a bug in Kurtosis itself", dockerNetwork.ID)
	}
	firstIpamConfig := dockerNetwork.IPAM.Config[0]

	_, ipAndMask, err := net.ParseCIDR(firstIpamConfig.Subnet)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing CIDR '%v'", firstIpamConfig.Subnet)
	}

	gatewayIp := firstIpamConfig.Gateway

	networkWrapper := docker_manager_types.NewNetwork(
		dockerNetwork.Name,
		dockerNetwork.ID,
		ipAndMask,
		gatewayIp,
		dockerNetwork.Labels,
	)

	return networkWrapper, nil
}

func (manager DockerManager) getFailedContainerLogsOrErrorString(ctx context.Context, containerId string) string {

	var containerLogs string

	containerLogsReadCloser, err := manager.GetContainerLogs(ctx, containerId, shouldFollowContainerLogsWhenGettingFailedContainerLogs)
	if err != nil {
		return fmt.Sprintf("An error occurred getting logs for container with ID '%v' error:\n%v", containerId, err)
	}
	defer func() {
		if err := containerLogsReadCloser.Close(); err != nil {
			logrus.Warningf("We tried to close the container logs read-closer, but doing so threw an error:\n%v", err)
		}
	}()

	containerLogsBufferSize := 10
	containerLogsBuffer := make([]byte, containerLogsBufferSize)

	for {
		numberOfBytesRead, err := containerLogsReadCloser.Read(containerLogsBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Sprintf("An error occurred reading logs for container with ID '%v' error:\n%v", containerId, err)
		}
		if numberOfBytesRead > 0 {
			newString := string(containerLogsBuffer[:numberOfBytesRead])
			containerLogs = containerLogs + newString
		}
	}
	return containerLogs
}

func convertMegabytesToBytes(value uint64) uint64 {
	return value * megabytesToBytesFactor
}

// Intakes millicpu unit and converts to NanoCPUs in Docker
// In Docker 1 CPU = 1000 millicpus = 1000000000 NanoCPUs
func convertMillicpusToNanoCPUs(value uint64) uint64 {
	return value * millicpusToNanoCPUsFactor
}

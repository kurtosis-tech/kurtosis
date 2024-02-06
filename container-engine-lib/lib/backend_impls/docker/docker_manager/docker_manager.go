/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package docker_manager

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/go-units"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/utils"
	"io"
	"math"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	kurtosis_sdk_version "github.com/kurtosis-tech/kurtosis/api/golang/kurtosis_version"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	docker_manager_types "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/compute_resources"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/concurrent_writer"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"

	bksession "github.com/moby/buildkit/session"
)

const (
	dockerClientTimeout = 30 * time.Second
	// We use a bridge network because, as of 2020-08-01, we're only running locally; however, this may need to change
	//  at some point in the future
	dockerNetworkDriver = "bridge"

	// Per https://docs.docker.com/engine/reference/commandline/kill/ , this seems to mean "the default
	//  kill signal"
	dockerKillSignal = "KILL"

	expectedHostIp = "0.0.0.0"

	// When Docker binds a container port to the host machine, it binds it to host interface 0.0.0.0
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

	containerNameSearchFilterKey      = "name"
	containerLabelSearchFilterKey     = "label"
	containerNetworkIdSearchFilterKey = "network"

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

	zombieProcessesCannotRemoveContainerErrMsg = "is zombie and can not be killed"
	defaultRemoveProcessesMaxRetries           = uint8(3)
	defaultRemoveContainerTimeBetweenRetries   = 10 * time.Second

	containerIsNotRunningErrMsg            = "is not running"
	cannotKillContainerErrMsg              = "cannot kill container"
	defaultKillContainerMaxRetries         = uint8(3)
	defaultKillContainerTimeBetweenRetries = 10 * time.Millisecond

	successfulExitCode = 0

	emptyNetworkAlias     = ""
	streamOutputDelimiter = '\n'

	isDockerNetworkAttachable = true

	linuxAmd64              = "linux/amd64"
	defaultPlatform         = ""
	architectureErrorString = "no matching manifest for linux/arm64/v8"

	onlyReturnContainerIds = true
	coresToMilliCores      = 1000
	bytesInMegaBytes       = 1000000
	dontStreamStats        = false

	kurtosisTagPrefix = "kurtosistech/"

	defaultContainerImageFile = "Dockerfile"

	// Per https://github.com/hashicorp/waypoint/pull/1937/files
	buildkitSessionSharedKey = ""
)

type RestartPolicy string

const (
	RestartAlways    = "always"
	RestartOnFailure = "on-failure"
	NoRestart        = ""
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
	// This client has a timeout so that request that should return quickly do not end up hanging forever.
	dockerClient *client.Client

	// We need to use a specific docker client with no timeout for long-running requests on docker, such as tailing
	// service logs for a long time, or even downloading large container images than can take longer than the timeout
	dockerClientNoTimeout *client.Client
}

/*
CreateDockerManager
Creates a new Docker manager for manipulating the Docker engine using the given client.

Args:

	dockerClient: The Docker client that will be used when interacting with the underlying Docker engine the Docker engine.
*/
func CreateDockerManager(dockerClientOpts []client.Opt) (*DockerManager, error) {
	optsWithTimeout := []client.Opt{
		client.WithTimeout(dockerClientTimeout),
	}
	optsWithTimeout = append(optsWithTimeout, dockerClientOpts...)
	dockerClient, err := client.NewClientWithOpts(optsWithTimeout...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error creating docker client")
	}
	dockerClientNoTimeout, err := client.NewClientWithOpts(dockerClientOpts...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error creating docker client")
	}

	return &DockerManager{
		dockerClient:          dockerClient,
		dockerClientNoTimeout: dockerClientNoTimeout,
	}, nil
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
func (manager *DockerManager) CreateNetwork(context context.Context, name string, subnetMask string, gatewayIP net.IP, labels map[string]string) (id string, err error) {
	ipamConfig := []network.IPAMConfig{{
		Subnet:     subnetMask,
		IPRange:    "",
		Gateway:    gatewayIP.String(),
		AuxAddress: nil,
	}}

	resp, err := manager.dockerClient.NetworkCreate(context, name, types.NetworkCreate{
		CheckDuplicate: false,
		Driver:         dockerNetworkDriver,
		Scope:          "",
		EnableIPv6:     false,
		IPAM: &network.IPAM{
			Driver:  "",
			Options: nil,
			Config:  ipamConfig,
		},
		Internal:   false,
		Attachable: isDockerNetworkAttachable,
		Ingress:    false,
		ConfigOnly: false,
		ConfigFrom: nil,
		Options: map[string]string{
			"com.docker.network.driver.mtu": "1440",
		},
		Labels: labels,
	})
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create network %s with subnet %s", name, subnetMask)
	}
	return resp.ID, nil
}

func (manager *DockerManager) ListNetworks(ctx context.Context) ([]types.NetworkResource, error) {
	networks, err := manager.dockerClient.NetworkList(ctx, types.NetworkListOptions{
		Filters: filters.Args{},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred listing the Docker networks")
	}
	// The Network objects that come back ostensibly, *should*  have the Containers field filled out... but they don't
	// If we ever need that field, we have to call an InspectNetwork, and even then it seems to have some amount of
	// nondeterminism (i.e. brand-new containers won't show up)
	return networks, nil
}

func (manager *DockerManager) PruneUnusedImages(ctx context.Context) ([]types.ImageSummary, error) {
	unusedImages, err := manager.ListUnusedImages(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list unused images")
	}
	logrus.Debugf("List of unused images to be pruned '%v'", unusedImages)
	successfulPrunedImages := []types.ImageSummary{}
	for _, image := range unusedImages {
		imagePruneResponse, err := manager.dockerClient.ImageRemove(ctx, image.ID, types.ImageRemoveOptions{}) //nolint:exhaustruct
		if err != nil {
			return successfulPrunedImages, stacktrace.Propagate(err, "Failed to remove image '%v'", image.ID)
		}
		logrus.Debugf("Pruned image '%v' with response '%v'", image, imagePruneResponse)
		successfulPrunedImages = append(successfulPrunedImages, image)
	}
	return successfulPrunedImages, nil
}

func containsSemVer(s string) bool {
	// Matches patterns like X.Y.Z
	semVerRegex := `\b(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)\b`
	matched, _ := regexp.MatchString(semVerRegex, s)
	return matched
}

func (manager *DockerManager) ListUnusedImages(ctx context.Context) ([]types.ImageSummary, error) {
	images, err := manager.dockerClient.ImageList(ctx, types.ImageListOptions{}) //nolint:exhaustruct
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list Docker images")
	}
	containers, err := manager.dockerClient.ContainerList(ctx, types.ContainerListOptions{All: true}) //nolint:exhaustruct
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list Docker images")
	}

	usedImages := make(map[string]bool)
	for _, cont := range containers {
		usedImages[cont.ImageID] = true
	}

	unusedImages := []types.ImageSummary{}
	for _, image := range images {
		if _, used := usedImages[image.ID]; used {
			logrus.Debugf("Skipping image '%v' since its in use", image.ID)
			continue
		}
		for _, tag := range image.RepoTags {
			if strings.Contains(tag, kurtosisTagPrefix) && containsSemVer(tag) && !strings.Contains(tag, kurtosis_sdk_version.KurtosisVersion) {
				unusedImages = append(unusedImages, image)
			}
		}
	}
	return unusedImages, nil
}

/*
GetNetworksByName
Returns Network list matching the given name (if any).
*/
// TODO Combine with GetNetworksByLabel using a search filter builder
func (manager *DockerManager) GetNetworksByName(ctx context.Context, name string) ([]*docker_manager_types.Network, error) {
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
func (manager *DockerManager) GetNetworksByLabels(ctx context.Context, labels map[string]string) ([]*docker_manager_types.Network, error) {
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

func (manager *DockerManager) GetContainerIdsConnectedToNetwork(context context.Context, networkId string) ([]string, error) {
	inspectResponse, err := manager.dockerClient.NetworkInspect(context, networkId, types.NetworkInspectOptions{
		Scope:   "",
		Verbose: false,
	})
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
func (manager *DockerManager) RemoveNetwork(context context.Context, networkId string) error {
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
func (manager *DockerManager) CreateVolume(context context.Context, volumeName string, labels map[string]string) error {
	volumeConfig := volume.CreateOptions{
		ClusterVolumeSpec: nil,
		Driver:            "",
		DriverOpts:        nil,
		Labels:            labels,
		Name:              volumeName,
	}

	return manager.createPersistentVolumeInternal(context, volumeConfig)
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
	listOptions := volume.ListOptions{Filters: filterArgs}
	resp, err := manager.dockerClient.VolumeList(ctx, listOptions)
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
func (manager *DockerManager) GetVolumesByLabels(ctx context.Context, labels map[string]string) ([]*volume.Volume, error) {
	labelsFilterArgs := getLabelsFilterArgs(volumeLabelSearchFilterKey, labels)
	listOptions := volume.ListOptions{Filters: labelsFilterArgs}
	resp, err := manager.dockerClient.VolumeList(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finding volumes with labels '%+v'", labels)
	}

	result := []*volume.Volume{}
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
func (manager *DockerManager) CreateAndStartContainer(
	ctx context.Context,
	args *CreateAndStartContainerArgs,
) (string, map[nat.Port]*nat.PortBinding, error) {

	// If the user passed in a Docker image that doesn't have a tag separator (indicating no tag was specified), manually append
	//  the Docker default tag so that when we search for the image we're searching for a very specific image
	dockerImage := args.dockerImage
	if !strings.Contains(dockerImage, dockerTagSeparatorChar) {
		dockerImage = dockerImage + dockerTagSeparatorChar + dockerDefaultTag
	}

	_, _, err := manager.FetchImage(ctx, dockerImage, args.imageRegistrySpec, args.imageDownloadMode)
	if err != nil {
		logrus.Debugf("Error occurred fetching image '%v'. Err:\n%v", dockerImage, err)
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

	var userStr string
	if args.user != nil {
		userStr = args.user.GetUIDGIDPairAsStr()
	}

	containerConfigPtr, err := manager.getContainerCfg(
		dockerImage,
		isInteractiveMode,
		usedPortsSet,
		args.entrypointArgs,
		args.cmdArgs,
		args.envVariables,
		args.labels,
		userStr,
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Failed to configure container from service.")
	}
	containerHostConfigPtr, err := manager.getContainerHostConfig(
		args.addedCapabilities,
		args.securityOpts,
		args.networkMode,
		args.bindMounts,
		args.volumeMounts,
		args.usedPorts,
		args.needsAccessToDockerHostMachine,
		args.cpuAllocationMillicpus,
		args.memoryAllocationMegabytes,
		args.loggingDriverConfig,
		args.containerInitEnabled,
		args.restartPolicy)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Failed to configure host to container mappings from service.")
	}

	// note a nil network config would connect to bridge network by default
	var networkConfig *network.NetworkingConfig
	if args.staticIp != nil && args.skipAddingToBridgeNetworkIfStaticIpIsSet {
		targetNetworkEndPointSettings := getEndpointSettingsForIpAddress(args.staticIp.String(), args.alias)
		endpointSettingsByNetworkId := map[string]*network.EndpointSettings{}
		endpointSettingsByNetworkId[args.networkId] = targetNetworkEndPointSettings
		networkConfig = &network.NetworkingConfig{
			EndpointsConfig: endpointSettingsByNetworkId,
		}
	}

	// This function dockerClient.ContainerCreate adds the container to the bridge network if the networkConfig is nil or if its empty
	// Ideally we'd start with an empty network config, add the target network if its supplied and add the bridge network if the person needs it
	// This logic breaks at two places
	// If a person doesn't need either of them, and we pass a nil(or empty) we get the bridge network for free
	// While starting the enclave, adding both bridge & enclave network to the networkConfig just fails
	// I tried creating the container with networkConfig - nil & args.NetworkMode set to none but that stopped me from adding the container to a network
	// using manager.ConnectContainerToNetwork
	containerCreateResp, err := manager.dockerClient.ContainerCreate(ctx, containerConfigPtr, containerHostConfigPtr, networkConfig, nil, args.name)
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

	// static ip is provided and the user wants the connection to bridge network to happen
	// in the container start the bridge network got connected and now we connect to target network
	if args.staticIp != nil && !args.skipAddingToBridgeNetworkIfStaticIpIsSet {
		if err = manager.ConnectContainerToNetwork(ctx, args.networkId, containerId, args.staticIp, args.alias); err != nil {
			return "", nil, stacktrace.Propagate(err, "Failed to connect container %s to network.", containerId)
		}
	}
	// TODO defer a disconnct-from-network if this function doesn't succeed??

	err = manager.StartContainer(ctx, containerId)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Could not start Docker container from image '%v'.", dockerImage)
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

	//Check if the container dies because sometimes users starts containers with a wrong configuration and these quickly dies
	didContainerStartSuccessfully, err := manager.didContainerStartSuccessfully(ctx, containerId, dockerImage)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred checking if container '%v' is running", containerId)
	}

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
	// on the host. We need to look up what those ports are, so we can return report them back to the user.
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

			if !didContainerStartSuccessfully {
				//Then, if the container is running, show the error related to the ports problem
				return "", nil, stacktrace.NewError(
					"%v ports were to be published to the host machine, but container '%v' never got host machine port"+
						" bindings on interface %v for all published ports even after %v checks with %v between checks.",
					numPublishedPorts,
					containerId,
					expectedHostIp,
					maxNumHostPortBindingChecks,
					timeBetweenHostPortBindingChecks,
				)
			}
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
func (manager *DockerManager) GetContainerIP(ctx context.Context, networkName string, containerId string) (string, error) {
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

/*
GetContainerIps
Gets the container's IPs on all networks
Returns a map of network ID to network IP address
*/
func (manager *DockerManager) GetContainerIps(ctx context.Context, containerId string) (map[string]string, error) {
	containerIps := map[string]string{}
	resp, err := manager.dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred inspecting container with ID '%v'", containerId)
	}
	allNetworkInfo := resp.NetworkSettings.Networks
	for _, networkInfo := range allNetworkInfo {
		containerIps[networkInfo.NetworkID] = networkInfo.IPAddress
	}
	return containerIps, nil
}

func (manager *DockerManager) AttachToContainer(ctx context.Context, containerId string) (types.HijackedResponse, error) {
	attachOpts := types.ContainerAttachOptions{
		Stream:     true,
		Stdin:      true,
		Stdout:     true,
		Stderr:     true,
		DetachKeys: "",
		Logs:       false,
	}
	hijackedResponse, err := manager.dockerClient.ContainerAttach(ctx, containerId, attachOpts)
	if err != nil {
		return types.HijackedResponse{}, stacktrace.Propagate(err, "An error occurred attaching to container '%v'", containerId)
	}
	return hijackedResponse, nil
}

/*
StartContainer
Starts the container with the given container ID

Args:

	context: The context that the starting runs in (useful for cancellation)
	containerId: ID of Docker container to start
*/
func (manager *DockerManager) StartContainer(context context.Context, containerId string) error {
	options := types.ContainerStartOptions{
		CheckpointID:  "",
		CheckpointDir: "",
	}
	err := manager.dockerClient.ContainerStart(context, containerId, options)
	if err != nil {
		containerLogs := manager.getFailedContainerLogsOrErrorString(context, containerId)
		containerLogsHeader := "\n--------------------- CONTAINER LOGS -----------------------\n"
		containerLogsFooter := "\n------------------- END CONTAINER LOGS --------------------"
		return stacktrace.Propagate(err, "Could not start Docker container with ID '%v'; logs are below:%v%v%v", containerId, containerLogsHeader, containerLogs, containerLogsFooter)
	}

	return nil
}

/*
StopContainer
Stops the container with the given container ID, waiting for the provided timeout before forcefully terminating the container

Args:

	context: The context that the stopping runs in (useful for cancellation)
	containerId: ID of Docker container to stop
	timeout: How long to wait for container stoppage before throwing an error
*/
func (manager *DockerManager) StopContainer(context context.Context, containerId string, timeout time.Duration) error {
	timeoutSeconds := int(timeout.Seconds())
	stopOpts := container.StopOptions{Signal: "", Timeout: &timeoutSeconds}
	err := manager.dockerClient.ContainerStop(context, containerId, stopOpts)
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
func (manager *DockerManager) KillContainer(ctx context.Context, containerId string) error {
	if err := manager.killContainerWithRetriesWhenErrorResponseFromDaemon(
		ctx,
		containerId,
		defaultKillContainerMaxRetries,
		defaultKillContainerTimeBetweenRetries,
	); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred killing container '%v', even after %v retries with '%v' in between retries",
			containerId,
			defaultKillContainerMaxRetries,
			defaultKillContainerTimeBetweenRetries,
		)
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
	removeOpts := &types.ContainerRemoveOptions{
		RemoveVolumes: shouldRemoveAnonymousVolumesWhenRemovingContainers,
		RemoveLinks:   shouldRemoveLinksWhenRemovingContainers,
		Force:         shouldKillContainersWhenRemovingContainers,
	}
	err := manager.removeContainerWithRetriesOnFailureForZombieProcesses(
		ctx,
		containerId,
		removeOpts,
		defaultRemoveProcessesMaxRetries,
		defaultRemoveContainerTimeBetweenRetries)
	if err != nil {
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
func (manager *DockerManager) WaitForExit(context context.Context, containerId string) (exitCode int64, err error) {
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
func (manager *DockerManager) GetContainerLogs(
	ctx context.Context,
	containerId string,
	shouldFollowLogs bool,
) (io.ReadCloser, error) {
	// As we're using the docker client with no timeout to be able to follow the logs for a long time, we quickly check
	// with the client that has  a timeout whether the docker engine is reachable.
	if _, err := manager.dockerClient.Ping(ctx); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred communicating with docker engine")
	}

	containerLogOpts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      "",
		Until:      "",
		Timestamps: false,
		Follow:     shouldFollowLogs,
		Tail:       "",
		Details:    false,
	}
	readCloser, err := manager.dockerClientNoTimeout.ContainerLogs(ctx, containerId, containerLogOpts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs for container ID '%v'", containerId)
	}
	return readCloser, nil
}

/*
RunExecCommand
Executes the given command inside the container with the given ID, blocking until the command completes
*/
func (manager *DockerManager) RunExecCommand(context context.Context, containerId string, command []string, logOutput io.Writer) (int32, error) {
	dockerClient := manager.dockerClient
	execConfig := types.ExecConfig{
		User:         "",
		Privileged:   false,
		Tty:          false,
		ConsoleSize:  nil,
		AttachStdin:  false,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
		DetachKeys:   "",
		Env:          nil,
		WorkingDir:   "",
		Cmd:          command,
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
		// Can not be run in detached mode or else response from ContainerExecAttach doesn't return output
		Detach:      false,
		Tty:         false,
		ConsoleSize: nil,
	}

	// IMPORTANT NOTE:
	// You'd think that we'd need to call ContainerExecStart separately after this ContainerExecAttach....
	//  ...but ContainerExecAttach **actually starts the exec command!!!!**
	// We used to be doing them both, but then we were hitting this occasional race condition: https://github.com/moby/moby/issues/42408
	// Therefore, we ONLY call Attach, without Start
	attachResp, err := dockerClient.ContainerExecAttach(context, execId, execStartConfig)
	if err != nil {
		return 0, stacktrace.Propagate(err, "An error occurred starting/attaching to the exec command")
	}
	defer attachResp.Close()

	// NOTE: We have to demultiplex the logs that come back
	// This will keep reading until it receives EOF
	concurrentWriter := concurrent_writer.NewConcurrentWriter(logOutput)
	if _, err := stdcopy.StdCopy(concurrentWriter, concurrentWriter, attachResp.Reader); err != nil {
		return 0, stacktrace.Propagate(err, "An error occurred copying the exec command output to the given output writer")
	}

	inspectResponse, err := dockerClient.ContainerExecInspect(context, execId)
	if err != nil {
		return 0, stacktrace.Propagate(err, "An error occurred inspecting the exec to get the response code")
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

func (manager *DockerManager) RunExecCommandWithStreamedOutput(context context.Context, containerId string, command []string) (chan string, chan *exec_result.ExecResult, error) {
	dockerClient := manager.dockerClient
	execConfig := types.ExecConfig{
		User:         "",
		Privileged:   false,
		Tty:          false,
		ConsoleSize:  nil,
		AttachStdin:  false,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
		DetachKeys:   "",
		Env:          nil,
		WorkingDir:   "",
		Cmd:          command,
	}

	createResp, err := dockerClient.ContainerExecCreate(context, containerId, execConfig)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating the exec process")
	}

	execId := createResp.ID
	if execId == "" {
		return nil, nil, stacktrace.NewError("Got back an empty exec ID when running '%v' on container '%v'", command, containerId)
	}

	execStartConfig := types.ExecStartCheck{
		// Can not be run in detached mode or else response from ContainerExecAttach doesn't return output
		Detach:      false,
		Tty:         false,
		ConsoleSize: nil,
	}

	execOutputChan := make(chan string)
	finalExecResultChan := make(chan *exec_result.ExecResult)
	go func() {
		defer func() {
			close(execOutputChan)
			close(finalExecResultChan)
		}()

		// IMPORTANT NOTE:
		// You'd think that we'd need to call ContainerExecStart separately after this ContainerExecAttach....
		//  ...but ContainerExecAttach **actually starts the exec command!!!!**
		// We used to be doing them both, but then we were hitting this occasional race condition: https://github.com/moby/moby/issues/42408
		// Therefore, we ONLY call Attach, without Start
		attachResp, err := dockerClient.ContainerExecAttach(context, execId, execStartConfig)
		if err != nil {
			execOutputChan <- err.Error()
			return
		}
		defer attachResp.Close()

		// Stream output from docker through output channel
		reader := bufio.NewReader(attachResp.Reader)
		for {
			execOutputLine, err := reader.ReadString(streamOutputDelimiter)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return
				}
			}

			execOutputChan <- execOutputLine
		}

		inspectResponse, err := dockerClient.ContainerExecInspect(context, execId)
		if err != nil {
			execOutputChan <- err.Error()
			return
		}
		if inspectResponse.Running {
			execOutputChan <- stacktrace.NewError("Expected exec to have stopped, but it's still running!").Error()
			return
		}
		unsizedExitCode := inspectResponse.ExitCode
		if unsizedExitCode > math.MaxInt32 || unsizedExitCode < math.MinInt32 {
			execOutputChan <- stacktrace.NewError("Could not cast unsized int '%v' to int32 because it does not fit", unsizedExitCode).Error()
			return
		}
		int32ExitCode := int32(unsizedExitCode)

		// Don't send output in final result because it was already streamed
		finalExecResultChan <- exec_result.NewExecResult(int32ExitCode, "")
	}()
	return execOutputChan, finalExecResultChan, nil
}

/*
ConnectContainerToNetwork
Connects the container with the given container ID to the network with the given network ID, using the given IP address
If the IP address passed is nil then we get a random ip address
*/
func (manager *DockerManager) ConnectContainerToNetwork(ctx context.Context, networkId string, containerId string, staticIpAddr net.IP, alias string) (err error) {
	logrus.Tracef(
		"Connecting container ID %v to network ID %v using static IP %v",
		containerId,
		networkId,
		staticIpAddr.String())

	staticIpAddressStr := ""
	if staticIpAddr != nil {
		staticIpAddressStr = staticIpAddr.String()
	}

	config := getEndpointSettingsForIpAddress(staticIpAddressStr, alias)

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

func (manager *DockerManager) DisconnectContainerFromNetwork(ctx context.Context, containerId string, networkId string) error {
	if err := manager.dockerClient.NetworkDisconnect(ctx, networkId, containerId, true); err != nil {
		return stacktrace.Propagate(err, "An error occurred disconnecting container '%v' from network '%v'", containerId, networkId)
	}
	return nil
}

// TODO Refactor this to be GetContainersByName - no need to be so specific now that we have the Container type we can return
func (manager *DockerManager) GetContainerIdsByName(ctx context.Context, nameStr string) ([]string, error) {
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

func (manager *DockerManager) GetContainersByLabels(ctx context.Context, labels map[string]string, shouldShowStoppedContainers bool) ([]*docker_manager_types.Container, error) {
	labelsFilterList := getLabelsFilterArgs(containerLabelSearchFilterKey, labels)
	result, err := manager.getContainersByFilterArgs(ctx, labelsFilterList, shouldShowStoppedContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting containers with labels '%+v'", labelsFilterList)
	}
	return result, nil
}

func (manager *DockerManager) GetContainersByNetworkId(ctx context.Context, networkId string, shouldShowStoppedContainers bool) ([]*docker_manager_types.Container, error) {
	filterArg := filters.Arg(containerNetworkIdSearchFilterKey, networkId)
	networkIdFilterList := filters.NewArgs(filterArg)
	result, err := manager.getContainersByFilterArgs(ctx, networkIdFilterList, shouldShowStoppedContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting containers with network id '%+v'", networkIdFilterList)
	}
	return result, nil
}

// [FetchImageIfMissing] uses the local [dockerImage] if it's available.
// If unavailable, will attempt to fetch the latest image.
// Returns error if local [dockerImage] is unavailable and pulling image fails.
func (manager *DockerManager) FetchImageIfMissing(ctx context.Context, dockerImage string, registrySpec *image_registry_spec.ImageRegistrySpec) (bool, error) {
	// if the image name doesn't have version information we concatenate `:latest`
	// this behavior is similar to CreateAndStartContainer above
	// this allows us to be deterministic in our behaviour
	if !strings.Contains(dockerImage, dockerTagSeparatorChar) {
		dockerImage = dockerImage + dockerTagSeparatorChar + dockerDefaultTag
	}
	logrus.Tracef("Checking if image '%v' is available locally...", dockerImage)
	doesImageExistLocally, err := manager.isImageAvailableLocally(dockerImage)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred checking for local availability of Docker image '%v'", dockerImage)
	}
	logrus.Tracef("Is image available locally?: %v", doesImageExistLocally)

	if !doesImageExistLocally {
		logrus.Tracef("Image doesn't exist locally, so attempting to pull it...")
		err = manager.pullImage(ctx, dockerImage, registrySpec)
		if err != nil {
			return false, stacktrace.Propagate(err, "Failed to pull Docker image '%v' from remote image repository", dockerImage)
		}
		logrus.Tracef("Image successfully pulled from remote to local")
	}

	return !doesImageExistLocally, nil
}

// [FetchLatestImage] always attempts to retrieve the latest [dockerImage].
// If retrieving the latest [dockerImage] fails, the local image will be used.
// Returns error, if no local image is available after retrieving latest fails.
func (manager *DockerManager) FetchLatestImage(ctx context.Context, dockerImage string, registrySpec *image_registry_spec.ImageRegistrySpec) error {
	// if the image name doesn't have version information we concatenate `:latest`
	// this behavior is similar to CreateAndStartContainer above
	// this allows us to be deterministic in our behaviour
	if !strings.Contains(dockerImage, dockerTagSeparatorChar) {
		dockerImage = dockerImage + dockerTagSeparatorChar + dockerDefaultTag
	}
	logrus.Tracef("Checking if image '%v' is available locally...", dockerImage)
	doesImageExistLocally, err := manager.isImageAvailableLocally(dockerImage)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking for local availability of Docker image '%v'", dockerImage)
	}
	logrus.Tracef("Is image available locally?: %v", doesImageExistLocally)

	// try and pull latest image even if image exists locally
	if doesImageExistLocally {
		logrus.Tracef("Image exists locally, but attempting to get latest from remote image repository.")
		err = manager.pullImage(ctx, dockerImage, registrySpec)
		if err != nil {
			logrus.Tracef("Failed to pull Docker image '%v' from remote image repository. Going to use available local image.", dockerImage)
		} else {
			logrus.Tracef("Latest image successfully pulled from remote to local.")
		}
	} else {
		err = manager.pullImage(ctx, dockerImage, registrySpec)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to pull Docker image '%v' from remote image repository.", dockerImage)
		}
	}

	return nil
}

func (manager *DockerManager) FetchImage(ctx context.Context, image string, registrySpec *image_registry_spec.ImageRegistrySpec, downloadMode image_download_mode.ImageDownloadMode) (bool, string, error) {
	var err error
	var pulledFromRemote bool = true
	logrus.Debugf("Fetching image '%s' with image download mode: %s", image, downloadMode)

	switch image_fetching := downloadMode; image_fetching {
	case image_download_mode.ImageDownloadMode_Always:
		err = manager.FetchLatestImage(ctx, image, registrySpec)
	case image_download_mode.ImageDownloadMode_Missing:
		pulledFromRemote, err = manager.FetchImageIfMissing(ctx, image, registrySpec)
	default:
		return false, "", stacktrace.NewError("Undefined image pulling mode: '%v'", image_fetching)
	}

	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred fetching image '%v'", image)
	}

	imageArchitecture, err := manager.getImagePlatform(ctx, image)
	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred while fetching the architecture of the image")
	}

	return pulledFromRemote, imageArchitecture, nil
}

func (manager *DockerManager) BuildImage(ctx context.Context, imageName string, imageBuildSpec *image_build_spec.ImageBuildSpec) (string, error) {
	buildContextDirPath := imageBuildSpec.GetBuildContextDir()
	buildContextTarReader, err := getBuildContextReader(buildContextDirPath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred retrieving the build context for '%v' at context directory path: %v", imageName, buildContextDirPath)
	}

	// Before instructing docker client to execute an image build, we need to create a connection to buildkit
	// buildkit is the daemon process that executes build workloads: https://docs.docker.com/build/architecture/#buildkit

	// Setup session to buildkit (eg. https://github.com/hashicorp/waypoint/pull/1937)
	uuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating a UUID to give the Docker Buildkit session")
	}
	sessionName := fmt.Sprintf("kurtosis-%s", uuidStr)

	// Generate a new session every time because per https://github.com/moby/buildkit/issues/1432 sharing sessions is an optimization
	// Don't bother reusing sessions so that we don't hit bugs
	buildkitSession, err := bksession.NewSession(ctx, sessionName, buildkitSessionSharedKey)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error generating a Docker Buildkit session with sessionName: %v", sessionName)
	}
	dialSessionFunc := func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
		return manager.dockerClientNoTimeout.DialHijack(ctx, "/session", proto, meta)
	}

	// Activate the session
	go func() {
		err := buildkitSession.Run(ctx, dialSessionFunc)
		if err != nil {
			logrus.Errorf("An error occurred running a buildkit session for building image '%v':\n%v", imageName, err)
		}
	}()
	defer buildkitSession.Close() //nolint

	imageBuildOpts := types.ImageBuildOptions{
		Tags:           []string{imageName},
		SuppressOutput: false,
		RemoteContext:  "",    // We don't have a remote context (we're uploading it)
		NoCache:        false, // needs to be false so image only rebuilds if docker detects changes to cached image
		Remove:         false,
		ForceRemove:    false,
		PullParent:     false,
		Isolation:      container.Isolation(""),
		CPUSetCPUs:     "",
		CPUSetMems:     "",
		CPUShares:      0,
		CPUQuota:       0,
		CPUPeriod:      0,
		Memory:         0,
		MemorySwap:     0,
		CgroupParent:   "",
		NetworkMode:    "",
		ShmSize:        0,
		Dockerfile:     defaultContainerImageFile,
		Ulimits:        []*units.Ulimit{},
		BuildArgs:      map[string]*string{},
		AuthConfigs:    map[string]registry.AuthConfig{},
		Context:        buildContextTarReader,
		// 0.0.0 label is a hack so that images by internal testsuite are cleaned up by kurtosis clean/PruneUnusedImages
		Labels:      map[string]string{},
		Squash:      false,
		CacheFrom:   []string{},
		SecurityOpt: []string{},
		ExtraHosts:  []string{},
		Target:      imageBuildSpec.GetTargetStage(),
		SessionID:   buildkitSession.ID(),
		Platform:    "",
		// Version specifies the version of the underlying builder to use
		Version: types.BuilderBuildKit, // Use 2 for BuildKit
		// BuildID is an optional identifier that can be passed together with the
		// build request. The same identifier can be used to gracefully cancel the
		// build with the cancel request.
		BuildID: "",
		// Outputs defines configurations for exporting build results. Only supported in BuildKit mode.
		Outputs: []types.ImageBuildOutput{},
	}
	imageBuildResponse, err := manager.dockerClientNoTimeout.ImageBuild(ctx, buildContextTarReader, imageBuildOpts)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred attempting to build image using Docker: %v", imageName)
	}
	defer imageBuildResponse.Body.Close()

	var imageBuildResponseBuffer bytes.Buffer
	_, err = io.Copy(&imageBuildResponseBuffer, imageBuildResponse.Body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while trying to pipe image build output to a buffer.")
	}
	imageBuildResponseBodyStr := imageBuildResponseBuffer.String()

	// ImageBuildResponse has no notion of success or error builds, so we check if the image is available locally and return the
	// response body if it is not found
	isImageAvailable, err := manager.isImageAvailableLocally(imageName)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to check if '%v' was built and available locally.", imageName)
	}
	if !isImageAvailable {
		return "", stacktrace.NewError("Image build for '%s' failed with the following output:\n%v", imageName, imageBuildResponseBodyStr)
	}

	imageArch, err := manager.getImagePlatform(ctx, imageName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred attempting to get image platform for '%v'.", imageName)
	}

	return imageArch, nil
}

// returns a reader to a tarball of [contextDirPath]
func getBuildContextReader(contextDirPath string) (io.Reader, error) {
	buildContext, _, _, err := utils.CompressPath(contextDirPath, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred compressing the path to context directory path '%v'", contextDirPath)
	}
	return buildContext, nil
}

func (manager *DockerManager) CreateContainerExec(context context.Context, containerId string, cmd []string) (*types.HijackedResponse, error) {
	config := types.ExecConfig{
		User:         "",
		Privileged:   false,
		Tty:          shouldAttachStandardStreamsToTtyWhenCreatingContainerExec,
		ConsoleSize:  nil,
		AttachStdin:  shouldAttachStdinWhenCreatingContainerExec,
		AttachStderr: shouldAttachStderrWhenCreatingContainerExec,
		AttachStdout: shouldAttachStdoutWhenCreatingContainerExec,
		Detach:       shouldExecuteInDetachModeWhenCreatingContainerExec,
		DetachKeys:   "",
		Env:          nil,
		WorkingDir:   "",
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
		Detach:      false,
		Tty:         true,
		ConsoleSize: nil,
	}

	hijackedResponse, err := manager.dockerClient.ContainerExecAttach(context, execID, execStartCheck)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was an error while attaching connection to the execution process with ID '%v' in container with ID '%v'", execID, containerId)
	}

	return &hijackedResponse, nil
}

// CopyFromContainer returns a io.ReadCloser representing the bytes of the TAR'd files at srcPath
// The caller must close the result
func (manager *DockerManager) CopyFromContainer(ctx context.Context, containerId string, srcPath string) (io.ReadCloser, error) {

	tarStreamReadCloser, _, err := manager.dockerClient.CopyFromContainer(
		ctx,
		containerId,
		srcPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred copying content '%v' from container with ID '%v'", srcPath, containerId)
	}

	return tarStreamReadCloser, nil
}

// GetAvailableCPUAndMemory returns free memory in megabytes, free cpu in millicores, information on whether cpu information is complete
func (manager *DockerManager) GetAvailableCPUAndMemory(ctx context.Context) (compute_resources.MemoryInMegaBytes, compute_resources.CpuMilliCores, error) {
	availableMemoryInBytes, availableCpuInMilliCores, err := getFreeMemoryAndCPU(ctx, manager.dockerClient)
	if err != nil {
		return 0, 0, stacktrace.Propagate(err, "an error occurred while getting available cpu and memory on docker")
	}
	// cpu isn't complete on windows but is complete on linux
	return compute_resources.MemoryInMegaBytes(availableMemoryInBytes), compute_resources.CpuMilliCores(availableCpuInMilliCores), nil
}

// =================================================================================================================
//
//	INSTANCE HELPER FUNCTIONS
//
// =================================================================================================================
func (manager *DockerManager) createPersistentVolumeInternal(context context.Context, volumeConfig volume.CreateOptions) error {
	/*
		We don't use the return value of VolumeCreate because there's not much useful information on there - Docker doesn't
		use UUIDs to identify volumes - only the name - so there's no UUID to retrieve, and the volume's Mountpoint (what you'd
		think would be the path of the volume on the local machine) isn't useful either because Docker itself runs inside a VM
		so *this path is only a path inside the Docker VM* (meaning we can't use it to read/write files). AFAICT, the only way
		to read/write data to a volume is to mount it in a container. ~ ktoday, 2020-07-01
	*/
	_, err := manager.dockerClient.VolumeCreate(context, volumeConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Could not create Docker volume for test controller")
	}

	return nil

}

func (manager *DockerManager) isImageAvailableLocally(imageName string) (bool, error) {
	// Own context for checking if the image is locally available because we do not want to cancel this works in case the main context in the request is cancelled
	// if the first request fails the image will be ready for following request making the process faster
	checkImageAvailabilityCtx := context.Background()
	referenceArg := filters.Arg("reference", imageName)
	filterArgs := filters.NewArgs(referenceArg)
	images, err := manager.dockerClient.ImageList(
		checkImageAvailabilityCtx,
		types.ImageListOptions{
			All:            true,
			Filters:        filterArgs,
			SharedSize:     false,
			ContainerCount: false,
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

func (manager *DockerManager) pullImage(context context.Context, imageName string, registrySpec *image_registry_spec.ImageRegistrySpec) error {
	// As we're using the docker client with no timeout to pull the image, we quickly check with the client that has
	// a timeout whether the docker engine is reachable.
	if _, err := manager.dockerClient.Ping(context); err != nil {
		return stacktrace.Propagate(err, "An error occurred communicating with docker engine")
	}
	logrus.Infof("Pulling image '%s'", imageName)
	err, retryWithLinuxAmd64 := pullImage(manager.dockerClientNoTimeout, imageName, registrySpec, defaultPlatform)
	if err == nil {
		return nil
	}
	if err != nil && !retryWithLinuxAmd64 {
		return stacktrace.Propagate(err, "Tried pulling image '%v' but failed", imageName)
	}
	// we retry with linux/amd64
	logrus.Debugf("Retrying pulling image '%s' for '%s'", imageName, linuxAmd64)
	err, _ = pullImage(manager.dockerClientNoTimeout, imageName, registrySpec, linuxAmd64)
	if err != nil {
		return stacktrace.Propagate(err, "Had previously failed with a manifest error so tried pulling image '%v' for platform '%v' but failed", imageName, linuxAmd64)
	}
	logrus.Warnf("Image '%s' successfully pulled for '%s' which is not the architecture this OS is running on.", imageName, linuxAmd64)
	return nil
}

func (manager *DockerManager) getNetworksByFilterArgs(ctx context.Context, args filters.Args) ([]types.NetworkResource, error) {
	// NOTE: Even though this returns a `NetworkResource` object which has a Containers field on it, this is a lie!!
	// For whatever insane reason, Docker doesn't fill this field out when NetworkList is used and there doesn't seem to
	// be a way to get it to do so. Instead, we'd have to do an InspectNetwork call.
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

func (manager *DockerManager) getImagePlatform(ctx context.Context, imageName string) (string, error) {
	imageInspect, _, err := manager.dockerClient.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return "", stacktrace.Propagate(err, "an error occurred while running image inspect on image '%v'", imageName)
	}

	return imageInspect.Architecture, nil
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
	securityOpts map[ContainerSecurityOpt]bool,
	networkMode DockerManagerNetworkMode,
	bindMounts map[string]string,
	volumeMounts map[string]string,
	usedPortsWithPublishSpec map[nat.Port]PortPublishSpec,
	needsToAccessDockerHostMachine bool,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
	loggingDriverConfig LoggingDriver,
	useInit bool,
	restartPolicy RestartPolicy,
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
				{
					HostIP:   "",
					HostPort: "",
				},
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

	securityOptsSlice := []string{}
	for securityOpt := range securityOpts {
		securityOptStr := string(securityOpt)
		securityOptsSlice = append(securityOptsSlice, securityOptStr)
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

	resources := container.Resources{
		CPUShares:            0,
		Memory:               0,
		NanoCPUs:             0,
		CgroupParent:         "",
		BlkioWeight:          0,
		BlkioWeightDevice:    nil,
		BlkioDeviceReadBps:   nil,
		BlkioDeviceWriteBps:  nil,
		BlkioDeviceReadIOps:  nil,
		BlkioDeviceWriteIOps: nil,
		CPUPeriod:            0,
		CPUQuota:             0,
		CPURealtimePeriod:    0,
		CPURealtimeRuntime:   0,
		CpusetCpus:           "",
		CpusetMems:           "",
		Devices:              nil,
		DeviceCgroupRules:    nil,
		DeviceRequests:       nil,
		KernelMemory:         0,
		KernelMemoryTCP:      0,
		MemoryReservation:    0,
		MemorySwap:           0,
		MemorySwappiness:     nil,
		OomKillDisable:       nil,
		PidsLimit:            nil,
		Ulimits:              nil,
		CPUCount:             0,
		CPUPercent:           0,
		IOMaximumIOps:        0,
		IOMaximumBandwidth:   0,
	}
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

	logConfig := container.LogConfig{
		Type:   "",
		Config: nil,
	}
	if loggingDriverConfig != nil {
		logConfig = loggingDriverConfig.GetLogConfig()
	}

	// NOTE: Do NOT use PublishAllPorts here!!!! This will work if a Dockerfile doesn't have an EXPOSE directive, but
	//  if the Dockerfile *does* have and EXPOSE directive then _only_ the ports with EXPOSE will be published
	// See also: https://www.ctl.io/developers/blog/post/docker-networking-rules/

	containerHostConfigPtr := &container.HostConfig{
		Binds:           bindsList,
		ContainerIDFile: "",
		LogConfig:       logConfig,
		NetworkMode:     container.NetworkMode(networkMode),
		PortBindings:    portMap,
		RestartPolicy: container.RestartPolicy{
			Name:              string(restartPolicy),
			MaximumRetryCount: 0,
		},
		AutoRemove:      false,
		VolumeDriver:    "",
		VolumesFrom:     nil,
		Annotations:     map[string]string{},
		CapAdd:          addedCapabilitiesSlice,
		CapDrop:         nil,
		CgroupnsMode:    "",
		DNS:             nil,
		DNSOptions:      nil,
		DNSSearch:       nil,
		ExtraHosts:      extraHosts,
		GroupAdd:        nil,
		IpcMode:         "",
		Cgroup:          "",
		Links:           nil,
		OomScoreAdj:     0,
		PidMode:         "",
		Privileged:      false,
		PublishAllPorts: false,
		ReadonlyRootfs:  false,
		SecurityOpt:     securityOptsSlice,
		StorageOpt:      nil,
		Tmpfs:           nil,
		UTSMode:         "",
		UsernsMode:      "",
		ShmSize:         0,
		Sysctls:         nil,
		Runtime:         "",
		ConsoleSize:     [2]uint{},
		Isolation:       "",
		Resources:       resources,
		Mounts:          nil,
		MaskedPaths:     nil,
		ReadonlyPaths:   nil,
		Init:            &useInit,
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
	labels map[string]string,
	user string) (config *container.Config, err error) {

	envVariablesSlice := make([]string, 0, len(envVariables))
	for key, val := range envVariables {
		envVariablesSlice = append(envVariablesSlice, fmt.Sprintf("%v=%v", key, val))
	}

	nodeConfigPtr := &container.Config{
		Hostname:        "",
		Domainname:      "",
		User:            user,
		AttachStdin:     isInteractiveMode, // Analogous to `-a STDIN` option to `docker run`
		AttachStdout:    isInteractiveMode, // Analogous to `-a STDOUT` option to `docker run`
		AttachStderr:    isInteractiveMode, // Analogous to `-a STDERR` option to `docker run`
		ExposedPorts:    usedPorts,
		Tty:             isInteractiveMode, // Analogous to the `-t` option to `docker run`
		OpenStdin:       true,              // Analogous to the `-i` option to `docker run`
		StdinOnce:       false,
		Env:             envVariablesSlice,
		Cmd:             cmdArgs,
		Healthcheck:     nil,
		ArgsEscaped:     false,
		Image:           dockerImage,
		Volumes:         nil,
		WorkingDir:      "",
		Entrypoint:      entrypointArgs,
		NetworkDisabled: false,
		MacAddress:      "",
		OnBuild:         nil,
		Labels:          labels,
		StopSignal:      "",
		StopTimeout:     nil,
		Shell:           nil,
	}
	return nodeConfigPtr, nil
}

func (manager *DockerManager) killContainerWithRetriesWhenErrorResponseFromDaemon(
	ctx context.Context,
	containerId string,
	maxRetries uint8,
	timeBetweenRetries time.Duration,
) error {

	var err error

	for i := uint8(0); i < maxRetries; i++ {
		if err = manager.dockerClient.ContainerKill(ctx, containerId, dockerKillSignal); err != nil {

			errMsg := strings.ToLower(err.Error())

			// For some stupid reason, ContainerKill throws an error if the container isn't running (even though
			//  ContainerStop does not)
			if strings.Contains(errMsg, containerIsNotRunningErrMsg) {
				return nil
			}

			//Container wasn't killed, waits and retry
			if strings.Contains(errMsg, cannotKillContainerErrMsg) {
				time.Sleep(timeBetweenRetries)
				continue
			}
			break
		}
		return nil
	}

	return stacktrace.Propagate(err, "An error occurred killing container with ID '%v'", containerId)
}

func (manager *DockerManager) removeContainerWithRetriesOnFailureForZombieProcesses(
	ctx context.Context,
	containerId string,
	options *types.ContainerRemoveOptions,
	maxRetries uint8,
	timeBetweenRetries time.Duration,
) error {
	var err error
	for i := uint8(0); i < maxRetries; i++ {
		if err = manager.dockerClient.ContainerRemove(ctx, containerId, *options); err != nil {

			errMsg := strings.ToLower(err.Error())

			// For some stupid reason, ContainerKill throws an error if the container isn't running (even though
			//  ContainerStop does not)
			if strings.Contains(errMsg, zombieProcessesCannotRemoveContainerErrMsg) {
				logrus.Warnf("Container with ID '%s' has zombie processes and cannot be removed. Removal will be retried in %f seconds", containerId, timeBetweenRetries.Seconds())
				time.Sleep(timeBetweenRetries)
				continue
			} else {
				return err
			}
		}
		return nil
	}
	return stacktrace.Propagate(err, "All %d attempts to remove container with ID '%s' failed", maxRetries, containerId)
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

func (manager *DockerManager) getContainersByFilterArgs(ctx context.Context, filterArgs filters.Args, shouldShowStoppedContainers bool) ([]*docker_manager_types.Container, error) {
	opts := types.ContainerListOptions{
		Size:    false,
		All:     shouldShowStoppedContainers,
		Latest:  false,
		Since:   "",
		Before:  "",
		Limit:   0,
		Filters: filterArgs,
	}
	dockerContainers, err := manager.dockerClient.ContainerList(ctx, opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the docker containers with filter args '%+v'", filterArgs)
	}
	dockerContainersDetails := []types.ContainerJSON{}
	for _, dockerContainer := range dockerContainers {
		dockerContainerDetails, err := manager.InspectContainer(ctx, dockerContainer.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred inspecting the docker container with ID '%s'", dockerContainer.ID)
		}
		dockerContainersDetails = append(dockerContainersDetails, dockerContainerDetails)
	}
	containers, err := newContainersListFromDockerContainersList(dockerContainersDetails)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new containers list from Docker containers list")
	}
	return containers, nil
}

// didContainerStartSuccessfully there are 4 return cases:
// 1- sometime goes wrong during the method's execution, for example an error calling manager.InspectContainer
// 2- the container is running which will return true and nil
// 2- the container runs successfully and exited with 0, in this case this method will also return true and nil
// 3- the container dies, in this case this method will return false and the error with the container logs
func (manager *DockerManager) didContainerStartSuccessfully(ctx context.Context, containerId string, dockerImage string) (bool, error) {

	containerJson, err := manager.InspectContainer(ctx, containerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting container JSON info for container with ID '%v'", containerId)
	}
	containerState := containerJson.State
	containerStatus, err := getContainerStatusByDockerContainerState(containerState.Status)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting ContainerStatus from Docker container state '%v'", containerState.Status)
	}

	//check if the container run successfully, could be those cases were the container execute a small task, like configuration and then exits
	if containerStatus == docker_manager_types.ContainerStatus_Exited && containerState.ExitCode == successfulExitCode {
		return true, nil
	}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return false, stacktrace.NewError("No is-running designation found for API container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}

	if !isContainerRunning {
		containerLogs := manager.getFailedContainerLogsOrErrorString(ctx, containerId)
		containerLogsHeader := "\n--------------------- CONTAINER LOGS -----------------------\n"
		containerLogsFooter := "\n------------------- END CONTAINER LOGS --------------------"
		return false, stacktrace.NewError("Container '%v' (with image '%v') die with a non zero exit code rapidly after it was started. This likely indicates a misconfiguration with how the container was started. Container should either exit gracefully or keep running for Kurtosis to consider it in a good state; logs are below:%v%v%v", containerId, dockerImage, containerLogsHeader, containerLogs, containerLogsFooter)
	}

	return true, nil
}

func newContainersListFromDockerContainersList(dockerContainers []types.ContainerJSON) ([]*docker_manager_types.Container, error) {
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

func newContainerFromDockerContainer(dockerContainer types.ContainerJSON) (*docker_manager_types.Container, error) {
	containerStatus, err := getContainerStatusByDockerContainerState(dockerContainer.State.Status)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting ContainerStatus from Docker container state '%v'", dockerContainer.State.Status)
	}
	containerHostPortBindings := getHostPortBindingsOnExpectedInterface(dockerContainer.NetworkSettings.Ports)
	containerEnvArgs := map[string]string{}
	for _, env := range dockerContainer.Config.Env {
		envSlice := strings.Split(env, "=")
		containerEnvArgs[envSlice[0]] = envSlice[1]
	}

	newContainer := docker_manager_types.NewContainer(
		dockerContainer.ID,
		dockerContainer.Name,
		dockerContainer.Config.Labels,
		containerStatus,
		containerHostPortBindings,
		dockerContainer.Config.Image,
		dockerContainer.Config.Entrypoint,
		dockerContainer.Config.Cmd,
		containerEnvArgs,
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

func (manager *DockerManager) getFailedContainerLogsOrErrorString(ctx context.Context, containerId string) string {

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

func getEndpointSettingsForIpAddress(ipAddress string, alias string) *network.EndpointSettings {
	ipamConfig := &network.EndpointIPAMConfig{
		IPv4Address:  ipAddress,
		IPv6Address:  "",
		LinkLocalIPs: nil,
	}

	config := &network.EndpointSettings{
		IPAMConfig:          ipamConfig,
		Links:               nil,
		NetworkID:           "",
		EndpointID:          "",
		Gateway:             "",
		IPAddress:           "",
		Aliases:             nil,
		IPPrefixLen:         0,
		IPv6Gateway:         "",
		GlobalIPv6Address:   "",
		GlobalIPv6PrefixLen: 0,
		MacAddress:          "",
		DriverOpts:          nil,
	}

	if alias != emptyNetworkAlias {
		// docker treats [""] differently from []
		config.Aliases = []string{alias}
	}

	return config
}

func pullImage(dockerClient *client.Client, imageName string, registrySpec *image_registry_spec.ImageRegistrySpec, platform string) (error, bool) {
	// Own context for pulling images because we do not want to cancel this works in case the main context in the request is cancelled
	// if the fist request fails the image will be ready for following request making the process faster
	pullImageCtx := context.Background()
	logrus.Tracef("Starting pulling '%s' for platform '%s'", imageName, platform)
	imagePullOptions := types.ImagePullOptions{
		All:           false,
		RegistryAuth:  "",
		PrivilegeFunc: nil,
		Platform:      platform,
	}
	if registrySpec != nil {
		authConfig := registry.AuthConfig{
			Username:      registrySpec.GetUsername(),
			Password:      registrySpec.GetPassword(),
			Email:         "",
			Auth:          "",
			ServerAddress: registrySpec.GetRegistryAddr(),
			IdentityToken: "",
			RegistryToken: "",
		}
		encodedAuthConfig, err := registry.EncodeAuthConfig(authConfig)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting registry auth to base64"), false
		}
		imagePullOptions.RegistryAuth = encodedAuthConfig

	}
	out, err := dockerClient.ImagePull(pullImageCtx, imageName, imagePullOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Tried pulling image '%v' with platform '%v' but failed", imageName, platform), false
	}
	defer out.Close()
	logrus.Tracef("Finished pulling '%s' for platform '%s'. Analyzing response to check for errors", imageName, platform)
	responseDecoder := json.NewDecoder(out)
	for {
		jsonMessage := new(jsonmessage.JSONMessage)
		err = responseDecoder.Decode(&jsonMessage)
		if err == io.EOF {
			break
		}
		if err != nil {
			return stacktrace.Propagate(err, "ImagePull for '%s' on platform '%s' failed with an unexpected error", imageName, platform), false
		}
		if jsonMessage.Error != nil {
			return stacktrace.NewError("ImagePull failed with the following error '%v'", jsonMessage.Error.Message), strings.HasPrefix(jsonMessage.Error.Message, architectureErrorString)
		}
	}
	logrus.Tracef("No error pulling '%s' for platform '%s'. Returning", imageName, platform)
	return nil, false
}

// getFreeMemoryAndCPU returns free memory in bytes and free cpu in MilliCores
// this is a best effort calculation, it creates a list of containers and then adds up resources on that list
// if a container dies during list creation this just ignores it
func getFreeMemoryAndCPU(ctx context.Context, dockerClient *client.Client) (compute_resources.MemoryInMegaBytes, compute_resources.CpuMilliCores, error) {
	info, err := dockerClient.Info(ctx)
	if err != nil {
		return 0, 0, stacktrace.Propagate(err, "An error occurred while running info on docker")
	}
	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{
		Size:    false,
		All:     false,
		Latest:  false,
		Since:   "",
		Before:  "",
		Limit:   0,
		Filters: filters.Args{},
	})
	if err != nil {
		return 0, 0, stacktrace.Propagate(err, "an error occurred while getting a list of all containers")
	}
	totalFreeMemory := uint64(info.MemTotal)
	totalUsedMemory := uint64(0)
	cpuUsageAsFractionOfAvailableCpu := float64(0)
	totalCPUs := info.NCPU

	var wg sync.WaitGroup
	resourceMutex := sync.Mutex{}

	for _, maybeRunningContainer := range containers {
		wg.Add(1)
		go func(containerId string) {
			defer wg.Done()
			containerStatsResponse, err := dockerClient.ContainerStats(ctx, containerId, dontStreamStats)
			if err != nil {
				if strings.Contains(err.Error(), "No such container") {
					logrus.Warnf("Container with '%v' was in the list of containers for which we wanted to calculate consumed resources but it vanished in the meantime.", containerId)
				}
				logrus.Errorf("An unexpected error occured while fetching information about container '%v':\n%v", containerId, err)
				return
			}
			var containerStats types.Stats
			if err = json.NewDecoder(containerStatsResponse.Body).Decode(&containerStats); err != nil {
				logrus.Errorf("an error occurred while unmarshalling stats response for container with id '%v':\n%v", containerId, err)
				return
			}
			resourceMutex.Lock()
			totalUsedMemory += containerStats.MemoryStats.Usage
			cpuUsageAsFractionOfAvailableCpu += float64(containerStats.CPUStats.CPUUsage.TotalUsage-containerStats.PreCPUStats.CPUUsage.TotalUsage) / float64(containerStats.CPUStats.SystemUsage-containerStats.PreCPUStats.SystemUsage)
			resourceMutex.Unlock()
		}(maybeRunningContainer.ID)
	}
	wg.Wait()
	return compute_resources.MemoryInMegaBytes((totalFreeMemory - totalUsedMemory) / bytesInMegaBytes), compute_resources.CpuMilliCores(float64(totalCPUs*coresToMilliCores) * (1 - cpuUsageAsFractionOfAvailableCpu)), nil
}

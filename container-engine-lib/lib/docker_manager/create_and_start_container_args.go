package docker_manager

import (
	"github.com/docker/go-connections/nat"
	"net"
)

// See CreateAndStartContainerArgsBuilder for detailed documentation on the fields
type CreateAndStartContainerArgs struct {
	dockerImage string
	name string
	alias string
	interactiveModeTtySize *InteractiveModeTtySize // If nil interactive mode will be disabled; if non-nil then interactive mode will be enabled
	networkId string
	staticIp net.IP
	addedCapabilities map[ContainerCapability]bool
	networkMode DockerManagerNetworkMode
	usedPortsSet map[nat.Port]bool
	shouldPublishAllPorts bool
	entrypointArgs []string
	cmdArgs []string
	envVariables map[string]string
	bindMounts map[string]string
	volumeMounts map[string]string
	needsAccessToDockerHostMachine bool
}

// Builder for creating CreateAndStartContainerArgs object
type CreateAndStartContainerArgsBuilder struct {
	dockerImage string
	name string
	alias string
	interactiveModeTtySize *InteractiveModeTtySize // If nil interactive mode will be disabled; if non-nil then interactive mode will be enabled
	networkId string
	staticIp net.IP
	addedCapabilities map[ContainerCapability]bool
	networkMode DockerManagerNetworkMode
	usedPortsSet map[nat.Port]bool
	shouldPublishAllPorts bool
	entrypointArgs []string
	cmdArgs []string
	envVariables map[string]string
	bindMounts map[string]string
	volumeMounts map[string]string
	needsAccessToDockerHostMachine bool
}

/*
Args:
	dockerImage: Image to start
	name: The name to give the container to be created
	networkId: The ID of the Docker network that this container should be attached to
 */
func NewCreateAndStartContainerArgsBuilder(dockerImage string, name string, networkId string) *CreateAndStartContainerArgsBuilder {
	return &CreateAndStartContainerArgsBuilder{
		dockerImage: dockerImage,
		name: name,
		alias: "",
		interactiveModeTtySize: nil,
		networkId: networkId,
		staticIp: nil,
		addedCapabilities: map[ContainerCapability]bool{},
		networkMode: DefaultNetworkMode,
		usedPortsSet: map[nat.Port]bool{},
		shouldPublishAllPorts: false,
		entrypointArgs: nil,
		cmdArgs: nil,
		envVariables: map[string]string{},
		bindMounts: map[string]string{},
		volumeMounts: map[string]string{},
		needsAccessToDockerHostMachine: false,
	}
}

func (builder *CreateAndStartContainerArgsBuilder) Build() *CreateAndStartContainerArgs {
	return &CreateAndStartContainerArgs{
		dockerImage:                    builder.dockerImage,
		name:                           builder.name,
		alias:                          builder.alias,
		interactiveModeTtySize:         builder.interactiveModeTtySize,
		networkId:                      builder.networkId,
		staticIp:                       builder.staticIp,
		addedCapabilities:              builder.addedCapabilities,
		networkMode:                    builder.networkMode,
		usedPortsSet:                   builder.usedPortsSet,
		shouldPublishAllPorts:          builder.shouldPublishAllPorts,
		entrypointArgs:                 builder.entrypointArgs,
		cmdArgs:                        builder.cmdArgs,
		envVariables:                   builder.envVariables,
		bindMounts:                     builder.bindMounts,
		volumeMounts:                   builder.volumeMounts,
		needsAccessToDockerHostMachine: builder.needsAccessToDockerHostMachine,
	}
}

// Alias to give the container, so that other machines can reference the container by this alias to connect to it
func (builder *CreateAndStartContainerArgsBuilder) WithAlias(alias string) *CreateAndStartContainerArgsBuilder {
	builder.alias = alias
	return builder
}

// If non-nil, the container will be started in interactive mode, with a container TTY
//  set to the specified dimensions
func (builder *CreateAndStartContainerArgsBuilder) WithInteractiveModeTtySize(size *InteractiveModeTtySize) *CreateAndStartContainerArgsBuilder {
	builder.interactiveModeTtySize = size
	return builder
}

// IP the container will be assigned (leave nil to not assign any IP, which only works with the bridge network)
func (builder *CreateAndStartContainerArgsBuilder) WithStaticIP(ip net.IP) *CreateAndStartContainerArgsBuilder {
	builder.staticIp = ip
	return builder
}

// A "set" of capabilities to add to the container, corresponding to the --cap-add Docker flag
// For more info, see the --cap-add section of https://docs.docker.com/engine/reference/run/
func (builder *CreateAndStartContainerArgsBuilder) WithAddedCapabilities(capabilities map[ContainerCapability]bool) *CreateAndStartContainerArgsBuilder {
	builder.addedCapabilities = capabilities
	return builder
}

// When a non-empty string, sets the Docker --network flag to be this given string
func (builder *CreateAndStartContainerArgsBuilder) WithNetworkMode(mode DockerManagerNetworkMode) *CreateAndStartContainerArgsBuilder {
	builder.networkMode = mode
	return builder
}

// A set of ports that the container will listen on
func (builder *CreateAndStartContainerArgsBuilder) WithUsedPorts(portsSet map[nat.Port]bool) *CreateAndStartContainerArgsBuilder {
	builder.usedPortsSet = portsSet
	return builder
}

// If true, we'll publish all the exposed ports to the Docker host so that the outside world can connect
//  to the container
func (builder *CreateAndStartContainerArgsBuilder) ShouldPublishAllPorts(shouldPublishAllPorts bool) *CreateAndStartContainerArgsBuilder {
	builder.shouldPublishAllPorts = shouldPublishAllPorts
	return builder
}

// The args that will be used to override the ENTRYPOINT of the image (leave as nil to not override)
func (builder *CreateAndStartContainerArgsBuilder) WithEntrypointArgs(args []string) *CreateAndStartContainerArgsBuilder {
	builder.entrypointArgs = args
	return builder
}


// The args that will be used to run the container (leave as nil to run the CMD in the image)
func (builder *CreateAndStartContainerArgsBuilder) WithCmdArgs(args []string) *CreateAndStartContainerArgsBuilder {
	builder.cmdArgs = args
	return builder
}

// A key-value mapping of Docker environment variables which will be passed to the container during startup
func (builder *CreateAndStartContainerArgsBuilder) WithEnvironmentVariables(envVars map[string]string) *CreateAndStartContainerArgsBuilder {
	builder.envVariables = envVars
	return builder
}

// Mapping of (host file) -> (mountpoint on container) that will be mounted on container startup
func (builder *CreateAndStartContainerArgsBuilder) WithBindMounts(bindMounts map[string]string) *CreateAndStartContainerArgsBuilder {
	builder.bindMounts = bindMounts
	return builder
}

// Mounts: Mapping of (volume name) -> (mountpoint on container) to mount during container launch
func (builder *CreateAndStartContainerArgsBuilder) WithVolumeMounts(volumeMounts map[string]string) *CreateAndStartContainerArgsBuilder {
	builder.volumeMounts = volumeMounts
	return builder
}

// Will provide the container with a magic "host.docker.internal" domain name
//  that it can use to access ports of the machine running Docker itself (useful if, e.g., the container
//  needs to check the host machine's free ports)
func (builder *CreateAndStartContainerArgsBuilder) NeedsAccessToDockerHostMachine(needsAccess bool) *CreateAndStartContainerArgsBuilder {
	builder.needsAccessToDockerHostMachine = needsAccess
	return builder
}
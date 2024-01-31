package docker_manager

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_user"
	"net"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
)

// See CreateAndStartContainerArgsBuilder for detailed documentation on the fields
type CreateAndStartContainerArgs struct {
	dockerImage                              string
	name                                     string
	alias                                    string
	interactiveModeTtySize                   *InteractiveModeTtySize // If nil interactive mode will be disabled; if non-nil then interactive mode will be enabled
	networkId                                string
	staticIp                                 net.IP
	addedCapabilities                        map[ContainerCapability]bool
	securityOpts                             map[ContainerSecurityOpt]bool
	networkMode                              DockerManagerNetworkMode
	usedPorts                                map[nat.Port]PortPublishSpec
	entrypointArgs                           []string
	cmdArgs                                  []string
	envVariables                             map[string]string
	bindMounts                               map[string]string
	volumeMounts                             map[string]string
	needsAccessToDockerHostMachine           bool
	labels                                   map[string]string
	cpuAllocationMillicpus                   uint64
	memoryAllocationMegabytes                uint64
	loggingDriverConfig                      LoggingDriver
	skipAddingToBridgeNetworkIfStaticIpIsSet bool
	containerInitEnabled                     bool
	restartPolicy                            RestartPolicy
	imageDownloadMode                        image_download_mode.ImageDownloadMode
	user                                     *service_user.ServiceUser
	imageRegistrySpec                        *image_registry_spec.ImageRegistrySpec
}

// Builder for creating CreateAndStartContainerArgs object
type CreateAndStartContainerArgsBuilder struct {
	dockerImage                              string
	name                                     string
	alias                                    string
	interactiveModeTtySize                   *InteractiveModeTtySize // If nil interactive mode will be disabled; if non-nil then interactive mode will be enabled
	networkId                                string
	staticIp                                 net.IP
	addedCapabilities                        map[ContainerCapability]bool
	securityOpts                             map[ContainerSecurityOpt]bool
	networkMode                              DockerManagerNetworkMode
	usedPorts                                map[nat.Port]PortPublishSpec
	entrypointArgs                           []string
	cmdArgs                                  []string
	envVariables                             map[string]string
	bindMounts                               map[string]string
	volumeMounts                             map[string]string
	needsAccessToDockerHostMachine           bool
	labels                                   map[string]string
	cpuAllocationMillicpus                   uint64
	memoryAllocationMegabytes                uint64
	loggingDriverCnfg                        LoggingDriver
	skipAddingToBridgeNetworkIfStaticIpIsSet bool
	containerInitEnabled                     bool
	restartPolicy                            RestartPolicy
	imageDownloadMode                        image_download_mode.ImageDownloadMode
	user                                     *service_user.ServiceUser
	imageRegistrySpec                        *image_registry_spec.ImageRegistrySpec
}

/*
Args:

	dockerImage: Image to start
	name: The name to give the container to be created
	networkId: The ID of the Docker network that this container should be attached to
*/
func NewCreateAndStartContainerArgsBuilder(dockerImage string, name string, networkId string) *CreateAndStartContainerArgsBuilder {
	return &CreateAndStartContainerArgsBuilder{
		dockerImage:                              dockerImage,
		name:                                     name,
		alias:                                    "",
		interactiveModeTtySize:                   nil,
		networkId:                                networkId,
		staticIp:                                 nil,
		addedCapabilities:                        map[ContainerCapability]bool{},
		securityOpts:                             map[ContainerSecurityOpt]bool{},
		networkMode:                              DefaultNetworkMode,
		usedPorts:                                map[nat.Port]PortPublishSpec{},
		entrypointArgs:                           nil,
		cmdArgs:                                  nil,
		envVariables:                             map[string]string{},
		bindMounts:                               map[string]string{},
		volumeMounts:                             map[string]string{},
		needsAccessToDockerHostMachine:           false,
		labels:                                   map[string]string{},
		cpuAllocationMillicpus:                   0,
		memoryAllocationMegabytes:                0,
		loggingDriverCnfg:                        nil,
		skipAddingToBridgeNetworkIfStaticIpIsSet: false,
		containerInitEnabled:                     false,
		restartPolicy:                            NoRestart,
		imageDownloadMode:                        image_download_mode.ImageDownloadMode_Missing,
		user:                                     nil,
		imageRegistrySpec:                        nil,
	}
}

func (builder *CreateAndStartContainerArgsBuilder) Build() *CreateAndStartContainerArgs {
	return &CreateAndStartContainerArgs{
		dockerImage:                              builder.dockerImage,
		name:                                     builder.name,
		labels:                                   builder.labels,
		alias:                                    builder.alias,
		interactiveModeTtySize:                   builder.interactiveModeTtySize,
		networkId:                                builder.networkId,
		staticIp:                                 builder.staticIp,
		addedCapabilities:                        builder.addedCapabilities,
		securityOpts:                             builder.securityOpts,
		networkMode:                              builder.networkMode,
		usedPorts:                                builder.usedPorts,
		entrypointArgs:                           builder.entrypointArgs,
		cmdArgs:                                  builder.cmdArgs,
		envVariables:                             builder.envVariables,
		bindMounts:                               builder.bindMounts,
		volumeMounts:                             builder.volumeMounts,
		needsAccessToDockerHostMachine:           builder.needsAccessToDockerHostMachine,
		cpuAllocationMillicpus:                   builder.cpuAllocationMillicpus,
		memoryAllocationMegabytes:                builder.memoryAllocationMegabytes,
		loggingDriverConfig:                      builder.loggingDriverCnfg,
		skipAddingToBridgeNetworkIfStaticIpIsSet: builder.skipAddingToBridgeNetworkIfStaticIpIsSet,
		containerInitEnabled:                     builder.containerInitEnabled,
		restartPolicy:                            builder.restartPolicy,
		imageDownloadMode:                        builder.imageDownloadMode,
		user:                                     builder.user,
		imageRegistrySpec:                        builder.imageRegistrySpec,
	}
}

// Alias to give the container, so that other machines can reference the container by this alias to connect to it
func (builder *CreateAndStartContainerArgsBuilder) WithAlias(alias string) *CreateAndStartContainerArgsBuilder {
	builder.alias = alias
	return builder
}

// If non-nil, the container will be started in interactive mode, with a container TTY
// set to the specified dimensions
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

// A "set" of security options to add to the container, corresponding to the --security-opt Docker flag
// For more info, see https://docs.docker.com/engine/reference/commandline/container_run/#security-opt
func (builder *CreateAndStartContainerArgsBuilder) WithSecurityOpts(securityOpts map[ContainerSecurityOpt]bool) *CreateAndStartContainerArgsBuilder {
	builder.securityOpts = securityOpts
	return builder
}

// When a non-empty string, sets the Docker --network flag to be this given string
func (builder *CreateAndStartContainerArgsBuilder) WithNetworkMode(mode DockerManagerNetworkMode) *CreateAndStartContainerArgsBuilder {
	builder.networkMode = mode
	return builder
}

// A set of ports that the container will listen on, mapped to a specification for how they should be published to the host machine (if at all)
func (builder *CreateAndStartContainerArgsBuilder) WithUsedPorts(usedPorts map[nat.Port]PortPublishSpec) *CreateAndStartContainerArgsBuilder {
	builder.usedPorts = usedPorts
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
// that it can use to access ports of the machine running Docker itself (useful if, e.g., the container
// needs to check the host machine's free ports)
func (builder *CreateAndStartContainerArgsBuilder) NeedsAccessToDockerHostMachine(needsAccess bool) *CreateAndStartContainerArgsBuilder {
	builder.needsAccessToDockerHostMachine = needsAccess
	return builder
}

// A key-value map that represents labels to give the container, for use in searching later
func (builder *CreateAndStartContainerArgsBuilder) WithLabels(labels map[string]string) *CreateAndStartContainerArgsBuilder {
	builder.labels = labels
	return builder
}

// Corresponds to millicpus where 1000 millicpus = 1 CPU in Docker, this gets converted and set to NanoCPUs in the underlying container
// 0 is the empty value, meaning if the value is 0, this field is ignored
// https://pkg.go.dev/github.com/docker/docker@v20.10.17+incompatible/api/types/container#Resources
func (builder *CreateAndStartContainerArgsBuilder) WithCPUAllocationMillicpus(cpuAllocationMillicpus uint64) *CreateAndStartContainerArgsBuilder {
	builder.cpuAllocationMillicpus = cpuAllocationMillicpus
	return builder
}

// Corresponds to `--memory` limit in Docker in megabytes, used to set Memory and MemorySwap resource in the underlying container
// 0 is the empty value, meaning if the value is 0, this field is ignored
// https://pkg.go.dev/github.com/docker/docker@v20.10.17+incompatible/api/types/container#Resources
func (builder *CreateAndStartContainerArgsBuilder) WithMemoryAllocationMegabytes(memoryAllocationMegabytes uint64) *CreateAndStartContainerArgsBuilder {
	builder.memoryAllocationMegabytes = memoryAllocationMegabytes
	return builder
}

// Will configure the container to use and specific logging driver which can be configured using the different implementations
func (builder *CreateAndStartContainerArgsBuilder) WithLoggingDriver(loggingDriverConfig LoggingDriver) *CreateAndStartContainerArgsBuilder {
	builder.loggingDriverCnfg = loggingDriverConfig
	return builder
}

// Will configure the container restart policy (restart on failure, no restart)
func (builder *CreateAndStartContainerArgsBuilder) WithRestartPolicy(restartPolicy RestartPolicy) *CreateAndStartContainerArgsBuilder {
	builder.restartPolicy = restartPolicy
	return builder
}

// WithSkipAddingToBridgeNetworkIfStaticIpIsSet Allows you to skip adding the container to the bridge network assuming the static ip address is set
// With this option set to false
// 1. We connect a container to the bridge network by default when it starts
// 2. We then connect it to the network with the StaticIPAddress (target network)
// With this set to true
// 1. We start the container and connect it to the network with the static ip address by default, and don't connect it to the bridge network
// If the static ip address isn't set, then this has no effect, as Docker defaults to adding to the bridge network if no network is provided
func (builder *CreateAndStartContainerArgsBuilder) WithSkipAddingToBridgeNetworkIfStaticIpIsSet(skipAddingToBridgeNetworkIfStaticIpIsSet bool) *CreateAndStartContainerArgsBuilder {
	builder.skipAddingToBridgeNetworkIfStaticIpIsSet = skipAddingToBridgeNetworkIfStaticIpIsSet
	return builder
}

// WithContainerInitEnabled allows you to set the `--init` options when a container is started in Docker.
// `--init` option wraps the main container process with `tini`, making sure zombie processes are handled properly
// at the container level.
// If not used, a buggy container image can create zombie processes which can prevents docker from removing the
// container. tini makes sure it keeps track of all processes and can kill the zombie ones appropriately.
func (builder *CreateAndStartContainerArgsBuilder) WithContainerInitEnabled(containerInitEnabled bool) *CreateAndStartContainerArgsBuilder {
	builder.containerInitEnabled = containerInitEnabled
	return builder
}

func (builder *CreateAndStartContainerArgsBuilder) WithFetchingLatestImageAlways() *CreateAndStartContainerArgsBuilder {
	builder.imageDownloadMode = image_download_mode.ImageDownloadMode_Always
	return builder
}

func (builder *CreateAndStartContainerArgsBuilder) WithFetchingLatestImageIfMissing() *CreateAndStartContainerArgsBuilder {
	builder.imageDownloadMode = image_download_mode.ImageDownloadMode_Missing
	return builder
}

func (builder *CreateAndStartContainerArgsBuilder) WithUser(user *service_user.ServiceUser) *CreateAndStartContainerArgsBuilder {
	builder.user = user
	return builder
}

func (builder *CreateAndStartContainerArgsBuilder) WithImageRegistrySpec(imageRegistrySpec *image_registry_spec.ImageRegistrySpec) *CreateAndStartContainerArgsBuilder {
	builder.imageRegistrySpec = imageRegistrySpec
	return builder
}

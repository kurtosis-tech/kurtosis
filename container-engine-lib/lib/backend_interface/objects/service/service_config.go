package service

import (
	"encoding/json"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/nix_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_user"
	"github.com/kurtosis-tech/stacktrace"
	v1 "k8s.io/api/core/v1"
)

// Config options for the underlying container of a service
type ServiceConfig struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateServiceConfig *privateServiceConfig
}

type privateServiceConfig struct {
	ContainerImageName string

	// Configuration for container engine to build image for this service
	// If nil, container engine won't be able to build image for this service
	ImageBuildSpec *image_build_spec.ImageBuildSpec

	// Configuration for container engine to pull an in a private registry behind authentication
	// If nil, we will use the ContainerImageName and not use any auth
	// Mutually exclusive from ImageBuildSpec, ContainerImageName, NixBuildSpec
	ImageRegistrySpec *image_registry_spec.ImageRegistrySpec

	// Configuration for container engine to using Nix
	// If nil, we will use the ContainerImageName and not use any Nix
	// Mutually exclusive from ImageBuildSpec, ContainerImageName, ImageRegistrySpec
	NixBuildSpec *nix_build_spec.NixBuildSpec

	PrivatePorts map[string]*port_spec.PortSpec

	PublicPorts map[string]*port_spec.PortSpec //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution

	EntrypointArgs []string

	CmdArgs []string

	EnvVars map[string]string

	// Leave as nil to not do any files artifact expansion
	FilesArtifactExpansion *service_directory.FilesArtifactsExpansion

	PersistentDirectories *service_directory.PersistentDirectories

	CpuAllocationMillicpus uint64

	MemoryAllocationMegabytes uint64

	PrivateIPAddrPlaceholder string

	MinCpuAllocationMilliCpus uint64

	MinMemoryAllocationMegabytes uint64

	Labels map[string]string

	User *service_user.ServiceUser

	// TODO replace this with an abstraction that we own
	Tolerations []v1.Toleration

	NodeSelectors map[string]string

	ImageDownloadMode image_download_mode.ImageDownloadMode

	FilesToBeMoved map[string]string

	TiniEnabled bool

	TtyEnabled bool

	Devices []string

	PublishUdp bool
}

func CreateServiceConfig(
	containerImageName string,
	imageBuildSpec *image_build_spec.ImageBuildSpec,
	imageRegistrySpec *image_registry_spec.ImageRegistrySpec,
	nixBuildSpec *nix_build_spec.NixBuildSpec,
	privatePorts map[string]*port_spec.PortSpec,
	publicPorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	filesArtifactExpansion *service_directory.FilesArtifactsExpansion,
	persistentDirectories *service_directory.PersistentDirectories,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
	privateIPAddrPlaceholder string,
	minCpuMilliCpus uint64,
	minMemoryMegaBytes uint64,
	labels map[string]string,
	user *service_user.ServiceUser,
	tolerations []v1.Toleration,
	nodeSelectors map[string]string,
	imageDownloadMode image_download_mode.ImageDownloadMode,
	tiniEnabled bool,
	ttyEnabled bool,
	devices []string,
	publishUdp bool) (*ServiceConfig, error) {
	if err := ValidateServiceConfigLabels(labels); err != nil {
		return nil, stacktrace.Propagate(err, "Invalid service config labels '%+v'", labels)
	}

	internalServiceConfig := &privateServiceConfig{
		ContainerImageName:        containerImageName,
		ImageBuildSpec:            imageBuildSpec,
		ImageRegistrySpec:         imageRegistrySpec,
		NixBuildSpec:              nixBuildSpec,
		PrivatePorts:              privatePorts,
		PublicPorts:               publicPorts,
		EntrypointArgs:            entrypointArgs,
		CmdArgs:                   cmdArgs,
		EnvVars:                   envVars,
		FilesArtifactExpansion:    filesArtifactExpansion,
		PersistentDirectories:     persistentDirectories,
		CpuAllocationMillicpus:    cpuAllocationMillicpus,
		MemoryAllocationMegabytes: memoryAllocationMegabytes,
		PrivateIPAddrPlaceholder:  privateIPAddrPlaceholder,
		// The minimum resources specification is only available for kubernetes
		MinCpuAllocationMilliCpus:    minCpuMilliCpus,
		MinMemoryAllocationMegabytes: minMemoryMegaBytes,
		Labels:                       labels,
		User:                         user,
		Tolerations:                  tolerations,
		NodeSelectors:                nodeSelectors,
		ImageDownloadMode:            imageDownloadMode,
		FilesToBeMoved:               map[string]string{},
		TiniEnabled:                  tiniEnabled,
		TtyEnabled:                   ttyEnabled,
		Devices:                      devices,
		PublishUdp:                   publishUdp,
	}
	return &ServiceConfig{internalServiceConfig}, nil
}

func (serviceConfig *ServiceConfig) GetContainerImageName() string {
	return serviceConfig.privateServiceConfig.ContainerImageName
}

func (serviceConfig *ServiceConfig) SetContainerImageName(containerImage string) {
	serviceConfig.privateServiceConfig.ContainerImageName = containerImage
}

func (serviceConfig *ServiceConfig) GetImageBuildSpec() *image_build_spec.ImageBuildSpec {
	return serviceConfig.privateServiceConfig.ImageBuildSpec
}

func (serviceConfig *ServiceConfig) SetImageBuildSpec(imageBuildSpec *image_build_spec.ImageBuildSpec) {
	serviceConfig.privateServiceConfig.ImageBuildSpec = imageBuildSpec
}

func (serviceConfig *ServiceConfig) GetImageRegistrySpec() *image_registry_spec.ImageRegistrySpec {
	return serviceConfig.privateServiceConfig.ImageRegistrySpec
}

func (serviceConfig *ServiceConfig) SetImageRegistrySpec(imageRegistrySpec *image_registry_spec.ImageRegistrySpec) {
	serviceConfig.privateServiceConfig.ImageRegistrySpec = imageRegistrySpec
}

func (serviceConfig *ServiceConfig) GetNixBuildSpec() *nix_build_spec.NixBuildSpec {
	return serviceConfig.privateServiceConfig.NixBuildSpec
}

func (serviceConfig *ServiceConfig) SetNixBuildSpec(nixBuildSpec *nix_build_spec.NixBuildSpec) {
	serviceConfig.privateServiceConfig.NixBuildSpec = nixBuildSpec
}

func (serviceConfig *ServiceConfig) GetPrivatePorts() map[string]*port_spec.PortSpec {
	return serviceConfig.privateServiceConfig.PrivatePorts
}

func (serviceConfig *ServiceConfig) GetPublicPorts() map[string]*port_spec.PortSpec {
	return serviceConfig.privateServiceConfig.PublicPorts
}

func (serviceConfig *ServiceConfig) GetEntrypointArgs() []string {
	return serviceConfig.privateServiceConfig.EntrypointArgs
}

func (serviceConfig *ServiceConfig) GetCmdArgs() []string {
	return serviceConfig.privateServiceConfig.CmdArgs
}

func (serviceConfig *ServiceConfig) GetEnvVars() map[string]string {
	return serviceConfig.privateServiceConfig.EnvVars
}

func (serviceConfig *ServiceConfig) GetFilesArtifactsExpansion() *service_directory.FilesArtifactsExpansion {
	return serviceConfig.privateServiceConfig.FilesArtifactExpansion
}

func (serviceConfig *ServiceConfig) GetPersistentDirectories() *service_directory.PersistentDirectories {
	return serviceConfig.privateServiceConfig.PersistentDirectories
}

func (serviceConfig *ServiceConfig) GetCPUAllocationMillicpus() uint64 {
	return serviceConfig.privateServiceConfig.CpuAllocationMillicpus
}

func (serviceConfig *ServiceConfig) SetCPUAllocationMillicpus(cpuAllocation uint64) {
	serviceConfig.privateServiceConfig.CpuAllocationMillicpus = cpuAllocation
}

func (serviceConfig *ServiceConfig) GetMemoryAllocationMegabytes() uint64 {
	return serviceConfig.privateServiceConfig.MemoryAllocationMegabytes
}

func (serviceConfig *ServiceConfig) SetMemoryAllocationMegabytes(memoryAllocation uint64) {
	serviceConfig.privateServiceConfig.MemoryAllocationMegabytes = memoryAllocation
}

func (serviceConfig *ServiceConfig) GetPrivateIPAddrPlaceholder() string {
	return serviceConfig.privateServiceConfig.PrivateIPAddrPlaceholder
}

// only available for Kubernetes
func (serviceConfig *ServiceConfig) GetMinCPUAllocationMillicpus() uint64 {
	return serviceConfig.privateServiceConfig.MinCpuAllocationMilliCpus
}

func (serviceConfig *ServiceConfig) SetMinCPUAllocationMillicpus(cpuAllocation uint64) {
	serviceConfig.privateServiceConfig.MinCpuAllocationMilliCpus = cpuAllocation
}

// only available for Kubernetes
func (serviceConfig *ServiceConfig) GetMinMemoryAllocationMegabytes() uint64 {
	return serviceConfig.privateServiceConfig.MinMemoryAllocationMegabytes
}

func (serviceConfig *ServiceConfig) SetMinMemoryAllocationMegabytes(memoryAllocation uint64) {
	serviceConfig.privateServiceConfig.MemoryAllocationMegabytes = memoryAllocation
}

func (serviceConfig *ServiceConfig) GetUser() *service_user.ServiceUser {
	return serviceConfig.privateServiceConfig.User
}

func (serviceConfig *ServiceConfig) SetUser(user *service_user.ServiceUser) {
	serviceConfig.privateServiceConfig.User = user
}

func (serviceConfig *ServiceConfig) GetLabels() map[string]string {
	return serviceConfig.privateServiceConfig.Labels
}

func (serviceConfig *ServiceConfig) SetLabels(labels map[string]string) {
	serviceConfig.privateServiceConfig.Labels = labels
}

func (serviceConfig *ServiceConfig) GetTolerations() []v1.Toleration {
	return serviceConfig.privateServiceConfig.Tolerations
}

func (serviceConfig *ServiceConfig) SetTolerations(tolerations []v1.Toleration) {
	serviceConfig.privateServiceConfig.Tolerations = tolerations
}

func (serviceConfig *ServiceConfig) GetImageDownloadMode() image_download_mode.ImageDownloadMode {
	return serviceConfig.privateServiceConfig.ImageDownloadMode
}

func (serviceConfig *ServiceConfig) SetImageDownloadMode(mode image_download_mode.ImageDownloadMode) {
	serviceConfig.privateServiceConfig.ImageDownloadMode = mode
}

func (serviceConfig *ServiceConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(serviceConfig.privateServiceConfig)
}

func (serviceConfig *ServiceConfig) GetNodeSelectors() map[string]string {
	return serviceConfig.privateServiceConfig.NodeSelectors
}

func (serviceConfig *ServiceConfig) SetFilesToBeMoved(filesToBeMoved map[string]string) {
	serviceConfig.privateServiceConfig.FilesToBeMoved = filesToBeMoved
}

func (serviceConfig *ServiceConfig) GetFilesToBeMoved() map[string]string {
	return serviceConfig.privateServiceConfig.FilesToBeMoved
}

func (serviceConfig *ServiceConfig) GetTiniEnabled() bool {
	return serviceConfig.privateServiceConfig.TiniEnabled
}

func (serviceConfig *ServiceConfig) GetTtyEnabled() bool {
	return serviceConfig.privateServiceConfig.TtyEnabled
}

func (serviceConfig *ServiceConfig) UnmarshalJSON(data []byte) error {
	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateServiceConfig{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	serviceConfig.privateServiceConfig = unmarshalledPrivateStructPtr
	return nil
}

func GetEmptyServiceConfig() *ServiceConfig {
	emptyServiceConfig, _ := CreateServiceConfig(
		"",
		nil,
		nil,
		nil,
		map[string]*port_spec.PortSpec{},
		map[string]*port_spec.PortSpec{},
		[]string{},
		[]string{},
		map[string]string{},
		&service_directory.FilesArtifactsExpansion{
			ExpanderImage:                        "",
			ExpanderEnvVars:                      nil,
			ServiceDirpathsToArtifactIdentifiers: nil,
			ExpanderDirpathsToServiceDirpaths:    nil,
		},
		&service_directory.PersistentDirectories{
			ServiceDirpathToPersistentDirectory: map[string]service_directory.PersistentDirectory{},
		},
		0,
		0,
		"",
		0,
		0,
		map[string]string{},
		nil,
		[]v1.Toleration{},
		map[string]string{},
		image_download_mode.ImageDownloadMode_Always,
		false,
		false,
		[]string{},
		false,
	)
	return emptyServiceConfig
}

func (serviceConfig *ServiceConfig) GetDevices() []string {
	return serviceConfig.privateServiceConfig.Devices
}

func (serviceConfig *ServiceConfig) GetPublishUdp() bool {
	return serviceConfig.privateServiceConfig.PublishUdp
}

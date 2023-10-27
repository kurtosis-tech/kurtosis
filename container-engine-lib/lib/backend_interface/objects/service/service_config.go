package service

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/stacktrace"
)

// Config options for the underlying container of a service
type ServiceConfig struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateServiceConfig *privateServiceConfig
}

type privateServiceConfig struct {
	ContainerImageName string

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
}

func CreateServiceConfig(
	containerImageName string,
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
	minCpuMilliCores uint64,
	minMemoryMegaBytes uint64,
	labels map[string]string,
) (*ServiceConfig, error) {

	if err := ValidateServiceConfigLabels(labels); err != nil {
		return nil, stacktrace.Propagate(err, "Invalid service config labels '%+v'", labels)
	}

	internalServiceConfig := &privateServiceConfig{
		ContainerImageName:        containerImageName,
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
		MinCpuAllocationMilliCpus:    minCpuMilliCores,
		MinMemoryAllocationMegabytes: minMemoryMegaBytes,
		Labels:                       labels,
	}
	return &ServiceConfig{internalServiceConfig}, nil
}

func (serviceConfig *ServiceConfig) GetContainerImageName() string {
	return serviceConfig.privateServiceConfig.ContainerImageName
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

func (serviceConfig *ServiceConfig) GetMemoryAllocationMegabytes() uint64 {
	return serviceConfig.privateServiceConfig.MemoryAllocationMegabytes
}

func (serviceConfig *ServiceConfig) GetPrivateIPAddrPlaceholder() string {
	return serviceConfig.privateServiceConfig.PrivateIPAddrPlaceholder
}

// only available for Kubernetes
func (serviceConfig *ServiceConfig) GetMinCPUAllocationMillicpus() uint64 {
	return serviceConfig.privateServiceConfig.MinCpuAllocationMilliCpus
}

// only available for Kubernetes
func (serviceConfig *ServiceConfig) GetMinMemoryAllocationMegabytes() uint64 {
	return serviceConfig.privateServiceConfig.MinMemoryAllocationMegabytes
}

func (serviceConfig *ServiceConfig) GetLabels() map[string]string {
	return serviceConfig.privateServiceConfig.Labels
}

func (serviceConfig *ServiceConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(serviceConfig.privateServiceConfig)
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

package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
)

// Config options for the underlying container of a service
type ServiceConfig struct {
	containerImageName string

	privatePorts map[string]*port_spec.PortSpec

	publicPorts map[string]*port_spec.PortSpec //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution

	entrypointArgs []string

	cmdArgs []string

	envVars map[string]string

	// Leave as nil to not do any files artifact expansion
	filesArtifactExpansion *files_artifacts_expansion.FilesArtifactsExpansion

	cpuAllocationMillicpus uint64

	memoryAllocationMegabytes uint64

	privateIPAddrPlaceholder string
}

func NewServiceConfig(
	containerImageName string,
	privatePorts map[string]*port_spec.PortSpec,
	publicPorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	filesArtifactExpansion *files_artifacts_expansion.FilesArtifactsExpansion,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
	privateIPAddrPlaceholder string) *ServiceConfig {
	return &ServiceConfig{
		containerImageName:        containerImageName,
		privatePorts:              privatePorts,
		publicPorts:               publicPorts,
		entrypointArgs:            entrypointArgs,
		cmdArgs:                   cmdArgs,
		envVars:                   envVars,
		filesArtifactExpansion:    filesArtifactExpansion,
		cpuAllocationMillicpus:    cpuAllocationMillicpus,
		memoryAllocationMegabytes: memoryAllocationMegabytes,
		privateIPAddrPlaceholder:  privateIPAddrPlaceholder,
	}
}

func (serviceConfig *ServiceConfig) GetContainerImageName() string {
	return serviceConfig.containerImageName
}

func (serviceConfig *ServiceConfig) GetPrivatePorts() map[string]*port_spec.PortSpec {
	return serviceConfig.privatePorts
}

func (serviceConfig *ServiceConfig) GetPublicPorts() map[string]*port_spec.PortSpec {
	return serviceConfig.publicPorts
}

func (serviceConfig *ServiceConfig) GetEntrypointArgs() []string {
	return serviceConfig.entrypointArgs
}

func (serviceConfig *ServiceConfig) GetCmdArgs() []string {
	return serviceConfig.cmdArgs
}

func (serviceConfig *ServiceConfig) GetEnvVars() map[string]string {
	return serviceConfig.envVars
}

func (serviceConfig *ServiceConfig) GetFilesArtifactsExpansion() *files_artifacts_expansion.FilesArtifactsExpansion {
	return serviceConfig.filesArtifactExpansion
}

func (serviceConfig *ServiceConfig) GetCPUAllocationMillicpus() uint64 {
	return serviceConfig.cpuAllocationMillicpus
}

func (serviceConfig *ServiceConfig) GetMemoryAllocationMegabytes() uint64 {
	return serviceConfig.memoryAllocationMegabytes
}

func (serviceConfig *ServiceConfig) GetPrivateIPAddrPlaceholder() string {
	return serviceConfig.privateIPAddrPlaceholder
}
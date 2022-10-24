package services

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
)

type ServiceConfigBuilder struct {
	containerImageName         string
	privatePorts               map[string]*kurtosis_core_rpc_api_bindings.Port
	publicPorts                map[string]*kurtosis_core_rpc_api_bindings.Port
	entrypointArgs             []string
	cmdArgs                    []string
	envVars                    map[string]string
	filesArtifactMountDirpaths map[string]string
	cpuAllocationMillicpus     uint64
	memoryAllocationMegabytes  uint64
	privateIPAddrPlaceholder   string
}

func NewServiceConfigBuilder(containerImageName string) *ServiceConfigBuilder {
	return &ServiceConfigBuilder{
		containerImageName:         containerImageName,
		privatePorts:               map[string]*kurtosis_core_rpc_api_bindings.Port{},
		publicPorts:                map[string]*kurtosis_core_rpc_api_bindings.Port{},
		entrypointArgs:             []string{},
		cmdArgs:                    []string{},
		envVars:                    map[string]string{},
		filesArtifactMountDirpaths: map[string]string{},
		cpuAllocationMillicpus:     0,
		memoryAllocationMegabytes:  0,
		privateIPAddrPlaceholder:   defaultPrivateIPAddrPlaceholder,
	}
}

func (builder *ServiceConfigBuilder) WithPrivatePorts(privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port) *ServiceConfigBuilder {
	builder.privatePorts = privatePorts
	return builder
}

func (builder *ServiceConfigBuilder) WithEntryPointArgs(entryPointArgs []string) *ServiceConfigBuilder {
	builder.entrypointArgs = entryPointArgs
	return builder
}

func (builder *ServiceConfigBuilder) WithCmdArgs(cmdArgs []string) *ServiceConfigBuilder {
	builder.cmdArgs = cmdArgs
	return builder
}

func (builder *ServiceConfigBuilder) WithEnvVars(envVars map[string]string) *ServiceConfigBuilder {
	builder.envVars = envVars
	return builder
}

func (builder *ServiceConfigBuilder) WithFilesArtifactMountDirpaths(filesArtifactMountDirpaths map[string]string) *ServiceConfigBuilder {
	builder.filesArtifactMountDirpaths = filesArtifactMountDirpaths
	return builder
}

func (builder *ServiceConfigBuilder) Build() *kurtosis_core_rpc_api_bindings.ServiceConfig {
	return binding_constructors.NewServiceConfig(
		builder.containerImageName,
		builder.privatePorts,
		builder.publicPorts,
		builder.entrypointArgs,
		builder.cmdArgs,
		builder.envVars,
		builder.filesArtifactMountDirpaths,
		builder.cpuAllocationMillicpus,
		builder.memoryAllocationMegabytes,
		builder.privateIPAddrPlaceholder,
	)
}

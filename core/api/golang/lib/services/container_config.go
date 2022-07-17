/*
 *    Copyright 2021 Kurtosis Technologies Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package services

// The UUID of an artifact containing files that should be mounted into a service container
type FilesArtifactUUID string

// ====================================================================================================
//                                    Config Object
// ====================================================================================================
// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
type ContainerConfig struct {
	image                        string
	usedPorts                    map[string]*PortSpec
	publicPorts                  map[string]*PortSpec //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	filesArtifactMountpoints     map[FilesArtifactUUID]string
	entrypointOverrideArgs       []string
	cmdOverrideArgs              []string
	environmentVariableOverrides map[string]string
	cpuAllocationMillicpus       uint64
	memoryAllocationMegabytes    uint64
}

func (config *ContainerConfig) GetImage() string {
	return config.image
}

func (config *ContainerConfig) GetUsedPorts() map[string]*PortSpec {
	return config.usedPorts
}

func (config *ContainerConfig) GetFilesArtifactMountpoints() map[FilesArtifactUUID]string {
	return config.filesArtifactMountpoints
}

func (config *ContainerConfig) GetEntrypointOverrideArgs() []string {
	return config.entrypointOverrideArgs
}

func (config *ContainerConfig) GetCmdOverrideArgs() []string {
	return config.cmdOverrideArgs
}

func (config *ContainerConfig) GetEnvironmentVariableOverrides() map[string]string {
	return config.environmentVariableOverrides
}

func (config *ContainerConfig) GetCPUAllocationMillicpus() uint64 {
	return config.cpuAllocationMillicpus
}

func (config *ContainerConfig) GetMemoryAllocationMegabytes() uint64 {
	return config.memoryAllocationMegabytes
}

//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
func (config *ContainerConfig) GetPublicPorts() map[string]*PortSpec {
	return config.publicPorts
}

// ====================================================================================================
//                                      Builder
// ====================================================================================================
// TODO Defensive copies on all these With... functions???
// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
type ContainerConfigBuilder struct {
	image                        string
	usedPorts                    map[string]*PortSpec
	publicPorts                  map[string]*PortSpec //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	filesArtifactMountpoints     map[FilesArtifactUUID]string
	entrypointOverrideArgs       []string
	cmdOverrideArgs              []string
	environmentVariableOverrides map[string]string
	cpuAllocationMillicpus       uint64
	memoryAllocationMegabytes    uint64
}

func NewContainerConfigBuilder(image string) *ContainerConfigBuilder {
	return &ContainerConfigBuilder{
		image:                        image,
		usedPorts:                    map[string]*PortSpec{},
		filesArtifactMountpoints:     map[FilesArtifactUUID]string{},
		entrypointOverrideArgs:       nil,
		cmdOverrideArgs:              nil,
		environmentVariableOverrides: map[string]string{},
		cpuAllocationMillicpus:       0,
		memoryAllocationMegabytes:    0,
	}
}

func (builder *ContainerConfigBuilder) WithUsedPorts(usedPorts map[string]*PortSpec) *ContainerConfigBuilder {
	builder.usedPorts = usedPorts
	return builder
}

func (builder *ContainerConfigBuilder) WithFiles(filesArtifactMountpoints map[FilesArtifactUUID]string) *ContainerConfigBuilder {
	builder.filesArtifactMountpoints = filesArtifactMountpoints
	return builder
}

func (builder *ContainerConfigBuilder) WithEntrypointOverride(args []string) *ContainerConfigBuilder {
	builder.entrypointOverrideArgs = args
	return builder
}

func (builder *ContainerConfigBuilder) WithCmdOverride(args []string) *ContainerConfigBuilder {
	builder.cmdOverrideArgs = args
	return builder
}

func (builder *ContainerConfigBuilder) WithEnvironmentVariableOverrides(envVars map[string]string) *ContainerConfigBuilder {
	builder.environmentVariableOverrides = envVars
	return builder
}

//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
func (builder *ContainerConfigBuilder) WithPublicPorts(publicPorts map[string]*PortSpec) *ContainerConfigBuilder {
	builder.publicPorts = publicPorts
	return builder
}

func (builder *ContainerConfigBuilder) WithCPUAllocationMillicpus(cpuAllocationMillicpus uint64) *ContainerConfigBuilder {
	builder.cpuAllocationMillicpus = cpuAllocationMillicpus
	return builder
}

func (builder *ContainerConfigBuilder) WithMemoryAllocationMegabytes(memoryAllocationMegabytes uint64) *ContainerConfigBuilder {
	builder.memoryAllocationMegabytes = memoryAllocationMegabytes
	return builder
}

func (builder *ContainerConfigBuilder) Build() *ContainerConfig {
	return &ContainerConfig{
		image:                        builder.image,
		usedPorts:                    builder.usedPorts,
		filesArtifactMountpoints:     builder.filesArtifactMountpoints,
		entrypointOverrideArgs:       builder.entrypointOverrideArgs,
		cmdOverrideArgs:              builder.cmdOverrideArgs,
		environmentVariableOverrides: builder.environmentVariableOverrides,
		//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
		publicPorts:      builder.publicPorts,
		cpuAllocationMillicpus:    builder.cpuAllocationMillicpus,
		memoryAllocationMegabytes: builder.memoryAllocationMegabytes,
	}
}

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

// The ID of an artifact containing files that should be mounted into a service container
type FilesArtifactID string

// ====================================================================================================
//                                    Config Object
// ====================================================================================================
// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
type ContainerConfig struct {
	image                        string
	usedPorts                   map[string]*PortSpec
	// TODO REMOVE
	oldFilesArtifactMountpoints map[FilesArtifactID]string
	filesArtifactMountpoints    map[FilesArtifactID]string
	entrypointOverrideArgs      []string
	cmdOverrideArgs              []string
	environmentVariableOverrides map[string]string
}

func (config *ContainerConfig) GetImage() string {
	return config.image
}

func (config *ContainerConfig) GetUsedPorts() map[string]*PortSpec {
	return config.usedPorts
}

func (config *ContainerConfig) GetFilesArtifactMountpoints() map[FilesArtifactID]string {
	return config.filesArtifactMountpoints
}

// TODO REMOVE
func (config *ContainerConfig) GetOldFilesArtifactMountpoints() map[FilesArtifactID]string {
	return config.oldFilesArtifactMountpoints
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

// ====================================================================================================
//                                      Builder
// ====================================================================================================
// TODO Defensive copies on all these With... functions???
// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
type ContainerConfigBuilder struct {
	image                        string
	usedPorts                   map[string]*PortSpec
	// TODO REMOVE
	oldFilesArtifactMountpoints map[FilesArtifactID]string
	filesArtifactMountpoints  map[FilesArtifactID]string
	entrypointOverrideArgs      []string
	cmdOverrideArgs              []string
	environmentVariableOverrides map[string]string
}

func NewContainerConfigBuilder(image string) *ContainerConfigBuilder {
	return &ContainerConfigBuilder{
		image:                        image,
		usedPorts:                    map[string]*PortSpec{},
		oldFilesArtifactMountpoints:  map[FilesArtifactID]string{},
		filesArtifactMountpoints:     map[FilesArtifactID]string{},
		entrypointOverrideArgs:       nil,
		cmdOverrideArgs:              nil,
		environmentVariableOverrides: map[string]string{},
	}
}

func (builder *ContainerConfigBuilder) WithUsedPorts(usedPorts map[string]*PortSpec) *ContainerConfigBuilder {
	builder.usedPorts = usedPorts
	return builder
}

// TODO REMOVE THIS
func (builder *ContainerConfigBuilder) WithFilesArtifacts(filesArtifactMountpoints map[FilesArtifactID]string) *ContainerConfigBuilder {
	builder.oldFilesArtifactMountpoints = filesArtifactMountpoints
	return builder
}

func (builder *ContainerConfigBuilder) WithFiles(filesArtifactMountpoints map[FilesArtifactID]string) *ContainerConfigBuilder {
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

func (builder *ContainerConfigBuilder) Build() *ContainerConfig {
	return &ContainerConfig{
		image:                        builder.image,
		usedPorts:                    builder.usedPorts,
		oldFilesArtifactMountpoints:  builder.oldFilesArtifactMountpoints,
		filesArtifactMountpoints:     builder.filesArtifactMountpoints,
		entrypointOverrideArgs:       builder.entrypointOverrideArgs,
		cmdOverrideArgs:              builder.cmdOverrideArgs,
		environmentVariableOverrides: builder.environmentVariableOverrides,
	}
}

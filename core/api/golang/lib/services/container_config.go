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
// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
type ContainerConfig struct {
	image                        string
	usedPortsSet                 map[string]bool
	filesArtifactMountpoints     map[FilesArtifactID]string
	entrypointOverrideArgs       []string
	cmdOverrideArgs              []string
	environmentVariableOverrides map[string]string
}

func (config *ContainerConfig) GetImage() string {
	return config.image
}

func (config *ContainerConfig) GetUsedPortsSet() map[string]bool {
	return config.usedPortsSet
}

func (config *ContainerConfig) GetFilesArtifactMountpoints() map[FilesArtifactID]string {
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

// ====================================================================================================
//                                      Builder
// ====================================================================================================
// TODO Defensive copies on all these With... functions???
// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
type ContainerConfigBuilder struct {
	image                        string
	usedPortsSet                 map[string]bool
	filesArtifactMountpoints     map[FilesArtifactID]string
	entrypointOverrideArgs       []string
	cmdOverrideArgs              []string
	environmentVariableOverrides map[string]string
}

func NewContainerConfigBuilder(image string) *ContainerConfigBuilder {
	return &ContainerConfigBuilder{
		image:                        image,
		usedPortsSet:                 map[string]bool{},
		filesArtifactMountpoints:     map[FilesArtifactID]string{},
		entrypointOverrideArgs:       nil,
		cmdOverrideArgs:              nil,
		environmentVariableOverrides: map[string]string{},
	}
}

func (builder *ContainerConfigBuilder) WithUsedPorts(usedPortsSet map[string]bool) *ContainerConfigBuilder {
	builder.usedPortsSet = usedPortsSet
	return builder
}

func (builder *ContainerConfigBuilder) WithFilesArtifacts(filesArtifactMountpoints map[FilesArtifactID]string) *ContainerConfigBuilder {
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
		usedPortsSet:                 builder.usedPortsSet,
		filesArtifactMountpoints:     builder.filesArtifactMountpoints,
		entrypointOverrideArgs:       builder.entrypointOverrideArgs,
		cmdOverrideArgs:              builder.cmdOverrideArgs,
		environmentVariableOverrides: builder.environmentVariableOverrides,
	}
}

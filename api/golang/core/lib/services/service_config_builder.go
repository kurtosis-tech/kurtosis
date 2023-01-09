package services

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"strings"
)

type FilesArtifactUUID string

const (
	defaultSubnetwork               = "default"
	defaultPrivateIPAddrPlaceholder = "KURTOSIS_IP_ADDR_PLACEHOLDER"
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
	subnetwork                 string
}

func NewServiceConfigBuilder(containerImageName string) *ServiceConfigBuilder {
	return &ServiceConfigBuilder{
		containerImageName:         containerImageName,
		privatePorts:               map[string]*kurtosis_core_rpc_api_bindings.Port{},
		publicPorts:                map[string]*kurtosis_core_rpc_api_bindings.Port{},
		entrypointArgs:             nil,
		cmdArgs:                    nil,
		envVars:                    map[string]string{},
		filesArtifactMountDirpaths: map[string]string{},
		cpuAllocationMillicpus:     0,
		memoryAllocationMegabytes:  0,
		privateIPAddrPlaceholder:   defaultPrivateIPAddrPlaceholder,
		subnetwork:                 defaultSubnetwork,
	}
}

func (builder *ServiceConfigBuilder) WithPrivatePorts(privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port) *ServiceConfigBuilder {
	builder.privatePorts = privatePorts
	return builder
}

func (builder *ServiceConfigBuilder) WithPublicPorts(publicPorts map[string]*kurtosis_core_rpc_api_bindings.Port) *ServiceConfigBuilder {
	builder.publicPorts = publicPorts
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

func (builder *ServiceConfigBuilder) WithPrivateIPAddressPlaceholder(privateIPAddrPlaceholder string) *ServiceConfigBuilder {
	if privateIPAddrPlaceholder == "" {
		privateIPAddrPlaceholder = defaultPrivateIPAddrPlaceholder
	}
	builder.privateIPAddrPlaceholder = privateIPAddrPlaceholder
	return builder
}

func (builder *ServiceConfigBuilder) WithSubnetwork(subnetwork string) *ServiceConfigBuilder {
	if subnetwork != "" {
		builder.subnetwork = subnetwork
	}
	return builder
}

func (builder *ServiceConfigBuilder) WithCpuAllocationMillicpus(cpuAllocationMillicpus uint64) *ServiceConfigBuilder {
	builder.cpuAllocationMillicpus = cpuAllocationMillicpus
	return builder
}

func (builder *ServiceConfigBuilder) WithMemoryAllocationMegabytes(memoryAllocationMegabytes uint64) *ServiceConfigBuilder {
	builder.memoryAllocationMegabytes = memoryAllocationMegabytes
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
		builder.subnetwork,
	)
}

func portToStarlark(port *kurtosis_core_rpc_api_bindings.Port) string {
	starlarkFields := []string{}
	starlarkFields = append(starlarkFields, fmt.Sprintf(`number=%d`, port.GetNumber()))
	if port.GetMaybeApplicationProtocol() != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`application_protocol="%s"`, port.GetMaybeApplicationProtocol()))
	}
	if port.GetTransportProtocol() != kurtosis_core_rpc_api_bindings.Port_TCP {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`transport_protocol="%s"`, port.GetTransportProtocol().String()))
	}
	return fmt.Sprintf("PortSpec(%s)", strings.Join(starlarkFields, ","))
}

func GetServiceConfigStarlark(
	containerImageName string,
	privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port,
	fileArtifactMountPoints map[string]string,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	subnetwork string,
	privateIpAddrPlaceholder string,
	cpuAllocationMillicpus int,
	memoryAllocationMegabytes int) string {
	starlarkFields := []string{}

	starlarkFields = append(starlarkFields, fmt.Sprintf(`image=%q`, containerImageName))

	portStrings := []string{}
	for portId, port := range privatePorts {
		portStrings = append(portStrings, fmt.Sprintf(`%q: %s`, portId, portToStarlark(port)))
	}
	starlarkFields = append(starlarkFields, fmt.Sprintf(`ports={%s}`, strings.Join(portStrings, ",")))

	fileStrings := []string{}
	for filePath, artifactName := range fileArtifactMountPoints {
		fileStrings = append(fileStrings, fmt.Sprintf(`%q: %q`, filePath, artifactName))
	}
	starlarkFields = append(starlarkFields, fmt.Sprintf(`files={%s}`, strings.Join(fileStrings, ",")))

	quotedEntrypointArgs := []string{}
	for _, entrypointArg := range entrypointArgs {
		quotedEntrypointArgs = append(quotedEntrypointArgs, fmt.Sprintf(`%q`, entrypointArg))
	}
	starlarkFields = append(starlarkFields, fmt.Sprintf(`entrypoint=[%s]`, strings.Join(quotedEntrypointArgs, ", ")))

	quotedCmdArgs := []string{}
	for _, cmdArg := range cmdArgs {
		quotedCmdArgs = append(quotedCmdArgs, fmt.Sprintf(`%q`, cmdArg))
	}
	starlarkFields = append(starlarkFields, fmt.Sprintf(`cmd=[%s]`, strings.Join(quotedCmdArgs, ", ")))

	envVarStrings := []string{}
	for envVar, envVarValue := range envVars {
		envVarStrings = append(envVarStrings, fmt.Sprintf(`%q: %q`, envVar, envVarValue))
	}
	starlarkFields = append(starlarkFields, fmt.Sprintf(`env_vars={%s}`, strings.Join(envVarStrings, ",")))

	if subnetwork != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`subnetwork=%q`, subnetwork))
	}
	if privateIpAddrPlaceholder != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`private_ip_address_placeholder=%q`, privateIpAddrPlaceholder))
	}
	if cpuAllocationMillicpus != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`cpu_allocation=%d`, cpuAllocationMillicpus))
	}
	if memoryAllocationMegabytes != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`memory_allocation=%d`, memoryAllocationMegabytes))
	}
	return fmt.Sprintf("ServiceConfig(%s)", strings.Join(starlarkFields, ","))
}

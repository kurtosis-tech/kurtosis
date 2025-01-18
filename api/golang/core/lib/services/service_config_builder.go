package services

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/kurtosis_core_rpc_api_bindings"
	"strings"
)

type FilesArtifactUUID string
type FileArtifactName string

func portToStarlark(port *kurtosis_core_rpc_api_bindings.Port) string {
	starlarkFields := []string{}
	starlarkFields = append(starlarkFields, fmt.Sprintf(`number=%d`, port.GetNumber()))
	if port.GetMaybeApplicationProtocol() != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`application_protocol="%s"`, port.GetMaybeApplicationProtocol()))
	}
	if port.GetTransportProtocol() != kurtosis_core_rpc_api_bindings.Port_TCP {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`transport_protocol="%s"`, port.GetTransportProtocol().String()))
	}
	if port.GetMaybeWaitTimeout() != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`wait="%s"`, port.GetMaybeWaitTimeout()))
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
	privateIpAddrPlaceholder string,
	cpuAllocationMillicpus int,
	memoryAllocationMegabytes int,
	minCpuMilliCores int,
	minMemoryMegaBytes int,
) string {
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

	if privateIpAddrPlaceholder != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`private_ip_address_placeholder=%q`, privateIpAddrPlaceholder))
	}
	if cpuAllocationMillicpus != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`cpu_allocation=%d`, cpuAllocationMillicpus))
	}
	if memoryAllocationMegabytes != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`memory_allocation=%d`, memoryAllocationMegabytes))
	}
	if minCpuMilliCores != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`min_cpu=%d`, minCpuMilliCores))
	}
	if minMemoryMegaBytes != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`min_memory=%d`, minMemoryMegaBytes))
	}

	return fmt.Sprintf("ServiceConfig(%s)", strings.Join(starlarkFields, ","))
}

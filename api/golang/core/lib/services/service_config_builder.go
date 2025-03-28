package services

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"strings"
)

type FilesArtifactUUID string
type FileArtifactName string

type Port struct {
	Number                   uint32 `json:"number" yaml:"number"`
	Transport                int    `json:"transport" yaml:"transport"` // e.g. "TCP", "UDP"
	MaybeApplicationProtocol string `json:"maybe_application_protocol,omitempty" yaml:"maybe_application_protocol,omitempty"`
	Wait                     string `json:"wait,omitempty" yaml:"wait,omitempty"`
}

type User struct {
	UID uint32 `json:"uid" yaml:"uid"`
	GID uint32 `json:"gid,omitempty" yaml:"gid,omitempty"`
}

type Toleration struct {
	Key               string `json:"key" yaml:"key"`
	Value             string `json:"value" yaml:"value"`
	Operator          string `json:"operator" yaml:"operator"`
	Effect            string `json:"effect" yaml:"effect"`
	TolerationSeconds int64  `json:"toleration_seconds" yaml:"toleration_seconds"`
}

type ServiceConfig struct {
	Image                       string            `json:"image" yaml:"image"`
	PrivatePorts                map[string]Port   `json:"ports,omitempty" yaml:"ports,omitempty"`
	PublicPorts                 map[string]Port   `json:"public_ports,omitempty" yaml:"public_ports,omitempty"`
	Files                       map[string]string `json:"files,omitempty" yaml:"files,omitempty"`
	Entrypoint                  []string          `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Cmd                         []string          `json:"cmd,omitempty" yaml:"cmd,omitempty"`
	EnvVars                     map[string]string `json:"env_vars,omitempty" yaml:"env_vars,omitempty"`
	PrivateIPAddressPlaceholder string            `json:"private_ip_address_placeholder,omitempty" yaml:"private_ip_address_placeholder,omitempty"`
	MaxMillicpus                uint32            `json:"max_cpu,omitempty" yaml:"max_cpu,omitempty"`
	MinMillicpus                uint32            `json:"min_cpu,omitempty" yaml:"min_cpu,omitempty"`
	MaxMemory                   uint32            `json:"max_memory,omitempty" yaml:"max_memory,omitempty"`
	MinMemory                   uint32            `json:"min_memory,omitempty" yaml:"min_memory,omitempty"`
	User                        *User             `json:"user,omitempty" yaml:"user,omitempty"`
	Tolerations                 []Toleration      `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	Labels                      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	NodeSelectors               map[string]string `json:"node_selectors,omitempty" yaml:"node_selectors,omitempty"`
	TiniEnabled                 *bool             `json:"tini_enabled,omitempty" yaml:"tini_enabled,omitempty"`
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
	if port.GetMaybeWaitTimeout() != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`wait="%s"`, port.GetMaybeWaitTimeout()))
	}
	return fmt.Sprintf("PortSpec(%s)", strings.Join(starlarkFields, ","))
}

func GetSimpleServiceConfigStarlark(
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

func GetFullServiceConfigStarlark(
	containerImageName string,
	privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port,
	fileArtifactMountPoints map[string]string,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	cpuAllocationMillicpus uint32,
	memoryAllocationMegabytes uint32,
	minCpuMilliCores uint32,
	minMemoryMegaBytes uint32,
	user *User,
	tolerations []Toleration,
	nodeSelectors map[string]string,
	labels map[string]string,
	tiniEnabled *bool,
	privateIpAddrPlaceholder string,
) string {
	starlarkFields := []string{}
	starlarkFields = append(starlarkFields, fmt.Sprintf(`image=%q`, containerImageName))

	// Ports
	portStrings := []string{}
	for portId, port := range privatePorts {
		portStrings = append(portStrings, fmt.Sprintf(`%q: %s`, portId, portToStarlark(port)))
	}
	if len(portStrings) > 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`ports={%s}`, strings.Join(portStrings, ",")))
	}

	// Files
	fileStrings := []string{}
	for filePath, artifactName := range fileArtifactMountPoints {
		fileStrings = append(fileStrings, fmt.Sprintf(`%q: %q`, filePath, artifactName))
	}
	if len(fileStrings) > 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`files={%s}`, strings.Join(fileStrings, ",")))
	}

	// Entrypoint
	if len(entrypointArgs) > 0 {
		quotedEntrypointArgs := []string{}
		for _, arg := range entrypointArgs {
			quotedEntrypointArgs = append(quotedEntrypointArgs, fmt.Sprintf(`%q`, arg))
		}
		starlarkFields = append(starlarkFields, fmt.Sprintf(`entrypoint=[%s]`, strings.Join(quotedEntrypointArgs, ", ")))
	}

	// Cmd
	if len(cmdArgs) > 0 {
		quotedCmdArgs := []string{}
		for _, arg := range cmdArgs {
			quotedCmdArgs = append(quotedCmdArgs, fmt.Sprintf(`%q`, arg))
		}
		starlarkFields = append(starlarkFields, fmt.Sprintf(`cmd=[%s]`, strings.Join(quotedCmdArgs, ", ")))
	}

	// Env Vars
	if len(envVars) > 0 {
		envVarStrings := []string{}
		for k, v := range envVars {
			envVarStrings = append(envVarStrings, fmt.Sprintf(`%q: %q`, k, v))
		}
		starlarkFields = append(starlarkFields, fmt.Sprintf(`env_vars={%s}`, strings.Join(envVarStrings, ",")))
	}

	// Optional simple fields
	if cpuAllocationMillicpus != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`max_cpu=%d`, cpuAllocationMillicpus))
	}
	if memoryAllocationMegabytes != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`max_memory=%d`, memoryAllocationMegabytes))
	}
	if minCpuMilliCores != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`min_cpu=%d`, minCpuMilliCores))
	}
	if minMemoryMegaBytes != 0 {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`min_memory=%d`, minMemoryMegaBytes))
	}

	if privateIpAddrPlaceholder != "" {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`private_ip_address_placeholder=%q`, privateIpAddrPlaceholder))
	}

	// User
	if user != nil {
		userStr := fmt.Sprintf("User(uid=%d", user.UID)
		if user.GID != 0 {
			userStr += fmt.Sprintf(", gid=%d", user.GID)
		}
		userStr += ")"
		starlarkFields = append(starlarkFields, fmt.Sprintf(`user=%s`, userStr))
	}

	// Tolerations
	if len(tolerations) > 0 {
		tolerationStrs := []string{}
		for _, t := range tolerations {
			tolerationStrs = append(tolerationStrs, fmt.Sprintf(
				`Toleration(key=%q, value=%q, operator=%q, effect=%q, toleration_seconds=%d)`,
				t.Key, t.Value, t.Operator, t.Effect, t.TolerationSeconds,
			))
		}
		starlarkFields = append(starlarkFields, fmt.Sprintf(`tolerations=[%s]`, strings.Join(tolerationStrs, ", ")))
	}

	// Node selectors
	if len(nodeSelectors) > 0 {
		selectorStrs := []string{}
		for k, v := range nodeSelectors {
			selectorStrs = append(selectorStrs, fmt.Sprintf(`%q: %q`, k, v))
		}
		starlarkFields = append(starlarkFields, fmt.Sprintf(`node_selectors={%s}`, strings.Join(selectorStrs, ", ")))
	}

	// Labels
	if len(labels) > 0 {
		labelStrs := []string{}
		for k, v := range labels {
			labelStrs = append(labelStrs, fmt.Sprintf(`%q: %q`, k, v))
		}
		starlarkFields = append(starlarkFields, fmt.Sprintf(`labels={%s}`, strings.Join(labelStrs, ", ")))
	}

	// Tini
	if tiniEnabled != nil {
		starlarkFields = append(starlarkFields, fmt.Sprintf(`tini_enabled=%t`, *tiniEnabled))
	}

	return fmt.Sprintf("ServiceConfig(%s)", strings.Join(starlarkFields, ", "))
}

func ConvertJsonPortToApiPort(jsonPorts map[string]Port) map[string]*kurtosis_core_rpc_api_bindings.Port {
	apiPorts := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	for portId, port := range jsonPorts {
		apiPort := &kurtosis_core_rpc_api_bindings.Port{
			Number:                   port.Number,
			TransportProtocol:        kurtosis_core_rpc_api_bindings.Port_TransportProtocol(port.Transport),
			MaybeApplicationProtocol: port.MaybeApplicationProtocol,
			MaybeWaitTimeout:         port.Wait,
			Locked:                   nil,
			Alias:                    nil,
		}
		apiPorts[portId] = apiPort
	}
	return apiPorts
}

func ConvertApiPortToJsonPort(apiPorts map[string]*kurtosis_core_rpc_api_bindings.Port) map[string]Port {
	jsonPorts := map[string]Port{}
	for portId, apiPort := range apiPorts {
		jsonPort := Port{
			Number:                   apiPort.GetNumber(),
			Transport:                int(apiPort.TransportProtocol),
			MaybeApplicationProtocol: apiPort.GetMaybeApplicationProtocol(),
			Wait:                     apiPort.GetMaybeWaitTimeout(),
		}
		jsonPorts[portId] = jsonPort
	}
	return jsonPorts
}

func ConvertApiFilesArtifactsToJsonFiles(serviceDirPathsToFilesArtifactsList map[string]*kurtosis_core_rpc_api_bindings.FilesArtifactsList) map[string]string {
	serviceDirPathsToFilesArtifacts := map[string]string{}
	for serviceDirPath, filesArtifactsList := range serviceDirPathsToFilesArtifactsList {
		filesArtifactsIdentifers := filesArtifactsList.GetFilesArtifactsIdentifiers()
		serviceDirPathsToFilesArtifacts[serviceDirPath] = filesArtifactsIdentifers[0]
	}
	return serviceDirPathsToFilesArtifacts
}

func ConvertApiUserToJsonUser(user *kurtosis_core_rpc_api_bindings.User) *User {
	return &User{
		UID: user.GetUid(),
		GID: user.GetGid(),
	}
}

func ConvertApiTolerationsToJsonTolerations(tolerations []*kurtosis_core_rpc_api_bindings.Toleration) []Toleration {
	jsonTolerations := []Toleration{}
	for _, apiToleration := range tolerations {
		jsonTolerations = append(jsonTolerations, Toleration{
			Key:               apiToleration.GetKey(),
			Value:             apiToleration.GetValue(),
			Operator:          apiToleration.GetOperator(),
			Effect:            apiToleration.GetEffect(),
			TolerationSeconds: apiToleration.GetTolerationSeconds(),
		})
	}
	return jsonTolerations
}

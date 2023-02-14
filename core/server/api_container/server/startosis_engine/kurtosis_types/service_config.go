package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"strings"
)

const (
	ServiceConfigTypeName = "ServiceConfig"

	serviceConfigImageAttr                       = "image"
	serviceConfigPortsAttr                       = "ports"
	serviceConfigPublicPortsAttr                 = "public_ports"
	serviceConfigFilesAttr                       = "files"
	serviceConfigEntrypointAttr                  = "entrypoint"
	serviceConfigCmdAttr                         = "cmd"
	serviceConfigEnvVarsAttr                     = "env_vars"
	serviceConfigPrivateIpAddressPlaceholderAttr = "private_ip_address_placeholder"
	serviceConfigSubnetworkAttr                  = "subnetwork"
	serviceConfigCpuAllocationAttr               = "cpu_allocation"
	serviceConfigMemoryAllocationAttr            = "memory_allocation"

	emptyPrivateIpAddressPlaceholderValue = ""
	emptySubnetworkValue                  = ""

	argumentPatternStr = "%s=%s"
)

var (
	emptyCpuAllocationValue    starlark.Int
	emptyMemoryAllocationValue starlark.Int
)

// ServiceConfig A starlark.Value that represents a service config used in the add_service instruction
type ServiceConfig struct {
	image                       starlark.String
	ports                       *starlark.Dict
	publicPorts                 *starlark.Dict
	files                       *starlark.Dict
	entrypoint                  *starlark.List
	cmd                         *starlark.List
	envVars                     *starlark.Dict
	privateIpAddressPlaceholder *starlark.String
	subnetwork                  *starlark.String
	cpuAllocation               *starlark.Int
	memoryAllocation            *starlark.Int
}

func NewServiceConfig(image starlark.String,
	ports *starlark.Dict,
	publicPorts *starlark.Dict,
	files *starlark.Dict,
	entrypoint *starlark.List,
	cmd *starlark.List,
	envVars *starlark.Dict,
	privateIpAddressPlaceholder *starlark.String,
	subnetwork *starlark.String,
	cpuAllocation *starlark.Int,
	memoryAllocation *starlark.Int) *ServiceConfig {
	return &ServiceConfig{
		image:                       image,
		ports:                       ports,
		publicPorts:                 publicPorts,
		entrypoint:                  entrypoint,
		files:                       files,
		cmd:                         cmd,
		envVars:                     envVars,
		privateIpAddressPlaceholder: privateIpAddressPlaceholder,
		subnetwork:                  subnetwork,
		cpuAllocation:               cpuAllocation,
		memoryAllocation:            memoryAllocation,
	}
}

func MakeServiceConfig(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		image                       starlark.String
		ports                       *starlark.Dict
		publicPorts                 *starlark.Dict
		files                       *starlark.Dict
		entrypoint                  *starlark.List
		cmd                         *starlark.List
		envVars                     *starlark.Dict
		privateIpAddressPlaceholder starlark.String
		subnetwork                  starlark.String
		cpuAllocation               starlark.Int
		memoryAllocation            starlark.Int
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		serviceConfigImageAttr, &image,
		MakeOptional(serviceConfigPortsAttr), &ports,
		MakeOptional(serviceConfigPublicPortsAttr), &publicPorts,
		MakeOptional(serviceConfigFilesAttr), &files,
		MakeOptional(serviceConfigEntrypointAttr), &entrypoint,
		MakeOptional(serviceConfigCmdAttr), &cmd,
		MakeOptional(serviceConfigEnvVarsAttr), &envVars,
		MakeOptional(serviceConfigPrivateIpAddressPlaceholderAttr), &privateIpAddressPlaceholder,
		MakeOptional(serviceConfigSubnetworkAttr), &subnetwork,
		MakeOptional(serviceConfigCpuAllocationAttr), &cpuAllocation,
		MakeOptional(serviceConfigMemoryAllocationAttr), &memoryAllocation,
	); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Cannot construct '%s' from the provided arguments.", ServiceConfigTypeName)
	}
	var privateIpAddressPlaceholderMaybe *starlark.String
	if privateIpAddressPlaceholder != emptyPrivateIpAddressPlaceholderValue {
		privateIpAddressPlaceholderMaybe = &privateIpAddressPlaceholder
	}
	var subnetworkMaybe *starlark.String
	if subnetwork != emptySubnetworkValue {
		subnetworkMaybe = &subnetwork
	}
	var cpuAllocationMaybe *starlark.Int
	if cpuAllocation != emptyCpuAllocationValue {
		cpuAllocationMaybe = &cpuAllocation
	}
	var memoryAllocationMaybe *starlark.Int
	if memoryAllocation != emptyMemoryAllocationValue {
		memoryAllocationMaybe = &memoryAllocation
	}
	return NewServiceConfig(image, ports, publicPorts, files, entrypoint, cmd, envVars, privateIpAddressPlaceholderMaybe, subnetworkMaybe, cpuAllocationMaybe, memoryAllocationMaybe), nil
}

// String the starlark.Value interface
func (serviceConfig *ServiceConfig) String() string {
	args := []string{
		// add image arg straight as it is required
		fmt.Sprintf(argumentPatternStr, serviceConfigImageAttr, serviceConfig.image.String()),
	}
	if serviceConfig.ports != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigPortsAttr, serviceConfig.ports.String()))
	}
	if serviceConfig.publicPorts != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigPublicPortsAttr, serviceConfig.publicPorts.String()))
	}
	if serviceConfig.files != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigFilesAttr, serviceConfig.files.String()))
	}
	if serviceConfig.entrypoint != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigEntrypointAttr, serviceConfig.entrypoint.String()))
	}
	if serviceConfig.cmd != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigCmdAttr, serviceConfig.cmd.String()))
	}
	if serviceConfig.envVars != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigEnvVarsAttr, serviceConfig.envVars.String()))
	}
	if serviceConfig.privateIpAddressPlaceholder != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigPrivateIpAddressPlaceholderAttr, serviceConfig.privateIpAddressPlaceholder.String()))
	}
	if serviceConfig.subnetwork != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigSubnetworkAttr, serviceConfig.subnetwork.String()))
	}
	if serviceConfig.cpuAllocation != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigCpuAllocationAttr, serviceConfig.cpuAllocation.String()))
	}
	if serviceConfig.memoryAllocation != nil {
		args = append(args, fmt.Sprintf(argumentPatternStr, serviceConfigMemoryAllocationAttr, serviceConfig.memoryAllocation.String()))
	}
	return fmt.Sprintf("%s(%s)", ServiceConfigTypeName, strings.Join(args, ", "))
}

// Type implements the starlark.Value interface
func (serviceConfig *ServiceConfig) Type() string {
	return ServiceConfigTypeName
}

// Freeze implements the starlark.Value interface
func (serviceConfig *ServiceConfig) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (serviceConfig *ServiceConfig) Truth() starlark.Bool {
	// image is the only required attribute
	return serviceConfig.image != ""
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed
func (serviceConfig *ServiceConfig) Hash() (uint32, error) {
	return 0, startosis_errors.NewInterpretationError("unhashable type: '%s'", ServiceConfigTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (serviceConfig *ServiceConfig) Attr(name string) (starlark.Value, error) {
	switch name {
	case serviceConfigImageAttr:
		return serviceConfig.image, nil
	case serviceConfigPortsAttr:
		if serviceConfig.ports == nil {
			break
		}
		return serviceConfig.ports, nil
	case serviceConfigPublicPortsAttr:
		if serviceConfig.publicPorts == nil {
			break
		}
		return serviceConfig.publicPorts, nil
	case serviceConfigFilesAttr:
		if serviceConfig.files == nil {
			break
		}
		return serviceConfig.files, nil
	case serviceConfigEntrypointAttr:
		if serviceConfig.entrypoint == nil {
			break
		}
		return serviceConfig.entrypoint, nil
	case serviceConfigCmdAttr:
		if serviceConfig.cmd == nil {
			break
		}
		return serviceConfig.cmd, nil
	case serviceConfigEnvVarsAttr:
		if serviceConfig.envVars == nil {
			break
		}
		return serviceConfig.envVars, nil
	case serviceConfigPrivateIpAddressPlaceholderAttr:
		if serviceConfig.privateIpAddressPlaceholder == nil {
			break
		}
		return *serviceConfig.privateIpAddressPlaceholder, nil
	case serviceConfigSubnetworkAttr:
		if serviceConfig.subnetwork == nil {
			break
		}
		return *serviceConfig.subnetwork, nil
	case serviceConfigCpuAllocationAttr:
		if serviceConfig.cpuAllocation == nil {
			break
		}
		return *serviceConfig.cpuAllocation, nil
	case serviceConfigMemoryAllocationAttr:
		if serviceConfig.memoryAllocation == nil {
			break
		}
		return *serviceConfig.memoryAllocation, nil
	}
	return nil, startosis_errors.NewInterpretationError("'%s' has no attribute '%s'", ServiceConfigTypeName, name)
}

// AttrNames implements the starlark.HasAttrs interface.
func (serviceConfig *ServiceConfig) AttrNames() []string {
	attrs := []string{
		serviceConfigImageAttr, // only required attribute
	}
	if serviceConfig.ports != nil {
		attrs = append(attrs, serviceConfigPortsAttr)
	}
	if serviceConfig.publicPorts != nil {
		attrs = append(attrs, serviceConfigPublicPortsAttr)
	}
	if serviceConfig.files != nil {
		attrs = append(attrs, serviceConfigFilesAttr)
	}
	if serviceConfig.entrypoint != nil {
		attrs = append(attrs, serviceConfigEntrypointAttr)
	}
	if serviceConfig.cmd != nil {
		attrs = append(attrs, serviceConfigCmdAttr)
	}
	if serviceConfig.envVars != nil {
		attrs = append(attrs, serviceConfigEnvVarsAttr)
	}
	if serviceConfig.privateIpAddressPlaceholder != nil {
		attrs = append(attrs, serviceConfigPrivateIpAddressPlaceholderAttr)
	}
	if serviceConfig.subnetwork != nil {
		attrs = append(attrs, serviceConfigSubnetworkAttr)
	}
	if serviceConfig.cpuAllocation != nil {
		attrs = append(attrs, serviceConfigCpuAllocationAttr)
	}
	if serviceConfig.memoryAllocation != nil {
		attrs = append(attrs, serviceConfigMemoryAllocationAttr)
	}
	return attrs
}

func (serviceConfig *ServiceConfig) ToKurtosisType() (*kurtosis_core_rpc_api_bindings.ServiceConfig, *startosis_errors.InterpretationError) {
	builder := services.NewServiceConfigBuilder(serviceConfig.image.GoString())

	if serviceConfig.ports != nil && serviceConfig.ports.Len() > 0 {
		privatePorts := make(map[string]*kurtosis_core_rpc_api_bindings.Port, serviceConfig.ports.Len())
		for _, portItem := range serviceConfig.ports.Items() {
			portKey, portValue, interpretationError := convertPortMapEntry(portItem[0], portItem[1], serviceConfig.ports)
			if interpretationError != nil {
				return nil, interpretationError
			}
			privatePorts[portKey] = portValue
		}
		builder.WithPrivatePorts(privatePorts)
	}

	if serviceConfig.publicPorts != nil && serviceConfig.publicPorts.Len() > 0 {
		publicPorts := make(map[string]*kurtosis_core_rpc_api_bindings.Port, serviceConfig.publicPorts.Len())
		for _, portItem := range serviceConfig.publicPorts.Items() {
			portKey, portValue, interpretationError := convertPortMapEntry(portItem[0], portItem[1], serviceConfig.publicPorts)
			if interpretationError != nil {
				return nil, interpretationError
			}
			publicPorts[portKey] = portValue
		}
		builder.WithPublicPorts(publicPorts)
	}

	if serviceConfig.files != nil && serviceConfig.files.Len() > 0 {
		filesArtifactMountDirpaths, interpretationErr := SafeCastToMapStringString(serviceConfig.files, serviceConfigFilesAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builder.WithFilesArtifactMountDirpaths(filesArtifactMountDirpaths)
	}

	if serviceConfig.entrypoint != nil && serviceConfig.entrypoint.Len() > 0 {
		entryPointArgs, interpretationErr := SafeCastToStringSlice(serviceConfig.entrypoint, serviceConfigEntrypointAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builder.WithEntryPointArgs(entryPointArgs)
	}

	if serviceConfig.cmd != nil && serviceConfig.cmd.Len() > 0 {
		cmdArgs, interpretationErr := SafeCastToStringSlice(serviceConfig.cmd, serviceConfigCmdAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builder.WithCmdArgs(cmdArgs)
	}

	if serviceConfig.envVars != nil && serviceConfig.envVars.Len() > 0 {
		envVars, interpretationErr := SafeCastToMapStringString(serviceConfig.envVars, serviceConfigEnvVarsAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builder.WithEnvVars(envVars)
	}

	if serviceConfig.privateIpAddressPlaceholder != nil && serviceConfig.privateIpAddressPlaceholder.GoString() != "" {
		builder.WithPrivateIPAddressPlaceholder(serviceConfig.privateIpAddressPlaceholder.GoString())
	}

	if serviceConfig.subnetwork != nil && serviceConfig.subnetwork.GoString() != "" {
		builder.WithSubnetwork(serviceConfig.subnetwork.GoString())
	}

	if serviceConfig.cpuAllocation != nil {
		cpuAllocation, ok := serviceConfig.cpuAllocation.Uint64()
		if !ok {
			return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", serviceConfigCpuAllocationAttr, serviceConfig.cpuAllocation)
		}
		builder.WithCpuAllocationMillicpus(cpuAllocation)
	}

	if serviceConfig.memoryAllocation != nil {
		memoryAllocation, ok := serviceConfig.memoryAllocation.Uint64()
		if !ok {
			return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", serviceConfigMemoryAllocationAttr, serviceConfig.memoryAllocation)
		}
		builder.WithMemoryAllocationMegabytes(memoryAllocation)
	}

	return builder.Build(), nil
}

func convertPortMapEntry(key starlark.Value, value starlark.Value, dictForLogging *starlark.Dict) (string, *kurtosis_core_rpc_api_bindings.Port, *startosis_errors.InterpretationError) {
	keyStr, ok := key.(starlark.String)
	if !ok {
		return "", nil, startosis_errors.NewInterpretationError("Unable to convert key of '%s' dictionary '%v' to string", serviceConfigPortsAttr, dictForLogging)
	}
	valuePortSpec, ok := value.(*PortSpec)
	if !ok {
		return "", nil, startosis_errors.NewInterpretationError("Unable to convert value of '%s' dictionary '%v' to a port object", serviceConfigPortsAttr, dictForLogging)
	}
	return keyStr.GoString(), valuePortSpec.ToKurtosisType(), nil
}

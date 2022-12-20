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

	emptyPrivateIpAddressPlaceholderValue = ""
	emptySubnetworkValue                  = ""

	argumentPatternStr = "%s=%s"
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
}

func NewServiceConfig(image starlark.String,
	ports *starlark.Dict,
	publicPorts *starlark.Dict,
	files *starlark.Dict,
	entrypoint *starlark.List,
	cmd *starlark.List,
	envVars *starlark.Dict,
	privateIpAddressPlaceholder *starlark.String,
	subnetwork *starlark.String) *ServiceConfig {
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
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		serviceConfigImageAttr, &image,
		makeOptional(serviceConfigPortsAttr), &ports,
		makeOptional(serviceConfigPublicPortsAttr), &publicPorts,
		makeOptional(serviceConfigFilesAttr), &files,
		makeOptional(serviceConfigEntrypointAttr), &entrypoint,
		makeOptional(serviceConfigCmdAttr), &cmd,
		makeOptional(serviceConfigEnvVarsAttr), &envVars,
		makeOptional(serviceConfigPrivateIpAddressPlaceholderAttr), &privateIpAddressPlaceholder,
		makeOptional(serviceConfigSubnetworkAttr), &subnetwork,
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
	return NewServiceConfig(image, ports, publicPorts, files, entrypoint, cmd, envVars, privateIpAddressPlaceholderMaybe, subnetworkMaybe), nil
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
		return serviceConfig.ports, nil
	case serviceConfigPublicPortsAttr:
		return serviceConfig.publicPorts, nil
	case serviceConfigFilesAttr:
		return serviceConfig.files, nil
	case serviceConfigEntrypointAttr:
		return serviceConfig.entrypoint, nil
	case serviceConfigCmdAttr:
		return serviceConfig.cmd, nil
	case serviceConfigEnvVarsAttr:
		return serviceConfig.envVars, nil
	case serviceConfigPrivateIpAddressPlaceholderAttr:
		return serviceConfig.privateIpAddressPlaceholder, nil
	case serviceConfigSubnetworkAttr:
		return serviceConfig.subnetwork, nil
	default:
		return nil, startosis_errors.NewInterpretationError("'%s' has no attribute '%s'", ServiceConfigTypeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (serviceConfig *ServiceConfig) AttrNames() []string {
	return []string{
		serviceConfigImageAttr,
		serviceConfigPortsAttr,
		serviceConfigPublicPortsAttr,
		serviceConfigFilesAttr,
		serviceConfigEntrypointAttr,
		serviceConfigCmdAttr,
		serviceConfigEnvVarsAttr,
		serviceConfigPrivateIpAddressPlaceholderAttr,
		serviceConfigSubnetworkAttr,
	}
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

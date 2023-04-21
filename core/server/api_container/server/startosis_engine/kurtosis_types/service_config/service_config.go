package service_config

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_warning"
	"go.starlark.net/starlark"
	"math"
)

const (
	ServiceConfigTypeName = "ServiceConfig"

	ImageAttr                       = "image"
	PortsAttr                       = "ports"
	PublicPortsAttr                 = "public_ports"
	FilesAttr                       = "files"
	EntrypointAttr                  = "entrypoint"
	CmdAttr                         = "cmd"
	EnvVarsAttr                     = "env_vars"
	PrivateIpAddressPlaceholderAttr = "private_ip_address_placeholder"
	SubnetworkAttr                  = "subnetwork"
	CpuAllocationAttr               = "cpu_allocation"
	MemoryAllocationAttr            = "memory_allocation"
	ReadyConditionsAttr             = "ready_conditions"
)

func NewServiceConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ServiceConfigTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ImageAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ImageAttr)
					},
				},
				{
					Name:              PortsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              PublicPortsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              FilesAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              EntrypointAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator:         nil,
				},
				{
					Name:              CmdAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator:         nil,
				},
				{
					Name:              EnvVarsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              PrivateIpAddressPlaceholderAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, PrivateIpAddressPlaceholderAttr)
					},
				},
				{
					Name:              SubnetworkAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, SubnetworkAttr)
					},
				},
				{
					Name:              CpuAllocationAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, CpuAllocationAttr, 0, math.MaxUint64)
					},
				},
				{
					Name:              MemoryAllocationAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, MemoryAllocationAttr, 6, math.MaxUint64)
					},
				},
				{
					Name:              ReadyConditionsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*ReadyCondition],
					Validator:         nil,
				},
			},
		},

		Instantiate: instantiateServiceConfig,
	}
}

func instantiateServiceConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ServiceConfigTypeName, arguments)
	startosis_warning.Printf("Service Config Warning!!")
	if err != nil {
		return nil, err
	}
	return &ServiceConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// ServiceConfig is a starlark.Value that represents a service config used in the add_service instruction
type ServiceConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *ServiceConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ServiceConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (config *ServiceConfig) ToKurtosisType() (*kurtosis_core_rpc_api_bindings.ServiceConfig, *startosis_errors.InterpretationError) {
	image, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](config.KurtosisValueTypeDefault, ImageAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			SubnetworkAttr, ServiceConfigTypeName)
	}

	builder := services.NewServiceConfigBuilder(image.GoString())

	portsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, PortsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && portsStarlark.Len() > 0 {
		ports := make(map[string]*kurtosis_core_rpc_api_bindings.Port, portsStarlark.Len())
		for _, portItem := range portsStarlark.Items() {
			portKey, portValue, interpretationError := convertPortMapEntry(PortsAttr, portItem[0], portItem[1], portsStarlark)
			if interpretationError != nil {
				return nil, interpretationError
			}
			ports[portKey] = portValue
		}
		builder.WithPrivatePorts(ports)
	}

	publicPortsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, PublicPortsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && publicPortsStarlark.Len() > 0 {
		publicPorts := make(map[string]*kurtosis_core_rpc_api_bindings.Port, publicPortsStarlark.Len())
		for _, portItem := range publicPortsStarlark.Items() {
			portKey, portValue, interpretationError := convertPortMapEntry(PublicPortsAttr, portItem[0], portItem[1], publicPortsStarlark)
			if interpretationError != nil {
				return nil, interpretationError
			}
			publicPorts[portKey] = portValue
		}
		builder.WithPublicPorts(publicPorts)
	}

	filesStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, FilesAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && filesStarlark.Len() > 0 {
		filesArtifactMountDirpaths, interpretationErr := kurtosis_types.SafeCastToMapStringString(filesStarlark, FilesAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builder.WithFilesArtifactMountDirpaths(filesArtifactMountDirpaths)
	}

	entrypointStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](config.KurtosisValueTypeDefault, EntrypointAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && entrypointStarlark.Len() > 0 {
		entryPointArgs, interpretationErr := kurtosis_types.SafeCastToStringSlice(entrypointStarlark, EntrypointAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builder.WithEntryPointArgs(entryPointArgs)
	}

	cmdStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](config.KurtosisValueTypeDefault, CmdAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && cmdStarlark.Len() > 0 {
		cmdArgs, interpretationErr := kurtosis_types.SafeCastToStringSlice(cmdStarlark, CmdAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builder.WithCmdArgs(cmdArgs)
	}

	envVarsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](config.KurtosisValueTypeDefault, EnvVarsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && envVarsStarlark.Len() > 0 {
		envVars, interpretationErr := kurtosis_types.SafeCastToMapStringString(envVarsStarlark, EnvVarsAttr)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builder.WithEnvVars(envVars)
	}

	privateIpAddressPlaceholderStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](config.KurtosisValueTypeDefault, PrivateIpAddressPlaceholderAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && privateIpAddressPlaceholderStarlark.GoString() != "" {
		builder.WithPrivateIPAddressPlaceholder(privateIpAddressPlaceholderStarlark.GoString())
	}

	subnetworkStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](config.KurtosisValueTypeDefault, SubnetworkAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found && subnetworkStarlark.GoString() != "" {
		builder.WithSubnetwork(subnetworkStarlark.GoString())
	}

	cpuAllocationStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](config.KurtosisValueTypeDefault, CpuAllocationAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found {
		cpuAllocation, ok := cpuAllocationStarlark.Uint64()
		if !ok {
			return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", CpuAllocationAttr, cpuAllocationStarlark)
		}
		builder.WithCpuAllocationMillicpus(cpuAllocation)
	}

	memoryAllocationStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](config.KurtosisValueTypeDefault, MemoryAllocationAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if found {
		memoryAllocation, ok := memoryAllocationStarlark.Uint64()
		if !ok {
			return nil, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", MemoryAllocationAttr, memoryAllocationStarlark)
		}
		builder.WithMemoryAllocationMegabytes(memoryAllocation)
	}

	return builder.Build(), nil
}

func (config *ServiceConfig) GetReadyCondition() (*ReadyCondition, *startosis_errors.InterpretationError) {
	readyConditions, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*ReadyCondition](config.KurtosisValueTypeDefault, ReadyConditionsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	return readyConditions, nil
}

func convertPortMapEntry(attrNameForLogging string, key starlark.Value, value starlark.Value, dictForLogging *starlark.Dict) (string, *kurtosis_core_rpc_api_bindings.Port, *startosis_errors.InterpretationError) {
	keyStr, ok := key.(starlark.String)
	if !ok {
		return "", nil, startosis_errors.NewInterpretationError("Unable to convert key of '%s' dictionary '%v' to string", attrNameForLogging, dictForLogging)
	}
	valuePortSpec, ok := value.(*port_spec.PortSpec)
	if !ok {
		return "", nil, startosis_errors.NewInterpretationError("Unable to convert value of '%s' dictionary '%v' to a port object", attrNameForLogging, dictForLogging)
	}
	apiPortSpec, interpretationErr := valuePortSpec.ToKurtosisType()
	if interpretationErr != nil {
		return "", nil, interpretationErr
	}
	return keyStr.GoString(), apiPortSpec, nil
}

package kurtosis_instruction

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"reflect"
	"strings"
)

const (
	serviceIdArgName     = "service_id"
	serviceConfigArgName = "service_config"
	commandsArgKey       = "command"

	containerImageNameKey = "container_image_name"
	usedPortsKey          = "used_ports"

	portNumberKey   = "number"
	portProtocolKey = "protocol"

	maxPortNumber = 65535
)

func ParseServiceId(serviceIdRaw starlark.String) (service.ServiceID, *startosis_errors.InterpretationError) {
	// TODO(gb): maybe prohibit certain characters for service ids
	serviceId, interpretationErr := safeCastToString(serviceIdRaw, serviceIdArgName)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if len(serviceId) == 0 {
		return "", startosis_errors.NewInterpretationError("Service ID cannot be empty")
	}
	return service.ServiceID(serviceId), nil
}

func ParseServiceConfigArg(serviceConfig *starlarkstruct.Struct) (*kurtosis_core_rpc_api_bindings.ServiceConfig, *startosis_errors.InterpretationError) {
	containerImageName, interpretationErr := parseServiceConfigContainerImageName(serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	privatePorts, interpretationErr := parseServiceConfigPrivatePorts(serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	builtConfig := services.NewServiceConfigBuilder(containerImageName).WithPrivatePorts(
		privatePorts,
	).Build()

	return builtConfig, nil
}

func ParseCommand(commandsRaw *starlark.List) ([]string, *startosis_errors.InterpretationError) {
	commandArgs, interpretationErr := safeCastToStringSlice(commandsRaw, commandsArgKey)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if len(commandArgs) == 0 {
		return nil, startosis_errors.NewInterpretationError("Command cannot be empty")
	}
	return commandArgs, nil
}

func parseServiceConfigContainerImageName(serviceConfig *starlarkstruct.Struct) (string, *startosis_errors.InterpretationError) {
	// containerImageName should be a simple string
	containerImageName, interpretationErr := extractStringValue(serviceConfig, containerImageNameKey, serviceConfigArgName)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return containerImageName, nil
}

func parseServiceConfigPrivatePorts(serviceConfig *starlarkstruct.Struct) (map[string]*kurtosis_core_rpc_api_bindings.Port, *startosis_errors.InterpretationError) {
	privatePortsRawArg, err := serviceConfig.Attr(usedPortsKey)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Missing `%s` as part of the service config", usedPortsKey))
	}
	privatePortsArg, ok := privatePortsRawArg.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Argument `%s` is expected to be a dictionary", usedPortsKey))
	}

	var privatePorts = make(map[string]*kurtosis_core_rpc_api_bindings.Port)
	for _, portNameRaw := range privatePortsArg.Keys() {
		portDefinitionRaw, found, err := privatePortsArg.Get(portNameRaw)
		if !found || err != nil {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Unable to find a value in a dict associated with a key that exists (key = '%s') - this is a product bug", portNameRaw))
		}

		portName, interpretationErr := safeCastToString(portNameRaw, usedPortsKey)
		if interpretationErr != nil {
			return nil, interpretationErr
		}

		portDefinition, ok := portDefinitionRaw.(*starlarkstruct.Struct)
		if !ok {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Port definition `%s` is expected to be a struct", portNameRaw))
		}

		port, interpretationErr := parsePort(portDefinition)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		privatePorts[portName] = port
	}
	return privatePorts, nil
}

func parsePort(portArg *starlarkstruct.Struct) (*kurtosis_core_rpc_api_bindings.Port, *startosis_errors.InterpretationError) {
	portNumber, interpretationErr := extractUint32Value(portArg, portNumberKey, usedPortsKey)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if portNumber > maxPortNumber {
		return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Port number should be less than or equal to %d", maxPortNumber))
	}

	protocolRaw, interpretationErr := extractStringValue(portArg, portProtocolKey, usedPortsKey)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	protocol, interpretationErr := parsePortProtocol(protocolRaw)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return binding_constructors.NewPort(portNumber, protocol), nil
}

func parsePortProtocol(portProtocol string) (kurtosis_core_rpc_api_bindings.Port_Protocol, *startosis_errors.InterpretationError) {
	parsedPortProtocol, err := port_spec.PortProtocolString(portProtocol)
	if err != nil {
		return -1, startosis_errors.NewInterpretationError(fmt.Sprintf("Port protocol should be one of %s", strings.Join(port_spec.PortProtocolStrings(), ", ")))
	}

	// TODO(gb): once we stop exposing this in the API, use only port_spec.PortProtocol enum and remove the below
	switch parsedPortProtocol {
	case port_spec.PortProtocol_TCP:
		return kurtosis_core_rpc_api_bindings.Port_TCP, nil
	case port_spec.PortProtocol_SCTP:
		return kurtosis_core_rpc_api_bindings.Port_SCTP, nil
	case port_spec.PortProtocol_UDP:
		return kurtosis_core_rpc_api_bindings.Port_UDP, nil
	}
	return -1, startosis_errors.NewInterpretationError(fmt.Sprintf("Port protocol should be one of %s", strings.Join(port_spec.PortProtocolStrings(), ", ")))
}

func extractStringValue(structField *starlarkstruct.Struct, key string, argNameForLogging string) (string, *startosis_errors.InterpretationError) {
	value, err := structField.Attr(key)
	if err != nil {
		return "", startosis_errors.NewInterpretationError(fmt.Sprintf("Missing value '%s' as element of the struct object '%s'", key, argNameForLogging))
	}
	stringValue, interpretationErr := safeCastToString(value, key)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return stringValue, nil
}

func extractUint32Value(structField *starlarkstruct.Struct, key string, argNameForLogging string) (uint32, *startosis_errors.InterpretationError) {
	value, err := structField.Attr(key)
	if err != nil {
		return 0, startosis_errors.NewInterpretationError(fmt.Sprintf("Missing value '%s' as element of the struct object '%s'", key, argNameForLogging))
	}
	uint32Value, interpretationErr := safeCastToUint32(value, key)
	if interpretationErr != nil {
		return 0, interpretationErr
	}
	return uint32Value, nil
}

func safeCastToString(expectedValueString starlark.Value, argNameForLogging string) (string, *startosis_errors.InterpretationError) {
	castValue, ok := expectedValueString.(starlark.String)
	if !ok {
		return "", startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' argument is expected to be a string. Got %s", argNameForLogging, reflect.TypeOf(expectedValueString)))
	}
	return castValue.GoString(), nil
}

func safeCastToUint32(expectedValueString starlark.Value, argNameForLogging string) (uint32, *startosis_errors.InterpretationError) {
	castValue, ok := expectedValueString.(starlark.Int)
	if !ok {
		return 0, startosis_errors.NewInterpretationError(fmt.Sprintf("Argument '%s' is expected to be an integer. Got %s", argNameForLogging, reflect.TypeOf(expectedValueString)))
	}

	uint64Value, ok := castValue.Uint64()
	if !ok || uint64Value != uint64(uint32(uint64Value)) {
		// second clause if to safeguard against "overflow"
		return 0, startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' argument is expected to be a an integer greater than 0 and lower than %d", argNameForLogging, ^uint32(0)))
	}
	return uint32(uint64Value), nil

}

func safeCastToStringSlice(expectedValueList starlark.Value, argNameForLogging string) ([]string, *startosis_errors.InterpretationError) {
	listValue, ok := expectedValueList.(*starlark.List)
	if !ok {
		return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' argument is expected to be a list. Got %s", argNameForLogging, reflect.TypeOf(expectedValueList)))
	}
	var castValue []string
	listIterator := listValue.Iterate()
	var value starlark.Value
	var index = 0
	for listIterator.Next(&value) {
		stringValue, err := safeCastToString(value, fmt.Sprintf("%v[%v]", argNameForLogging, index))
		if err != nil {
			return nil, err
		}
		castValue = append(castValue, stringValue)
		index += 1
	}
	return castValue, nil
}

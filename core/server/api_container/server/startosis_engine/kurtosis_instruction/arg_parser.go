package kurtosis_instruction

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"math"
	"reflect"
	"strings"
)

const (
	serviceIdArgName     = "service_id"
	serviceConfigArgName = "service_config"

	containerImageNameKey = "container_image_name"
	usedPortsKey          = "used_ports"
	entryPointArgsKey     = "entry_point_args"
	cmdArgsKey            = "cmd_args"
	envVarArgsKey         = "env_vars"

	portNumberKey   = "number"
	portProtocolKey = "protocol"

	commandArgName          = "command"
	expectedExitCodeArgName = "expected_exit_code"

	srcPathArgName = "src_path"

	templatesAndDataArgName = "template_and_data_by_dest_rel_filepath"
	templateFieldKey        = "template"
	templateDataFieldKey    = "template_data"

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

	entryPointArgs, interpretationErr := parseEntryPointArgs(serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	cmdArgs, interpretationErr := parseCmdArgs(serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	envVars, interpretationErr := parseEnvVars(serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	builtConfig := services.NewServiceConfigBuilder(containerImageName).WithPrivatePorts(
		privatePorts,
	).WithEntryPointArgs(
		entryPointArgs,
	).WithCmdArgs(
		cmdArgs,
	).WithEnvVars(
		envVars,
	).Build()

	return builtConfig, nil
}

func ParseCommand(commandsRaw *starlark.List) ([]string, *startosis_errors.InterpretationError) {
	commandArgs, interpretationErr := safeCastToStringSlice(commandsRaw, commandArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if len(commandArgs) == 0 {
		return nil, startosis_errors.NewInterpretationError("Command cannot be empty")
	}
	return commandArgs, nil
}

func ParseExpectedExitCode(expectedExitCodeRaw starlark.Int) (int32, *startosis_errors.InterpretationError) {
	expectedExitCode, interpretationErr := safeCastToInt32(expectedExitCodeRaw, expectedExitCodeArgName)
	if interpretationErr != nil {
		return 0, interpretationErr
	}
	return expectedExitCode, nil
}

func ParseSrcPath(serviceIdRaw starlark.String) (string, *startosis_errors.InterpretationError) {
	srcPath, interpretationErr := safeCastToString(serviceIdRaw, srcPathArgName)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if len(srcPath) == 0 {
		return "", startosis_errors.NewInterpretationError("Source path cannot be empty")
	}
	return srcPath, nil
}

func ParseTemplatesAndData(templatesAndData *starlark.Dict) (map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, *startosis_errors.InterpretationError) {
	templateAndDataByDestRelFilepath := make(map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData)
	for _, key := range templatesAndData.Keys() {
		stringKey, castErr := safeCastToString(key, fmt.Sprintf("%v.key:%v", templatesAndDataArgName, key))
		if castErr != nil {
			return nil, castErr
		}
		value, found, dictErr := templatesAndData.Get(key)
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' key in dict '%s' doesn't have a value we could retrieve. This is a Kurtosis bug.", key.String(), templatesAndDataArgName))
		}
		dictValue, ok := value.(*starlark.Dict)
		if !ok {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Expected %v[\"%v\"] to be a dict. Got '%s'", templatesAndData, stringKey, reflect.TypeOf(value)))
		}
		template, found, dictErr := dictValue.Get(starlark.String(templateFieldKey))
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Expected values in '%v' to have a '%v' field", templatesAndDataArgName, templateFieldKey))
		}
		templateStr, castErr := safeCastToString(template, fmt.Sprintf("%v[\"%v\"][\"%v\"]", templatesAndDataArgName, stringKey, templateFieldKey))
		if castErr != nil {
			return nil, castErr
		}
		templateDataStarlarkValue, found, dictErr := dictValue.Get(starlark.String(templateDataFieldKey))
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Expected values in '%v' to have a '%v' field", templatesAndDataArgName, templateDataFieldKey))
		}
		// Massive Hack
		// We do this for a couple of reasons,
		// 1. There is no good way to go from Starlark to JSON, the `String()` method seems to work, this ensures that it does
		// 2. Unmarshalling followed by Marshalling, allows for the non-scientific notation of floats to be preserved
		// 3. Don't have to write a custom way to jsonify Starlark
		// 4. This behaves as close to marshalling primitives in Golang as possible
		var temporaryUnmarshalledValue interface{}
		err := json.Unmarshal([]byte(templateDataStarlarkValue.String()), &temporaryUnmarshalledValue)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("An error occurred while converting the template data as string '%v' to temporary object", templateDataStarlarkValue))
		}
		templateDataJson, err := json.Marshal(temporaryUnmarshalledValue)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("An error occurred while converting the template data '%v' to JSON", temporaryUnmarshalledValue))
		}
		// end Massive Hack
		templateAndData := binding_constructors.NewTemplateAndData(templateStr, string(templateDataJson))
		templateAndDataByDestRelFilepath[stringKey] = templateAndData
	}
	return templateAndDataByDestRelFilepath, nil
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

func parseEntryPointArgs(serviceConfig *starlarkstruct.Struct) ([]string, *startosis_errors.InterpretationError) {
	_, err := serviceConfig.Attr(entryPointArgsKey)
	//an error here means that no argument was found which is alright as this is an optional
	if err != nil {
		return []string{}, nil
	}
	entryPointArgs, interpretationErr := extractStringSliceValue(serviceConfig, entryPointArgsKey, serviceConfigArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return entryPointArgs, nil
}

func parseCmdArgs(serviceConfig *starlarkstruct.Struct) ([]string, *startosis_errors.InterpretationError) {
	_, err := serviceConfig.Attr(cmdArgsKey)
	//an error here means that no argument was found which is alright as this is an optional
	if err != nil {
		return []string{}, nil
	}
	entryPointArgs, interpretationErr := extractStringSliceValue(serviceConfig, cmdArgsKey, serviceConfigArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return entryPointArgs, nil
}

func parseEnvVars(serviceConfig *starlarkstruct.Struct) (map[string]string, *startosis_errors.InterpretationError) {
	_, err := serviceConfig.Attr(envVarArgsKey)
	//an error here means that no argument was found which is alright as this is an optional
	if err != nil {
		return map[string]string{}, nil
	}
	envVarArgs, interpretationErr := extractMapStringStringValue(serviceConfig, envVarArgsKey, serviceConfigArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return envVarArgs, nil
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

func extractStringSliceValue(structField *starlarkstruct.Struct, key string, argNameForLogging string) ([]string, *startosis_errors.InterpretationError) {
	value, err := structField.Attr(key)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Missing value '%s' as element of the struct object '%s'", key, argNameForLogging))
	}
	stringSliceValue, interpretationErr := safeCastToStringSlice(value, key)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return stringSliceValue, nil
}

func extractMapStringStringValue(structField *starlarkstruct.Struct, key string, argNameForLogging string) (map[string]string, *startosis_errors.InterpretationError) {
	value, err := structField.Attr(key)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("Missing value '%s' as element of the struct object '%s'", key, argNameForLogging))
	}
	mapStringStringValue, interpretationErr := safeCastToMapStringString(value, key)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return mapStringStringValue, nil
}

func safeCastToString(expectedValueString starlark.Value, argNameForLogging string) (string, *startosis_errors.InterpretationError) {
	castValue, ok := expectedValueString.(starlark.String)
	if !ok {
		return "", startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' is expected to be a string. Got %s", argNameForLogging, reflect.TypeOf(expectedValueString)))
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
		return 0, startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' argument is expected to be a an integer greater than 0 and lower than %d", argNameForLogging, math.MaxUint32))
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

func safeCastToMapStringString(expectedValue starlark.Value, argNameForLogging string) (map[string]string, *startosis_errors.InterpretationError) {
	dictValue, ok := expectedValue.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' argument is expected to be a dict. Got %s", argNameForLogging, reflect.TypeOf(expectedValue)))
	}
	castValue := make(map[string]string)
	for _, key := range dictValue.Keys() {
		stringKey, castErr := safeCastToString(key, fmt.Sprintf("%v.key:%v", argNameForLogging, key))
		if castErr != nil {
			return nil, castErr
		}
		value, found, dictErr := dictValue.Get(key)
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' key in dict '%s' doesn't have a value we could retrieve. This is a Kurtosis bug.", key.String(), argNameForLogging))
		}
		stringValue, castErr := safeCastToString(value, fmt.Sprintf("%v[\"%v\"]", argNameForLogging, stringKey))
		if castErr != nil {
			return nil, castErr
		}
		castValue[stringKey] = stringValue
	}
	return castValue, nil
}

func safeCastToInt32(expectedValueString starlark.Value, argNameForLogging string) (int32, *startosis_errors.InterpretationError) {
	castValue, ok := expectedValueString.(starlark.Int)
	if !ok {
		return 0, startosis_errors.NewInterpretationError(fmt.Sprintf("Argument '%s' is expected to be an integer. Got %s", argNameForLogging, reflect.TypeOf(expectedValueString)))
	}

	int64Value, ok := castValue.Int64()
	if !ok || int64Value != int64(int32(int64Value)) {
		// second clause if to safeguard against "overflow"
		return 0, startosis_errors.NewInterpretationError(fmt.Sprintf("'%s' argument is expected to be a an integer greater than %d and lower than %d", argNameForLogging, math.MinInt32, math.MaxInt32))
	}
	return int32(int64Value), nil

}

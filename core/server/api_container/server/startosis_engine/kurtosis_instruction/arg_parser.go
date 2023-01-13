package kurtosis_instruction

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"math"
	"reflect"
)

const (
	serviceIdArgName     = "service_id"
	serviceConfigArgName = "config"
	defineFactArgName    = "define_fact"
	requestArgName       = "request"
	execArgName          = "exec"
	subnetworksArgName   = "subnetworks"

	containerImageNameKey = "image"
	usedPortsKey          = "ports"
	subnetworkKey         = "subnetwork"
	// TODO remove this when we have the Portal as this is a temporary hack to meet the NEAR use case
	serviceIdKey   = "service_id"
	contentTypeKey = "content_type"
	bodyKey        = "body"

	publicPortsKey                 = "public_ports"
	entryPointArgsKey              = "entrypoint"
	cmdArgsKey                     = "cmd"
	envVarArgsKey                  = "env_vars"
	filesArtifactMountDirpathsKey  = "files"
	portIdKey                      = "port_id"
	requestEndpointKey             = "endpoint"
	requestMethodEndpointKey       = "method"
	privateIPAddressPlaceholderKey = "private_ip_address_placeholder"

	httpRequestExtractorsKey = "extract"
	commandArgName           = "command"
	expectedExitCodeArgName  = "expected_exit_code"

	templatesAndDataArgName = "config"
	templateFieldKey        = "template"
	templateDataFieldKey    = "data"

	getRequestMethod  = "GET"
	postRequestMethod = "POST"

	jsonParsingThreadName = "Unused thread name"
	jsonParsingModuleId   = "Unused module id"
)

func ParseServiceId(serviceIdRaw starlark.String) (service.ServiceID, *startosis_errors.InterpretationError) {
	// TODO(gb): maybe prohibit certain characters for service ids
	serviceId, interpretationErr := kurtosis_types.SafeCastToString(serviceIdRaw, serviceIdArgName)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if len(serviceId) == 0 {
		return "", startosis_errors.NewInterpretationError("Service ID cannot be empty")
	}
	return service.ServiceID(serviceId), nil
}

//TODO: remove this method when we stop supporting struct for recipe defn
func ParseHttpRequestRecipe(recipeConfig *starlarkstruct.Struct) (*recipe.HttpRequestRecipe, *startosis_errors.InterpretationError) {
	serviceId, interpretationErr := extractStringValue(recipeConfig, serviceIdKey, requestArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	portId, interpretationErr := extractStringValue(recipeConfig, portIdKey, requestArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	endpoint, interpretationErr := extractStringValue(recipeConfig, requestEndpointKey, requestArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	method, interpretationErr := extractStringValue(recipeConfig, requestMethodEndpointKey, requestArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	extractors, interpretationErr := parseHttpRequestExtractors(recipeConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	if method == getRequestMethod {
		builtConfig := recipe.NewGetHttpRequestRecipe(service.ServiceID(serviceId), portId, endpoint, extractors)
		return builtConfig, nil
	} else if method == postRequestMethod {
		contentType, interpretationErr := extractStringValue(recipeConfig, contentTypeKey, defineFactArgName)
		if interpretationErr != nil {
			return nil, interpretationErr
		}

		body, interpretationErr := extractStringValue(recipeConfig, bodyKey, defineFactArgName)
		if interpretationErr != nil {
			return nil, interpretationErr
		}

		builtConfig := recipe.NewPostHttpRequestRecipe(service.ServiceID(serviceId), portId, contentType, endpoint, body, extractors)
		return builtConfig, nil
	} else {
		return nil, startosis_errors.NewInterpretationError("Define fact HTTP method not recognized")
	}
}

//TODO: remove this method when we stop supporting struct for recipe defn
func ParseExecRecipe(recipeConfig *starlarkstruct.Struct) (*recipe.ExecRecipe, *startosis_errors.InterpretationError) {
	serviceId, interpretationErr := extractStringValue(recipeConfig, serviceIdKey, execArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	command, interpretationErr := extractStringSliceValue(recipeConfig, commandArgName, execArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return recipe.NewExecRecipe(service.ServiceID(serviceId), command), nil
}

func ParseServiceConfigArg(serviceConfig *starlarkstruct.Struct) (*kurtosis_core_rpc_api_bindings.ServiceConfig, *startosis_errors.InterpretationError) {
	containerImageName, interpretationErr := extractStringValue(serviceConfig, containerImageNameKey, serviceConfigArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	privatePorts, interpretationErr := parseServiceConfigPorts(serviceConfig, usedPortsKey)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	publicPorts, interpretationErr := parseServiceConfigPorts(serviceConfig, publicPortsKey)
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

	filesArtifactMountDirpaths, interpretationErr := parseFilesArtifactMountDirpaths(serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	privateIPAddressPlaceholder, interpretationErr := parsePrivateIPAddressPlaceholder(serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	subnetwork, interpretationErr := parseServiceConfigSubnetwork(serviceConfig)
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
	).WithFilesArtifactMountDirpaths(
		filesArtifactMountDirpaths,
	).WithPrivateIPAddressPlaceholder(
		privateIPAddressPlaceholder,
	).WithPublicPorts(
		publicPorts,
	).WithSubnetwork(
		subnetwork,
	).Build()

	return builtConfig, nil
}

func ParseExpectedExitCode(expectedExitCodeRaw starlark.Int) (int32, *startosis_errors.InterpretationError) {
	expectedExitCode, interpretationErr := safeCastToInt32(expectedExitCodeRaw, expectedExitCodeArgName)
	if interpretationErr != nil {
		return 0, interpretationErr
	}
	return expectedExitCode, nil
}

func ParseNonEmptyString(argName string, argValue starlark.Value) (string, *startosis_errors.InterpretationError) {
	strArgValue, interpretationErr := kurtosis_types.SafeCastToString(argValue, argName)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if len(strArgValue) == 0 {
		return "", startosis_errors.NewInterpretationError("Expected non empty string for argument '%s'", argName)
	}
	return strArgValue, nil
}

func ParseTemplatesAndData(templatesAndData *starlark.Dict) (map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, *startosis_errors.InterpretationError) {
	templateAndDataByDestRelFilepath := make(map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData)
	for _, relPathInFilesArtifactKey := range templatesAndData.Keys() {
		relPathInFilesArtifactStr, castErr := kurtosis_types.SafeCastToString(relPathInFilesArtifactKey, fmt.Sprintf("%v.key:%v", templatesAndDataArgName, relPathInFilesArtifactKey))
		if castErr != nil {
			return nil, castErr
		}
		value, found, dictErr := templatesAndData.Get(relPathInFilesArtifactKey)
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError("'%s' key in dict '%s' doesn't have a value we could retrieve. This is a Kurtosis bug.", relPathInFilesArtifactKey.String(), templatesAndDataArgName)
		}
		structValue, ok := value.(*starlarkstruct.Struct)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Expected %v[\"%v\"] to be a dict. Got '%s'", templatesAndData, relPathInFilesArtifactStr, reflect.TypeOf(value))
		}
		template, err := structValue.Attr(templateFieldKey)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Expected values in '%v' to have a '%v' field", templatesAndDataArgName, templateFieldKey)
		}
		templateStr, castErr := kurtosis_types.SafeCastToString(template, fmt.Sprintf("%v[\"%v\"][\"%v\"]", templatesAndDataArgName, relPathInFilesArtifactStr, templateFieldKey))
		if castErr != nil {
			return nil, castErr
		}
		templateDataStarlarkValue, err := structValue.Attr(templateDataFieldKey)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Expected values in '%v' to have a '%v' field", templatesAndDataArgName, templateDataFieldKey)
		}

		templateDataJSONStrValue, encodingError := encodeStarlarkObjectAsJSON(templateDataStarlarkValue, templateDataFieldKey)
		if encodingError != nil {
			return nil, encodingError
		}
		// Massive Hack
		// We do this for a couple of reasons,
		// 1. Unmarshalling followed by Marshalling, allows for the non-scientific notation of floats to be preserved
		// 2. Don't have to write a custom way to jsonify Starlark
		// 3. This behaves as close to marshalling primitives in Golang as possible
		// 4. Allows us to validate that string input is valid JSON
		var temporaryUnmarshalledValue interface{}
		err = json.Unmarshal([]byte(templateDataJSONStrValue), &temporaryUnmarshalledValue)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Template data for file '%v', '%v' isn't valid JSON", relPathInFilesArtifactStr, templateDataJSONStrValue)
		}
		templateDataJson, err := json.Marshal(temporaryUnmarshalledValue)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Template data for file '%v', '%v' isn't valid JSON", relPathInFilesArtifactStr, templateDataJSONStrValue)
		}
		// end Massive Hack
		templateAndData := binding_constructors.NewTemplateAndData(templateStr, string(templateDataJson))
		templateAndDataByDestRelFilepath[relPathInFilesArtifactStr] = templateAndData
	}
	return templateAndDataByDestRelFilepath, nil
}

func ParseSubnetworks(subnetworksTuple starlark.Tuple) (service_network_types.PartitionID, service_network_types.PartitionID, *startosis_errors.InterpretationError) {
	subnetworksStr, interpretationErr := kurtosis_types.SafeCastToStringSlice(subnetworksTuple, subnetworksArgName)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}
	if len(subnetworksStr) != 2 {
		return "", "", startosis_errors.NewInterpretationError("Subnetworks tuple should contain exactly 2 subnetwork names. %d was/were provided", len(subnetworksStr))
	}
	subnetwork1 := service_network_types.PartitionID(subnetworksStr[0])
	subnetwork2 := service_network_types.PartitionID(subnetworksStr[1])
	return subnetwork1, subnetwork2, nil
}
func parseServiceConfigSubnetwork(serviceConfig *starlarkstruct.Struct) (string, *startosis_errors.InterpretationError) {
	// subnetwork, if present, should be a simple string
	_, err := serviceConfig.Attr(subnetworkKey)
	if err != nil {
		// subnetwork is optional, if it's not present -> return empty string
		return "", nil
	}
	subnetwork, interpretationErr := extractStringValue(serviceConfig, subnetworkKey, serviceConfigArgName)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return subnetwork, nil
}

func parseServiceConfigPorts(serviceConfig *starlarkstruct.Struct, portsKey string) (map[string]*kurtosis_core_rpc_api_bindings.Port, *startosis_errors.InterpretationError) {
	privatePortsRawArg, err := serviceConfig.Attr(portsKey)
	if err != nil {
		// not all services need to create ports, this being empty is okay
		return map[string]*kurtosis_core_rpc_api_bindings.Port{}, nil
	}
	privatePortsArg, ok := privatePortsRawArg.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Argument `%s` is expected to be a dictionary", usedPortsKey)
	}

	var privatePorts = make(map[string]*kurtosis_core_rpc_api_bindings.Port)
	for _, portNameRaw := range privatePortsArg.Keys() {
		portDefinitionRaw, found, err := privatePortsArg.Get(portNameRaw)
		if !found || err != nil {
			return nil, startosis_errors.NewInterpretationError("Unable to find a value in a dict associated with a key that exists (key = '%s') - this is a product bug", portNameRaw)
		}

		portName, interpretationErr := kurtosis_types.SafeCastToString(portNameRaw, portsKey)
		if interpretationErr != nil {
			return nil, interpretationErr
		}

		portDefinition, ok := portDefinitionRaw.(*kurtosis_types.PortSpec)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Port definition `%s` is expected to be a PortSpec", portDefinitionRaw)
		}
		privatePorts[portName] = portDefinition.ToKurtosisType()
	}
	return privatePorts, nil
}

func parseEntryPointArgs(serviceConfig *starlarkstruct.Struct) ([]string, *startosis_errors.InterpretationError) {
	_, err := serviceConfig.Attr(entryPointArgsKey)
	//an error here means that no argument was found which is alright as this is an optional
	if err != nil {
		return nil, nil
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
		return nil, nil
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

func parseFilesArtifactMountDirpaths(serviceConfig *starlarkstruct.Struct) (map[string]string, *startosis_errors.InterpretationError) {
	_, err := serviceConfig.Attr(filesArtifactMountDirpathsKey)
	//an error here means that no argument was found which is alright as this is an optional
	if err != nil {
		return map[string]string{}, nil
	}
	filesArtifactMountDirpathsArg, interpretationErr := extractMapStringStringValue(serviceConfig, filesArtifactMountDirpathsKey, serviceConfigArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return filesArtifactMountDirpathsArg, nil
}

//TODO: remove this method when we stop supporting struct for recipe defn
func parseHttpRequestExtractors(recipe *starlarkstruct.Struct) (map[string]string, *startosis_errors.InterpretationError) {
	_, err := recipe.Attr(httpRequestExtractorsKey)
	//an error here means that no argument was found which is alright as this is an optional
	if err != nil {
		return map[string]string{}, nil
	}
	httpRequestExtractorsArg, interpretationErr := extractMapStringStringValue(recipe, httpRequestExtractorsKey, requestArgName)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return httpRequestExtractorsArg, nil
}

func parsePrivateIPAddressPlaceholder(serviceConfig *starlarkstruct.Struct) (string, *startosis_errors.InterpretationError) {
	_, err := serviceConfig.Attr(privateIPAddressPlaceholderKey)
	//an error here means that no argument was found which is alright as this is an optional
	if err != nil {
		return "", nil
	}
	privateIpAddressPlaceholder, interpretationErr := extractStringValue(serviceConfig, privateIPAddressPlaceholderKey, serviceConfigArgName)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return privateIpAddressPlaceholder, nil
}

func extractStringValue(structField *starlarkstruct.Struct, key string, argNameForLogging string) (string, *startosis_errors.InterpretationError) {
	value, err := structField.Attr(key)
	if err != nil {
		return "", startosis_errors.NewInterpretationError("Missing value '%s' as element of the struct object '%s'", key, argNameForLogging)
	}
	stringValue, interpretationErr := kurtosis_types.SafeCastToString(value, key)
	if interpretationErr != nil {
		return "", startosis_errors.WrapWithInterpretationError(interpretationErr, "Error casting value '%s' as element of the struct object '%s'", key, argNameForLogging)
	}
	return stringValue, nil
}

func extractUint32Value(structField *starlarkstruct.Struct, key string, argNameForLogging string) (uint32, *startosis_errors.InterpretationError) {
	value, err := structField.Attr(key)
	if err != nil {
		return 0, startosis_errors.NewInterpretationError("Missing value '%s' as element of the struct object '%s'", key, argNameForLogging)
	}
	uint32Value, interpretationErr := safeCastToUint32(value, key)
	if interpretationErr != nil {
		return 0, startosis_errors.WrapWithInterpretationError(interpretationErr, "Error casting value '%s' as element of the struct object '%s'", key, argNameForLogging)
	}
	return uint32Value, nil
}

func extractStringSliceValue(structField *starlarkstruct.Struct, key string, argNameForLogging string) ([]string, *startosis_errors.InterpretationError) {
	value, err := structField.Attr(key)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("Missing value '%s' as element of the struct object '%s'", key, argNameForLogging)
	}
	stringSliceValue, interpretationErr := kurtosis_types.SafeCastToStringSlice(value, key)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "Error casting value '%s' as element of the struct object '%s'", key, argNameForLogging)
	}
	return stringSliceValue, nil
}

func extractMapStringStringValue(structField *starlarkstruct.Struct, key string, argNameForLogging string) (map[string]string, *startosis_errors.InterpretationError) {
	value, err := structField.Attr(key)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("Missing value '%s' as element of the struct object '%s'", key, argNameForLogging)
	}
	mapStringStringValue, interpretationErr := kurtosis_types.SafeCastToMapStringString(value, key)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "Error casting value '%s' as element of the struct object '%s'", key, argNameForLogging)
	}
	return mapStringStringValue, nil
}

func safeCastToUint32(expectedValueString starlark.Value, argNameForLogging string) (uint32, *startosis_errors.InterpretationError) {
	castValue, ok := expectedValueString.(starlark.Int)
	if !ok {
		return 0, startosis_errors.NewInterpretationError("Argument '%s' is expected to be an integer. Got %s", argNameForLogging, reflect.TypeOf(expectedValueString))
	}

	uint64Value, ok := castValue.Uint64()
	if !ok || uint64Value != uint64(uint32(uint64Value)) {
		// second clause if to safeguard against "overflow"
		return 0, startosis_errors.NewInterpretationError("'%s' argument is expected to be a an integer greater than 0 and lower than %d", argNameForLogging, math.MaxUint32)
	}
	return uint32(uint64Value), nil

}

func safeCastToInt32(expectedValueString starlark.Value, argNameForLogging string) (int32, *startosis_errors.InterpretationError) {
	castValue, ok := expectedValueString.(starlark.Int)
	if !ok {
		return 0, startosis_errors.NewInterpretationError("Argument '%s' is expected to be an integer. Got %s", argNameForLogging, reflect.TypeOf(expectedValueString))
	}

	int64Value, ok := castValue.Int64()
	if !ok || int64Value != int64(int32(int64Value)) {
		// second clause if to safeguard against "overflow"
		return 0, startosis_errors.NewInterpretationError("'%s' argument is expected to be a an integer greater than %d and lower than %d", argNameForLogging, math.MinInt32, math.MaxInt32)
	}
	return int32(int64Value), nil

}

func encodeStarlarkObjectAsJSON(object starlark.Value, argNameForLogging string) (string, *startosis_errors.InterpretationError) {
	jsonifiedVersion := ""
	thread := &starlark.Thread{
		Name:       jsonParsingThreadName,
		OnMaxSteps: nil,
		Print: func(_ *starlark.Thread, msg string) {
			jsonifiedVersion = msg
		},
		Load: nil,
	}

	predeclared := &starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
	}

	// We do a print here as if we return the encoded variable we get extra quotes and slashes
	// {"fizz": "buzz"} becomes "{\"fizz": \"buzz"\}"
	scriptToRun := fmt.Sprintf(`encoded_json = json.encode(%v)
print(encoded_json)`, object.String())

	_, err := starlark.ExecFile(thread, jsonParsingModuleId, scriptToRun, *predeclared)

	if err != nil {
		return "", startosis_errors.NewInterpretationError("Error converting '%v' with string value '%v' to JSON", argNameForLogging, object.String())
	}

	return jsonifiedVersion, nil
}

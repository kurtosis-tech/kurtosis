package kurtosis_instruction

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"reflect"
)

const (
	serviceNameArgName       = "service_name"
	requestArgName           = "request"
	subnetworksArgName       = "subnetworks"
	httpRequestExtractorsKey = "extract"
	templatesAndDataArgName  = "config"
	templateFieldKey         = "template"
	templateDataFieldKey     = "data"
	jsonParsingThreadName    = "Unused thread name"
	jsonParsingModuleId      = "Unused module id"
)

func ParseServiceName(serviceIdRaw starlark.String) (service.ServiceName, *startosis_errors.InterpretationError) {
	// TODO(gb): maybe prohibit certain characters for service ids
	serviceName, interpretationErr := kurtosis_types.SafeCastToString(serviceIdRaw, serviceNameArgName)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if len(serviceName) == 0 {
		return "", startosis_errors.NewInterpretationError("Service Name cannot be empty")
	}
	return service.ServiceName(serviceName), nil
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

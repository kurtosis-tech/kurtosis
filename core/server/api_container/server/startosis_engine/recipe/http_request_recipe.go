package recipe

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/itchyny/gojq"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"io"
	"net/http"
	"reflect"
	"strings"
)

const (
	statusCodeKey    = "code"
	bodyKey          = "body"
	extractKeyPrefix = "extract"

	// Common attributes for both [Get|Post]HttpRequestRecipe
	PortIdAttr   = "port_id"
	EndpointAttr = "endpoint"
	ExtractAttr  = "extract"
)

type HttpRequestRecipe interface {
	builtin_argument.KurtosisValueType

	Recipe
}

func executeInternal(
	ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	serviceName service.ServiceName,
	requestBody string,
	portId string,
	method string,
	contentType string,
	endpoint string,
	extractors map[string]string,
) (map[string]starlark.Comparable, error) {
	var response *http.Response
	var err error
	recipeBodyWithRuntimeValue, err := magic_string_helper.ReplaceRuntimeValueInString(requestBody, runtimeValueStore)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while replacing runtime values in the body of the http recipe")
	}

	serviceNameStr := string(serviceName)
	if serviceNameStr == "" {
		return nil, stacktrace.NewError("The service name parameter can't be an empty string")
	}

	response, err = serviceNetwork.HttpRequestService(
		ctx,
		serviceNameStr,
		portId,
		method,
		contentType,
		endpoint,
		recipeBodyWithRuntimeValue,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when running HTTP request recipe")
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			logrus.Errorf("An error occurred when closing response body: %v", err)
		}
	}()
	responseBody, err := io.ReadAll(response.Body)
	logrus.Debugf("Got response '%v'", string(responseBody))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while reading HTTP response body")
	}
	resultDict := map[string]starlark.Comparable{
		bodyKey:       starlark.String(responseBody),
		statusCodeKey: starlark.MakeInt(response.StatusCode),
	}
	extractDict, err := extractInternal(responseBody, extractors)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while running extractors on HTTP response body")
	}
	for extractorKey, extractorValue := range extractDict {
		resultDict[fmt.Sprintf("%v.%v", extractKeyPrefix, extractorKey)] = extractorValue
	}
	return resultDict, nil
}

func extractInternal(responseBody []byte, extractors map[string]string) (map[string]starlark.Comparable, error) {
	if len(extractors) == 0 {
		return map[string]starlark.Comparable{}, nil
	}
	logrus.Debug("Executing extract recipe")
	var jsonBody interface{}
	err := json.Unmarshal(responseBody, &jsonBody)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when parsing JSON response body")
	}
	extractorResult := map[string]starlark.Comparable{}
	for extractorKey, extractor := range extractors {
		logrus.Debugf("Running against '%v' '%v'", jsonBody, extractor)
		query, err := gojq.Parse(extractor)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred when parsing field extractor '%v'", extractor)
		}
		iter := query.Run(jsonBody)
		foundMatch := false
		for {
			matchValue, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := matchValue.(error); ok {
				logrus.Errorf("HTTP request recipe extract emitted error '%v'", err)
			}
			if matchValue != nil {
				var parsedMatchValue starlark.Comparable
				logrus.Debug("Start parsing...")
				switch value := matchValue.(type) {
				case int:
					parsedMatchValue = starlark.MakeInt(value)
				case string:
					parsedMatchValue = starlark.String(value)
				case float32:
					parsedMatchValue = starlark.Float(value)
				case float64:
					parsedMatchValue = starlark.Float(value)
				default:
					parsedMatchValue = starlark.String(fmt.Sprintf("%v", value))
				}
				logrus.Debugf("Parsed successfully %v %v", matchValue, parsedMatchValue)
				extractorResult[extractorKey] = parsedMatchValue
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			return nil, stacktrace.NewError("No field '%s' was found on input '%s'", extractor, string(responseBody))
		}
	}
	logrus.Debugf("Extractor result map '%v'", extractorResult)
	return extractorResult, nil
}

func resultMapToStringInternal(resultMap map[string]starlark.Comparable) string {
	statusCode := resultMap[statusCodeKey]
	body := resultMap[bodyKey]
	extractedFieldString := strings.Builder{}
	for resultKey, resultValue := range resultMap {
		if strings.Contains(resultKey, extractKeyPrefix) {
			extractedFieldString.WriteString(fmt.Sprintf("\n'%v': %v", resultKey, resultValue))
		}
	}
	if extractedFieldString.Len() == 0 {
		return fmt.Sprintf("Request had response code '%v' and body %v", statusCode, body)
	} else {
		return fmt.Sprintf("Request had response code '%v' and body %v, with extracted fields:%s", statusCode, body, extractedFieldString.String())
	}
}

func createStarlarkReturnValueInternal(resultUuid string, extractors map[string]string) (*starlark.Dict, *startosis_errors.InterpretationError) {
	dict := &starlark.Dict{}
	err := dict.SetKey(starlark.String(bodyKey), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, bodyKey)))
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error has occurred when creating return value for request recipe, setting field '%v'", bodyKey)
	}
	err = dict.SetKey(starlark.String(statusCodeKey), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, statusCodeKey)))
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error has occurred when creating return value for request recipe, setting field '%v'", statusCodeKey)
	}
	for extractorKey := range extractors {
		fullExtractorKey := fmt.Sprintf("%v.%v", extractKeyPrefix, extractorKey)
		err = dict.SetKey(starlark.String(fullExtractorKey), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, fullExtractorKey)))
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("An error has occurred when creating return value for request recipe, setting field '%v'", fullExtractorKey)
		}
	}
	dict.Freeze()
	return dict, nil
}

func convertExtractorsToDict(isAttrSet bool, extractorsValue starlark.Value) (map[string]string, *startosis_errors.InterpretationError) {
	extractorStringMap := map[string]string{}
	if !isAttrSet {
		return extractorStringMap, nil
	}
	extractorsDict, ok := extractorsValue.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Attribute '%s' on '%s' is expected to be a dictionary of strings, got '%s'", ExtractAttr, GetHttpRecipeTypeName, reflect.TypeOf(extractorsValue))
	}

	for _, extractorKey := range extractorsDict.Keys() {
		extractorValue, found, err := extractorsDict.Get(extractorKey)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unexpected error iterating on dictionary. Value associated to key '%v' could not be found", extractorKey)
		} else if !found {
			return nil, startosis_errors.NewInterpretationError("Unexpected error iterating on dictionary. Value associated to key '%v' could not be found", extractorKey)
		}

		extractorKeyStr, ok := extractorKey.(starlark.String)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Key in '%s' dictionary was expected to be a string, got '%s'", ExtractAttr, reflect.TypeOf(extractorKey))
		}
		extractorValueStr, ok := extractorValue.(starlark.String)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Value associated to key '%s' in dictionary '%s' was expected to be a string, got '%s'", extractorKeyStr, ExtractAttr, reflect.TypeOf(extractorsValue))
		}
		extractorStringMap[extractorKeyStr.GoString()] = extractorValueStr.GoString()
	}
	return extractorStringMap, nil
}

package recipe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"golang.org/x/exp/maps"
)

const (
	statusCodeKey = "code"
	bodyKey       = "body"

	// Common attributes for both [Get|Post]HttpRequestRecipe
	PortIdAttr   = "port_id"
	EndpointAttr = "endpoint"
	HeadersAttr  = "headers"
	ExtractAttr  = "extract"
)

type HttpRequestRecipe interface {
	builtin_argument.KurtosisValueType

	Recipe

	// RequestType as of 2023-04-18 this only exists so that ExecRecipe doesn't implement HttpRequestRecipe
	RequestType() string
}

func executeInternal(
	ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	service *service.Service,
	requestBody string,
	portId string,
	method string,
	contentType string,
	endpoint string,
	extractors map[string]string,
	headers map[string]string,
) (map[string]starlark.Comparable, error) {
	var response *http.Response
	var err error
	recipeBodyWithRuntimeValue, err := magic_string_helper.ReplaceRuntimeValueInString(requestBody, runtimeValueStore)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while replacing runtime values in the body of the http recipe")
	}

	// service, err := serviceNetwork.GetService(ctx, serviceNameStr)
	// if err != nil {
	// 	return nil, stacktrace.Propagate(err, "An error occurred when getting service '%v'", serviceNameStr)
	// }

	response, err = serviceNetwork.HttpRequestServiceObject(ctx, service, portId, method, contentType, endpoint, recipeBodyWithRuntimeValue, headers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when running HTTP request recipe.")
	}

	// response, err = serviceNetwork.HttpRequestService(ctx, serviceNameStr, portId, method, contentType, endpoint, recipeBodyWithRuntimeValue, headers)
	// if err != nil {
	// 	return nil, stacktrace.Propagate(err, "An error occurred when running HTTP request recipe")
	// }
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
	extractDict, err := runExtractors(responseBody, extractors)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while running extractors from HTTP recipe")
	}
	maps.Copy(resultDict, extractDict)
	return resultDict, nil
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

func convertHeadersToMapStringString(isSet bool, headersStarlarkValue starlark.Value) (map[string]string, *startosis_errors.InterpretationError) {
	if !isSet {
		return map[string]string{}, nil
	}
	headersDict, ok := headersStarlarkValue.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("expected '%v' to be a starlark dict but got '%v'", headersStarlarkValue, reflect.TypeOf(headersStarlarkValue))
	}
	headers, interpretationErr := kurtosis_types.SafeCastToMapStringString(headersDict, HeadersAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return headers, nil
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

func ValidateHttpRequestRecipe(httpRequestRecipe HttpRequestRecipe, serviceName service.ServiceName, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	portIdValue, err := httpRequestRecipe.Attr(PortIdAttr)
	if err != nil {
		return startosis_errors.NewValidationError("Tried fetching port ID for request on service '%s' but failed", serviceName)
	}
	portIdStringValue, ok := starlark.AsString(portIdValue)
	if !ok {
		return startosis_errors.NewValidationError("Tried getting string value for port ID '%v' for request to service '%s' but failed", portIdValue, serviceName)
	}
	if portIdExists := validatorEnvironment.DoesPrivatePortIDExistForService(portIdStringValue, serviceName); !portIdExists {
		return startosis_errors.NewValidationError("Request required port ID '%v' to exist on service '%v' but it doesn't", portIdStringValue, serviceName)
	}
	return nil
}

package recipe

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/itchyny/gojq"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"io"
	"net/http"
	"strings"
)

const (
	postMethod        = "POST"
	getMethod         = "GET"
	emptyBody         = ""
	unusedContentType = ""

	statusCodeKey    = "code"
	bodyKey          = "body"
	extractKeyPrefix = "extract"

	portIdAttr      = "port_id"
	serviceIdAttr   = "service_id"
	endpointAttr    = "endpoint"
	methodAttr      = "method"
	contentTypeAttr = "content_type"

	PostHttpRecipeTypeName = "PostHttpRequestRecipe"
	GetHttpRecipeTypeName  = "GetHttpRequestRecipe"

	HttpRecipeTypeName = "HttpRequestRecipe"
)

type HttpRequestRecipe struct {
	serviceId   service.ServiceName
	portId      string
	contentType string
	endpoint    string
	method      string
	body        string
	extractors  map[string]string
}

func NewPostHttpRequestRecipe(serviceId service.ServiceName, portId string, contentType string, endpoint string, body string, extractors map[string]string) *HttpRequestRecipe {
	return &HttpRequestRecipe{
		serviceId:   serviceId,
		portId:      portId,
		method:      postMethod,
		contentType: contentType,
		endpoint:    endpoint,
		body:        body,
		extractors:  extractors,
	}
}

func NewGetHttpRequestRecipe(serviceId service.ServiceName, portId string, endpoint string, extractors map[string]string) *HttpRequestRecipe {
	return &HttpRequestRecipe{
		serviceId:   serviceId,
		portId:      portId,
		method:      getMethod,
		contentType: unusedContentType,
		endpoint:    endpoint,
		body:        emptyBody,
		extractors:  extractors,
	}
}

// String the starlark.Value interface
func (recipe *HttpRequestRecipe) String() string {
	buffer := new(strings.Builder)
	instanceName := recipe.GetInstanceName()

	buffer.WriteString(instanceName + "(")
	buffer.WriteString(portIdAttr + "=")
	buffer.WriteString(fmt.Sprintf("%q, ", recipe.portId))
	buffer.WriteString(serviceIdAttr + "=")
	buffer.WriteString(fmt.Sprintf("%q, ", recipe.serviceId))
	buffer.WriteString(endpointAttr + "=")
	buffer.WriteString(fmt.Sprintf("%q, ", recipe.endpoint))

	if recipe.method == postMethod {
		buffer.WriteString(bodyKey + "=")
		buffer.WriteString(fmt.Sprintf("%q, ", recipe.body))
		buffer.WriteString(contentTypeAttr + "=")
		buffer.WriteString(fmt.Sprintf("%q, ", recipe.contentType))
	}

	buffer.WriteString(extractKeyPrefix + "=")
	extractors, err := convertMapToStarlarkDict(recipe.extractors)

	if err != nil {
		logrus.Errorf("Error occurred while accessing extractors")
	}

	if extractors.Len() > 0 {
		buffer.WriteString(fmt.Sprintf("%q)", extractors))
	} else {
		buffer.WriteString(fmt.Sprintf("%q)", ""))
	}
	return buffer.String()
}

// Type implements the starlark.Value interface
func (recipe *HttpRequestRecipe) Type() string {
	return HttpRecipeTypeName
}

// Freeze implements the starlark.Value interface
func (recipe *HttpRequestRecipe) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (recipe *HttpRequestRecipe) Truth() starlark.Bool {
	truth := recipe.portId != "" && recipe.serviceId != "" && recipe.endpoint != "" && recipe.method != ""
	if recipe.method == postMethod {
		truth = truth && recipe.body != "" && recipe.contentType != ""
	}
	return starlark.Bool(truth)
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed, users should use a portId instead
func (recipe *HttpRequestRecipe) Hash() (uint32, error) {
	return 0, startosis_errors.NewInterpretationError("unhashable type: '%v'", HttpRecipeTypeName)
}

func (recipe *HttpRequestRecipe) GetInstanceName() string {
	instanceName := GetHttpRecipeTypeName
	if recipe.method == postMethod {
		instanceName = PostHttpRecipeTypeName
	}
	return instanceName
}

// Attr implements the starlark.HasAttrs interface.
func (recipe *HttpRequestRecipe) Attr(name string) (starlark.Value, error) {
	switch name {
	case portIdAttr:
		return starlark.String(recipe.portId), nil
	case serviceIdAttr:
		return starlark.String(recipe.serviceId), nil
	case extractKeyPrefix:
		return convertMapToStarlarkDict(recipe.extractors)
	case bodyKey:
		return starlark.String(recipe.body), nil
	case contentTypeAttr:
		return starlark.String(recipe.contentType), nil
	case methodAttr:
		return starlark.String(recipe.method), nil
	case endpointAttr:
		return starlark.String(recipe.endpoint), nil
	default:
		return nil, startosis_errors.NewInterpretationError("'%v' has no attribute '%v;", HttpRecipeTypeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (recipe *HttpRequestRecipe) AttrNames() []string {
	return []string{portIdAttr, serviceIdAttr, extractKeyPrefix, endpointAttr, contentTypeAttr, methodAttr, bodyKey}
}

func MakeGetHttpRequestRecipe(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var portId string
	var endpoint string
	var serviceId string
	var maybeExtractField starlark.Value

	if err := starlark.UnpackArgs(builtin.Name(), args, kwargs,
		serviceIdAttr, &serviceId,
		portIdAttr, &portId,
		endpointAttr, &endpoint,
		kurtosis_types.MakeOptional(extractKeyPrefix), &maybeExtractField,
	); err != nil {
		return nil, startosis_errors.NewInterpretationError(err.Error())
	}

	extractedMap := map[string]string{}
	var err *startosis_errors.InterpretationError

	if maybeExtractField != nil {
		extractedMap, err = kurtosis_types.SafeCastToMapStringString(maybeExtractField, extractKeyPrefix)
		if err != nil {
			return nil, err
		}
	}
	recipe := NewGetHttpRequestRecipe(service.ServiceName(serviceId), portId, endpoint, extractedMap)
	return recipe, nil
}

func MakePostHttpRequestRecipe(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var portId string
	var endpoint string
	var serviceId string

	var body string
	var contentType string
	var maybeExtractField starlark.Value

	if err := starlark.UnpackArgs(builtin.Name(), args, kwargs,
		serviceIdAttr, &serviceId,
		portIdAttr, &portId,
		endpointAttr, &endpoint,
		bodyKey, &body,
		contentTypeAttr, &contentType,
		kurtosis_types.MakeOptional(extractKeyPrefix), &maybeExtractField,
	); err != nil {
		return nil, startosis_errors.NewInterpretationError("%v", err.Error())
	}

	extractedMap := map[string]string{}
	var err *startosis_errors.InterpretationError

	if maybeExtractField != nil {
		extractedMap, err = kurtosis_types.SafeCastToMapStringString(maybeExtractField, extractKeyPrefix)
		if err != nil {
			return nil, err
		}
	}

	recipe := NewPostHttpRequestRecipe(service.ServiceName(serviceId), portId, contentType, endpoint, body, extractedMap)
	return recipe, nil
}

func (recipe *HttpRequestRecipe) Execute(ctx context.Context, serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) (map[string]starlark.Comparable, error) {
	var response *http.Response
	var err error
	logrus.Debugf("Running HTTP request recipe '%v'", recipe)
	maybeRecipeBodyWithIPAddress, err := magic_string_helper.ReplaceIPAddressInString(recipe.body, serviceNetwork, bodyKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while replacing IP address in the body of the http recipe")
	}
	maybeRecipeBodyWithIPAddressAndRuntimeValue, err := magic_string_helper.ReplaceRuntimeValueInString(maybeRecipeBodyWithIPAddress, runtimeValueStore)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while replacing runtime values in the body of the http recipe")
	}
	response, err = serviceNetwork.HttpRequestService(
		ctx,
		string(recipe.serviceId),
		recipe.portId,
		recipe.method,
		recipe.contentType,
		recipe.endpoint,
		maybeRecipeBodyWithIPAddressAndRuntimeValue,
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
	body, err := io.ReadAll(response.Body)
	logrus.Debugf("Got response '%v'", string(body))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while reading HTTP response body")
	}
	resultDict := map[string]starlark.Comparable{
		bodyKey:       starlark.String(body),
		statusCodeKey: starlark.MakeInt(response.StatusCode),
	}
	extractDict, err := recipe.extract(body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while running extractors on HTTP response body")
	}
	for extractorKey, extractorValue := range extractDict {
		resultDict[fmt.Sprintf("%v.%v", extractKeyPrefix, extractorKey)] = extractorValue
	}
	return resultDict, nil
}

func (recipe *HttpRequestRecipe) extract(body []byte) (map[string]starlark.Comparable, error) {
	if len(recipe.extractors) == 0 {
		return map[string]starlark.Comparable{}, nil
	}
	logrus.Debug("Executing extract recipe")
	var jsonBody interface{}
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when parsing JSON response body")
	}
	extractorResult := map[string]starlark.Comparable{}
	for extractorKey, extractor := range recipe.extractors {
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
			return nil, stacktrace.NewError("No field '%v' was found on input '%v'", extractor, body)
		}
	}
	logrus.Debugf("Extractor result map '%v'", extractorResult)
	return extractorResult, nil
}

func (recipe *HttpRequestRecipe) ResultMapToString(resultMap map[string]starlark.Comparable) string {
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

func (recipe *HttpRequestRecipe) CreateStarlarkReturnValue(resultUuid string) (*starlark.Dict, *startosis_errors.InterpretationError) {
	dict := &starlark.Dict{}
	err := dict.SetKey(starlark.String(bodyKey), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, bodyKey)))
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error has occurred when creating return value for request recipe, setting field '%v'", bodyKey)
	}
	err = dict.SetKey(starlark.String(statusCodeKey), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, statusCodeKey)))
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error has occurred when creating return value for request recipe, setting field '%v'", statusCodeKey)
	}
	for extractorKey := range recipe.extractors {
		fullExtractorKey := fmt.Sprintf("%v.%v", extractKeyPrefix, extractorKey)
		err = dict.SetKey(starlark.String(fullExtractorKey), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, fullExtractorKey)))
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("An error has occurred when creating return value for request recipe, setting field '%v'", fullExtractorKey)
		}
	}
	dict.Freeze()
	return dict, nil
}

func convertMapToStarlarkDict(inputMap map[string]string) (*starlark.Dict, *startosis_errors.InterpretationError) {
	sizeOfExtractors := len(inputMap)
	dict := starlark.NewDict(sizeOfExtractors)
	for key, val := range inputMap {
		err := dict.SetKey(starlark.String(key), starlark.String(val))
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Error occurred while converting extractor map to starlark type")
		}
	}
	return dict, nil
}

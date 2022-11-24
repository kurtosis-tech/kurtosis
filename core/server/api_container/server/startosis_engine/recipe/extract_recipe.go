package recipe

import (
	"encoding/json"
	"fmt"
	"github.com/itchyny/gojq"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

type ExtractRecipe struct {
	fieldExtractor string
}

func NewExtractRecipe(fieldExtractor string) *ExtractRecipe {
	return &ExtractRecipe{
		fieldExtractor: fieldExtractor,
	}
}

func (recipe *ExtractRecipe) Execute(input string) (map[string]starlark.Comparable, error) {
	logrus.Debug("Executing extract recipe")
	var jsonBody interface{}
	err := json.Unmarshal([]byte(input), &jsonBody)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when parsing JSON response body")
	}
	logrus.Debugf("Running against '%v' '%v' '%v'", input, jsonBody, recipe.fieldExtractor)
	query, err := gojq.Parse(recipe.fieldExtractor)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when parsing field extractor '%v'", recipe.fieldExtractor)
	}
	iter := query.Run(jsonBody)
	for {
		matchValue, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := matchValue.(error); ok {
			logrus.Errorf("%v", err)
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
			return map[string]starlark.Comparable{
				"match": parsedMatchValue,
			}, nil
		}
	}
	return nil, stacktrace.NewError("No field '%v' was found on input '%v'", recipe.fieldExtractor, input)
}

func CreateStarlarkReturnValueFromExtractRuntimeValue(resultUuid string) starlark.String {
	return starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, "match"))
}

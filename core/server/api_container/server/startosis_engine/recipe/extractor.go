package recipe

import (
	"encoding/json"
	"fmt"
	"github.com/itchyny/gojq"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const extractKeyPrefix = "extract"

func extract(input []byte, query string) (starlark.Comparable, error) {
	logrus.Debugf("Running extractor against query '%v' and input '%v'", string(input), query)
	jqQuery, err := gojq.Parse(query)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when parsing field extractor '%v'", query)
	}
	var jsonBody interface{}
	err = json.Unmarshal(input, &jsonBody)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when parsing JSON response body:\n'%v'", string(input))
	}
	matchIterator := jqQuery.Run(jsonBody)
	parsedMatchList := []starlark.Value{}
	for {
		matchValue, ok := matchIterator.Next()
		if !ok {
			break
		}
		logrus.Debugf("Found match '%v'", matchValue)
		if err, ok := matchValue.(error); ok {
			logrus.Warnf("Match '%v' was error type '%T'", err, err)
			continue
		}
		if matchValue != nil {
			parsedMatchValue := parseJsonValueToStarlark(matchValue)
			logrus.Debugf("Parsed successfully from %v to %v", matchValue, parsedMatchValue)
			parsedMatchList = append(parsedMatchList, parsedMatchValue)
		}
	}
	if len(parsedMatchList) == 0 {
		return nil, stacktrace.NewError("No field '%v' was found on input '%v'", query, string(input))
	}
	if len(parsedMatchList) == 1 {
		return parsedMatchList[0].(starlark.Comparable), nil
	}
	return starlark.NewList(parsedMatchList), nil
}

func parseJsonValueToStarlark(value any) starlark.Value {
	switch value := value.(type) {
	case int:
		return starlark.MakeInt(value)
	case string:
		return starlark.String(value)
	case float64:
		if float64(int(value)) == value {
			return starlark.MakeInt(int(value))
		}
		return starlark.Float(value)
	case []any:
		list := []starlark.Value{}
		for _, element := range value {
			list = append(list, parseJsonValueToStarlark(element))
		}
		return starlark.NewList(list)
	case map[string]any:
		parsedMap := starlark.NewDict(len(value))
		for k, v := range value {
			_ = parsedMap.SetKey(parseJsonValueToStarlark(k), parseJsonValueToStarlark(v))
		}
		return parsedMap
	case bool:
		return starlark.Bool(value)
	default:
		logrus.Warnf("Type %T has no cast defined to Starlark on extract", value)
		return starlark.String(fmt.Sprintf("%v", value))
	}
}

// runExtractors takes in `input` and a map of `extractors`. Each entry of extractors has an `id` and a `query` string.
// For each extractor, we run `query` against `input`, returning a map with key being extract.id and the Starlark result.
func runExtractors(input []byte, extractors map[string]string) (map[string]starlark.Comparable, error) {
	extractResult := map[string]starlark.Comparable{}
	for extractorName, query := range extractors {
		extractedValue, err := extract(input, query)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred running extractor '%v' on recipe", query)
		}
		extractResult[fmt.Sprintf("%v.%v", extractKeyPrefix, extractorName)] = extractedValue
	}
	return extractResult, nil
}

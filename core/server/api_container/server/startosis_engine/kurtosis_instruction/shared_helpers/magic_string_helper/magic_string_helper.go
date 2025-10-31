package magic_string_helper

import (
	"regexp"
	"strings"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	unlimitedMatches = -1
	singleMatch      = 1

	serviceNameSubgroupName = "name"
	allSubgroupName         = "all"
	kurtosisNamespace       = "kurtosis"

	runtimeValueSubgroupName      = "runtime_value"
	runtimeValueFieldSubgroupName = "runtime_value_field"
	runtimeValueKeyRegexp         = "[a-zA-Z0-9-_\\.]+"
	uuidFormat                    = "[a-f0-9]{32}"

	runtimeValueReplacementRegex             = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + runtimeValueSubgroupName + ">" + uuidFormat + ")" + ":(?P<" + runtimeValueFieldSubgroupName + ">" + runtimeValueKeyRegexp + ")\\.runtime_value\\}\\})"
	RuntimeValueReplacementPlaceholderFormat = "{{" + kurtosisNamespace + ":%v:%v.runtime_value}}"

	subExpNotFound = -1
)

// The compiled regular expression to do IP address replacements
// Treat this as a constant
var compiledRuntimeValueReplacementRegex = regexp.MustCompile(runtimeValueReplacementRegex)

func ReplaceRuntimeValueInString(originalString string, recipeEngine *runtime_value_store.RuntimeValueStore) (string, error) {
	matches := compiledRuntimeValueReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		selectedRuntimeValue, err := getRuntimeValueFromRegexMatch(match, recipeEngine)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error happened getting runtime value from regex match '%v'", match)
		}
		allMatchIndex := compiledRuntimeValueReplacementRegex.SubexpIndex(allSubgroupName)
		if allMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceNameSubgroupName, compiledRuntimeValueReplacementRegex.String())
		}
		allMatch := match[allMatchIndex]
		switch value := selectedRuntimeValue.(type) {
		case starlark.String:
			replacedString = strings.Replace(replacedString, allMatch, value.GoString(), singleMatch)
		default:
			replacedString = strings.Replace(replacedString, allMatch, value.String(), singleMatch)
		}
	}
	return replacedString, nil
}

func GetOrReplaceRuntimeValueFromString(originalString string, runtimeValueStore *runtime_value_store.RuntimeValueStore) (starlark.Comparable, error) {
	matches := compiledRuntimeValueReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	if len(matches) == 1 && len(matches[0][0]) == len(originalString) {
		return getRuntimeValueFromRegexMatch(matches[0], runtimeValueStore)
	} else {
		runtimeValue, err := ReplaceRuntimeValueInString(originalString, runtimeValueStore)
		return starlark.String(runtimeValue), err
	}
}

func ContainsRuntimeValue(originalString string) ([]string, bool) {
	matches := compiledRuntimeValueReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	if len(matches) == 0 {
		return []string{}, false
	}
	runtimeValues := make([]string, len(matches))
	for i, match := range matches {
		runtimeValues[i] = match[0]
	}
	return runtimeValues, true
}

func getRuntimeValueFromRegexMatch(match []string, runtimeValueStore *runtime_value_store.RuntimeValueStore) (starlark.Comparable, error) {
	runtimeValueMatchIndex := compiledRuntimeValueReplacementRegex.SubexpIndex(runtimeValueSubgroupName)
	if runtimeValueMatchIndex == subExpNotFound {
		return nil, stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", runtimeValueSubgroupName, compiledRuntimeValueReplacementRegex.String())
	}
	runtimeValueFieldMatchIndex := compiledRuntimeValueReplacementRegex.SubexpIndex(runtimeValueFieldSubgroupName)
	if runtimeValueFieldMatchIndex == subExpNotFound {
		return nil, stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", runtimeValueFieldSubgroupName, compiledRuntimeValueReplacementRegex.String())
	}
	runtimeValue, err := runtimeValueStore.GetValue(match[runtimeValueMatchIndex])
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error happened getting runtime value '%v'", match[runtimeValueMatchIndex])
	}
	selectedRuntimeValue, found := runtimeValue[match[runtimeValueFieldMatchIndex]]
	if !found {
		return nil, stacktrace.NewError("An error happened getting runtime value field '%v' '%v'", match[runtimeValueMatchIndex], match[runtimeValueFieldMatchIndex])
	}
	return selectedRuntimeValue, nil
}

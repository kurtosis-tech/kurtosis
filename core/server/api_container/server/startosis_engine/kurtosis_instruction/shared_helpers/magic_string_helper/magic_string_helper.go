package magic_string_helper

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe_executor"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"regexp"
	"strings"
)

const (
	unlimitedMatches = -1
	singleMatch      = 1

	serviceIdSubgroupName = "service_id"
	allSubgroupName       = "all"
	kurtosisNamespace     = "kurtosis"
	// The placeholder format & regex should align
	ipAddressReplacementRegex             = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + serviceIdSubgroupName + ">" + service.ServiceIdRegexp + ")\\.ip_address\\}\\})"
	IpAddressReplacementPlaceholderFormat = "{{" + kurtosisNamespace + ":%v.ip_address}}"

	factNameArgName      = "fact_name"
	factNameSubgroupName = "fact_name"

	factReplacementRegex             = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + serviceIdSubgroupName + ">" + service.ServiceIdRegexp + ")" + ":(?P<" + factNameArgName + ">" + service.ServiceIdRegexp + ")\\.fact\\}\\})"
	FactReplacementPlaceholderFormat = "{{" + kurtosisNamespace + ":%v:%v.fact}}"

	runtimeValueSubgroupName      = "runtime_value"
	runtimeValueFieldSubgroupName = "runtime_value_field"

	runtimeValueReplacementRegex             = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + runtimeValueSubgroupName + ">" + service.ServiceIdRegexp + ")" + ":(?P<" + runtimeValueFieldSubgroupName + ">" + service.ServiceIdRegexp + ")\\.runtime_value\\}\\})"
	RuntimeValueReplacementPlaceholderFormat = "{{" + kurtosisNamespace + ":%v:%v.runtime_value}}"

	subExpNotFound = -1
)

// The compiled regular expression to do IP address replacements
// Treat this as a constant
var (
	compiledRegex                        = regexp.MustCompile(ipAddressReplacementRegex)
	compiledFactReplacementRegex         = regexp.MustCompile(factReplacementRegex)
	compiledRuntimeValueReplacementRegex = regexp.MustCompile(runtimeValueReplacementRegex)
)

func MakeWaitInterpretationReturnValue(serviceId service.ServiceID, factName string) starlark.String {
	fact := starlark.String(fmt.Sprintf(FactReplacementPlaceholderFormat, serviceId, factName))
	return fact
}

func ReplaceIPAddressInString(originalString string, network service_network.ServiceNetwork, argNameForLogigng string) (string, error) {
	matches := compiledRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		serviceIdMatchIndex := compiledRegex.SubexpIndex(serviceIdSubgroupName)
		if serviceIdMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledRegex.String())
		}
		serviceId := service.ServiceID(match[serviceIdMatchIndex])
		ipAddress, found := network.GetIPAddressForService(serviceId)
		if !found {
			return "", stacktrace.NewError("'%v' depends on the IP address of '%v' but we don't have any registrations for it", argNameForLogigng, serviceId)
		}
		ipAddressStr := ipAddress.String()
		allMatchIndex := compiledRegex.SubexpIndex(allSubgroupName)
		if allMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledRegex.String())
		}
		allMatch := match[allMatchIndex]
		replacedString = strings.Replace(replacedString, allMatch, ipAddressStr, singleMatch)
	}
	return replacedString, nil
}

func ReplaceFactsInString(originalString string, factsEngine *facts_engine.FactsEngine) (string, error) {
	matches := compiledFactReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		serviceIdMatchIndex := compiledFactReplacementRegex.SubexpIndex(serviceIdSubgroupName)
		if serviceIdMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledFactReplacementRegex.String())
		}
		factNameMatchIndex := compiledFactReplacementRegex.SubexpIndex(factNameSubgroupName)
		if factNameMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledFactReplacementRegex.String())
		}
		factValues, err := factsEngine.FetchLatestFactValues(facts_engine.GetFactId(match[serviceIdMatchIndex], match[factNameMatchIndex]))
		if err != nil {
			return "", stacktrace.Propagate(err, "There was an error fetching fact value while replacing string '%v' '%v' ", match[serviceIdMatchIndex], match[factNameMatchIndex])
		}
		allMatchIndex := compiledFactReplacementRegex.SubexpIndex(allSubgroupName)
		if allMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledFactReplacementRegex.String())
		}
		allMatch := match[allMatchIndex]
		replacedString = strings.Replace(replacedString, allMatch, factValues[len(factValues)-1].GetStringValue(), singleMatch)
	}
	return replacedString, nil
}

func ReplaceRuntimeValueInString(originalString string, recipeEngine *recipe_executor.RuntimeValueStore) (string, error) {
	matches := compiledRuntimeValueReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		runtimeValueMatchIndex := compiledRuntimeValueReplacementRegex.SubexpIndex(runtimeValueSubgroupName)
		if runtimeValueMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", runtimeValueSubgroupName, compiledRuntimeValueReplacementRegex.String())
		}
		runtimeValueFieldMatchIndex := compiledRuntimeValueReplacementRegex.SubexpIndex(runtimeValueFieldSubgroupName)
		if runtimeValueFieldMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", runtimeValueFieldSubgroupName, compiledRuntimeValueReplacementRegex.String())
		}
		runtimeValue, err := recipeEngine.GetValue(match[runtimeValueMatchIndex])
		if err != nil {
			return "", stacktrace.Propagate(err, "An error happened getting runtime value '%v'", match[runtimeValueMatchIndex])
		}
		allMatchIndex := compiledRuntimeValueReplacementRegex.SubexpIndex(allSubgroupName)
		if allMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledFactReplacementRegex.String())
		}
		allMatch := match[allMatchIndex]
		replacedString = strings.Replace(replacedString, allMatch, runtimeValue[match[runtimeValueFieldMatchIndex]].String(), singleMatch)
	}
	return replacedString, nil
}

func GetRuntimeValueFromString(originalString string, runtimeValueStore *recipe_executor.RuntimeValueStore) (starlark.Comparable, error) {
	matches := compiledRuntimeValueReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	if len(matches) != 1 {
		return nil, stacktrace.NewError("More than one match was found on regexp '%v'. Match count '%v'", compiledRuntimeValueReplacementRegex.String(), len(matches))

	}
	match := matches[0]
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

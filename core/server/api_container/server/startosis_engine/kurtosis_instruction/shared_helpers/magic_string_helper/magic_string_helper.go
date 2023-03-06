package magic_string_helper

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"regexp"
	"strings"
)

const (
	unlimitedMatches = -1
	singleMatch      = 1

	serviceNameSubgroupName = "name"
	allSubgroupName         = "all"
	kurtosisNamespace       = "kurtosis"
	// The placeholder format & regex should align
	ipAddressReplacementRegex = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + serviceNameSubgroupName + ">" + service.ServiceNameRegex + ")\\.ip_address\\}\\})"
	hostnameReplacementRegex  = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + serviceNameSubgroupName + ">" + service.ServiceNameRegex + ")\\.hostname\\}\\})"

	runtimeValueSubgroupName      = "runtime_value"
	runtimeValueFieldSubgroupName = "runtime_value_field"
	runtimeValueKeyRegexp         = "[a-zA-Z0-9-_\\.]+"

	runtimeValueReplacementRegex             = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + runtimeValueSubgroupName + ">" + service.ServiceNameRegex + ")" + ":(?P<" + runtimeValueFieldSubgroupName + ">" + runtimeValueKeyRegexp + ")\\.runtime_value\\}\\})"
	RuntimeValueReplacementPlaceholderFormat = "{{" + kurtosisNamespace + ":%v:%v.runtime_value}}"

	subExpNotFound = -1
)

// The compiled regular expression to do IP address replacements
// Treat this as a constant
var (
	compiledIpAddressReplacementRegex    = regexp.MustCompile(ipAddressReplacementRegex)
	compiledHostnameReplacementRegex     = regexp.MustCompile(hostnameReplacementRegex)
	compiledRuntimeValueReplacementRegex = regexp.MustCompile(runtimeValueReplacementRegex)
)

func ReplaceIPAddressAndHostnameInString(originalString string, network service_network.ServiceNetwork, argNameForLogging string) (string, error) {
	stringWithIpAddressReplaced, err := replaceRegexpMatchesWithString(
		compiledIpAddressReplacementRegex,
		originalString,
		argNameForLogging,
		network,
		func(serviceRegistration *service.ServiceRegistration) string {
			return serviceRegistration.GetPrivateIP().String()
		},
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred replacing the IP address")
	}

	stringWithIpAddressAndHostnameReplaced, err := replaceRegexpMatchesWithString(
		compiledHostnameReplacementRegex,
		stringWithIpAddressReplaced,
		argNameForLogging,
		network,
		func(serviceRegistration *service.ServiceRegistration) string {
			return serviceRegistration.GetHostname()
		},
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred replacing the hostname")
	}
	return stringWithIpAddressAndHostnameReplaced, nil
}

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

func GetRuntimeValueFromString(originalString string, runtimeValueStore *runtime_value_store.RuntimeValueStore) (starlark.Comparable, error) {
	matches := compiledRuntimeValueReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	if len(matches) == 1 && len(matches[0][0]) == len(originalString) {
		return getRuntimeValueFromRegexMatch(matches[0], runtimeValueStore)
	} else {
		runtimeValue, err := ReplaceRuntimeValueInString(originalString, runtimeValueStore)
		return starlark.String(runtimeValue), err
	}
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

func replaceRegexpMatchesWithString(regexpToReplace *regexp.Regexp, originalString string, argNameForLogigng string, network service_network.ServiceNetwork, stringExtractor func(serviceRegistration *service.ServiceRegistration) string) (string, error) {
	matches := regexpToReplace.FindAllStringSubmatch(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		serviceNameMatchIndex := compiledIpAddressReplacementRegex.SubexpIndex(serviceNameSubgroupName)
		if serviceNameMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceNameSubgroupName, regexpToReplace.String())
		}
		serviceName := service.ServiceName(match[serviceNameMatchIndex])
		serviceRegistration, found := network.GetServiceRegistration(serviceName)
		if !found {
			return "", stacktrace.NewError("'%v' depends on the hostname and IP address of '%v' but we don't have any registrations for it", argNameForLogigng, serviceName)
		}
		stringToInject := stringExtractor(serviceRegistration)
		allMatchIndex := regexpToReplace.SubexpIndex(allSubgroupName)
		if allMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceNameSubgroupName, regexpToReplace.String())
		}
		allMatch := match[allMatchIndex]
		replacedString = strings.Replace(replacedString, allMatch, stringToInject, singleMatch)
	}
	return replacedString, nil
}

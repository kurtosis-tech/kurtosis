package shared_helpers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe_executor"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"regexp"
	"strings"
)

const (
	unlimitedMatches = -1
	singleMatch      = 1

	callerPosition = 1

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

// GetCallerPositionFromThread gets you the position (line, col, filename) from where this function is called
// We pick the first position on the stack based on this https://github.com/google/starlark-go/blob/eaacdf22efa54ae03ea2ec60e248be80d0cadda0/starlark/eval.go#L136
// As far as I understand, when we call this function from any of the `GenerateXXXBuiltIn` the following occurs
// 1. the 0th position is the builtin, the pos/col on it are 0,0
// 2. the 1st position is the function itself, line is the line of the function, col is the position of the opening parenthesis
// 3. the 2nd position is whatever calls the function, so if its nested in another function its that
// I reckon the stack is built on top of a queue or something, otherwise I'd expect the last item to contain the calling function too
func GetCallerPositionFromThread(thread *starlark.Thread) *kurtosis_instruction.InstructionPosition {
	// TODO(gb): can do better by returning the entire callstack positions, but it's a good start
	// As the bottom of the stack is guaranteed to be a built in based on above,
	// The 2nd item is the caller, this should always work when called from a GenerateXXXBuiltIn context
	// We panic to eject early in case a bug occurs.
	if thread.CallStackDepth() < 2 {
		panic("Call stack needs to contain at least 2 items for us to get the callers position. This is a Kurtosis Bug.")
	}
	callFrame := thread.CallStack().At(callerPosition)
	return kurtosis_instruction.NewInstructionPosition(callFrame.Pos.Line, callFrame.Pos.Col, callFrame.Pos.Filename())
}

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

func ReplaceRuntimeValueInString(originalString string, recipeEngine *recipe_executor.RecipeExecutor) (string, error) {
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
		runtimeValue := recipeEngine.GetValue(match[runtimeValueMatchIndex])
		allMatchIndex := compiledRuntimeValueReplacementRegex.SubexpIndex(allSubgroupName)
		if allMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledFactReplacementRegex.String())
		}
		allMatch := match[allMatchIndex]
		replacedString = strings.Replace(replacedString, allMatch, runtimeValue[match[runtimeValueFieldMatchIndex]], singleMatch)
	}
	return replacedString, nil
}

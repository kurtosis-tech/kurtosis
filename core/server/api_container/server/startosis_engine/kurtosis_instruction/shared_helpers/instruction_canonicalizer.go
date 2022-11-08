package shared_helpers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"sort"
	"strings"
)

const (
	initialIndentationLevel  = 1
	starlarkTimeKeyComponent = "unix"
)

func CanonicalizeInstruction(instructionName string, serializedKwargs starlark.StringDict, position *kurtosis_instruction.InstructionPosition) string {
	buffer := new(strings.Builder)
	buffer.WriteString(fmt.Sprintf("# from: %s\n", position.String()))
	buffer.WriteString(instructionName)
	buffer.WriteString("(")

	//sort the key of the map for determinism
	var sortedArgName []string
	for argName := range serializedKwargs {
		sortedArgName = append(sortedArgName, argName)
	}
	sort.Strings(sortedArgName)

	// print each arg depending on its type
	for _, argName := range sortedArgName {
		genericArgValue, found := serializedKwargs[argName]
		if !found {
			panic(fmt.Sprintf("Couldn't find a value for the key '%s' in the canonical instruction argument map ('%v'). This is unexpected and a bug in Kurtosis", argName, serializedKwargs))
		}
		buffer.WriteString(fmt.Sprintf("\n%s%s=%s,", indentPrefixString(initialIndentationLevel), argName, canonicalizeArgValue(genericArgValue, false, 1)))
	}
	buffer.WriteString("\n)")
	return buffer.String()
}

func canonicalizeArgValue(genericArgValue starlark.Value, newline bool, indent int) string {
	var stringifiedArg string
	switch argValue := genericArgValue.(type) {
	case starlark.NoneType, starlark.Bool, starlark.String, starlark.Bytes, starlark.Int, starlark.Float:
		stringifiedArg = argValue.String()
	case time.Time:
		timestamp, err := argValue.Attr(starlarkTimeKeyComponent)
		if err != nil {
			panic(fmt.Sprintf("Unable to retrieve '%s' component from Starlark time object '%s'. This is unexpected", starlarkTimeKeyComponent, argValue.String()))
		}
		// This can be made more readable by doing the full time.time(year, month, day, ....)
		// For now it works just fine
		stringifiedArg = fmt.Sprintf("time.from_timestamp(%d)", timestamp)
	case time.Duration:
		stringifiedArg = "time.parse_duration(" + argValue.String() + ")"
	case *starlark.List:
		stringifiedList := stringifyIterable(argValue, argValue.Len(), indent)
		stringifiedArg = fmt.Sprintf("[%s\n%s]", strings.Join(stringifiedList, ","), indentPrefixString(indent))
	case *starlark.Set:
		stringifiedSet := stringifyIterable(argValue, argValue.Len(), indent)
		stringifiedArg = fmt.Sprintf("{%s\n%s}", strings.Join(stringifiedSet, ","), indentPrefixString(indent))
	case starlark.Tuple:
		stringifiedTuple := stringifyIterable(argValue, argValue.Len(), indent)
		stringifiedArg = fmt.Sprintf("(%s\n%s)", strings.Join(stringifiedTuple, ","), indentPrefixString(indent))
	case *starlark.Dict:
		allKeys := argValue.Keys()
		stringifiedElement := make([]string, len(allKeys))
		idx := 0
		for _, key := range allKeys {
			value, found, err := argValue.Get(key)
			if err != nil || !found {
				panic(fmt.Sprintf("Iterating over all keys from the struct, the key '%s' could not be found ('%v'). This is unexpected and a bug in Kurtosis", key, argValue))
			}
			stringifiedElement[idx] = fmt.Sprintf("%s: %s", canonicalizeArgValue(key, true, indent+1), canonicalizeArgValue(value, false, indent+1))
			idx++
		}
		sort.Strings(stringifiedElement)
		stringifiedArg = fmt.Sprintf("{%s\n%s}", strings.Join(stringifiedElement, ","), indentPrefixString(indent))
	case *starlarkstruct.Struct:
		allAttributes := argValue.AttrNames()
		sort.Strings(allAttributes)
		stringifiedElement := make([]string, len(allAttributes))
		idx := 0
		for _, attributeName := range allAttributes {
			attributeValue, err := argValue.Attr(attributeName)
			if err != nil {
				panic(fmt.Sprintf("Iterating over all keys from the struct, the key '%s' could not be found ('%v'). This is unexpected and a bug in Kurtosis", attributeName, argValue))
			}
			stringifiedElement[idx] = fmt.Sprintf("\n%s%s=%s", indentPrefixString(indent+1), attributeName, canonicalizeArgValue(attributeValue, false, indent+1))
			idx++
		}
		stringifiedArg = fmt.Sprintf("struct(%s\n%s)", strings.Join(stringifiedElement, ","), indentPrefixString(indent))
	default:
		stringifiedArg = fmt.Sprintf("UNSUPPORTED_TYPE['%v']", argValue)
	}

	var resultBuffer strings.Builder
	if newline {
		resultBuffer.WriteString("\n")
		resultBuffer.WriteString(indentPrefixString(indent))
	}
	resultBuffer.WriteString(stringifiedArg)
	return resultBuffer.String()
}

func stringifyIterable(iterable starlark.Iterable, length int, currentIndentationLevel int) []string {
	stringifiedIterable := make([]string, length)
	iterator := iterable.Iterate()
	defer iterator.Done()
	var item starlark.Value
	for idx := 0; iterator.Next(&item); idx++ {
		stringifiedIterable[idx] = canonicalizeArgValue(item, true, currentIndentationLevel+1)
	}
	return stringifiedIterable
}

func indentPrefixString(indent int) string {
	var resultBuffer strings.Builder
	for idx := 0; idx < indent; idx++ {
		resultBuffer.WriteString("\t")
	}
	return resultBuffer.String()
}

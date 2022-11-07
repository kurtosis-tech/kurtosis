package shared_helpers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"sort"
	"strings"
)

const (
	initialIndentationLevel = 1
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
	case starlark.String:
		stringifiedArg = argValue.String()
	case starlark.Int:
		stringifiedArg = argValue.String()
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
	case *starlark.List:
		stringifiedElement := make([]string, argValue.Len())
		for idx := 0; idx < argValue.Len(); idx++ {
			attributeValue := argValue.Index(idx)
			stringifiedElement[idx] = canonicalizeArgValue(attributeValue, true, indent+1)
		}
		stringifiedArg = fmt.Sprintf("[%s\n%s]", strings.Join(stringifiedElement, ","), indentPrefixString(indent))
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

func indentPrefixString(indent int) string {
	var resultBuffer strings.Builder
	for idx := 0; idx < indent; idx++ {
		resultBuffer.WriteString("\t")
	}
	return resultBuffer.String()
}

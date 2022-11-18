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
	initialIndentationLevel  = 0
	starlarkTimeKeyComponent = "unix"
)

var (
	SingleLineCanonicalizer = newSingleLineInstructionCanonicalizer()
	MultiLineCanonicalizer  = newMultiLineInstructionCanonicalizer()
)

type kurtosisInstructionCanonicalizer struct {
	singleLineMode bool
}

func newSingleLineInstructionCanonicalizer() *kurtosisInstructionCanonicalizer {
	return &kurtosisInstructionCanonicalizer{
		singleLineMode: true,
	}
}

func newMultiLineInstructionCanonicalizer() *kurtosisInstructionCanonicalizer {
	return &kurtosisInstructionCanonicalizer{
		singleLineMode: false,
	}
}

func (canonicalizer *kurtosisInstructionCanonicalizer) CanonicalizeInstruction(instructionName string, serializedArgs []starlark.Value, serializedKwargs starlark.StringDict, position *kurtosis_instruction.InstructionPosition) string {
	buffer := new(strings.Builder)
	if !canonicalizer.singleLineMode {
		buffer.WriteString(fmt.Sprintf("# from: %s\n", position.String()))
	}
	buffer.WriteString(canonicalizer.canonicalizeInstruction(instructionName, serializedArgs, serializedKwargs, initialIndentationLevel))
	return buffer.String()
}

func (canonicalizer *kurtosisInstructionCanonicalizer) canonicalizeInstruction(instructionName string, serializedArgs []starlark.Value, serializedKwargs starlark.StringDict, indentLevel int) string {
	buffer := new(strings.Builder)
	buffer.WriteString(instructionName)
	buffer.WriteString(fmt.Sprintf("(%s", canonicalizer.newlineIndent(indentLevel+1)))

	// print each positional arg
	canonicalizedArgs := make([]string, len(serializedArgs)+len(serializedKwargs))
	for idx, genericArgValue := range serializedArgs {
		canonicalizedArgs[idx] = canonicalizer.canonicalizeArgValue(genericArgValue, indentLevel+1)
	}

	// print each named arg, sorting them first for determinism
	var sortedKwargName []string
	for kwargName := range serializedKwargs {
		sortedKwargName = append(sortedKwargName, kwargName)
	}
	sort.Strings(sortedKwargName)

	idx := len(serializedArgs)
	for _, kwargName := range sortedKwargName {
		genericKwargValue, found := serializedKwargs[kwargName]
		if !found {
			panic(fmt.Sprintf("Couldn't find a value for the key '%s' in the canonical instruction argument map ('%v'). This is unexpected and a bug in Kurtosis", kwargName, serializedKwargs))
		}
		canonicalizedArgs[idx] = fmt.Sprintf("%s=%s", kwargName, canonicalizer.canonicalizeArgValue(genericKwargValue, indentLevel+1))
		idx++
	}
	buffer.WriteString(strings.Join(canonicalizedArgs, canonicalizer.separator(indentLevel+1)))

	// finalize function closing the parenthesis
	buffer.WriteString(fmt.Sprintf("%s)", canonicalizer.newlineIndent(indentLevel)))
	return buffer.String()
}

func (canonicalizer *kurtosisInstructionCanonicalizer) canonicalizeArgValue(genericArgValue starlark.Value, indent int) string {
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
		stringifiedList := canonicalizer.stringifyIterable(argValue, argValue.Len(), indent)
		stringifiedArg = fmt.Sprintf("[%s%s%s]", canonicalizer.newlineIndent(indent+1), strings.Join(stringifiedList, canonicalizer.separator(indent+1)), canonicalizer.newlineIndent(indent))
	case *starlark.Set:
		stringifiedSet := canonicalizer.stringifyIterable(argValue, argValue.Len(), indent)
		stringifiedArg = fmt.Sprintf("{%s%s%s}", canonicalizer.newlineIndent(indent+1), strings.Join(stringifiedSet, canonicalizer.separator(indent+1)), canonicalizer.newlineIndent(indent))
	case starlark.Tuple:
		stringifiedTuple := canonicalizer.stringifyIterable(argValue, argValue.Len(), indent)
		stringifiedArg = fmt.Sprintf("(%s%s%s)", canonicalizer.newlineIndent(indent+1), strings.Join(stringifiedTuple, canonicalizer.separator(indent+1)), canonicalizer.newlineIndent(indent))
	case *starlark.Dict:
		allKeys := argValue.Keys()
		stringifiedElement := make([]string, len(allKeys))
		idx := 0
		for _, key := range allKeys {
			value, found, err := argValue.Get(key)
			if err != nil || !found {
				panic(fmt.Sprintf("Iterating over all keys from the struct, the key '%s' could not be found ('%v'). This is unexpected and a bug in Kurtosis", key, argValue))
			}
			stringifiedElement[idx] = fmt.Sprintf("%s: %s", canonicalizer.canonicalizeArgValue(key, indent+1), canonicalizer.canonicalizeArgValue(value, indent+1))
			idx++
		}
		sort.Strings(stringifiedElement)
		stringifiedArg = fmt.Sprintf("{%s%s%s}", canonicalizer.newlineIndent(indent+1), strings.Join(stringifiedElement, canonicalizer.separator(indent+1)), canonicalizer.newlineIndent(indent))
	case *starlarkstruct.Struct:
		// building struct is just calling the function with the argument matching the struct attributes
		structKwargs := starlark.StringDict{}
		for _, attributeName := range argValue.AttrNames() {
			attributeValue, err := argValue.Attr(attributeName)
			if err != nil {
				panic(fmt.Sprintf("Iterating over all keys from the struct, the key '%s' could not be found ('%v'). This is unexpected and a bug in Kurtosis", attributeName, argValue))
			}
			structKwargs[attributeName] = attributeValue
		}
		return canonicalizer.canonicalizeInstruction(argValue.Type(), kurtosis_instruction.NoArgs, structKwargs, indent)
	default:
		stringifiedArg = fmt.Sprintf("UNSUPPORTED_TYPE['%v']", argValue)
	}

	var resultBuffer strings.Builder
	resultBuffer.WriteString(stringifiedArg)
	return resultBuffer.String()
}

func (canonicalizer *kurtosisInstructionCanonicalizer) stringifyIterable(iterable starlark.Iterable, length int, currentIndentationLevel int) []string {
	stringifiedIterable := make([]string, length)
	iterator := iterable.Iterate()
	defer iterator.Done()
	var item starlark.Value
	for idx := 0; iterator.Next(&item); idx++ {
		stringifiedIterable[idx] = canonicalizer.canonicalizeArgValue(item, currentIndentationLevel+1)
	}
	return stringifiedIterable
}

func (canonicalizer *kurtosisInstructionCanonicalizer) newlineIndent(indent int) string {
	if canonicalizer.singleLineMode {
		return ""
	}
	var resultBuffer strings.Builder
	resultBuffer.WriteString("\n")
	for idx := 0; idx < indent; idx++ {
		resultBuffer.WriteString("\t")
	}
	return resultBuffer.String()
}

func (canonicalizer *kurtosisInstructionCanonicalizer) separator(indent int) string {
	if canonicalizer.singleLineMode {
		return ", "
	}
	return fmt.Sprintf(",%s", canonicalizer.newlineIndent(indent))
}

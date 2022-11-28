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
	starlarkTimeKeyComponent = "unix"

	argSeparator = ", "
)

func CanonicalizeInstruction(instructionName string, serializedArgs []starlark.Value, serializedKwargs starlark.StringDict) string {
	buffer := new(strings.Builder)
	buffer.WriteString(instructionName)
	buffer.WriteString("(")

	// print each positional arg
	canonicalizedArgs := make([]string, len(serializedArgs)+len(serializedKwargs))
	for idx, genericArgValue := range serializedArgs {
		canonicalizedArgs[idx] = CanonicalizeArgValue(genericArgValue)
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
		canonicalizedArgs[idx] = fmt.Sprintf("%s=%s", kwargName, CanonicalizeArgValue(genericKwargValue))
		idx++
	}
	buffer.WriteString(strings.Join(canonicalizedArgs, ", "))

	// finalize function closing the parenthesis
	buffer.WriteString(")")
	return buffer.String()
}

func CanonicalizeArgValue(genericArgValue starlark.Value) string {
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
		stringifiedList := stringifyIterable(argValue, argValue.Len())
		stringifiedArg = fmt.Sprintf("[%s]", strings.Join(stringifiedList, argSeparator))
	case *starlark.Set:
		stringifiedSet := stringifyIterable(argValue, argValue.Len())
		stringifiedArg = fmt.Sprintf("{%s}", strings.Join(stringifiedSet, argSeparator))
	case starlark.Tuple:
		stringifiedTuple := stringifyIterable(argValue, argValue.Len())
		stringifiedArg = fmt.Sprintf("(%s)", strings.Join(stringifiedTuple, argSeparator))
	case *starlark.Dict:
		allKeys := argValue.Keys()
		stringifiedElement := make([]string, len(allKeys))
		idx := 0
		for _, key := range allKeys {
			value, found, err := argValue.Get(key)
			if err != nil || !found {
				panic(fmt.Sprintf("Iterating over all keys from the struct, the key '%s' could not be found ('%v'). This is unexpected and a bug in Kurtosis", key, argValue))
			}
			stringifiedElement[idx] = fmt.Sprintf("%s: %s", CanonicalizeArgValue(key), CanonicalizeArgValue(value))
			idx++
		}
		sort.Strings(stringifiedElement)
		stringifiedArg = fmt.Sprintf("{%s}", strings.Join(stringifiedElement, argSeparator))
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
		return CanonicalizeInstruction(argValue.Type(), kurtosis_instruction.NoArgs, structKwargs)
	default:
		argValueStr := fmt.Sprintf("UNSUPPORTED_TYPE[%s]", argValue)
		stringifiedArg = fmt.Sprintf("%q", argValueStr)
	}
	return stringifiedArg
}

func stringifyIterable(iterable starlark.Iterable, length int) []string {
	stringifiedIterable := make([]string, length)
	iterator := iterable.Iterate()
	defer iterator.Done()
	var item starlark.Value
	for idx := 0; iterator.Next(&item); idx++ {
		stringifiedIterable[idx] = CanonicalizeArgValue(item)
	}
	return stringifiedIterable
}

package builtin_argument

import (
	"fmt"
	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"sort"
	"strings"
)

const (
	argSeparator = ", "

	starlarkTimeKeyComponent = "unix"
)

// StringifyArgumentValue converts to a Starlark-executable string any starlark.Value object
//
// Its logic is mostly about calling starlark.Value#String(), except for a few complex types like iterables or struct,
// which are handled manually
func StringifyArgumentValue(genericArgValue starlark.Value) string {
	var stringifiedArg string
	switch argValue := genericArgValue.(type) {
	case starlarktime.Time:
		timestamp, err := argValue.Attr(starlarkTimeKeyComponent)
		if err != nil {
			panic(fmt.Sprintf("Unable to retrieve '%s' component from Starlark time object '%s'. This is unexpected", starlarkTimeKeyComponent, argValue.String()))
		}
		// This can be made more readable by doing the full time.time(year, month, day, ....)
		// For now it works just fine
		stringifiedArg = fmt.Sprintf("time.from_timestamp(%d)", timestamp)
	case starlarktime.Duration:
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
			stringifiedElement[idx] = fmt.Sprintf("%s: %s", StringifyArgumentValue(key), StringifyArgumentValue(value))
			idx++
		}
		sort.Strings(stringifiedElement)
		stringifiedArg = fmt.Sprintf("{%s}", strings.Join(stringifiedElement, argSeparator))
	case *starlarkstruct.Struct:
		// serialize constructor
		var structConstructor string
		switch constructor := argValue.Constructor().(type) {
		case starlark.String:
			structConstructor = constructor.GoString()
		default:
			structConstructor = constructor.String()
		}
		// serialize struct attributes
		allAttrs := argValue.AttrNames()
		stringifiedComponents := make([]string, len(allAttrs))
		idx := 0
		for _, attributeName := range allAttrs {
			attributeValue, err := argValue.Attr(attributeName)
			if err != nil {
				panic(fmt.Sprintf("Iterating over all keys from the struct, the key '%s' could not be found ('%v'). This is unexpected and a bug in Kurtosis", attributeName, argValue))
			}
			stringifiedComponents[idx] = fmt.Sprintf("%s=%s", attributeName, StringifyArgumentValue(attributeValue))
			idx++
		}
		stringifiedArg = fmt.Sprintf("%s(%s)", structConstructor, strings.Join(stringifiedComponents, argSeparator))
	default:
		stringifiedArg = argValue.String()
	}
	return stringifiedArg
}

func stringifyIterable(iterable starlark.Iterable, length int) []string {
	stringifiedIterable := make([]string, length)
	iterator := iterable.Iterate()
	defer iterator.Done()
	var item starlark.Value
	for idx := 0; iterator.Next(&item); idx++ {
		stringifiedIterable[idx] = StringifyArgumentValue(item)
	}
	return stringifiedIterable
}

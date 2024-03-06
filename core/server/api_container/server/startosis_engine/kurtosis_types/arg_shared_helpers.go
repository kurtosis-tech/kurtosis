package kurtosis_types

import (
	"fmt"
	"reflect"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

func MakeOptional(argName string) string {
	return fmt.Sprintf("%s?", argName)
}

// TODO: make private once arg_parser don't need it anymore
func SafeCastToStringSlice(expectedStringIterable starlark.Value, argNameForLogging string) ([]string, *startosis_errors.InterpretationError) {
	iterableValue, ok := expectedStringIterable.(starlark.Iterable)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("'%s' argument is expected to be an iterable. Got %s", argNameForLogging, reflect.TypeOf(expectedStringIterable))
	}
	var castValue []string
	iterator := iterableValue.Iterate()
	defer iterator.Done()
	var value starlark.Value
	var index = 0
	for iterator.Next(&value) {
		stringValue, err := SafeCastToString(value, fmt.Sprintf("%v[%v]", argNameForLogging, index))
		if err != nil {
			return nil, err
		}
		castValue = append(castValue, stringValue)
		index += 1
	}
	return castValue, nil
}

// TODO: make private once arg_parser don't need it anymore
func SafeCastToMapStringString(expectedValue starlark.Value, argNameForLogging string) (map[string]string, *startosis_errors.InterpretationError) {
	dictValue, ok := expectedValue.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("'%s' argument is expected to be a dict. Got %s", argNameForLogging, reflect.TypeOf(expectedValue))
	}
	castValue := make(map[string]string)
	for _, key := range dictValue.Keys() {
		stringKey, castErr := SafeCastToString(key, fmt.Sprintf("%v.key:%v", argNameForLogging, key))
		if castErr != nil {
			return nil, castErr
		}
		value, found, dictErr := dictValue.Get(key)
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError("'%s' key in dict '%s' doesn't have a value we could retrieve. This is a Kurtosis bug.", key.String(), argNameForLogging)
		}
		stringValue, castErr := SafeCastToString(value, fmt.Sprintf("%v[\"%v\"]", argNameForLogging, stringKey))
		if castErr != nil {
			return nil, castErr
		}
		castValue[stringKey] = stringValue
	}
	return castValue, nil
}

// TODO: make private once arg_parser don't need it anymore
func SafeCastToMapStringStringPtr(expectedValue starlark.Value, argNameForLogging string) (map[string]*string, *startosis_errors.InterpretationError) {
	dictValue, ok := expectedValue.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("'%s' argument is expected to be a dict. Got %s", argNameForLogging, reflect.TypeOf(expectedValue))
	}
	castValue := make(map[string]*string)
	for _, key := range dictValue.Keys() {
		stringKey, castErr := SafeCastToString(key, fmt.Sprintf("%v.key:%v", argNameForLogging, key))
		if castErr != nil {
			return nil, castErr
		}
		value, found, dictErr := dictValue.Get(key)
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError("'%s' key in dict '%s' doesn't have a value we could retrieve. This is a Kurtosis bug.", key.String(), argNameForLogging)
		}
		stringValue, castErr := SafeCastToString(value, fmt.Sprintf("%v[\"%v\"]", argNameForLogging, stringKey))
		if castErr != nil {
			return nil, castErr
		}
		castValue[stringKey] = &stringValue
	}
	return castValue, nil
}

// TODO: make private once arg_parser don't need it anymore
func SafeCastToMapStringStringSlice(expectedValue starlark.Value, argNameForLogging string) (map[string][]string, *startosis_errors.InterpretationError) {
	dictValue, ok := expectedValue.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("'%s' argument is expected to be a dict. Got %s", argNameForLogging, reflect.TypeOf(expectedValue))
	}
	castValue := make(map[string][]string)
	for _, key := range dictValue.Keys() {
		stringKey, castErr := SafeCastToString(key, fmt.Sprintf("%v.key:%v", argNameForLogging, key))
		if castErr != nil {
			return nil, castErr
		}
		value, found, dictErr := dictValue.Get(key)
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError("'%s' key in dict '%s' doesn't have a value we could retrieve. This is a Kurtosis bug.", key.String(), argNameForLogging)
		}
		stringSliceValue, castErr := SafeCastToStringSlice(value, fmt.Sprintf("%v[\"%v\"]", argNameForLogging, stringKey))
		if castErr != nil {
			return nil, castErr
		}
		castValue[stringKey] = stringSliceValue
	}
	return castValue, nil
}

// TODO: make private once arg_parser don't need it anymore
func SafeCastToString(expectedValueString starlark.Value, argNameForLogging string) (string, *startosis_errors.InterpretationError) {
	castValue, ok := expectedValueString.(starlark.String)
	if !ok {
		return "", startosis_errors.NewInterpretationError("'%s' is expected to be a string. Got %s", argNameForLogging, reflect.TypeOf(expectedValueString))
	}
	return castValue.GoString(), nil
}

func SafeCastToIntegerSlice(starlarkList *starlark.List) ([]int64, error) {
	slice := []int64{}
	for i := 0; i < starlarkList.Len(); i++ {
		value := starlarkList.Index(i)
		starlarkCastedValue, ok := value.(starlark.Int)
		if !ok {
			return nil, stacktrace.NewError("An error occurred when casting element '%v' from slice '%v' to integer", value, starlarkList)
		}
		castedValue, ok := starlarkCastedValue.Int64()
		if !ok {
			return nil, stacktrace.NewError("An error occurred when casting element '%v' to Go integer", castedValue)
		}
		slice = append(slice, castedValue)
	}
	return slice, nil
}

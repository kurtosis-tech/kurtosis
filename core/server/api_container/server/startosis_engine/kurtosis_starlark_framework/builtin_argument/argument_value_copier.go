package builtin_argument

import (
	"fmt"
	"github.com/sirupsen/logrus"
	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"reflect"
)

func DeepCopyArgumentValue[StarlarkValueType starlark.Value](genericArgValue StarlarkValueType) (StarlarkValueType, error) {
	var copiedValue StarlarkValueType
	genericCopiedValue, err := deepCopyArgumentValueInternal(genericArgValue)
	if err != nil {
		return copiedValue, err
	}
	copiedValue, ok := genericCopiedValue.(StarlarkValueType)
	if !ok {
		return copiedValue, fmt.Errorf("Error copying argument value '%s'. Unexpected type returned (original type: '%s', returned type: '%s')", genericArgValue, reflect.TypeOf(genericArgValue), reflect.TypeOf(genericCopiedValue))
	}
	return copiedValue, nil
}

func deepCopyArgumentValueInternal(genericArgValue starlark.Value) (starlark.Value, error) {
	var valueCopy starlark.Value
	var err error
	switch argValue := genericArgValue.(type) {
	case starlark.Bytes,
		starlark.Bool,
		starlark.Float,
		starlark.Int,
		starlark.NoneType,
		starlark.String,
		starlarktime.Time,
		starlarktime.Duration:
		valueCopy = argValue
	case *starlark.List:
		copiedList := make([]starlark.Value, argValue.Len())
		for idx := 0; idx < argValue.Len(); idx++ {
			copiedItem, err := deepCopyArgumentValueInternal(argValue.Index(idx))
			if err != nil {
				return nil, fmt.Errorf("Cannot copy starlark.List object: '%s'. Item at index '%d' failed to be copied. Error was: %s", argValue.String(), idx, err.Error())
			}
			copiedList[idx] = copiedItem
		}
		valueCopy = starlark.NewList(copiedList)
	case *starlark.Set:
		copiedSet := starlark.NewSet(argValue.Len())
		iterator := argValue.Iterate()
		defer iterator.Done()
		var item starlark.Value
		for idx := 0; iterator.Next(&item); idx++ {
			copiedItem, err := deepCopyArgumentValueInternal(item)
			if err != nil {
				return nil, fmt.Errorf("Cannot copy starlark.Set object: '%s'. Item '%s' failed to be copied. Error was: %s", argValue.String(), item, err.Error())
			}
			if err = copiedSet.Insert(copiedItem); err != nil {
				return nil, fmt.Errorf("Cannot copy starlark.Set object: '%s'. Item '%s' could not be persisted to the copy. Error was: %s", argValue.String(), item, err.Error())
			}
		}
		valueCopy = copiedSet
	case starlark.Tuple:
		copiedTuple := make([]starlark.Value, argValue.Len())
		iterator := argValue.Iterate()
		defer iterator.Done()
		var item starlark.Value
		for idx := 0; iterator.Next(&item); idx++ {
			copiedItem, err := deepCopyArgumentValueInternal(item)
			if err != nil {
				return nil, fmt.Errorf("Cannot copy starlark.Tuple object: '%s'. Item '%s' failed to be copied. Error was: %s", argValue.String(), item, err.Error())
			}
			copiedTuple[idx] = copiedItem
		}
		valueCopy = starlark.Tuple(copiedTuple)
	case *starlark.Dict:
		allKeys := argValue.Keys()
		dictCopy := starlark.NewDict(argValue.Len())
		for _, key := range allKeys {
			value, found, err := argValue.Get(key)
			if err != nil || !found {
				return nil, fmt.Errorf("Iterating over all keys from the dictionary: '%s', the key '%s' could not be found. Error was: '%s'", argValue, key, err.Error())
			}
			copiedKey, err := deepCopyArgumentValueInternal(key)
			if err != nil {
				return nil, fmt.Errorf("Iterating over all keys from the dictionary: '%s', the key '%s' could not be copied. Error was: '%s'", argValue, key, err.Error())
			}
			copiedValue, err := deepCopyArgumentValueInternal(value)
			if err != nil {
				return nil, fmt.Errorf("Iterating over all keys from the dictionary: '%s', the value '%s' associated with key '%s' could not be copied. Error was: '%s'", argValue, value, key, err.Error())
			}
			err = dictCopy.SetKey(copiedKey, copiedValue)
			if err != nil {
				return nil, fmt.Errorf("Iterating over all keys from the dictionary: '%s', the value '%s' associated with key '%s' could not be persisted to the copied object. Error was: '%s'", argValue, value, key, err.Error())
			}
		}
		valueCopy = dictCopy
	case *starlarkstruct.Struct:
		copiedStructDict := starlark.StringDict{}
		argValue.ToStringDict(copiedStructDict)
		valueCopy = starlarkstruct.FromStringDict(argValue.Constructor(), copiedStructDict)
	case KurtosisValueType:
		valueCopy, err = argValue.Copy()
	default:
		logrus.Warnf("Cannot copy value of argument '%s' as the type is not handled, returning the original "+
			"object but it might provoke unexpected behaviour downstream", argValue.String())
		valueCopy = argValue
	}
	return valueCopy, err
}

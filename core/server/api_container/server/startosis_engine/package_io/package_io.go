package package_io

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"reflect"
)

const (
	decoderKey = "decode"
)

var (
	noKwargs []starlark.Tuple
)

// DeserializeArgs deserializes the Kurtosis package args, which should be serialized JSON, into a starlark.Value type.
//
// It tries to convert starlark.Dict into starlarkstruct.Struct to allow users to do things like `args.my_param`
// in their code package in place of `args["my_param"]`. See convertValueToStructIfPossible below for more info.
func DeserializeArgs(thread *starlark.Thread, serializedJsonArgs string) (starlark.Value, *startosis_errors.InterpretationError) {
	if !starlarkjson.Module.Members.Has(decoderKey) {
		return nil, startosis_errors.NewInterpretationError("Unable to deserialize package input because Starlark deserializer was not found.")
	}
	decoder, ok := starlarkjson.Module.Members[decoderKey].(*starlark.Builtin)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Unable to deserialize package input because Starlark deserializer could not be loaded.")
	}

	args := []starlark.Value{
		starlark.String(serializedJsonArgs),
	}
	deserializedInputValue, err := decoder.CallInternal(thread, args, noKwargs)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to deserialize package input. Is it a valid JSON?")
	}

	processedValue, interpretationError := convertValueToStructIfPossible(deserializedInputValue)
	if interpretationError != nil {
		// error is not fatal here, we just pass the deserialized starlark.Value object with no transformation
		logrus.Warnf("JSON input successfully deserialized but it failed to be processed by Kurtosis. It will be passed to the package with no transformation. Intput was: \n%v\nError was: \n%v", deserializedInputValue, interpretationError.Error())
		return deserializedInputValue, nil
	}
	return processedValue, nil
}

// convertValueToStructIfPossible tries to convert a starlark.Value type to a starlarkstruct.Struct.
// It no-ops for most Starlark "simple" types, like string, integer, even iterables.
// It is expected to successfully convert starlark.Dict to starlarkstruct.Struct. Note however that there are certain
// cases where this is not possible, like when the dictionary contain non-string keys for example. In this case, this
// function throws an error. This error should never be hit here because what is being passed comes from serialized
// JSON, which cannot contain non-string keys.
func convertValueToStructIfPossible(genericValue starlark.Value) (starlark.Value, *startosis_errors.InterpretationError) {
	switch value := genericValue.(type) {
	case starlark.NoneType, starlark.Bool, starlark.String, starlark.Bytes, starlark.Int, starlark.Float:
		return value, nil
	case *starlark.List, *starlark.Set, starlark.Tuple:
		return value, nil
	case *starlark.Dict:
		// Dictionaries returned by JSON deserialization should have strings as keys. We therefore convert them to struct to facilitate reading from them in Starlark
		stringDict := starlark.StringDict{}
		for _, key := range value.Keys() {
			stringKey, ok := key.(starlark.String)
			if !ok {
				return nil, startosis_errors.NewInterpretationError("JSON input was deserialized in an unexpected manner. It seems some JSON keys were not string, which is currently not supported in Kurtosis (key: '%s', type: '%s')", key, reflect.TypeOf(key))
			}
			genericDictValue, found, err := value.Get(key)
			if !found {
				return nil, startosis_errors.NewInterpretationError("Unexpected error postprocessing JSON input. No value associated with key '%s'", key)
			}
			if err != nil {
				return nil, startosis_errors.NewInterpretationError("Unexpected error postprocessing JSON input (key: '%s')", key)

			}
			postProcessedValue, interpretationError := convertValueToStructIfPossible(genericDictValue)
			if err != nil {
				// do not wrap the interpretation error here as it's coming from a recursive call.
				return nil, interpretationError
			}
			stringDict[stringKey.GoString()] = postProcessedValue
		}
		return starlarkstruct.FromStringDict(starlarkstruct.Default, stringDict), nil
	case *starlarkstruct.Struct:
		return value, nil
	default:
		return nil, startosis_errors.NewInterpretationError("Unexpected type when trying to deserialize package input. Data will be passed to the package with no processing (unsupported type was: '%s').", reflect.TypeOf(genericValue))
	}
}

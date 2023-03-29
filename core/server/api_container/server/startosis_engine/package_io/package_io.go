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
	decoderKey  = "decode"
	encoderKey  = "encode"
	indenterKey = "indent"
)

var (
	noKwargs []starlark.Tuple
)

// DeserializeArgs deserializes the Kurtosis package args, which should be serialized JSON, into a starlark.Value type.
//
// It tries to convert starlark.Dict into starlarkstruct.Struct to allow users to do things like `args.my_param`
// in their code package in place of `args["my_param"]`. See convertValueToDictIfPossible below for more info.
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

	processedValue, interpretationError := convertValueToDictIfPossible(deserializedInputValue)
	if interpretationError != nil {
		// error is not fatal here, we just pass the deserialized starlark.Value object with no transformation
		logrus.Warnf("JSON input successfully deserialized but it failed to be processed by Kurtosis. It will be passed to the package with no transformation. Intput was: \n%v\nError was: \n%v", deserializedInputValue, interpretationError.Error())
		return deserializedInputValue, nil
	}
	return processedValue, nil
}

func SerializeOutputObject(thread *starlark.Thread, outputObject starlark.Value) (string, *startosis_errors.InterpretationError) {
	if !starlarkjson.Module.Members.Has(encoderKey) {
		return "", startosis_errors.NewInterpretationError("Unable to serialize output object because Starlark serializer was not found.")
	}
	encoder, ok := starlarkjson.Module.Members[encoderKey].(*starlark.Builtin)
	if !ok {
		return "", startosis_errors.NewInterpretationError("Unable to serialize output object because Starlark serializer could not be loaded.")
	}

	serializerArgs := []starlark.Value{
		outputObject,
	}
	serializedOutputObject, err := encoder.CallInternal(thread, serializerArgs, noKwargs)
	// remaining errors are not fatal, we just call the String() method instead
	if err != nil {
		logrus.Warnf("Output object couldn't be serialized to JSON. It will be returned as a raw string instead. Intput was: \n%v\nError was: \n%v", outputObject, err)
		return outputObject.String(), nil
	}

	maybeIndentedOutputObject := tryIndentJson(thread, serializedOutputObject)
	maybeIndentedOutputObjectStr, ok := maybeIndentedOutputObject.(starlark.String)
	if !ok {
		logrus.Warnf("Output object was successfully serialier to JSON but the output of the serialized wasn't a string. This is unexpected. It will be returned as a raw string instead. Intput was: \n%v\nSerialized object was:\n%v\nError was: \n%v", outputObject, serializedOutputObject, err)
		return outputObject.String(), nil
	}
	return maybeIndentedOutputObjectStr.GoString(), nil
}

// convertValueToDictIfPossible tries to convert a starlark.Value type to a starlarkstruct.Struct.
// It no-ops for most Starlark "simple" types, like string, integer, even iterables.
// It is expected to successfully convert starlark.Dict to starlarkstruct.Struct. Note however that there are certain
// cases where this is not possible, like when the dictionary contain non-string keys for example. In this case, this
// function throws an error. This error should never be hit here because what is being passed comes from serialized
// JSON, which cannot contain non-string keys.
func convertValueToDictIfPossible(genericValue starlark.Value) (starlark.Value, *startosis_errors.InterpretationError) {
	switch value := genericValue.(type) {
	case starlark.NoneType, starlark.Bool, starlark.String, starlark.Bytes, starlark.Int, starlark.Float:
		return value, nil
	case *starlark.List, *starlark.Set, starlark.Tuple:
		return value, nil
	case *starlark.Dict:
		// Dictionaries returned by JSON deserialization should have strings as keys. We therefore convert them to struct to facilitate reading from them in Starlark
		dict := starlark.NewDict(value.Len())
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
			postProcessedValue, interpretationError := convertValueToDictIfPossible(genericDictValue)
			if interpretationError != nil {
				// do not wrap the interpretation error here as it's coming from a recursive call.
				return nil, interpretationError
			}
			err = dict.SetKey(stringKey, postProcessedValue)
			if err != nil {
				return nil, startosis_errors.NewInterpretationError("Unexpected error setting dictionary key '%s' to value '%v'", stringKey, postProcessedValue)
			}
		}
		return dict, nil
	case *starlarkstruct.Struct:
		return value, nil
	default:
		return nil, startosis_errors.NewInterpretationError("Unexpected type when trying to deserialize package input. Data will be passed to the package with no processing (unsupported type was: '%s').", reflect.TypeOf(genericValue))
	}
}

func tryIndentJson(thread *starlark.Thread, unindentedJson starlark.Value) starlark.Value {
	if !starlarkjson.Module.Members.Has(indenterKey) {
		logrus.Warn("Unable to find Starlark indenter. Serialized result won't be indented")
		return unindentedJson
	}
	indenter, ok := starlarkjson.Module.Members[indenterKey].(*starlark.Builtin)
	if !ok {
		logrus.Warn("Unable to load Starlark indenter. Serialized result won't be indented")
		return unindentedJson
	}

	indenterArgs := []starlark.Value{
		unindentedJson,
	}
	indentedJson, err := indenter.CallInternal(thread, indenterArgs, noKwargs)
	if err != nil {
		logrus.Warnf("Could not indent serialized JSON. Serialized result won't be indented. Error was: \n%v", err.Error())
		return unindentedJson
	}
	return indentedJson
}

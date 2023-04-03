package package_io

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
)

const (
	decoderKey  = "decode"
	encoderKey  = "encode"
	indenterKey = "indent"
)

var (
	noKwargs []starlark.Tuple
)

// DeserializeArgs deserializes the Kurtosis package args, which should be serialized JSON, into a *starlark.Dict type.
func DeserializeArgs(thread *starlark.Thread, serializedJsonArgs string) (*starlark.Dict, *startosis_errors.InterpretationError) {
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
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to deserialize package input '%v'. Is it a valid JSON?", serializedJsonArgs)
	}
	parsedDeserializedInputValue, ok := deserializedInputValue.(*starlark.Dict)
	if !ok {
		// TODO: we could easily support any kind of starlark.Value here
		return nil, startosis_errors.NewInterpretationError("Unable to parse package input '%v' into a dictionary. JSON other than dictionaries aren't support right now.", deserializedInputValue)
	}
	return parsedDeserializedInputValue, nil
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

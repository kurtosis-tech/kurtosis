package builtins

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"strings"
)

const (
	KurtosisModuleName = "kurtosis"

	decoderKey = "decode"
)

var (
	noKwargs []starlark.Tuple
)

func KurtosisModule(thread *starlark.Thread, enclaveEnvVars string) (*starlarkstruct.Module, *startosis_errors.InterpretationError) {
	enclaveEnvVarsStringDict, interpretationErr := convertEnvVarsToDict(thread, enclaveEnvVars)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return &starlarkstruct.Module{
		Name:    KurtosisModuleName,
		Members: *enclaveEnvVarsStringDict,
	}, nil
}

func convertEnvVarsToDict(thread *starlark.Thread, enclaveEnvVars string) (*starlark.StringDict, *startosis_errors.InterpretationError) {
	envVarsDict := starlark.StringDict{}
	if enclaveEnvVars == "" {
		return &envVarsDict, nil
	}
	if !starlarkjson.Module.Members.Has(decoderKey) {
		return nil, startosis_errors.NewInterpretationError("Unable to deserialize enclave env vars because Starlark deserializer was not found.")
	}
	decoder, ok := starlarkjson.Module.Members[decoderKey].(*starlark.Builtin)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Unable to deserialize enclave env vars because Starlark deserializer could not be loaded.")
	}

	args := []starlark.Value{
		starlark.String(enclaveEnvVars),
	}
	deserializedInputValue, err := decoder.CallInternal(thread, args, noKwargs)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to deserialize enclave env vars '%v'. Is it a valid JSON?", enclaveEnvVars)
	}
	parsedDeserializedInputValue, ok := deserializedInputValue.(*starlark.Dict)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Unable to parse enclave env vars '%v' into a dictionary. JSON other than dictionaries aren't support right now.", deserializedInputValue)
	}

	for _, rawEnvVarName := range parsedDeserializedInputValue.Keys() {
		envVarName, ok := rawEnvVarName.(starlark.String)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Environment variable name '%v' was not a string. This is an unexpected bug", rawEnvVarName)
		}
		envVarValue, found, err := parsedDeserializedInputValue.Get(envVarName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "An unexpected error occurred converting the environment variables dictionary for key '%s'", envVarName)
		}
		if !found {
			return nil, startosis_errors.NewInterpretationError("No value found for key '%s' converting the environment variables dictionary", envVarName)
		}
		lowercaseEnvVarName := strings.ToLower(envVarName.GoString())
		envVarsDict[lowercaseEnvVarName] = envVarValue
	}
	return &envVarsDict, nil
}

package builtins

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	KurtosisModuleName = "kurtosis"

	EnvironmentVarsKey = "env"
)

const (
	decoderKey = "decode"
)

var (
	noKwargs []starlark.Tuple
)

// TODO This module was created for storing Kurtosis constatns that then can be used for any Kurtosis module or package
// TODO we use to store the contants related with subnetworks (BLOCKED, ALLOWED) here but these were removed since we deprecate the network partitioning feature
// TODO it's planned to store some another constants like the port protols here, so we are leaving it here for this reason.
func KurtosisModule(thread *starlark.Thread, enclaveEnvVars string) (*starlarkstruct.Module, *startosis_errors.InterpretationError) {
	enclaveEnvVarsDict, interpretationErr := convertEnvVarsToDict(thread, enclaveEnvVars)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return &starlarkstruct.Module{
		Name: KurtosisModuleName,
		Members: starlark.StringDict{
			EnvironmentVarsKey: enclaveEnvVarsDict,
		},
	}, nil
}

func convertEnvVarsToDict(thread *starlark.Thread, enclaveEnvVars string) (*starlark.Dict, *startosis_errors.InterpretationError) {
	if enclaveEnvVars == "" {
		return starlark.NewDict(0), nil
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
	return parsedDeserializedInputValue, nil
}

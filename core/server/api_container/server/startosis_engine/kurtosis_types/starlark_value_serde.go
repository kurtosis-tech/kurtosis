package kurtosis_types

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

type StarlarkValueSerde struct {
	thread      *starlark.Thread
	starlarkEnv starlark.StringDict
}

func NewStarlarkValueSerde(thread *starlark.Thread, starlarkEnv starlark.StringDict) *StarlarkValueSerde {
	return &StarlarkValueSerde{
		thread:      thread,
		starlarkEnv: starlarkEnv,
	}
}

func (serde *StarlarkValueSerde) Serialize(val starlark.Value) string {
	return val.String()
}

func (serde *StarlarkValueSerde) Deserialize(serializedStarlarkValue string) (starlark.Value, *startosis_errors.InterpretationError) {
	val, err := starlark.Eval(serde.thread, "", serializedStarlarkValue, serde.starlarkEnv)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to deserialize starlark value '%s'", serializedStarlarkValue)
	}
	return val, nil
}

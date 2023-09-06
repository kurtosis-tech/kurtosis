package starlark_value_serde

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	starlarkThreadName = "starlark-deserializer-thread"
)

var (
	cachedStarlarkThread *starlark.Thread

	cachedStarlarkEnv starlark.StringDict
)

type StarlarkValueSerde struct {
	predeclared              starlark.StringDict
	kurtosisTypeConstructors []*starlark.Builtin
}

func NewStarlarkValueSerde(predeclared starlark.StringDict, kurtosisTypeConstructors []*starlark.Builtin) *StarlarkValueSerde {
	return &StarlarkValueSerde{
		predeclared:              predeclared,
		kurtosisTypeConstructors: kurtosisTypeConstructors,
	}
}

func (serde *StarlarkValueSerde) SerializeStarlarkValue(val starlark.Value) string {
	return val.String()
}

func SerializeStarlarkValue(val starlark.Value) string {
	return val.String()
}

func (serde *StarlarkValueSerde) DeserializeStarlarkValue(serializedStarlarkValue string) (starlark.Value, *startosis_errors.InterpretationError) {
	if cachedStarlarkThread == nil {
		cachedStarlarkThread = &starlark.Thread{
			Name:       starlarkThreadName,
			Print:      nil,
			Load:       nil,
			OnMaxSteps: nil,
			Steps:      0,
		}
	}
	if cachedStarlarkEnv == nil {
		cachedStarlarkEnv = serde.predeclared
		builtins := serde.kurtosisTypeConstructors
		for _, builtin := range builtins {
			cachedStarlarkEnv[builtin.Name()] = builtin
		}
	}

	val, err := starlark.Eval(cachedStarlarkThread, "", serializedStarlarkValue, cachedStarlarkEnv)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to deserialize starlark value '%s'", serializedStarlarkValue)
	}
	return val, nil
}

func DeserializeStarlarkValue(serializedStarlarkValue string, predeclared starlark.StringDict, kurtosisTypeConstructors []*starlark.Builtin) (starlark.Value, *startosis_errors.InterpretationError) {
	if cachedStarlarkThread == nil {
		cachedStarlarkThread = &starlark.Thread{
			Name:       starlarkThreadName,
			Print:      nil,
			Load:       nil,
			OnMaxSteps: nil,
			Steps:      0,
		}
	}
	if cachedStarlarkEnv == nil {
		cachedStarlarkEnv = predeclared
		builtins := kurtosisTypeConstructors
		for _, builtin := range builtins {
			cachedStarlarkEnv[builtin.Name()] = builtin
		}
	}

	val, err := starlark.Eval(cachedStarlarkThread, "", serializedStarlarkValue, cachedStarlarkEnv)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to deserialize starlark value '%s'", serializedStarlarkValue)
	}
	return val, nil
}

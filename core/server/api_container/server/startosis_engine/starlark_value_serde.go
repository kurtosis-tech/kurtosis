package startosis_engine

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

func SerializeStarlarkValue(val starlark.Value) string {
	return val.String()
}

func DeserializeStarlarkValue(serializedStarlarkValue string) (starlark.Value, *startosis_errors.InterpretationError) {
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
		cachedStarlarkEnv = Predeclared()
		builtins := KurtosisTypeConstructors()
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

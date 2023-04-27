package builtin_argument

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

// BuiltinArgument is the expected way to declare an argument type at each builtin level.
type BuiltinArgument struct {
	Name string

	IsOptional bool

	ZeroValueProvider func() starlark.Value

	Validator func(argumentValue starlark.Value) *startosis_errors.InterpretationError

	Deprecation *starlark_warning.DeprecationNotice
}

func ZeroValueProvider[StarlarkValueType starlark.Value]() starlark.Value {
	var val StarlarkValueType
	return val
}

func (argument BuiltinArgument) IsDeprecated() bool {
	return argument.Deprecation != nil
}

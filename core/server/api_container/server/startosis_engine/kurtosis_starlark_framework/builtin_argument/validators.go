package builtin_argument

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"reflect"
)

func NonEmptyString(value starlark.Value, argNameForLogging string) *startosis_errors.InterpretationError {
	valueStr, ok := value.(starlark.String)
	if !ok {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be a starlark.String but was '%s'", argNameForLogging, reflect.TypeOf(value))
	}
	if len(valueStr.GoString()) == 0 {
		return startosis_errors.NewInterpretationError("Value for '%s' was an empty string. This is disallowed", argNameForLogging)
	}
	return nil
}

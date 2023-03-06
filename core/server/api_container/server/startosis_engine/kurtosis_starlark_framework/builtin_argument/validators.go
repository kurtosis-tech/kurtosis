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

func Uint64InRange(value starlark.Value, argNameForLogging string, min uint64, max uint64) *startosis_errors.InterpretationError {
	valueInt, ok := value.(starlark.Int)
	if !ok {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be an integer between %d and %d, but it was '%s'", argNameForLogging, min, max, reflect.TypeOf(value))
	}
	valueUint64, ok := valueInt.Uint64()
	if !ok {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be an integer between %d and %d, but it was %v", argNameForLogging, min, max, valueInt)
	}
	if valueUint64 < min || valueUint64 > max {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be an integer between %d and %d, but it was %v", argNameForLogging, min, max, valueUint64)
	}
	return nil
}

func FloatInRange(value starlark.Value, argNameForLogging string, min float64, max float64) *startosis_errors.InterpretationError {
	valueStarlarkFloat, ok := value.(starlark.Float)
	if !ok {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be a float between %f and %f, but it was '%s'", argNameForLogging, min, max, reflect.TypeOf(value))
	}
	valueFloat := float64(valueStarlarkFloat)
	if !ok {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be a float between %f and %f, but it was %v", argNameForLogging, min, max, valueFloat)
	}
	if valueFloat < min || valueFloat > max {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be a float between %f and %f, but it was %v", argNameForLogging, min, max, valueFloat)
	}
	return nil
}

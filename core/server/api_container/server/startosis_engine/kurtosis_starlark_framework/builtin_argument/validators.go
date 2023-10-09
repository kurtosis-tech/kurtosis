package builtin_argument

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"reflect"
	"regexp"
	"strings"
	"time"
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

func StringValues(value starlark.Value, argNameForLogging string, acceptableValues []string) *startosis_errors.InterpretationError {
	valueStr, ok := value.(starlark.String)
	if !ok {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be a starlark.String but was '%s'", argNameForLogging, reflect.TypeOf(value))
	}
	for _, acceptableValue := range acceptableValues {
		if acceptableValue == valueStr.GoString() {
			return nil
		}
	}
	return startosis_errors.NewInterpretationError("Invalid argument value for '%s': '%s'. Valid values are %s", argNameForLogging, valueStr.GoString(), strings.Join(acceptableValues, ", "))
}

func StringRegexp(value starlark.Value, argNameForLogging string, mustMatchRegexpStr string) *startosis_errors.InterpretationError {
	mustMatchRegexp := regexp.MustCompile(mustMatchRegexpStr)
	valueStr, ok := value.(starlark.String)
	if !ok {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be a starlark.String but was '%s'", argNameForLogging, reflect.TypeOf(value))
	}
	doesMatch := mustMatchRegexp.MatchString(valueStr.GoString())
	if doesMatch {
		return nil
	}
	return startosis_errors.NewInterpretationError(
		"Argument '%s' must match regexp: '%v'. Its value was '%s'",
		argNameForLogging,
		mustMatchRegexpStr,
		valueStr.GoString(),
	)
}

func Duration(value starlark.Value, attributeName string) *startosis_errors.InterpretationError {
	valueStarlarkStr, ok := value.(starlark.String)
	if !ok {
		return startosis_errors.NewInterpretationError("The '%s' attribute is not a valid string type (was '%s').", attributeName, reflect.TypeOf(value))
	}

	if valueStarlarkStr.GoString() == "" {
		return nil
	}

	_, parseErr := time.ParseDuration(valueStarlarkStr.GoString())
	if parseErr != nil {
		return startosis_errors.WrapWithInterpretationError(parseErr, "The value '%v' of '%s' attribute is not a valid duration string format", valueStarlarkStr.GoString(), attributeName)
	}

	return nil
}

func DurationOrNone(value starlark.Value, attributeName string) *startosis_errors.InterpretationError {
	// the value can be a string duration, or it can be a Starlark none value (which usually is used to disable the argument and his behaviour)
	if _, ok := value.(starlark.NoneType); !ok {
		// we do not accept empty string as a none wait config
		if interpretationErr := NonEmptyString(value, attributeName); interpretationErr != nil {
			return interpretationErr
		}
		return Duration(value, attributeName)
	}
	return nil
}

func RelativeOrRemoteAbsoluteLocator(locatorStarlarkValue starlark.Value, packageId string, argNameForLogging string) *startosis_errors.InterpretationError {

	if err := NonEmptyString(locatorStarlarkValue, argNameForLogging); err != nil {
		return err
	}

	locatorStarlarkStr, ok := locatorStarlarkValue.(starlark.String)
	if !ok {
		return startosis_errors.NewInterpretationError("Value for '%s' was expected to be a starlark.String but was '%s'", argNameForLogging, reflect.TypeOf(locatorStarlarkValue))
	}
	locatorStr := locatorStarlarkStr.GoString()

	logrus.Infof("MONDAY this is the information i have %v %v %v", locatorStarlarkValue, packageId, argNameForLogging)

	// if absolute and local return error
	if isLocalAbsoluteLocator(locatorStr, packageId) {
		return startosis_errors.NewInterpretationError("The locator '%s' set in attribute '%v' is not a 'local relative locator'. Local absolute locators are not allowed you should modified it to be a valid 'local relative locator'", locatorStarlarkStr, argNameForLogging)
	}
	return nil
}

func isLocalAbsoluteLocator(locatorStr string, packageId string) bool {
	logrus.Debug(packageId, locatorStr)
	return strings.HasPrefix(locatorStr, packageId)
}

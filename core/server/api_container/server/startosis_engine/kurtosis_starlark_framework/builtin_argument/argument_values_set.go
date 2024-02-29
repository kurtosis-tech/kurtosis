package builtin_argument

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"reflect"
	"strings"
)

// ArgumentValuesSet object stores both the definition of all argument types for a given builtin and the values passed
// by the Starlark interpreter. It can be seen as a concrete "instantiation" for a given schema
//
// It has multiple functions:
// - It provides a way to get an argument definition from its name
// - It provides a way to access the value of an argument from its name
type ArgumentValuesSet struct {
	argumentsDefinition []*BuiltinArgument

	values []starlark.Value
}

func NewArgumentValuesSet(argumentsDefinition []*BuiltinArgument, values []starlark.Value) *ArgumentValuesSet {
	return &ArgumentValuesSet{
		argumentsDefinition: argumentsDefinition,
		values:              values,
	}
}

func CreateNewArgumentValuesSet(builtinName string, argumentsDefinition []*BuiltinArgument, args starlark.Tuple, kwargs []starlark.Tuple) (*ArgumentValuesSet, *startosis_errors.InterpretationError) {
	argumentValues, err := parseArguments(argumentsDefinition, builtinName, args, kwargs)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Cannot construct '%s' from the provided arguments.", builtinName)
	}
	return &ArgumentValuesSet{
		argumentsDefinition: argumentsDefinition,
		values:              argumentValues,
	}, nil
}

// ExtractArgumentValue is a static generic wrapper around ArgumentValuesSet#ExtractArgumentValue to avoid the pain
// of creating an empty pointer to store the value to.
func ExtractArgumentValue[ArgumentValueType starlark.Value](arguments *ArgumentValuesSet, argumentName string) (ArgumentValueType, error) {
	var val ArgumentValueType
	if err := arguments.ExtractArgumentValue(argumentName, &val); err != nil {
		return val, err
	}
	return val, nil
}

func (arguments *ArgumentValuesSet) GetDefinition() []*BuiltinArgument {
	return arguments.argumentsDefinition
}

// IsSet returns whether an optional argument as been set or not. If the argument is mandatory, it will return true.
func (arguments *ArgumentValuesSet) IsSet(argumentName string) bool {
	argumentIdx, found := getArgumentIndex(arguments, argumentName)
	if !found {
		return false
	}
	argumentDefinition := arguments.argumentsDefinition[argumentIdx]
	if argumentDefinition.IsOptional && arguments.values[argumentIdx] == nil {
		return false
	}
	return true
}

// ExtractArgumentValue compute the value associated with the argumentName and store it in the argumentValuePointer
// It throws an exception if the argument is optional and unset in this ArgumentValuesSet
//
// See the static ExtractArgumentValue function above as an alternative to this function.
func (arguments *ArgumentValuesSet) ExtractArgumentValue(argumentName string, argumentValuePointer interface{}) error {
	argumentIdx, found := getArgumentIndex(arguments, argumentName)
	if !found {
		return fmt.Errorf("Argument '%s' could not be found in schema", argumentName)
	}

	if !arguments.IsSet(argumentName) {
		return fmt.Errorf("Argument '%s' should be set to extract its value", argumentName)
	}

	pointerValue := reflect.ValueOf(argumentValuePointer)
	if pointerValue.Kind() != reflect.Ptr {
		return fmt.Errorf("Unable to extract value for argument '%s'. Need a value pointer to value as input", argumentName)
	}
	paramVar := pointerValue.Elem()
	if !reflect.TypeOf(arguments.values[argumentIdx]).AssignableTo(paramVar.Type()) {
		return fmt.Errorf("Unable to extract value for argument '%s'. Types were not assignable (got '%s', expecting '%s')", argumentName, paramVar.Type(), reflect.TypeOf(arguments.values[argumentIdx]))
	}

	// Deep copy the value to avoid having the value modified downstream in the script execution
	copiedVal, err := DeepCopyArgumentValue(arguments.values[argumentIdx])
	if err != nil {
		return fmt.Errorf("Argument '%s' could not be copied to a separate value", argumentName)
	}

	paramVar.Set(reflect.ValueOf(copiedVal))
	return nil
}

func (arguments *ArgumentValuesSet) String() string {
	var serializedArguments []string
	for idx, argument := range arguments.argumentsDefinition {
		if !arguments.IsSet(argument.Name) {
			continue
		}
		value := arguments.values[idx]
		serializedArgument := fmt.Sprintf("%s=%s", argument.Name, StringifyArgumentValue(value))
		serializedArguments = append(serializedArguments, serializedArgument)
	}
	return fmt.Sprintf("(%s)", strings.Join(serializedArguments, ", "))
}

func parseArguments(argumentDefinitions []*BuiltinArgument, builtinName string, args starlark.Tuple, kwargs []starlark.Tuple) ([]starlark.Value, error) {
	storedValues := make([]starlark.Value, len(argumentDefinitions))
	var pairs []interface{}
	// We add the description argument; which comes with all instructions
	descriptionArgument := CreateDescriptionArgument()
	argumentDefinitions = append(argumentDefinitions, descriptionArgument)
	for idx, argumentDefinition := range argumentDefinitions {
		if argumentDefinition.IsOptional {
			pairs = append(pairs, makeOptional(argumentDefinition.Name))
		} else {
			pairs = append(pairs, argumentDefinition.Name)
		}
		pairs = append(pairs, &storedValues[idx])
	}
	if err := starlark.UnpackArgs(builtinName, args, kwargs, pairs...); err != nil {
		return nil, err
	}
	// validate each type matches the expected before returning
	invalidArgs := map[string]error{}
	for idx, argumentDefinition := range argumentDefinitions {
		argName := argumentDefinition.Name
		argValue := storedValues[idx]
		if argumentDefinition.IsOptional && argValue == nil {
			continue
		}
		if argumentDefinition.Validator != nil {
			if interpretationErr := argumentDefinition.Validator(argValue); interpretationErr != nil {
				// TODO: probably worth making parseArguments return an interpretationError to avoid conversion here
				invalidArgs[argName] = fmt.Errorf(interpretationErr.Error())
				continue
			}
		}
		if reflect.TypeOf(argumentDefinition.ZeroValueProvider()) == nil {
			// it means the type is an interface, we are not able to validate its concrete type at this step
			continue
		}
		if !reflect.TypeOf(argValue).AssignableTo(reflect.TypeOf(argumentDefinition.ZeroValueProvider())) {
			invalidArgs[argName] = fmt.Errorf("the argument '%s' could not be parsed because their type ('%s') did not match the expected ('%s')",
				argName,
				reflect.TypeOf(argValue),
				reflect.TypeOf(argumentDefinition.ZeroValueProvider()))
			continue
		}
	}
	if len(invalidArgs) > 0 {
		serializedErrors, err := serializeErrorsMap(invalidArgs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Some errors happened parsing the arguments, but those errors could not be serialized. This is a Kurtosis internal bug. Error was: %v", err)
		}
		return nil, fmt.Errorf("The following argument(s) could not be parsed or did not pass validation: %s", serializedErrors)
	}
	return storedValues, nil
}

func getArgumentIndex(arguments *ArgumentValuesSet, argumentName string) (int, bool) {
	found := false
	argumentIdx := -1
	for idx, argumentDefinition := range arguments.argumentsDefinition {
		if argumentDefinition.Name == argumentName {
			found = true
			argumentIdx = idx
			break
		}
	}
	return argumentIdx, found
}

func makeOptional(name string) string {
	return fmt.Sprintf("%s?", name)
}

func serializeErrorsMap(errors map[string]error) (string, error) {
	errorsStr := map[string]string{}
	for key, err := range errors {
		errorsStr[key] = err.Error()
	}
	serializedErrors, err := json.Marshal(errorsStr)
	if err != nil {
		return "", err
	}
	return string(serializedErrors), nil
}

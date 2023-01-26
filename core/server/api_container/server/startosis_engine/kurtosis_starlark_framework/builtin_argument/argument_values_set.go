package builtin_argument

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
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
		return nil, startosis_errors.WrapWithInterpretationError(err, "Error creating argument values set as arguments couldn't be parsed")
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
	paramVar.Set(reflect.ValueOf(arguments.values[argumentIdx]))
	return nil
}

func (arguments *ArgumentValuesSet) String() string {
	var serializedArguments []string
	for idx, argument := range arguments.argumentsDefinition {
		if !arguments.IsSet(argument.Name) {
			continue
		}
		value := arguments.values[idx]
		serializedArgument := fmt.Sprintf("%s=%s", argument.Name, shared_helpers.CanonicalizeArgValue(value))
		serializedArguments = append(serializedArguments, serializedArgument)
	}
	return fmt.Sprintf("(%s)", strings.Join(serializedArguments, ", "))
}

func parseArguments(argumentDefinitions []*BuiltinArgument, builtinName string, args starlark.Tuple, kwargs []starlark.Tuple) ([]starlark.Value, error) {
	storedValues := make([]starlark.Value, len(argumentDefinitions))
	var pairs []interface{}
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
	var invalidArgs []string
	for idx, argumentDefinition := range argumentDefinitions {
		if argumentDefinition.IsOptional && storedValues[idx] == nil {
			continue
		}
		if reflect.TypeOf(argumentDefinition.ZeroValueProvider()) == nil {
			// it means the type is an interface, we are not able to validate its concrete type at this step
			continue
		}
		if !reflect.TypeOf(storedValues[idx]).AssignableTo(reflect.TypeOf(argumentDefinition.ZeroValueProvider())) {
			invalidArgs = append(invalidArgs, argumentDefinition.Name)
		}
	}
	if len(invalidArgs) > 0 {
		return nil, fmt.Errorf("The following argument could not be parse because their type did not match the expected: %v", invalidArgs)
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

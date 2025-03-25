package kurtosis_type_constructor

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"reflect"
)

// KurtosisValueTypeDefault is a helper type to easily build a type that implements the KurtosisValueType interface
// straight from a builtin_argument.ArgumentValuesSet, without having to manually implement all the functions
type KurtosisValueTypeDefault struct {
	*starlarkstruct.Struct

	typeName  string
	arguments *builtin_argument.ArgumentValuesSet
}

func CreateKurtosisStarlarkTypeDefault(name string, arguments *builtin_argument.ArgumentValuesSet) (*KurtosisValueTypeDefault, *startosis_errors.InterpretationError) {
	underlyingStruct, interpretationErr := structFromKurtosisStarlarkType(name, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &KurtosisValueTypeDefault{
		Struct:    underlyingStruct,
		typeName:  name,
		arguments: arguments,
	}, nil
}

func ExtractAttrValue[AttrValueType starlark.Value](value *KurtosisValueTypeDefault, attrName string) (AttrValueType, bool, *startosis_errors.InterpretationError) {
	var result AttrValueType
	attrValue, err := value.Attr(attrName)
	if err != nil {
		// starlarkstruct.Struct.Attr() throws an error only when the attribute does not exist. Therefore, we know here
		// this is not a real error, it's just that the attribute name could not be found.
		return result, false, nil
	}
	result, ok := attrValue.(AttrValueType)
	if !ok {
		return result, true, startosis_errors.NewInterpretationError(
			"Attribute '%s' exists but is of type '%s' (requested: '%s')",
			attrValue,
			reflect.TypeOf(attrValue),
			reflect.TypeOf(result))
	}
	return result, true, nil
}

func (value *KurtosisValueTypeDefault) Copy() (*KurtosisValueTypeDefault, error) {
	copiedStructDict := starlark.StringDict{}
	value.ToStringDict(copiedStructDict)
	copiedStruct := starlarkstruct.FromStringDict(value.Constructor(), copiedStructDict)

	argumentValues := make([]starlark.Value, len(value.arguments.GetDefinition()))
	for idx, argumentDefinition := range value.arguments.GetDefinition() {
		if value.arguments.IsSet(argumentDefinition.Name) {
			argumentValue := argumentDefinition.ZeroValueProvider()
			err := value.arguments.ExtractArgumentValue(argumentDefinition.Name, &argumentValue)
			if err != nil {
				return nil, err
			}
			argumentValues[idx] = argumentValue
		}
	}
	return &KurtosisValueTypeDefault{
		Struct:    copiedStruct,
		typeName:  value.typeName,
		arguments: value.arguments,
	}, nil
}

func (value *KurtosisValueTypeDefault) Type() string {
	return value.typeName
}

func (value *KurtosisValueTypeDefault) String() string {
	return fmt.Sprintf("%s%s", value.typeName, value.arguments.String())
}

func structFromKurtosisStarlarkType(name string, arguments *builtin_argument.ArgumentValuesSet) (*starlarkstruct.Struct, *startosis_errors.InterpretationError) {
	structDict := starlark.StringDict{}
	for _, argumentDefinition := range arguments.GetDefinition() {
		attrName := argumentDefinition.Name
		attrValue := argumentDefinition.ZeroValueProvider()
		if argumentDefinition.IsOptional && !arguments.IsSet(attrName) {
			continue
		}
		if err := arguments.ExtractArgumentValue(attrName, &attrValue); err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Error building type '%s'. Argument '%s' could not be extracted.", name, attrName)
		}
		structDict[attrName] = attrValue
	}
	starlarkStruct := starlarkstruct.FromStringDict(starlark.String(name), structDict)
	return starlarkStruct, nil
}

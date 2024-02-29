package builtin_argument

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	DescriptionArgumentName       = "description"
	DescriptionArgumentIsOptional = true
)

// CreateDescriptionArgument This is an argument that gets injected into all instructions
func CreateDescriptionArgument(defaultValue string) *BuiltinArgument {
	return &BuiltinArgument{
		Name:              DescriptionArgumentName,
		IsOptional:        DescriptionArgumentIsOptional,
		ZeroValueProvider: ZeroValueProvider[starlark.String],
		Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
			return NonEmptyString(value, DescriptionArgumentName)
		},
		Deprecation: nil,
	}
}

func GetDescriptionOrFallBack(arguments *ArgumentValuesSet, fallback string) string {
	if arguments.IsSet(DescriptionArgumentName) {
		description, err := ExtractArgumentValue[starlark.String](arguments, DescriptionArgumentName)
		if err == nil {
			return description.GoString()
		}
		logrus.Debugf("An error occurred while extracting user supplied value for '%v'; using fallback '%v'", DescriptionArgumentName, fallback)
	}
	return fallback
}

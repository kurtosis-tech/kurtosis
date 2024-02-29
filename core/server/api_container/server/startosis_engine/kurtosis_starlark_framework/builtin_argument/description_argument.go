package builtin_argument

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	descriptionArgumentName       = "description"
	descriptionArgumentIsOptional = true
)

// createDescriptionArgument This is an argument that gets injected into all instructions
func createDescriptionArgument() *BuiltinArgument {
	return &BuiltinArgument{
		Name:              descriptionArgumentName,
		IsOptional:        descriptionArgumentIsOptional,
		ZeroValueProvider: ZeroValueProvider[starlark.String],
		Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
			return NonEmptyString(value, descriptionArgumentName)
		},
		Deprecation: nil,
	}
}

func GetDescriptionOrFallBack(arguments *ArgumentValuesSet, fallback string) string {
	if arguments.IsSet(descriptionArgumentName) {
		description, err := ExtractArgumentValue[starlark.String](arguments, descriptionArgumentName)
		if err == nil {
			return description.GoString()
		}
		logrus.Debugf("An error occurred while extracting user supplied value for '%v'; using fallback '%v'", descriptionArgumentName, fallback)
	}
	return fallback
}

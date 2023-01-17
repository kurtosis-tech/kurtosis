package kurtosis_starlark_framework

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	builtinName = "test_builtin"

	serviceIdArgName = "service_id"
	serviceId        = starlark.String("datastore-server")

	shouldStartArgName = "should_start"
	shouldStart        = starlark.Bool(true)

	fileName = "main.star"
)

func TestBasicFunctions(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	values := []starlark.Value{
		serviceId,
		shouldStart,
	}
	arguments := builtin_argument.NewArgumentValuesSet(argumentDefinitions, values)
	position := NewKurtosisBuiltinPosition(fileName, 12, 14)

	kurtosisBaseBuiltinInternal := newKurtosisBaseBuiltinInternal(builtinName, position, arguments)
	require.Equal(t, builtinName, kurtosisBaseBuiltinInternal.GetName())
	require.Equal(t, position, kurtosisBaseBuiltinInternal.GetPosition())

	expectedString := fmt.Sprintf(`%s(%s=%s, %s=%s)`, builtinName, serviceIdArgName, serviceId, shouldStartArgName, shouldStart)
	require.Equal(t, expectedString, kurtosisBaseBuiltinInternal.String())
}

func getArgumentDefinitionsWithServiceIdAndShouldStart() []*builtin_argument.BuiltinArgument {
	return []*builtin_argument.BuiltinArgument{
		{
			Name:              serviceIdArgName,
			IsOptional:        false,
			ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
			Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
				panic("should not be called")
			},
		},
		{
			Name:              shouldStartArgName,
			IsOptional:        true,
			ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Bool],
			Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
				panic("should not be called")
			},
		},
	}
}

package kurtosis_starlark_framework

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	builtinName = "test_builtin"

	serviceNameArgName = "service_name"
	serviceName        = starlark.String("datastore-server")

	shouldStartArgName = "should_start"
	shouldStart        = starlark.Bool(true)

	fileName = "main.star"
)

func TestBasicFunctions(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		shouldStart,
	}
	arguments := builtin_argument.NewArgumentValuesSet(argumentDefinitions, values)
	position := NewKurtosisBuiltinPosition(fileName, 12, 14)

	kurtosisBaseBuiltinInternal := newKurtosisBaseBuiltinInternal(builtinName, position, arguments)
	require.Equal(t, builtinName, kurtosisBaseBuiltinInternal.GetName())
	require.Equal(t, position, kurtosisBaseBuiltinInternal.GetPosition())

	expectedString := fmt.Sprintf(`%s(%s=%s, %s=%s)`, builtinName, serviceNameArgName, serviceName, shouldStartArgName, shouldStart)
	require.Equal(t, expectedString, kurtosisBaseBuiltinInternal.String())
}

func getArgumentDefinitionsWithServiceNameAndShouldStart() []*builtin_argument.BuiltinArgument {
	return []*builtin_argument.BuiltinArgument{
		{
			Name:              serviceNameArgName,
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

func Test_printWarningIfArgumentIsDeprecated(t *testing.T) {
	deprecatedDate := starlark_warning.DeprecationDate{
		Day: 20, Month: 11, Year: 2023,
	}

	deprecatedMitigation := "mitigation for config"
	// nolint:exhaustruct
	baseBuiltIn := KurtosisBaseBuiltin{
		Name: "kurtosis_builtin",
	}

	builtinArgs := []*builtin_argument.BuiltinArgument{
		{
			Name: "config",
			Deprecation: starlark_warning.Deprecation(
				deprecatedDate,
				deprecatedMitigation,
				nil,
			),
		},
		{
			Name: "another_key",
			Deprecation: starlark_warning.Deprecation(
				starlark_warning.DeprecationDate{
					Day: 20, Month: 11, Year: 2023,
				},
				"mitigation reason for another_key",
				nil,
			),
		},
	}

	argumentSet := builtin_argument.NewArgumentValuesSet(builtinArgs, []starlark.Value{starlark.String("test"), starlark.String("test")})
	err := printWarningForArguments(argumentSet, &baseBuiltIn)
	require.NoError(t, err)
	warnings := starlark_warning.GetContentFromWarningSet()
	require.Len(t, warnings, 2)

	// just checking only one warning to make sure that the string formatting works
	expectedWarning := fmt.Sprintf("[WARN]: %q field for %q will be deprecated by %v. %v",
		"config",
		"kurtosis_builtin",
		deprecatedDate.GetFormattedDate(),
		deprecatedMitigation,
	)
	require.Contains(t, warnings, expectedWarning)

}

func Test_printWarningForBuiltinIsDeprecated(t *testing.T) {
	deprecatedDate := starlark_warning.DeprecationDate{
		Day: 20, Month: 11, Year: 2023,
	}

	deprecatedMitigation := "mitigation instruction reason"
	// nolint:exhaustruct
	baseBuiltIn := KurtosisBaseBuiltin{
		Name: "kurtosis_builtin",
		Deprecation: starlark_warning.Deprecation(
			deprecatedDate,
			deprecatedMitigation,
			nil,
		),
	}

	builtinArgs := []*builtin_argument.BuiltinArgument{
		{
			Name: "config",
			Deprecation: starlark_warning.Deprecation(
				starlark_warning.DeprecationDate{
					Day: 20, Month: 11, Year: 2023,
				},
				"mitigation reason",
				nil,
			),
		},
	}

	argumentSet := builtin_argument.NewArgumentValuesSet(builtinArgs, []starlark.Value{starlark.String("test"), starlark.String("test")})

	err := printWarningForArguments(argumentSet, &baseBuiltIn)
	require.NoError(t, err)
	warnings := starlark_warning.GetContentFromWarningSet()
	require.Len(t, warnings, 1)
	require.Contains(t, warnings[0],
		fmt.Sprintf("%q instruction will be deprecated by %v. %v",
			"kurtosis_builtin",
			deprecatedDate.GetFormattedDate(),
			deprecatedMitigation,
		))
}

func Test_printWarningForInstructionNoWarning(t *testing.T) {
	// nolint:exhaustruct
	baseBuiltIn := KurtosisBaseBuiltin{
		Name: "KurtosisType",
	}

	builtinArgs := []*builtin_argument.BuiltinArgument{
		{
			Name: "config",
		},
	}

	argumentSet := builtin_argument.NewArgumentValuesSet(builtinArgs, []starlark.Value{starlark.String("test"), starlark.String("test")})
	err := printWarningForArguments(argumentSet, &baseBuiltIn)
	require.NoError(t, err)
	warnings := starlark_warning.GetContentFromWarningSet()
	require.Len(t, warnings, 0)
}

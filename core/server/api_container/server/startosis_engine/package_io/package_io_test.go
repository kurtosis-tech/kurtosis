package package_io

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

func TestPackageIo_simpleValue(t *testing.T) {
	result, err := convertValueToStructIfPossible(starlark.String("Hello World!"))
	require.Nil(t, err)

	expectedResult := starlark.String("Hello World!")
	require.Equal(t, expectedResult, result)
}

func TestPackageIo_listValue(t *testing.T) {
	listValue := starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.String("World!"),
	})
	result, err := convertValueToStructIfPossible(listValue)
	require.Nil(t, err)

	require.Equal(t, listValue, result)
}

func TestPackageIo_structValue(t *testing.T) {
	structValue := starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"greetings": starlark.String("bonjour!"),
	})
	result, err := convertValueToStructIfPossible(structValue)
	require.Nil(t, err)

	require.Equal(t, structValue, result)
}

func TestPackageIo_processSimpleDict(t *testing.T) {
	dict := starlark.NewDict(1)
	require.Nil(t, dict.SetKey(starlark.String("greetings"), starlark.String("bonjour!")))

	result, err := convertValueToStructIfPossible(dict)
	require.Nil(t, err)

	expectedResult := starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"greetings": starlark.String("bonjour!"),
	})
	require.Equal(t, expectedResult, result)
}

func TestPackageIo_processNestedDict(t *testing.T) {
	nested_dict := starlark.NewDict(1)
	require.Nil(t, nested_dict.SetKey(starlark.String("en_US"), starlark.String("Hello")))
	require.Nil(t, nested_dict.SetKey(starlark.String("fr_FR"), starlark.String("Bonjour")))

	dict := starlark.NewDict(1)
	require.Nil(t, dict.SetKey(starlark.String("greetings"), nested_dict))

	result, err := convertValueToStructIfPossible(dict)
	require.Nil(t, err)

	expectedResult := starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"greetings": starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
			"en_US": starlark.String("Hello"),
			"fr_FR": starlark.String("Bonjour"),
		}),
	})
	require.Equal(t, expectedResult, result)
}

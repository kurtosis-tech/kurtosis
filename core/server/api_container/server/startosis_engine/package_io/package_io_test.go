package package_io

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

const complexInputJson = `{
	"dict": {},
	"float": 3.4,
	"int": 1,
	"list": [
		"a",
		1,
		{}
	],
	"string": "Hello World!"
}`

func createDict(t *testing.T) *starlark.Dict {
	dict := starlark.NewDict(1)
	require.Nil(t, dict.SetKey(starlark.String("string"), starlark.String("Hello World!")))
	require.Nil(t, dict.SetKey(starlark.String("int"), starlark.MakeInt(1)))
	require.Nil(t, dict.SetKey(starlark.String("float"), starlark.Float(3.4)))
	require.Nil(t, dict.SetKey(starlark.String("dict"), starlark.NewDict(1)))
	require.Nil(t, dict.SetKey(starlark.String("list"), starlark.NewList([]starlark.Value{starlark.String("a"), starlark.MakeInt(1), starlark.NewDict(1)})))
	return dict
}

func TestPackageIo_DeserializeArgs(t *testing.T) {
	expectedResultDict := createDict(t)

	result, interpretationErr := DeserializeArgs(&starlark.Thread{}, complexInputJson) //nolint:exhaustruct
	require.Nil(t, interpretationErr)
	equal, err := starlark.Equal(expectedResultDict, result)
	require.Nil(t, err)
	require.True(t, equal)
}

func TestPackageIo_DeserializeArgsToStruct(t *testing.T) {
	complexInputJsonAsStruct := fmt.Sprintf(`{
	%q: %q,
	"struct": {},
	"float": 3.4,
	"int": 1,
	"list": [
		"a",
		1,
		{}
	],
	"string": "Hello World!"
}`, kurtosisParserKey, kurtosisParserStruct)

	expectedResultStringDict := starlark.StringDict{
		kurtosisParserKey: starlark.String(kurtosisParserStruct),
		"string":          starlark.String("Hello World!"),
		"int":             starlark.MakeInt(1),
		"float":           starlark.Float(3.4),
		"struct":          starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{}),
		"list":            starlark.NewList([]starlark.Value{starlark.String("a"), starlark.MakeInt(1), starlark.NewDict(1)}),
	}
	expectedResultStruct := starlarkstruct.FromStringDict(starlarkstruct.Default, expectedResultStringDict)

	result, interpretationErr := DeserializeArgs(&starlark.Thread{}, complexInputJsonAsStruct) //nolint:exhaustruct
	require.Nil(t, interpretationErr)

	equal, err := starlark.Equal(expectedResultStruct, result)
	require.Nil(t, err)
	require.True(t, equal)
}

func TestPackageIo_SerializeOutputObject(t *testing.T) {
	result, interpretationErr := SerializeOutputObject(&starlark.Thread{}, createDict(t)) //nolint:exhaustruct
	require.Nil(t, interpretationErr)

	require.Equal(t, result, complexInputJson)
}

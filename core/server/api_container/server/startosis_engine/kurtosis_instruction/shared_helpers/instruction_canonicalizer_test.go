package shared_helpers

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

func TestCanonicalizeInstruction(t *testing.T) {
	filename := "github.com/kurtosis-tech/module/main.star"
	position := kurtosis_instruction.NewInstructionPosition(12, 23, filename)

	arg1 := starlark.String("serviceA")
	arg2Dict := starlark.StringDict{}
	arg2Dict["hello"] = starlark.String("world")
	arg2Dict["bonjour"] = starlark.MakeInt(42)
	arg2 := starlarkstruct.FromStringDict(starlarkstruct.Default, arg2Dict)
	result := CanonicalizeInstruction("my_instruction", map[string]starlark.Value{
		"arg1": arg1,
		"arg2": arg2,
	}, position)
	expectedResult := `// from: github.com/kurtosis-tech/module/main.star-12:23
my_instruction(
	arg1="serviceA",
	arg2=struct(
		bonjour=42,
		hello="world"
	),
)`
	require.Equal(t, expectedResult, result)
}

func TestCanonicalizeArgValue_String(t *testing.T) {
	input := starlark.String("Hello")
	result := canonicalizeArgValue(input, false, 0)
	require.Equal(t, `"Hello"`, result) // notice the quotes here
}

func TestCanonicalizeArgValue_Int64(t *testing.T) {
	input := starlark.MakeInt64(1234)
	result := canonicalizeArgValue(input, false, 0)
	require.Equal(t, `1234`, result)
}

func TestCanonicalizeArgValue_Int(t *testing.T) {
	input := starlark.MakeInt(1234)
	result := canonicalizeArgValue(input, false, 0)
	require.Equal(t, `1234`, result)
}

func TestCanonicalizeArgValue_Slice(t *testing.T) {
	input := starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.String("World"),
		starlark.MakeInt(42),
	})
	result := canonicalizeArgValue(input, false, 0)
	expectedResult := `[
	"Hello",
	"World",
	42
]`
	require.Equal(t, expectedResult, result)
}

func TestCanonicalizeArgValue_SimpleMap(t *testing.T) {
	input := &starlark.Dict{}
	require.Nil(t, input.SetKey(starlark.String("hello"), starlark.String("world")))
	require.Nil(t, input.SetKey(starlark.String("bonjour"), starlark.MakeInt(1)))
	require.Nil(t, input.SetKey(starlark.MakeInt(42), starlark.String("bonjour")))
	result := canonicalizeArgValue(input, false, 0)
	expectedResult := `{
	"bonjour": 1,
	"hello": "world",
	42: "bonjour"
}`
	require.Equal(t, expectedResult, result)
}

func TestCanonicalizeArgValue_ComplexMap(t *testing.T) {
	input := &starlark.Dict{}
	nestedMap := &starlark.Dict{}
	require.Nil(t, nestedMap.SetKey(starlark.String("hello"), starlark.String("world")))
	nestedList := starlark.NewList([]starlark.Value{starlark.String("Hello"), starlark.MakeInt(42)})
	require.Nil(t, input.SetKey(starlark.String("nested_map"), nestedMap))
	require.Nil(t, input.SetKey(starlark.String("nested_list"), nestedList))
	result := canonicalizeArgValue(input, false, 0)
	expectedResult := `{
	"nested_list": [
		"Hello",
		42
	],
	"nested_map": {
		"hello": "world"
	}
}`
	require.Equal(t, expectedResult, result)
}

func TestCanonicalizeArgValue_SimpleStruct(t *testing.T) {
	inputDict := starlark.StringDict{}
	inputDict["hello"] = starlark.String("world")
	inputDict["bonjour"] = starlark.MakeInt(42)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)
	result := canonicalizeArgValue(input, false, 0)
	expectedResult := `struct(
	bonjour=42,
	hello="world"
)`
	require.Equal(t, expectedResult, result)
}

func TestCanonicalizeArgValue_ComplexStruct(t *testing.T) {

	nestedMap := &starlark.Dict{}
	require.Nil(t, nestedMap.SetKey(starlark.String("hello"), starlark.String("world")))
	nestedList := starlark.NewList([]starlark.Value{starlark.String("Hello"), starlark.MakeInt(42)})
	nestedStruct := starlark.StringDict{}
	nestedStruct["bonjour"] = starlark.MakeInt(42)

	inputDict := starlark.StringDict{}
	inputDict["nested_map"] = nestedMap
	inputDict["nested_list"] = nestedList
	inputDict["nested_struct"] = starlarkstruct.FromStringDict(starlarkstruct.Default, nestedStruct)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)
	result := canonicalizeArgValue(input, false, 0)
	expectedResult := `struct(
	nested_list=[
		"Hello",
		42
	],
	nested_map={
		"hello": "world"
	},
	nested_struct=struct(
		bonjour=42
	)
)`
	require.Equal(t, expectedResult, result)
}

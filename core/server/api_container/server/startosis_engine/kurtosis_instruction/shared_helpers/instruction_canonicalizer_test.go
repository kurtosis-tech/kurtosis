package shared_helpers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
	"time"
)

const (
	defaultInitialIndent = 0
)

func TestCanonicalizeInstruction(t *testing.T) {
	filename := "github.com/kurtosis-tech/module/main.star"
	position := kurtosis_instruction.NewInstructionPosition(12, 23, filename)

	args := []starlark.Value{
		starlark.String("foo"),
		starlark.NewList([]starlark.Value{starlark.String("bar")}),
	}

	kwarg1 := starlark.String("serviceA")
	kwarg2Dict := starlark.StringDict{}
	kwarg2Dict["hello"] = starlark.String("world")
	kwarg2Dict["bonjour"] = starlark.MakeInt(42)
	kwarg2 := starlarkstruct.FromStringDict(starlarkstruct.Default, kwarg2Dict)
	kwargs := map[string]starlark.Value{
		"kwarg1": kwarg1,
		"kwarg2": kwarg2,
	}

	multiLineResult := MultiLineCanonicalizer.CanonicalizeInstruction("my_instruction", args, kwargs, position)
	expectedMultiLineResult := `# from: github.com/kurtosis-tech/module/main.star[12:23]
my_instruction(
	"foo",
	[
		"bar"
	],
	kwarg1="serviceA",
	kwarg2=struct(
		bonjour=42,
		hello="world"
	)
)`
	require.Equal(t, expectedMultiLineResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.CanonicalizeInstruction("my_instruction", args, kwargs, position)
	expectedSingleLineResult := `my_instruction("foo", ["bar"], kwarg1="serviceA", kwarg2=struct(bonjour=42, hello="world"))`
	require.Equal(t, expectedSingleLineResult, singleLineResult)
}

func TestCanonicalizeArgValue_None(t *testing.T) {
	input := starlark.None
	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `None`, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `None`, singleLineResult)
}

func TestCanonicalizeArgValue_Bool(t *testing.T) {
	input := starlark.Bool(true)
	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `True`, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `True`, singleLineResult)
}

func TestCanonicalizeArgValue_String(t *testing.T) {
	input := starlark.String("Hello")
	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `"Hello"`, multiLineResult) // notice the quotes here

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `"Hello"`, singleLineResult) // notice the quotes here
}

func TestCanonicalizeArgValue_Bytes(t *testing.T) {
	input := starlark.Bytes("Hello")
	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `b"Hello"`, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `b"Hello"`, singleLineResult)
}

func TestCanonicalizeArgValue_Int64(t *testing.T) {
	input := starlark.MakeInt64(1234)
	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `1234`, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `1234`, singleLineResult)
}

func TestCanonicalizeArgValue_Int(t *testing.T) {
	input := starlark.MakeInt(1234)
	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `1234`, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `1234`, singleLineResult)
}

func TestCanonicalizeArgValue_Float(t *testing.T) {
	input := starlark.Float(3.14159)
	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `3.14159`, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	require.Equal(t, `3.14159`, singleLineResult)
}

func TestCanonicalizeArgValue_Time(t *testing.T) {
	paris, err := time.LoadLocation("Europe/Paris")
	require.Nil(t, err)
	goTime := time.Date(2022, 11, 8, 15, 14, 33, 999, paris)
	starlarkTime := starlarktime.Time(goTime)
	expectedResult := fmt.Sprintf("time.from_timestamp(%d)", goTime.Unix())

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(starlarkTime, defaultInitialIndent)
	require.Equal(t, expectedResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(starlarkTime, defaultInitialIndent)
	require.Equal(t, expectedResult, singleLineResult)
}

func TestCanonicalizeArgValue_Duration(t *testing.T) {
	goDuration, err := time.ParseDuration("2h34m09s999ns")
	require.Nil(t, err)
	starlarkDuration := starlarktime.Duration(goDuration)
	expectedResult := fmt.Sprintf("time.parse_duration(%s)", goDuration.String())

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(starlarkDuration, defaultInitialIndent)
	require.Equal(t, expectedResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(starlarkDuration, defaultInitialIndent)
	require.Equal(t, expectedResult, singleLineResult)
}

func TestCanonicalizeArgValue_List(t *testing.T) {
	input := starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.String("World"),
		starlark.MakeInt(42),
	})

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedMultiLineResult := `[
	"Hello",
	"World",
	42
]`
	require.Equal(t, expectedMultiLineResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedSingleLineResult := `["Hello", "World", 42]`
	require.Equal(t, expectedSingleLineResult, singleLineResult)
}

func TestCanonicalizeArgValue_Set(t *testing.T) {
	input := starlark.NewSet(3)
	require.Nil(t, input.Insert(starlark.String("Hello")))
	require.Nil(t, input.Insert(starlark.String("World")))
	require.Nil(t, input.Insert(starlark.MakeInt(42)))

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedMultiLineResult := `{
	"Hello",
	"World",
	42
}`
	require.Equal(t, expectedMultiLineResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedSingleLineResult := `{"Hello", "World", 42}`
	require.Equal(t, expectedSingleLineResult, singleLineResult)
}

func TestCanonicalizeArgValue_Tuple(t *testing.T) {
	input := starlark.Tuple{
		starlark.String("Hello"),
		starlark.String("World"),
		starlark.MakeInt(42),
	}

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedMultiLineResult := `(
	"Hello",
	"World",
	42
)`
	require.Equal(t, expectedMultiLineResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedSingleLineResult := `("Hello", "World", 42)`
	require.Equal(t, expectedSingleLineResult, singleLineResult)
}

func TestCanonicalizeArgValue_SimpleMap(t *testing.T) {
	input := &starlark.Dict{}
	require.Nil(t, input.SetKey(starlark.String("hello"), starlark.String("world")))
	require.Nil(t, input.SetKey(starlark.String("bonjour"), starlark.MakeInt(1)))
	require.Nil(t, input.SetKey(starlark.MakeInt(42), starlark.String("bonjour")))

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedMultiLineResult := `{
	"bonjour": 1,
	"hello": "world",
	42: "bonjour"
}`
	require.Equal(t, expectedMultiLineResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedSingleLineResult := `{"bonjour": 1, "hello": "world", 42: "bonjour"}`
	require.Equal(t, expectedSingleLineResult, singleLineResult)
}

func TestCanonicalizeArgValue_ComplexMap(t *testing.T) {
	input := &starlark.Dict{}
	nestedMap := &starlark.Dict{}
	require.Nil(t, nestedMap.SetKey(starlark.String("hello"), starlark.String("world")))
	nestedList := starlark.NewList([]starlark.Value{starlark.String("Hello"), starlark.MakeInt(42)})
	require.Nil(t, input.SetKey(starlark.String("nested_map"), nestedMap))
	require.Nil(t, input.SetKey(starlark.String("nested_list"), nestedList))

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedMultiLineResult := `{
	"nested_list": [
		"Hello",
		42
	],
	"nested_map": {
		"hello": "world"
	}
}`
	require.Equal(t, expectedMultiLineResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedSingleLineResult := `{"nested_list": ["Hello", 42], "nested_map": {"hello": "world"}}`
	require.Equal(t, expectedSingleLineResult, singleLineResult)
}

func TestCanonicalizeArgValue_SimpleStruct(t *testing.T) {
	inputDict := starlark.StringDict{}
	inputDict["hello"] = starlark.String("world")
	inputDict["bonjour"] = starlark.MakeInt(42)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedMultiLineResult := `struct(
	bonjour=42,
	hello="world"
)`
	require.Equal(t, expectedMultiLineResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedSingleLineResult := `struct(bonjour=42, hello="world")`
	require.Equal(t, expectedSingleLineResult, singleLineResult)
}

func TestCanonicalizeArgValue_ComplexStruct_MultiLine(t *testing.T) {

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

	multiLineResult := MultiLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedMultiLineResult := `struct(
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
	require.Equal(t, expectedMultiLineResult, multiLineResult)

	singleLineResult := SingleLineCanonicalizer.canonicalizeArgValue(input, defaultInitialIndent)
	expectedSingleLineResult := `struct(nested_list=["Hello", 42], nested_map={"hello": "world"}, nested_struct=struct(bonjour=42))`
	require.Equal(t, expectedSingleLineResult, singleLineResult)
}

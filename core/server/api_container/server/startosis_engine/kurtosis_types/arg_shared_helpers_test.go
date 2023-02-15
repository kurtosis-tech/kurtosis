package kurtosis_types

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestSafeCastToString_Success(t *testing.T) {
	input := starlark.String("blah")
	output, err := SafeCastToString(input, "test")
	require.Nil(t, err)
	require.Equal(t, "blah", output)
}

func TestSafeCastToString_Failure(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := SafeCastToString(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, "", output)
}

func TestSafeCastToStringSlice_Success(t *testing.T) {
	input := starlark.NewList([]starlark.Value{starlark.String("string_1"), starlark.String("string_2")})
	output, err := SafeCastToStringSlice(input, "test")
	require.Nil(t, err)
	require.Equal(t, []string{"string_1", "string_2"}, output)
}

func TestSafeCastToStringSlice_FailureWrongTypeInsideList(t *testing.T) {
	input := starlark.NewList([]starlark.Value{starlark.String("string_1"), starlark.MakeInt(42)})
	output, err := SafeCastToStringSlice(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test[1]' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, []string(nil), output)
}

func TestSafeCastToStringSlice_FailureNotList(t *testing.T) {
	input := starlark.MakeInt(42)
	output, err := SafeCastToStringSlice(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test' argument is expected to be an iterable. Got starlark.Int", err.Error())
	require.Equal(t, []string(nil), output)
}

func TestSafeCastToMapStringString_Success(t *testing.T) {
	input := starlark.NewDict(1)
	err := input.SetKey(starlark.String("key"), starlark.String("value"))
	require.Nil(t, err)
	output, err := SafeCastToMapStringString(input, "test")
	require.Nil(t, err)
	require.Equal(t, map[string]string{"key": "value"}, output)
}

func TestSafeCastToMapStringString_Failure(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := SafeCastToMapStringString(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test' argument is expected to be a dict. Got starlark.Int", err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestSafeCastToMapStringString_FailureValueIsNotString(t *testing.T) {
	input := starlark.NewDict(1)
	err := input.SetKey(starlark.String("key"), starlark.MakeInt(42))
	require.Nil(t, err)
	output, err := SafeCastToMapStringString(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test[\"key\"]' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestSafeCastToMapStringString_FailureKeyIsNotString(t *testing.T) {
	input := starlark.NewDict(1)
	err := input.SetKey(starlark.MakeInt(42), starlark.String("value"))
	require.Nil(t, err)
	output, err := SafeCastToMapStringString(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test.key:42' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, map[string]string(nil), output)
}

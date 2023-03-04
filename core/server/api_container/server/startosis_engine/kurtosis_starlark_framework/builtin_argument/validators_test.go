package builtin_argument

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"math"
	"testing"
)

func TestNonEmptyString_WrongType(t *testing.T) {
	value := starlark.MakeInt(12)
	err := NonEmptyString(value, "service_name")
	require.NotNil(t, err)
	require.Equal(t, "Value for 'service_name' was expected to be a starlark.String but was 'starlark.Int'", err.Error())
}

func TestNonEmptyString_EmptyStringFailure(t *testing.T) {
	value := starlark.String("")
	err := NonEmptyString(value, "service_name")
	require.NotNil(t, err)
	require.Equal(t, "Value for 'service_name' was an empty string. This is disallowed", err.Error())
}

func TestNonEmptyString_Valid(t *testing.T) {
	value := starlark.String("datastore-1")
	err := NonEmptyString(value, "service_name")
	require.Nil(t, err)

}

func TestUint64InRange_WrongType(t *testing.T) {
	value := starlark.String("25")
	err := Uint64InRange(value, "test_uint", 6, math.MaxUint64)
	require.NotNil(t, err)
	require.Equal(t, "Value for 'test_uint' was expected to be an integer between 6 and 18446744073709551615, but it was 'starlark.String'", err.Error())
}

func TestUint64InRange_OutOfRange(t *testing.T) {
	value := starlark.MakeInt(2)
	err := Uint64InRange(value, "test_uint", 6, math.MaxUint64)
	require.NotNil(t, err)
	require.Equal(t, "Value for 'test_uint' was expected to be an integer between 6 and 18446744073709551615, but it was 2", err.Error())
}

func TestUint64InRange_Valid(t *testing.T) {
	value := starlark.MakeInt(25)
	err := Uint64InRange(value, "test_uint", 6, math.MaxUint64)
	require.Nil(t, err)
}

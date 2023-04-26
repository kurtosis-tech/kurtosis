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

func TestStringValues_Valid(t *testing.T) {
	value := starlark.String("TCP")
	err := StringValues(value, "port_protocol", []string{"TCP", "UDP"})
	require.Nil(t, err)
}

func TestStringValues_Invalid(t *testing.T) {
	value := starlark.String("BLAH")
	err := StringValues(value, "port_protocol", []string{"TCP", "UDP"})
	require.NotNil(t, err)
	require.Equal(t, "Invalid argument value for 'port_protocol': 'BLAH'. Valid values are TCP, UDP", err.Error())
}

func TestStringRegexp_Valid(t *testing.T) {
	value := starlark.String("hello")
	err := StringRegexp(value, "port_protocol", "^[a-z]*$")
	require.Nil(t, err)
}

func TestStringRegexp_Invalid(t *testing.T) {
	value := starlark.String("HELLO")
	err := StringRegexp(value, "port_protocol", "^[a-z]*$")
	require.NotNil(t, err)
	require.Equal(t, "Argument 'port_protocol' must match regexp: '^[a-z]*$'. Its value was 'HELLO'", err.Error())
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

func TestFloatInRange_WrongType(t *testing.T) {
	value := starlark.String("25")
	err := FloatInRange(value, "test_uint", 6, math.MaxUint64)
	require.NotNil(t, err)
	require.Equal(t, "Value for 'test_uint' was expected to be a float between 6.000000 and 18446744073709551616.000000, but it was 'starlark.String'", err.Error())
}

func TestFloatInRange_OutOfRange(t *testing.T) {
	value := starlark.Float(2.75)
	err := FloatInRange(value, "test_uint", 6, math.MaxUint64)
	require.NotNil(t, err)
	require.Equal(t, "Value for 'test_uint' was expected to be a float between 6.000000 and 18446744073709551616.000000, but it was 2.75", err.Error())
}

func TestFloatInRange_Valid(t *testing.T) {
	value := starlark.Float(27.5)
	err := FloatInRange(value, "test_uint", 6, math.MaxUint64)
	require.Nil(t, err)
}

func TestDuration_Valid(t *testing.T) {
	value := starlark.String("2m")
	err := Duration(value, "test_duration")
	require.Nil(t, err)
}

func TestDuration_Invalid(t *testing.T) {
	value := starlark.String("2minutos")
	err := Duration(value, "test_duration")
	require.NotNil(t, err)
	require.Equal(t, "The value '2minutos' of 'test_duration' attribute is not a valid duration string format\n\tCaused by: time: unknown unit \"minutos\" in duration \"2minutos\"", err.Error())
}

func TestDurationOrNone_ValidDurationString(t *testing.T) {
	value := starlark.String("15s")
	err := DurationOrNone(value, "test_duration_or_none_valid_duration_string")
	require.Nil(t, err)
}

func TestDurationOrNone_ValidNoneType(t *testing.T) {
	value := starlark.None
	err := DurationOrNone(value, "test_duration_or_none_valid_none_type")
	require.Nil(t, err)
}

func TestDurationOrNone_InvalidEmptyString(t *testing.T) {
	value := starlark.String("")
	err := DurationOrNone(value, "test_duration_or_none_invalid_empty_string")
	require.NotNil(t, err)
	require.Equal(t, "Value for 'test_duration_or_none_invalid_empty_string' was an empty string. This is disallowed", err.Error())
}

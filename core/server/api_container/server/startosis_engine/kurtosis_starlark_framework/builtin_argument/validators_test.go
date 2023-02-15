package builtin_argument

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
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

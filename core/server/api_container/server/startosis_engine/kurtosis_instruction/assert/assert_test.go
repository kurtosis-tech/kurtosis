package assert

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestAssert_StringsEqual(t *testing.T) {
	currentValue := starlark.String("Hello")
	assertion := "=="
	targetValue := starlark.String("Hello")
	require.Nil(t, Assert(currentValue, assertion, targetValue))
}

func TestAssert_StringsNonEqual(t *testing.T) {
	currentValue := starlark.String("Hello")
	assertion := "=="
	targetValue := starlark.String("World")
	require.NotNil(t, Assert(currentValue, assertion, targetValue))
}

func TestAssert_IntsLt(t *testing.T) {
	currentValue := starlark.MakeInt(1)
	assertion := "<"
	targetValue := starlark.MakeInt(5)
	require.Nil(t, Assert(currentValue, assertion, targetValue))
}

func TestAssert_IntsGtFalse(t *testing.T) {
	currentValue := starlark.MakeInt(1)
	assertion := ">"
	targetValue := starlark.MakeInt(5)
	require.NotNil(t, Assert(currentValue, assertion, targetValue))
}

func TestAssert_ListIn(t *testing.T) {
	currentValue := starlark.String("Hello")
	assertion := "IN"
	targetValue := starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.String("World"),
	})
	require.Nil(t, Assert(currentValue, assertion, targetValue))
}

func TestAssert_ListInFalse(t *testing.T) {
	currentValue := starlark.String("Kurtosis")
	assertion := "IN"
	targetValue := starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.String("World"),
	})
	require.NotNil(t, Assert(currentValue, assertion, targetValue))
}

func TestAssert_ListNotIn(t *testing.T) {
	currentValue := starlark.String("Kurtosis")
	assertion := "NOT_IN"
	targetValue := starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.String("World"),
	})
	require.Nil(t, Assert(currentValue, assertion, targetValue))
}

func TestAssert_ListNotInFalse(t *testing.T) {
	currentValue := starlark.String("Hello")
	assertion := "NOT_IN"
	targetValue := starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.String("World"),
	})
	require.NotNil(t, Assert(currentValue, assertion, targetValue))
}

func TestAssert_InvalidToken(t *testing.T) {
	currentValue := starlark.String("Hello")
	assertion := "INVALID"
	targetValue := starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.String("World"),
	})
	require.NotNil(t, Assert(currentValue, assertion, targetValue))
}

package builtin_argument

import (
	"github.com/stretchr/testify/require"
	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"math"
	"testing"
	"time"
)

func TestCopyValueBool(t *testing.T) {
	value := starlark.Bool(true)
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
}

func TestCopyValueBytes(t *testing.T) {
	value := starlark.Bytes("Hello")
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
}

func TestCopyValueFloat(t *testing.T) {
	value := starlark.Float(math.Pi)
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
}

func TestCopyValueInt(t *testing.T) {
	value := starlark.MakeInt(42)
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
}

func TestCopyValueNoneType(t *testing.T) {
	value := starlark.None
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
}

func TestCopyValueString(t *testing.T) {
	value := starlark.String("Hello")
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
}

func TestCopyValueTime(t *testing.T) {
	value := starlarktime.Time(time.Now())
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
}

func TestCopyValueDuration(t *testing.T) {
	value := starlarktime.Duration(42 & time.Second)
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
}

func TestCopyValueList(t *testing.T) {
	value := starlark.NewList([]starlark.Value{
		starlark.String("hello"),
		starlark.MakeInt(42),
		starlark.None,
	})
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
	require.NotSame(t, valueCopy, value)
}

func TestCopyValueSet(t *testing.T) {
	value := starlark.NewSet(3)
	require.NoError(t, value.Insert(starlark.String("hello")))
	require.NoError(t, value.Insert(starlark.MakeInt(42)))
	require.NoError(t, value.Insert(starlark.None))
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
	require.NotSame(t, valueCopy, value)
}

func TestCopyValueTuple(t *testing.T) {
	value := starlark.Tuple([]starlark.Value{
		starlark.String("hello"),
		starlark.MakeInt(42),
		starlark.None,
	})
	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
	require.NotSame(t, valueCopy, value)
}

func TestCopyValueDict(t *testing.T) {
	value := starlark.NewDict(3)
	require.NoError(t, value.SetKey(starlark.String("hello"), starlark.Float(math.Pi)))
	require.NoError(t, value.SetKey(starlark.MakeInt(42), starlark.None))
	require.NoError(t, value.SetKey(starlark.String("nested"), starlark.Tuple([]starlark.Value{
		starlark.String("hello"),
		starlark.MakeInt(42),
		starlark.None,
	})))

	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
	require.NotSame(t, valueCopy, value)
}

func TestCopyValueStruct(t *testing.T) {
	valueDict := starlark.StringDict{}
	valueDict["hello"] = starlark.String("world")
	valueDict["bonjour"] = starlark.MakeInt(42)
	valueDict["nested"] = starlark.NewList([]starlark.Value{
		starlark.String("Hello"),
		starlark.MakeInt(42),
	})
	value := starlarkstruct.FromStringDict(starlarkstruct.Default, valueDict)

	valueCopy, err := DeepCopyArgumentValue(value)
	require.NoError(t, err)
	require.Equal(t, valueCopy, value)
	require.NotSame(t, valueCopy, value)
}

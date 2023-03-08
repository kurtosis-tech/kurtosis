package magic_string_helper

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testStringRuntimeValue         = starlark.String("test_string")
	testRuntimeValueField          = "field.subfield"
	testExpectedInterpolatedString = starlark.String("test_string is not 0")
)

var testIntRuntimeValue = starlark.MakeInt(0)

func TestGetOrReplaceRuntimeValueFromString_BasicFetch(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	intValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	runtimeValueStore.SetValue(intValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testIntRuntimeValue})
	fetchedStringValue, err := GetOrReplaceRuntimeValueFromString(fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField), runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, fetchedStringValue, testStringRuntimeValue)
	fetchedIntValue, err := GetOrReplaceRuntimeValueFromString(fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, intValueUuid, testRuntimeValueField), runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, fetchedIntValue, testIntRuntimeValue)
}

func TestGetOrReplaceRuntimeValueFromString_Interpolated(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	intValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	runtimeValueStore.SetValue(intValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testIntRuntimeValue})
	stringRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField)
	intRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, intValueUuid, testRuntimeValueField)
	interpolatedString := fmt.Sprintf("%v is not %v", stringRuntimeValue, intRuntimeValue)
	resolvedInterpolatedString, err := GetOrReplaceRuntimeValueFromString(interpolatedString, runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, resolvedInterpolatedString, testExpectedInterpolatedString)
}

func TestReplaceRuntimeValueFromString(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	intValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	runtimeValueStore.SetValue(intValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testIntRuntimeValue})
	stringRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField)
	intRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, intValueUuid, testRuntimeValueField)
	interpolatedString := fmt.Sprintf("%v is not %v", stringRuntimeValue, intRuntimeValue)
	resolvedInterpolatedString, err := ReplaceRuntimeValueInString(interpolatedString, runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, resolvedInterpolatedString, testExpectedInterpolatedString.GoString())
}

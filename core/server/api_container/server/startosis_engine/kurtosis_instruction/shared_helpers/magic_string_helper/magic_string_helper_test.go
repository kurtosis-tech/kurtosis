package magic_string_helper

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"os"
	"testing"
)

const (
	testStringRuntimeValue         = starlark.String("test_string")
	testRuntimeValueField          = "field.subfield"
	testExpectedInterpolatedString = starlark.String("test_string is not 0")
	starlarkThreadName             = "starlark-value-serde-for-test-in-magic-helper-thread"
)

var testIntRuntimeValue = starlark.MakeInt(0)

func TestGetOrReplaceRuntimeValueFromString_BasicFetch(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)

	dummySerde := newDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(t, err)
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	err = runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	require.NoError(t, err)
	intValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	err = runtimeValueStore.SetValue(intValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testIntRuntimeValue})
	require.NoError(t, err)
	fetchedStringValue, err := GetOrReplaceRuntimeValueFromString(fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField), runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, fetchedStringValue, testStringRuntimeValue)
	fetchedIntValue, err := GetOrReplaceRuntimeValueFromString(fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, intValueUuid, testRuntimeValueField), runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, fetchedIntValue, testIntRuntimeValue)
}

func TestGetOrReplaceRuntimeValueFromString_Interpolated(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)

	dummySerde := newDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(t, err)
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	err = runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	require.NoError(t, err)
	intValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	err = runtimeValueStore.SetValue(intValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testIntRuntimeValue})
	require.NoError(t, err)
	stringRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField)
	intRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, intValueUuid, testRuntimeValueField)
	interpolatedString := fmt.Sprintf("%v is not %v", stringRuntimeValue, intRuntimeValue)
	resolvedInterpolatedString, err := GetOrReplaceRuntimeValueFromString(interpolatedString, runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, resolvedInterpolatedString, testExpectedInterpolatedString)
}

func TestReplaceRuntimeValueFromString(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)

	dummySerde := newDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(t, err)
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	err = runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	require.NoError(t, err)
	intValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err)
	err = runtimeValueStore.SetValue(intValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testIntRuntimeValue})
	require.NoError(t, err)
	stringRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField)
	intRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, intValueUuid, testRuntimeValueField)
	interpolatedString := fmt.Sprintf("%v is not %v", stringRuntimeValue, intRuntimeValue)
	resolvedInterpolatedString, err := ReplaceRuntimeValueInString(interpolatedString, runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, resolvedInterpolatedString, testExpectedInterpolatedString.GoString())
}

func getEnclaveDBForTest(t *testing.T) *enclave_db.EnclaveDB {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer func() {
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.NoError(t, err)
	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}

	return enclaveDb
}

func newDummyStarlarkValueSerDeForTest() *kurtosis_types.StarlarkValueSerde {
	starlarkThread := &starlark.Thread{
		Name:       starlarkThreadName,
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := starlark.StringDict{}

	serde := kurtosis_types.NewStarlarkValueSerde(starlarkThread, starlarkEnv)

	return serde
}

package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"os"
	"testing"
)

const (
	resultStarlarkVar = "result"

	enclaveDbFilePerm = 0666
)

func getBasePredeclaredDict(t *testing.T, thread *starlark.Thread) starlark.StringDict {
	predeclared := startosis_engine.Predeclared()

	// Add all Kurtosis types
	for _, kurtosisTypeConstructor := range startosis_engine.KurtosisTypeConstructors() {
		predeclared[kurtosisTypeConstructor.Name()] = kurtosisTypeConstructor
	}
	return predeclared
}

func codeToExecute(builtinStarlarkCode string) string {
	return fmt.Sprintf("%s = %s", resultStarlarkVar, builtinStarlarkCode)
}

func extractResultValue(t *testing.T, globals starlark.StringDict) starlark.Value {
	value, found := globals[resultStarlarkVar]
	require.True(t, found, "Result variable could not be found in dictionary of global variables")
	return value
}

func newStarlarkThread(name string) *starlark.Thread {
	return &starlark.Thread{
		Name:       name,
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
}

func getEnclaveDBForTest(t *testing.T) *enclave_db.EnclaveDB {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer func() {
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	db, err := bolt.Open(file.Name(), enclaveDbFilePerm, nil)
	require.NoError(t, err)
	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}

	return enclaveDb
}

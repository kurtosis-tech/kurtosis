package runtime_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"os"
	"testing"
)

const (
	randomUuid = "random-uuid"
)

func TestRecipeResultSaveAndGet_Success(t *testing.T) {
	repository := getRecipeResultRepositoryForTest(t)

	myMap := map[string]stringifiedRecipeResultValue{
		"mykey": stringifiedRecipeResultValue(starlark.String("my-value").GoString()),
	}

	err := repository.Save(randomUuid, myMap)
	require.NoError(t, err)

	value, err := repository.Get(randomUuid)
	require.NoError(t, err)
	require.NotNil(t, value)

	require.Equal(t, myMap, value)

}

func getRecipeResultRepositoryForTest(t *testing.T) *recipeResultRepository {
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
	repository, err := getOrCreateNewRecipeResultRepository(enclaveDb)
	require.NoError(t, err)

	return repository
}

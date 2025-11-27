package runtime_value_store

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"os"
	"testing"
)

const (
	randomUuid = "abcd12a3948149d9afa2ef93abb4ec52"

	keyDoesNotExistOnRepositoryErrorMsg = "does not exist on the recipe result repository"

	firstKey            = "mykey"
	secondKey           = "mySecondKey"
	thirdKey            = "myThirdKey"
	starlarkStringValue = starlark.String("my-value")

	starlarkThreadName = "recipe-result-repository-test-starlark-thread"
)

var (
	starlarkIntValue  = starlark.MakeInt(30)
	starlarkBoolValue = starlark.Bool(true)
)

func TestRecipeResultSaveKey_Success(t *testing.T) {
	repository := getRecipeResultRepositoryForTest(t)

	err := repository.SaveKey(randomUuid)
	require.NoError(t, err)

	value, err := repository.Get(randomUuid)
	require.NoError(t, err)
	require.Empty(t, value)
}

func TestRecipeResultSaveAndGet_Success(t *testing.T) {
	repository := getRecipeResultRepositoryForTest(t)

	resultValue := map[string]starlark.Comparable{
		firstKey:  starlarkStringValue,
		secondKey: starlarkIntValue,
		thirdKey:  starlarkBoolValue,
	}

	err := repository.Save(randomUuid, resultValue)
	require.NoError(t, err)

	value, err := repository.Get(randomUuid)
	require.NoError(t, err)
	require.NotNil(t, value)

	require.Equal(t, resultValue, value)
}

func TestRecipeResultGet_DoesNotExist(t *testing.T) {
	repository := getRecipeResultRepositoryForTest(t)

	value, err := repository.Get(randomUuid)
	require.Error(t, err)
	require.ErrorContains(t, err, keyDoesNotExistOnRepositoryErrorMsg)
	require.Empty(t, value)
}

func TestDelete_Success(t *testing.T) {
	repository := getRecipeResultRepositoryForTest(t)

	resultValue := map[string]starlark.Comparable{
		firstKey:  starlarkStringValue,
		secondKey: starlarkIntValue,
	}

	err := repository.Save(randomUuid, resultValue)
	require.NoError(t, err)

	value, err := repository.Get(randomUuid)
	require.NoError(t, err)
	require.NotNil(t, value)

	require.Equal(t, resultValue, value)

	err = repository.Delete(randomUuid)
	require.NoError(t, err)

	value, err = repository.Get(randomUuid)
	require.Error(t, err)
	require.Empty(t, value)
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

	dummySerde := newDummyStarlarkValueSerDeForTest()

	repository, err := getOrCreateNewRecipeResultRepository(enclaveDb, dummySerde)
	require.NoError(t, err)

	return repository
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

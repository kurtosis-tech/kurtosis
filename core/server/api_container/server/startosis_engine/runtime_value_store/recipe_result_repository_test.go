package runtime_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"os"
	"testing"
)

const (
	randomUuid = "abcd12a3948149d9afa2ef93abb4ec52"

	notAcceptedComparableTypeErrorMsg   = "Unexpected comparable type"
	keyDoesNotExistOnRepositoryErrorMsg = "does not exist on the recipe result repository"

	firstKey            = "mykey"
	secondKey           = "mySecondKey"
	thirdKey            = "myThirdKey"
	starlarkStringValue = starlark.String("my-value")
)

var (
	starlarkIntValue = starlark.MakeInt(30)
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

func TestRecipeResultSave_ErrorWhenUsingNotStarlarkStringOrInt(t *testing.T) {
	repository := getRecipeResultRepositoryForTest(t)

	resultValue := map[string]starlark.Comparable{
		firstKey: starlark.Bool(true),
	}

	err := repository.Save(randomUuid, resultValue)
	require.Error(t, err)
	require.ErrorContains(t, err, notAcceptedComparableTypeErrorMsg)

	resultValue2 := map[string]starlark.Comparable{
		secondKey: directory.Directory{},
	}

	err = repository.Save(randomUuid, resultValue2)
	require.Error(t, err)
	require.ErrorContains(t, err, notAcceptedComparableTypeErrorMsg)

	resultValue3 := map[string]starlark.Comparable{
		thirdKey: &kurtosis_type_constructor.KurtosisValueTypeDefault{},
	}

	err = repository.Save(randomUuid, resultValue3)
	require.Error(t, err)
	require.ErrorContains(t, err, notAcceptedComparableTypeErrorMsg)

	resultValue4 := map[string]starlark.Comparable{
		thirdKey: &starlark.Dict{},
	}

	err = repository.Save(randomUuid, resultValue4)
	require.Error(t, err)
	require.ErrorContains(t, err, notAcceptedComparableTypeErrorMsg)
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
	repository, err := getOrCreateNewRecipeResultRepository(enclaveDb)
	require.NoError(t, err)

	return repository
}

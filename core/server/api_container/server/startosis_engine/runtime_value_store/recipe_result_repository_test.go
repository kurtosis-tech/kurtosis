package runtime_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_value_serde"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
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
	repository, err := getOrCreateNewRecipeResultRepository(enclaveDb, getStarlarkValueSerdeForTest())
	require.NoError(t, err)

	return repository
}

func getStarlarkValueSerdeForTest() *starlark_value_serde.StarlarkValueSerde {
	starlarkValueSerde := starlark_value_serde.NewStarlarkValueSerde(getPredeclaredForTest(), []*starlark.Builtin{})
	return starlarkValueSerde
}

// this should match the Predeclared() func in Kurtosis Builtins, which is not used here for import cycle reasons
func getPredeclaredForTest() starlark.StringDict {
	return starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark

		// go-starlark time module with time.now() disabled
		time.Module.Name: builtins.TimeModuleWithNowDisabled(),
	}
}

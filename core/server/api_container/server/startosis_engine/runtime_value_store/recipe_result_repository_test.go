package runtime_value_store

import (
	port_spec_core "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
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
	fourthKey           = "myFourthKey"
	fifthKey            = "fifthKey"
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

	startosisPortSpecType, interpretationErr := port_spec.CreatePortSpec(
		uint16(443),
		port_spec_core.TransportProtocol_TCP,
		nil,
		"10s",
	)
	require.Nil(t, interpretationErr)

	startosisDirectoryType, interpretationErr := directory.CreateDirectoryFromFilesArtifact("fake-file-artifact-name")
	require.Nil(t, interpretationErr)

	resultValue := map[string]string{
		firstKey:  starlarkStringValue.GoString(),
		secondKey: starlarkIntValue.String(),
		thirdKey:  starlarkBoolValue.String(),
		fourthKey: startosisPortSpecType.String(),
		fifthKey:  startosisDirectoryType.String(),
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

	resultValue := map[string]string{
		firstKey:  starlarkStringValue.GoString(),
		secondKey: starlarkIntValue.String(),
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

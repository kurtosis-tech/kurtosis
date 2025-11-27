package runtime_value_store

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
)

const (
	firstServiceName               = service.ServiceName("first-service-name")
	secondServiceName              = service.ServiceName("second-service-name")
	serviceWithoutAssociatedValues = service.ServiceName("no-values-service-name")

	firstServiceAssociatedValueUuid  = "cddc2ea3948149d9afa2ef93abb4ec52"
	secondServiceAssociatedValueUuid = "ae5c8bf2fbeb4de68f647280b1c79cbb"
)

func TestSaveAndGet_Success(t *testing.T) {
	repository := getAssociatedValuesRepositoryForTest(t)

	err := repository.Save(firstServiceName, firstServiceAssociatedValueUuid)
	require.NoError(t, err)

	err = repository.Save(secondServiceName, secondServiceAssociatedValueUuid)
	require.NoError(t, err)

	firstServiceAssociatedValues, err := repository.Get(firstServiceName)
	require.NoError(t, err)
	require.Equal(t, firstServiceAssociatedValueUuid, firstServiceAssociatedValues)

	secondServiceAssociatedValues, err := repository.Get(secondServiceName)
	require.NoError(t, err)
	require.Equal(t, secondServiceAssociatedValueUuid, secondServiceAssociatedValues)
}

func TestGet_NotAssociateValues(t *testing.T) {
	repository := getAssociatedValuesRepositoryForTest(t)

	serviceAssociatedValues, err := repository.Get(serviceWithoutAssociatedValues)
	require.NoError(t, err)
	require.Empty(t, serviceAssociatedValues)
}

func TestExist_Success(t *testing.T) {
	repository := getAssociatedValuesRepositoryForTest(t)

	err := repository.Save(firstServiceName, firstServiceAssociatedValueUuid)
	require.NoError(t, err)

	exist, err := repository.Exist(firstServiceName)
	require.NoError(t, err)
	require.True(t, exist)
}

func TestNotExist_Success(t *testing.T) {
	repository := getAssociatedValuesRepositoryForTest(t)

	exist, err := repository.Exist(serviceWithoutAssociatedValues)
	require.NoError(t, err)
	require.False(t, exist)
}

func getAssociatedValuesRepositoryForTest(t *testing.T) *serviceAssociatedValuesRepository {
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
	repository, err := getOrCreateNewServiceAssociatedValuesRepository(enclaveDb)
	require.NoError(t, err)

	return repository
}

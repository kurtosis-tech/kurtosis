package enclave_plan

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
)

const (
	uuidForTest = "abdc4cb3948149d9afa2ef93abb4ec56"
)

func TestSaveAndGetServiceRegistration_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalEnclavePlanInstruction := getEnclavePlanInstructionForTest()

	require.NotNil(t, originalEnclavePlanInstruction)

	err := repository.Save(uuidForTest, originalEnclavePlanInstruction)
	require.NoError(t, err)

	enclavePlanInstructionFromRepository, err := repository.Get(uuidForTest)
	require.NoError(t, err)

	require.Equal(t, originalEnclavePlanInstruction, enclavePlanInstructionFromRepository)
}

func TestSize_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalEnclavePlanInstruction := getEnclavePlanInstructionForTest()

	require.NotNil(t, originalEnclavePlanInstruction)

	err := repository.Save(uuidForTest, originalEnclavePlanInstruction)
	require.NoError(t, err)

	err = repository.Save("other-uuid", originalEnclavePlanInstruction)
	require.NoError(t, err)

	size, err := repository.Size()
	require.NoError(t, err)

	require.Equal(t, 2, size)
}

//TODO implement GetAll Test

func getRepositoryForTest(t *testing.T) *enclavePlanInstructionRepository {
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
	repository, err := getOrCreateNewEnclavePlanInstructionRepository(enclaveDb)
	require.NoError(t, err)

	return repository
}

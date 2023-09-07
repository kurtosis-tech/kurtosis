package enclave_plan_instruction

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
)

const (
	firstUuidForTest  = "abdc4cb3948149d9afa2ef93abb4ec56"
	secondUuidForTest = "31c6a6cb3948149d9afa2ef93abb47e9a"
	thirdUuidForTest  = "a5b6a6cb3948149d9afa2ef93abb5d6e"
	fourthUuidForTest = "c4c1b7cb3948149d9afa2ef93a4e5eb2a"
)

var allUuid = []string{firstUuidForTest, secondUuidForTest, thirdUuidForTest, fourthUuidForTest}

func TestSaveAndGet_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalEnclavePlanInstruction := getEnclavePlanInstructionForTest(1)[0]

	require.NotNil(t, originalEnclavePlanInstruction)

	err := repository.Save(firstUuidForTest, originalEnclavePlanInstruction)
	require.NoError(t, err)

	enclavePlanInstructionFromRepository, err := repository.Get(firstUuidForTest)
	require.NoError(t, err)

	require.Equal(t, originalEnclavePlanInstruction, enclavePlanInstructionFromRepository)
}

func getRepositoryForTest(t *testing.T) *EnclavePlanInstructionRepository {
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
	repository, err := GetOrCreateNewEnclavePlanInstructionRepository(enclaveDb)
	require.NoError(t, err)

	return repository
}

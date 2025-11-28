package service_identifiers

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
)

func TestAddServiceIdentifier_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalServiceIdentifiers := getServiceIdentifiersForTest()
	originalServiceIdentifiersPtr := &originalServiceIdentifiers

	jsonBytes, err := json.Marshal(originalServiceIdentifiersPtr)
	require.NoError(t, err)
	require.NotNil(t, jsonBytes)

	for _, serviceIdentifierObj := range originalServiceIdentifiers {
		err = repository.AddServiceIdentifier(serviceIdentifierObj)
		require.NoError(t, err)
	}

	serviceIdentifiersResult, err := repository.GetServiceIdentifiers()
	require.NoError(t, err)
	for index, serviceIdentifierObj := range serviceIdentifiersResult {
		require.EqualValues(t, originalServiceIdentifiers[index], serviceIdentifierObj)
	}
}

func getRepositoryForTest(t *testing.T) *ServiceIdentifiersRepository {
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
	repository, err := GetOrCreateNewServiceIdentifiersRepository(enclaveDb)
	require.NoError(t, err)

	return repository
}

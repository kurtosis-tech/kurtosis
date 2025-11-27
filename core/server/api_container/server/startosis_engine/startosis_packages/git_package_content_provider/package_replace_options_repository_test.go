package git_package_content_provider

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
)

var allPackageReplaceOptionsForTest = map[string]string{
	"github.com/kurtosis-tech/sample-dependency-package": "github.com/kurtosis-tech/another-sample-dependency-package",
	"github.com/ethpandaops/ethereum-package":            "github.com/another-org/ethereum-package",
}

func TestSaveAnGet_Success(t *testing.T) {
	repository := getPackageReplaceOptionsRepositoryForTest(t)

	err := repository.Save(allPackageReplaceOptionsForTest)
	require.NoError(t, err)

	historicalReplacePackageOptions, err := repository.Get()
	require.NoError(t, err)
	require.Equal(t, allPackageReplaceOptionsForTest, historicalReplacePackageOptions)
}

func TestSaveAnGet_OverwriteSuccess(t *testing.T) {
	repository := getPackageReplaceOptionsRepositoryForTest(t)

	err := repository.Save(allPackageReplaceOptionsForTest)
	require.NoError(t, err)

	randomReplaceOptionsForTest := map[string]string{
		"github.com/kurtosis-tech/sample-dependency-package": "github.com/kurtosis-tech/random-package",
		"github.com/kurtosis-tech/avalanche-package":         "github.com/another-org/avalanche-package",
	}

	err = repository.Save(randomReplaceOptionsForTest)
	require.NoError(t, err)

	expectedReplacePackageOptions := map[string]string{
		"github.com/kurtosis-tech/sample-dependency-package": "github.com/kurtosis-tech/random-package",
		"github.com/kurtosis-tech/avalanche-package":         "github.com/another-org/avalanche-package",
		"github.com/ethpandaops/ethereum-package":            "github.com/another-org/ethereum-package",
	}

	existingReplacePackageOptions, err := repository.Get()
	require.NoError(t, err)
	require.Equal(t, expectedReplacePackageOptions, existingReplacePackageOptions)
}

func TestSaveAnGet_SuccessForNoReplacePackageOptions(t *testing.T) {
	repository := getPackageReplaceOptionsRepositoryForTest(t)

	err := repository.Save(noPackageReplaceOptions)
	require.NoError(t, err)

	existingReplacePackageOptions, err := repository.Get()
	require.NoError(t, err)
	require.Equal(t, noPackageReplaceOptions, existingReplacePackageOptions)
}

func TestSave_ErrorWhenSavingNil(t *testing.T) {
	repository := getPackageReplaceOptionsRepositoryForTest(t)

	err := repository.Save(nil)
	require.Error(t, err)
}

func TestGet_SuccessEmptyRepository(t *testing.T) {
	repository := getPackageReplaceOptionsRepositoryForTest(t)

	existingReplacePackageOptions, err := repository.Get()
	require.NoError(t, err)
	require.Equal(t, noPackageReplaceOptions, existingReplacePackageOptions)
}

func getPackageReplaceOptionsRepositoryForTest(t *testing.T) *packageReplaceOptionsRepository {
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
	repository := newPackageReplaceOptionsRepository(enclaveDb)

	return repository
}

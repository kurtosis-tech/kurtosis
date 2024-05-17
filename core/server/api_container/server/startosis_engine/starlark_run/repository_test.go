package starlark_run

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
)

func TestSaveAndGetStarlarkRun_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	originalStarlarRunForTest := NewStarlarkRun(
		"github.com/kurtosis-tech/django-package",
		"postgres = import_module(\"github.com/kurtosis-tech/postgres-package/main.star\")\ndjango_app = import_module(\"/app/app.star\")\n\n# Postgres defaults\nDEFAULT_POSTGRES_USER = \"postgres\"\nDEFAULT_POSTGRES_PASSWORD = \"secretdatabasepassword\"\nDEFAULT_POSTGRES_DB_NAME = \"django-db\"\nDEFAULT_POSTGRES_SERVICE_NAME = \"postgres\"\n\ndef run(\n    plan,\n    postgres_user=DEFAULT_POSTGRES_USER,\n    postgres_password=DEFAULT_POSTGRES_PASSWORD,\n    postgres_db_name=DEFAULT_POSTGRES_DB_NAME,\n    postgres_service_name=DEFAULT_POSTGRES_SERVICE_NAME,\n):\n    \"\"\"\n    Starts this Django example application.\n\n    Args:\n        postgres_user (string): the Postgres's user name (default: postgres)\n        postgres_password (string): the Postgres's password (default: secretdatabasepassword)\n        postgres_db_name (string): the Postgres's db name (default: django-db)\n        postgres_service_name (string): the Postgres's service name (default: postgres)\n    \"\"\"\n\n    # run the application's db\n    postgres_db = postgres.run(\n        plan,\n        service_name=postgres_service_name,\n        user=postgres_user,\n        password=postgres_password,\n        database=postgres_db_name,\n    )\n\n    # run the application's backend service\n    django_app.run(plan, postgres_db, postgres_password)\n",
		"{}",
		4,
		"",
		"",
		[]int32{2},
		int32(1),
		"{}",
	)

	err := repository.Save(originalStarlarRunForTest)
	require.NoError(t, err)

	starlarkRunFromRepository, err := repository.Get()
	require.NoError(t, err)

	require.Equal(t, originalStarlarRunForTest, starlarkRunFromRepository)
}

func TestGetNilStarlarkRun_Success(t *testing.T) {
	repository := getRepositoryForTest(t)

	starlarkRunFromRepository, err := repository.Get()
	require.NoError(t, err)
	require.Nil(t, starlarkRunFromRepository)
}

func getRepositoryForTest(t *testing.T) *StarlarkRunRepository {
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
	repository, err := GetOrCreateNewStarlarkRunRepository(enclaveDb)
	require.NoError(t, err)

	return repository
}

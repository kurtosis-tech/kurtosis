package test_helpers

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	bolt "go.etcd.io/bbolt"
	"os"
)

const (
	defaultTmpDir               = ""
	dbSuffix                    = "*.db"
	readWriteEveryonePermission = 0666
)

// CreateEnclaveDbForTesting creates an enclaveDb for testing purposes
func CreateEnclaveDbForTesting() (*enclave_db.EnclaveDB, func(), error) {
	file, err := os.CreateTemp(defaultTmpDir, dbSuffix)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while creating temporary file to write enclave db too")
	}
	db, err := bolt.Open(file.Name(), readWriteEveryonePermission, nil)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while opening enclave db at '%v'", file.Name())
	}
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	cleanUpFunction := func() {
		os.Remove(file.Name())
		db.Close()
	}

	return enclaveDb, cleanUpFunction, nil
}

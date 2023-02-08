package enclave_db

import (
	"github.com/kurtosis-tech/stacktrace"
	bolt "go.etcd.io/bbolt"
	"sync"
)

const (
	readWritePermissionToDatabase = 0666
	enclaveDbFilePath             = "enclave.db"
)

var (
	openDatabaseOnce  sync.Once
	databaseInstance  *bolt.DB
	databaseOpenError error
)

type EnclaveDB struct {
	*bolt.DB
}

func GetOrCreateEnclaveDatabase() (*EnclaveDB, error) {
	openDatabaseOnce.Do(func() {
		databaseInstance, databaseOpenError = bolt.Open(enclaveDbFilePath, readWritePermissionToDatabase, &bolt.Options{
			Timeout:         0,
			NoGrowSync:      false,
			NoFreelistSync:  false,
			FreelistType:    "",
			ReadOnly:        false,
			MmapFlags:       0,
			InitialMmapSize: 0,
			PageSize:        0,
			NoSync:          false,
			OpenFile:        nil,
			Mlock:           false,
		})
	})
	if databaseOpenError != nil {
		return nil, stacktrace.Propagate(databaseOpenError, "An error occurred while opening the enclave database")
	}

	return &EnclaveDB{databaseInstance}, nil
}

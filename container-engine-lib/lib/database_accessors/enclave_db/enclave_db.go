package enclave_db

import (
	"os"
	"sync"
	"time"

	"github.com/kurtosis-tech/stacktrace"
	bolt "go.etcd.io/bbolt"
)

const (
	readWritePermissionToDatabase = 0666
	enclaveDbFilePath             = "enclave.db"
	timeOut                       = 10 * time.Second
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
			Timeout:         timeOut, //to fail if any other process is locking the file
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
			PreLoadFreelist: false,
		})
	})
	if databaseOpenError != nil {
		return nil, stacktrace.Propagate(databaseOpenError, "An error occurred while opening the enclave database")
	}

	return &EnclaveDB{databaseInstance}, nil
}

func EraseDatabase() error {
	path := databaseInstance.Path()
	err := databaseInstance.Close()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to close database during erase process")
	}
	err = os.Remove(path)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to erase database file during erase process '%v'", path)
	}
	return nil
}

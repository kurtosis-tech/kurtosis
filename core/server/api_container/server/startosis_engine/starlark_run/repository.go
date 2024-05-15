package starlark_run

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	starlarkRunBucketName = []byte("starlark-run-repository")
	// there is only one key because there is only one Starlark Run object per APIC
	starlarkRunKey = []byte("starlark-run-key")
)

type StarlarkRunRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func GetOrCreateNewStarlarkRunRepository(enclaveDb *enclave_db.EnclaveDB) (*StarlarkRunRepository, error) {
	if err := enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(starlarkRunBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the starlark run database bucket")
		}
		logrus.Debugf("Starlark run bucket: '%+v'", bucket)

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building the starlark run repository")
	}

	starlarkRunRepository := &StarlarkRunRepository{
		enclaveDb: enclaveDb,
	}

	return starlarkRunRepository, nil
}

// Get will return either the starlark run or nil
func (repository *StarlarkRunRepository) Get() (*StarlarkRun, error) {
	var (
		starlarkRunObj *StarlarkRun
		err            error
	)

	if err = repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(starlarkRunBucketName)

		// first get the bytes
		starlarkRunBytes := bucket.Get(starlarkRunKey)

		if starlarkRunBytes == nil {
			return nil
		}

		starlarkRunObj = &StarlarkRun{}
		if err = json.Unmarshal(starlarkRunBytes, starlarkRunObj); err != nil {
			return stacktrace.Propagate(err, "An error occurred unmarshalling starlark run")
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting starlark run from the repository")
	}
	return starlarkRunObj, nil
}

func (repository *StarlarkRunRepository) Save(
	starlarkRun *StarlarkRun,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(starlarkRunBucketName)

		jsonBytes, err := json.Marshal(starlarkRun)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling starlark run '%+v'", starlarkRun)
		}

		// save it to disk
		if err := bucket.Put(starlarkRunKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving starlark run '%+v' into the enclave db bucket", starlarkRun)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving starlark run '%+v' into the starlark run repository", starlarkRun)
	}
	return nil
}

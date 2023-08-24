package runtime_value_store

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type stringifiedRecipeResultValue string

var (
	recipeResultBucketName = []byte("recipe-result-repository")
)

type recipeResultRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func getOrCreateNewRecipeResultRepository(enclaveDb *enclave_db.EnclaveDB) (*recipeResultRepository, error) {
	if err := enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(recipeResultBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the recipe result database bucket")
		}
		logrus.Debugf("Recipe result bucket: '%+v'", bucket)

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building the recipe result repository")
	}

	repository := &recipeResultRepository{
		enclaveDb: enclaveDb,
	}

	return repository, nil
}

func (repository *recipeResultRepository) Save(
	uuid string,
	value map[string]stringifiedRecipeResultValue,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recipeResultBucketName)

		uuidKey := getUuidKey(uuid)

		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling value '%+v' in the recipe result repository", value)
		}

		// save it to disk
		if err := bucket.Put(uuidKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving recipe result value '%+v' with UUID '%s' into the enclave db bucket", value, uuid)
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving recipe result value '%+v' with UUID '%s' into the enclave db", value, uuid)
	}
	return nil
}

func (repository *recipeResultRepository) Get(
	uuid string,
) (map[string]stringifiedRecipeResultValue, error) {

	value := map[string]stringifiedRecipeResultValue{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recipeResultBucketName)

		uuidKey := getUuidKey(uuid)

		// first get the bytes
		jsonBytes := bucket.Get(uuidKey)

		if err := json.Unmarshal(jsonBytes, &value); err != nil {
			return stacktrace.Propagate(err, "An error occurred unmarshalling the recipe result value with UUID '%s' from the repository", uuid)
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the recipe result value with UUID '%s' from the enclave db", uuid)
	}
	return value, nil
}

func getUuidKey(uuid string) []byte {
	return []byte(uuid)
}

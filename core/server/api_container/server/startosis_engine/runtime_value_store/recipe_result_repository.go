package runtime_value_store

import (
	"encoding/json"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
)

var (
	recipeResultBucketName = []byte("recipe-result-repository")
	emptyValue             = []byte{}
)

type recipeResultRepository struct {
	enclaveDb          *enclave_db.EnclaveDB
	starlarkValueSerde *kurtosis_types.StarlarkValueSerde
}

func getOrCreateNewRecipeResultRepository(
	enclaveDb *enclave_db.EnclaveDB,
	starlarkValueSerde *kurtosis_types.StarlarkValueSerde,
) (*recipeResultRepository, error) {
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
		enclaveDb:          enclaveDb,
		starlarkValueSerde: starlarkValueSerde,
	}

	return repository, nil
}

func (repository *recipeResultRepository) SaveKey(
	uuid string,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recipeResultBucketName)

		uuidKey := getUuidKey(uuid)

		// save it to disk
		if err := bucket.Put(uuidKey, emptyValue); err != nil {
			return stacktrace.Propagate(err, "An error occurred while registering key UUID '%s' into the enclave db bucket", uuid)
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while registering key UUID '%s' into the enclave db", uuid)
	}
	return nil
}

func (repository *recipeResultRepository) Save(
	uuid string,
	value map[string]starlark.Comparable,
) error {
	logrus.Debugf("Saving recipe result '%v' with value '%v' repository", uuid, value)
	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recipeResultBucketName)

		uuidKey := getUuidKey(uuid)

		stringifiedValue := map[string]string{}

		for uuidStr, starlarkComparable := range value {
			starlarkValueStr := repository.starlarkValueSerde.Serialize(starlarkComparable)
			stringifiedValue[uuidStr] = starlarkValueStr
		}

		jsonBytes, err := json.Marshal(stringifiedValue)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling value '%+v' in the recipe result repository", stringifiedValue)
		}

		// save it to disk
		if err := bucket.Put(uuidKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving recipe result value '%+v' with UUID '%s' into the enclave db bucket", stringifiedValue, uuid)
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving recipe result value '%+v' with UUID '%s' into the enclave db", value, uuid)
	}
	logrus.Debugf("Succesfully saved recipe uuid '%v' on repository", uuid)
	return nil
}

func (repository *recipeResultRepository) Get(
	uuid string,
) (map[string]starlark.Comparable, error) {
	logrus.Debugf("Getting recipe result '%v' from repository", uuid)
	value := map[string]starlark.Comparable{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recipeResultBucketName)

		uuidKey := getUuidKey(uuid)

		// first get the bytes
		jsonBytes := bucket.Get(uuidKey)

		// check for existence
		if jsonBytes == nil {
			return stacktrace.NewError("Recipe result value with key UUID '%s' does not exist on the recipe result repository", uuid)
		}

		isEmptyValue := len(jsonBytes) == len(emptyValue)

		// this will the case if the key was saved with an empty value
		if isEmptyValue {
			return nil
		}

		stringifiedValue := map[string]string{}

		if err := json.Unmarshal(jsonBytes, &stringifiedValue); err != nil {
			return stacktrace.Propagate(err, "An error occurred unmarshalling the recipe result value with UUID '%s' from the repository", uuid)
		}

		for uuidStr, valueStr := range stringifiedValue {
			starlarkValue, err := repository.starlarkValueSerde.Deserialize(valueStr)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred deserializing the stringified value '%v'", valueStr)
			}

			starlarkComparable, ok := starlarkValue.(starlark.Comparable)
			if !ok {
				return stacktrace.NewError("Failed to cast Starlark value '%s' to Starlark comparable type", starlarkValue)
			}

			value[uuidStr] = starlarkComparable
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the recipe result value with UUID '%s' from the enclave db", uuid)
	}
	logrus.Debugf("Succesfully got recipe uuid '%v' with value '%v' from repository", uuid, value)
	return value, nil
}

func (repository *recipeResultRepository) Delete(uuid string) error {
	logrus.Debugf("Deleting recipe uuid '%v' from repository", uuid)
	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recipeResultBucketName)

		uuidKey := getUuidKey(uuid)
		if err := bucket.Delete(uuidKey); err != nil {
			return stacktrace.Propagate(err, "An error occurred deleting a recipe result with key '%v' from the recipe result bucket", uuidKey)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while deleting a recipe result with key '%v' from the recipe result repository", uuid)
	}
	logrus.Debugf("Succesfully deleted recipe uuid '%v' from repository", uuid)
	return nil
}

func getUuidKey(uuid string) []byte {
	return []byte(uuid)
}

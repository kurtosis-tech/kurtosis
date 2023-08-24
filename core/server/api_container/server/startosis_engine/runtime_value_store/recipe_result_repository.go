package runtime_value_store

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"strconv"
)

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

// Save store recipe result values into the repository, and it only accepts comparables of
// starlark.String and starlark.Int so far
func (repository *recipeResultRepository) Save(
	uuid string,
	value map[string]starlark.Comparable,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recipeResultBucketName)

		uuidKey := getUuidKey(uuid)

		stringifiedValue := map[string]string{}

		for key, comparableValue := range value {
			//TODO add more kind of comparable types if we want to extend the support
			//TODO now starlark.Int and starlark.String are enough so far
			switch comparableValue.(type) {
			case starlark.Int:
				stringifiedValue[key] = comparableValue.String()
			case starlark.String:
				comparableStr, ok := comparableValue.(starlark.String)
				if !ok {
					return stacktrace.NewError("An error occurred casting comparable type '%v' to Starlark string", comparableValue)
				}
				stringifiedValue[key] = comparableStr.GoString()
			default:
				return stacktrace.NewError("Unexpected comparable type on recipe result repository") //TODO improve error message by adding info expected types and received types
			}
		}

		jsonBytes, err := json.Marshal(stringifiedValue)
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
) (map[string]starlark.Comparable, error) {

	value := map[string]starlark.Comparable{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recipeResultBucketName)

		uuidKey := getUuidKey(uuid)

		// first get the bytes
		jsonBytes := bucket.Get(uuidKey)

		stringifiedValue := map[string]string{}

		if err := json.Unmarshal(jsonBytes, &stringifiedValue); err != nil {
			return stacktrace.Propagate(err, "An error occurred unmarshalling the recipe result value with UUID '%s' from the repository", uuid)
		}

		for key, stringifiedComparable := range stringifiedValue {
			var comparableValue starlark.Comparable
			comparableInt, err := strconv.Atoi(stringifiedComparable)
			if err != nil {
				comparableValue = starlark.String(stringifiedComparable)
			} else {
				comparableValue = starlark.MakeInt(comparableInt)
			}
			value[key] = comparableValue
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

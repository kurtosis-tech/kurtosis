package git_package_content_provider

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	packageReplaceOptionsBucketName = []byte("package-replace-options-repository")
)

type packageReplaceOptionsRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func newPackageReplaceOptionsRepository(
	enclaveDb *enclave_db.EnclaveDB,
) *packageReplaceOptionsRepository {
	return &packageReplaceOptionsRepository{
		enclaveDb: enclaveDb,
	}
}

func (repository *packageReplaceOptionsRepository) Save(
	allReplacePackageOptions map[string]string,
) error {

	if allReplacePackageOptions == nil {
		return stacktrace.NewError("Expected to receive a replace package options map but a nil value was received instead, this is a bug in Kurtosis")
	}

	logrus.Debugf("Saving package replace options '%v' in the repository...", allReplacePackageOptions)
	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists(packageReplaceOptionsBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the package replace options bucket")
		}

		for packageId, replaceOption := range allReplacePackageOptions {
			key := getKey(packageId)

			replaceOptionBytes := []byte(replaceOption)

			if err := bucket.Put(key, replaceOptionBytes); err != nil {
				return stacktrace.Propagate(err, "An error occurred while saving replace option '%s' for package ID '%s' into the enclave db bucket", replaceOption, packageId)
			}
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving package replace options '%v'  into the enclave db", allReplacePackageOptions)
	}
	logrus.Debugf("... replace options successfully saved.")
	return nil
}

func (repository *packageReplaceOptionsRepository) Get() (map[string]string, error) {
	allReplacePackageOptions := map[string]string{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(packageReplaceOptionsBucketName)
		if bucket == nil {
			return nil
		}

		if err := bucket.ForEach(func(packageIdKey, replaceOptionBytes []byte) error {
			packageId := string(packageIdKey)
			replaceOption := string(replaceOptionBytes)
			allReplacePackageOptions[packageId] = replaceOption

			return nil
		}); err != nil {
			return stacktrace.Propagate(err, "An error occurred while iterating the replace package options repository to get all the replace options")
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while iterating the replace package options from the package replace options repository")
	}

	return allReplacePackageOptions, nil
}

func getKey(keyString string) []byte {
	return []byte(keyString)
}

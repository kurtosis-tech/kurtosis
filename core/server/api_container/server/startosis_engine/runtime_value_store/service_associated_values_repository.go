package runtime_value_store

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	serviceAssociatedValuesBucketName = []byte("service-associated-values-repository")
)

type serviceAssociatedValuesRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func getOrCreateNewServiceAssociatedValuesRepository(enclaveDb *enclave_db.EnclaveDB) (*serviceAssociatedValuesRepository, error) {
	if err := enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(serviceAssociatedValuesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the service associated values database bucket")
		}
		logrus.Debugf("Service associated values bucket: '%+v'", bucket)

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building the service associated values repository")
	}

	serviceRuntimeValuesRepository := &serviceAssociatedValuesRepository{
		enclaveDb: enclaveDb,
	}

	return serviceRuntimeValuesRepository, nil
}

func (repository *serviceAssociatedValuesRepository) Save(
	serviceName service.ServiceName,
	uuid string,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceAssociatedValuesBucketName)

		serviceNameKey := getServiceNameKey(serviceName)

		// save it to disk
		if err := bucket.Put(serviceNameKey, []byte(uuid)); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving service associated value '%s' for service '%s' into the enclave db bucket", uuid, serviceName)
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding service associated value '%s' for service '%s' into the enclave db", uuid, serviceName)
	}
	return nil
}

func (repository *serviceAssociatedValuesRepository) Get(
	serviceName service.ServiceName,
) (string, error) {
	var uuid string

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceAssociatedValuesBucketName)

		serviceNameKey := getServiceNameKey(serviceName)

		// first get the bytes
		uuidBytes := bucket.Get(serviceNameKey)

		uuid = string(uuidBytes)

		return nil
	}); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting service associated values for service '%s' from the enclave db", serviceName)
	}
	return uuid, nil
}

func (repository *serviceAssociatedValuesRepository) Exist(
	serviceName service.ServiceName,
) (bool, error) {

	exist := false

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceAssociatedValuesBucketName)

		serviceNameKey := getServiceNameKey(serviceName)
		bytes := bucket.Get(serviceNameKey)
		if bytes != nil {
			exist = true
		}
		return nil
	}); err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while checking if there are associated values for service '%s' exist in service associated values repository", serviceName)
	}

	return exist, nil
}

func getServiceNameKey(serviceName service.ServiceName) []byte {
	return []byte(serviceName)
}
